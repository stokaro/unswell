#!/usr/bin/env python3
"""Run the built trim command on unused generated documents and choose paragraph pairs for the judges.

usage: product_trim.py <corpus workspace> <unswell binary> [--first N] [--last N] [--workers N]

Documents first..last (1-based) of the generated documents in hash order are
trimmed with each of two models through `unswell trim`. One pair per document
and model: the paragraph that holds the accepted trim with the lowest hash,
before and after all accepted trims in it. Writes study3/pairs.json and
study3/key.json in the shape of the earlier comparisons; the arm is "product".
"""
import argparse, concurrent.futures, hashlib, json, re, subprocess, sys, tempfile
from pathlib import Path

parser = argparse.ArgumentParser()
parser.add_argument("workspace")
parser.add_argument("binary")
parser.add_argument("--first", type=int, default=61)
parser.add_argument("--last", type=int, default=120)
parser.add_argument("--workers", type=int, default=4)
parser.add_argument("--per-model", type=int, default=20)
parser.add_argument("--study", default="study3")
options = parser.parse_args()
W = Path(options.workspace).resolve()
HUMAN = W / "human"
ASK, CACHE = str(W / "ask_model.py"), str(HUMAN / "cache")
documents = [json.loads(line) for line in open(W / "docs.jsonl")]
documents = [d for d in documents if d["origin"] == "generated"]
documents = sorted(documents, key=lambda d: hashlib.sha256(d["doc_id"].encode()).hexdigest())[options.first - 1:options.last]


def run(model, doc):
    target = HUMAN / options.study / "runs" / model / f"{doc['doc_id']}.json"
    target.parent.mkdir(parents=True, exist_ok=True)
    if target.exists():
        return json.loads(target.read_text())
    with tempfile.TemporaryDirectory() as scratch:
        done = subprocess.run([options.binary, "trim", "--stdin", "--filename", doc["doc_id"] + ".md", "--format", "markdown",
                               "--quiet", "--report", "json:-", "--request-timeout", "15m",
                               "--command", "python3", "--command-arg", ASK, "--command-arg", model, "--command-arg", CACHE],
                              input=doc["text"].encode(), cwd=scratch, capture_output=True)
    if done.returncode != 0:
        raise ValueError(done.stderr.decode("utf-8", "replace")[-300:])
    report = json.loads(done.stdout)
    target.write_text(json.dumps(report))
    return report


def plain(block):
    stripped = block.strip()
    if not stripped or stripped.startswith(("#", "|", "```", ">", "<")) or "```" in stripped or "\n|" in stripped:
        return False
    if re.match(r"^(\s*[-*+] |\s*\d+[.)] )", stripped):
        return False
    return 15 <= len(stripped.split()) <= 130


pairs, key, stats = [], {}, {}
for model in ("claude", "astra"):
    with concurrent.futures.ThreadPoolExecutor(options.workers) as pool:
        futures = {doc["doc_id"]: pool.submit(run, model, doc) for doc in documents}
    chosen = failed = trims = rejected = 0
    for doc in documents:
        try:
            report = futures[doc["doc_id"]].result()
        except ValueError as error:
            failed += 1
            print(model, doc["doc_id"], "FAILED", str(error)[:160])
            continue
        entry = report["documents"][0]
        trims += len(entry["trims"])
        rejected += len(entry["rejected"])
        if chosen == options.per_model or not entry["trims"]:
            continue
        raw = doc["text"].encode()
        # Paragraph bounds in bytes: blocks separated by blank lines.
        bounds, start = [], 0
        for match in re.finditer(rb"\n\s*\n", raw):
            bounds.append((start, match.start()))
            start = match.end()
        bounds.append((start, len(raw)))
        ordered = sorted(entry["trims"], key=lambda t: hashlib.sha256(f"{doc['doc_id']}|{t['span']['start']}".encode()).hexdigest())
        for trim in ordered:
            block = next(((s, e) for s, e in bounds if s <= trim["span"]["start"] and trim["span"]["end"] <= e), None)
            if block is None:
                continue
            before = raw[block[0]:block[1]]
            inside = sorted((t for t in entry["trims"] if block[0] <= t["span"]["start"] and t["span"]["end"] <= block[1]),
                            key=lambda t: -t["span"]["start"])
            after = before
            for t in inside:
                after = after[:t["span"]["start"] - block[0]] + t["trimmed"].encode() + after[t["span"]["end"] - block[0]:]
            before_text, after_text = before.decode(), after.decode()
            if not plain(before_text) or not plain(after_text):
                continue
            identifier = hashlib.sha256(f"product|{model}|{doc['doc_id']}".encode()).hexdigest()[:10]
            revised_left = int(hashlib.sha256(("side|" + identifier).encode()).hexdigest(), 16) % 2 == 0
            pairs.append({"id": identifier, "left": (after_text if revised_left else before_text).strip(),
                          "right": (before_text if revised_left else after_text).strip()})
            key[identifier] = {"reviewer": model, "arm": "product", "document": doc["doc_id"],
                               "revised": "left" if revised_left else "right", "trims": len(inside)}
            chosen += 1
            break
    stats[model] = {"documents": len(documents), "failed": failed, "trims": trims, "rejected": rejected, "pairs": chosen}
pairs.sort(key=lambda p: hashlib.sha256(("order|" + p["id"]).encode()).hexdigest())
(HUMAN / options.study / "pairs.json").write_text(json.dumps(pairs, ensure_ascii=False, indent=1) + "\n")
(HUMAN / options.study / "key.json").write_text(json.dumps(key, indent=1, sort_keys=True) + "\n")
(HUMAN / options.study / "stats.json").write_text(json.dumps(stats, indent=1, sort_keys=True) + "\n")
print(stats, "pairs", len(pairs))
