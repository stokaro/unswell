#!/usr/bin/env python3
"""Ask a judge model which version of a text is better written.

usage: judge.py <human directory> <task> <judge> [--workers N]

tasks:
  triplets       the LAMP preference triplets: rank three versions
  beemo          Beemo test rows: raw model answer or expert-edited answer
  pairs:<name>   rows of pairs/<name>.jsonl: {"id", "texts": {label: text}} with two labels

Labels are hidden. The order of the versions is fixed by a hash of the row
identifier. Answers are stored under judgments/<task>/<judge>/.
"""
import argparse, concurrent.futures, hashlib, json, os, subprocess, sys, tempfile
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))
from common import read_jsonl

RANK = """You are an experienced editor of English prose. Below are three versions of one passage, labeled A, B, and C.
Rank them from the best written to the worst written.

Judge the writing itself: clarity, precision, economy, and whether the wording sounds natural and not formulaic.
Do not reward length or ornament. Do not guess who or what wrote a version.

Answer with JSON only: {"ranking": ["<best>", "<middle>", "<worst>"], "reason": "<one sentence>"}
"""
PAIR = """You are an experienced editor of English prose. Below are two versions of one text, labeled A and B.
They carry the same content. Decide which version is better written.

Judge the writing itself: clarity, precision, economy, and whether the wording sounds natural and not formulaic.
Do not reward length or ornament. Do not guess who or what wrote a version.

Answer with JSON only: {"better": "A" or "B", "reason": "<one sentence>"}
"""

parser = argparse.ArgumentParser()
parser.add_argument("root")
parser.add_argument("task")
parser.add_argument("judge")
parser.add_argument("--workers", type=int, default=4)
options = parser.parse_args()
ROOT = Path(options.root).resolve()
ASK, CACHE = str(ROOT.parent / "ask_model.py"), str(ROOT / "cache")
out = ROOT / "judgments" / options.task.replace(":", "-") / options.judge
out.mkdir(parents=True, exist_ok=True)


def rows():
    if options.task == "triplets":
        seen = {}
        for row in read_jsonl(ROOT / "preferences.jsonl"):
            seen.setdefault(row["triplet"], {"id": row["triplet"], "texts": row["texts"]})
        return list(seen.values())
    if options.task == "beemo":
        edited = {f"beemo-{int(r['id']):04d}": r["human_edits"] for r in read_jsonl(ROOT / "raw/beemo.jsonl")}
        return [{"id": d["id"], "texts": {"raw": d["text"], "edited": edited[d["id"]]}}
                for d in read_jsonl(ROOT / "docs.jsonl") if d["set"] == "beemo" and d["split"] == "test"]
    return read_jsonl(ROOT / "pairs" / (options.task.split(":", 1)[1] + ".jsonl"))


def order(row):
    """Return the labels in a fixed pseudo-random order."""
    return sorted(row["texts"], key=lambda label: hashlib.sha256((row["id"] + "|" + label).encode()).hexdigest())


def ask(prompt):
    last = None
    for _ in range(3):
        with tempfile.TemporaryDirectory() as scratch:
            done = subprocess.run(["python3", ASK, options.judge, CACHE], input=prompt.encode(), cwd=scratch, capture_output=True)
        if done.returncode != 0:
            last = done.stderr.decode("utf-8", "replace")[-200:]
            continue
        answer = done.stdout.decode()
        try:
            return json.loads(answer[answer.index("{"):answer.rindex("}") + 1])
        except ValueError as error:
            last = str(error)
    raise ValueError(last)


def one(row):
    target = out / f"{row['id']}.json"
    if target.exists():
        return "cached"
    labels = order(row)
    letters = "ABC"[:len(labels)]
    body = "".join(f"\n===== VERSION {letter} =====\n{row['texts'][label]}\n" for letter, label in zip(letters, labels))
    try:
        answer = ask((RANK if len(labels) == 3 else PAIR) + body)
        if len(labels) == 3:
            ranking = [str(x).strip().upper()[:1] for x in answer["ranking"]]
            if sorted(ranking) != list(letters):
                raise ValueError("ranking is not a permutation: " + str(answer.get("ranking")))
            result = {"order": [labels[letters.index(x)] for x in ranking]}
        else:
            better = str(answer["better"]).strip().upper()[:1]
            if better not in letters:
                raise ValueError("no choice: " + str(answer.get("better")))
            result = {"better": labels[letters.index(better)], "position": better}
    except (ValueError, KeyError) as error:
        return "FAILED " + str(error)[:200]
    result["reason"] = str(answer.get("reason", ""))[:400]
    temporary = target.with_suffix(f".{os.getpid()}.tmp")
    temporary.write_text(json.dumps(result, ensure_ascii=False))
    temporary.rename(target)
    return "ok"


items = rows()
failed = 0
with concurrent.futures.ThreadPoolExecutor(options.workers) as pool:
    for row, outcome in zip(items, pool.map(one, items)):
        if outcome.startswith("FAILED"):
            failed += 1
            print(row["id"], outcome, flush=True)
print(f"{options.task} {options.judge}: {len(items)} rows, {failed} failed", flush=True)
