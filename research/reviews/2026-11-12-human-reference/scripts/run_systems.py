#!/usr/bin/env python3
"""Run one system over the development and test documents and store its spans.

usage: run_systems.py <human directory> <unswell binary> <system> [--workers N] [--splits dev,test]

systems:
  rules-default          unswell check with the default rules
  rules-opener           the same with a build whose scaffold.chat-preamble lists "Sure," and "Sure!"
  rules-all              unswell check with every rule enabled (needs RULES_CONFIG)
  review-<reviewer>      unswell review with one reviewer (claude or astra)
  review-<reviewer>-<confirmer>   the same with a confirming model
  lamp5-<reviewer>       the span-detection prompt of the LAMP authors, five examples (needs LAMP_PROMPT)
  trim-<model>           unswell trim with one model (claude or astra); only accepted trims are stored

Each document is sent on standard input from an empty directory. Model answers
are stored by prompt under cache/, so a confirmed run judges the flags of the
unconfirmed run. A finished document is not run again.
"""
import argparse, concurrent.futures, json, os, re, subprocess, sys, tempfile
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))
from common import byte_span, read_jsonl

parser = argparse.ArgumentParser()
parser.add_argument("root")
parser.add_argument("binary")
parser.add_argument("system")
parser.add_argument("--workers", type=int, default=4)
parser.add_argument("--splits", default="dev,test")
parser.add_argument("--sets", default="lamp,beemo,prs")
options = parser.parse_args()
ROOT = Path(options.root).resolve()
ASK = str(ROOT.parent / "ask_model.py")
CACHE = str(ROOT / "cache")
out = ROOT / "runs" / options.system
out.mkdir(parents=True, exist_ok=True)
docs = [d for d in read_jsonl(ROOT / "docs.jsonl")
        if d["split"] in options.splits.split(",") and d["set"] in options.sets.split(",")]
EXTENSION = {"text": "txt", "markdown": "md"}


def run(args, text):
    with tempfile.TemporaryDirectory() as scratch:
        return subprocess.run(args, input=text.encode(), cwd=scratch, capture_output=True)


def source(doc):
    return ["--stdin", "--filename", f"{doc['id']}.{EXTENSION[doc['format']]}", "--format", doc["format"]]


def rules(doc, extra):
    done = run([options.binary, "check", *source(doc), "--no-gate", "--allow-empty", "--report", "json:-", *extra], doc["text"])
    if done.returncode != 0:
        raise ValueError(done.stderr.decode("utf-8", "replace")[-300:])
    report = json.loads(done.stdout)
    return {"spans": [[f["primary"]["span"]["start"], f["primary"]["span"]["end"]] for f in report.get("findings") or []],
            "rules": [f["rule_id"] for f in report.get("findings") or []]}


def program(prefix, kind):
    return [f"--{prefix}command", "python3", f"--{prefix}command-arg", ASK, f"--{prefix}command-arg", kind,
            f"--{prefix}command-arg", CACHE]


def review(doc, reviewer, confirmer):
    args = [options.binary, "review", *source(doc), "--quiet", "--report", "json:-", "--request-timeout", "15m",
            *program("", reviewer)]
    if confirmer:
        args += program("confirm-", confirmer)
    done = run(args, doc["text"])
    if done.returncode != 0 and b"no selected document has prose" not in done.stderr:
        raise ValueError(done.stderr.decode("utf-8", "replace")[-300:])
    document = json.loads(done.stdout)["documents"][0]
    return {"spans": [[f["span"]["start"], f["span"]["end"]] for f in document["findings"]],
            "categories": [f["categories"] for f in document["findings"]],
            "unconfirmed": [[f["span"]["start"], f["span"]["end"]] for f in document["unconfirmed"]],
            "flags": document["flags"], "unbound": len(document["unbound"])}


def trim(doc, model):
    done = run([options.binary, "trim", *source(doc), "--quiet", "--report", "json:-", "--request-timeout", "15m",
                *program("", model)], doc["text"])
    if done.returncode != 0 and b"no selected document has a sentence to trim" not in done.stderr:
        raise ValueError(done.stderr.decode("utf-8", "replace")[-300:])
    document = json.loads(done.stdout)["documents"][0]
    return {"spans": [[t["span"]["start"], t["span"]["end"]] for t in document["trims"]],
            "trims": [{"span": [t["span"]["start"], t["span"]["end"]], "original": t["original"], "trimmed": t["trimmed"],
                       "removed": t["removed"]} for t in document["trims"]],
            "rejected": [r["problem"] for r in document["rejected"]],
            "sentences": document["sentences"], "proposed": document["proposed"]}


def lamp_prompt(doc, reviewer):
    prompt = Path(os.environ["LAMP_PROMPT"]).read_text().replace("[[PARAGRAPH]]", doc["text"])
    last = None
    for _ in range(3):
        done = run(["python3", ASK, reviewer, CACHE], prompt)
        if done.returncode != 0:
            last = done.stderr.decode("utf-8", "replace")[-200:]
            continue
        answer = done.stdout.decode()
        try:
            parsed = json.loads(answer[answer.index("{"):answer.rindex("}") + 1])
            break
        except ValueError as error:
            last = str(error)
    else:
        raise ValueError(last)
    text, spans, missing = doc["text"], [], 0
    for entry in parsed.get("problematic_spans", []):
        quote = str(entry.get("span", "")).strip()
        at = text.find(quote) if quote else -1
        if at < 0 and quote:
            pattern = r"\s+".join(re.escape(part) for part in quote.split())
            match = re.search(pattern, text)
            at, quote = (match.start(), match.group()) if match else (-1, quote)
        if at < 0:
            missing += 1
            continue
        spans.append(list(byte_span(text, at, at + len(quote))))
    return {"spans": spans, "flags": len(parsed.get("problematic_spans", [])), "unbound": missing}


def one(doc):
    target = out / f"{doc['id']}.json"
    if target.exists():
        return "cached"
    try:
        name = options.system
        if name in ("rules-default", "rules-opener"):
            # rules-opener is the default rules of a build whose chat-preamble rule lists the opener "Sure".
            result = rules(doc, [])
        elif name == "rules-all":
            result = rules(doc, ["--config", os.environ["RULES_CONFIG"], "--allow-config-outside-root"])
        elif name.startswith("review-"):
            parts = name.split("-")[1:]
            result = review(doc, parts[0], parts[1] if len(parts) > 1 else None)
        elif name.startswith("trim-"):
            result = trim(doc, name.split("-")[1])
        elif name.startswith("lamp5-"):
            result = lamp_prompt(doc, name.split("-")[1])
        else:
            raise SystemExit("unknown system " + name)
    except ValueError as error:
        return "FAILED " + str(error)[:200]
    temporary = target.with_suffix(f".{os.getpid()}.tmp")
    temporary.write_text(json.dumps(result))
    temporary.rename(target)
    return f"ok {len(result['spans'])}"


failed = 0
with concurrent.futures.ThreadPoolExecutor(options.workers) as pool:
    for doc, outcome in zip(docs, pool.map(one, docs)):
        if outcome.startswith("FAILED"):
            failed += 1
            print(doc["id"], outcome, flush=True)
print(f"{options.system}: {len(docs)} documents, {failed} failed", flush=True)
