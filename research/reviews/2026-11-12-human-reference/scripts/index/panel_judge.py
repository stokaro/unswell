#!/usr/bin/env python3
"""Ask one model which of two versions of a page reads more as written by a language model.

usage: panel_judge.py <workspace> <pairs jsonl> <model kind> <out jsonl> [--workers N]

A pair row holds "id", "first", and "second" (two versions of one page) and "first_is" with the side that "first" is.
The two versions are shown as A and B in an order fixed by the id, without any hint of which is which. The answer is
one of A, B, or same. Answers are cached by prompt under <out>.cache.
"""
import concurrent.futures, hashlib, json, re, subprocess, sys
from pathlib import Path

W, pairs_path, kind, out = sys.argv[1].rstrip("/") + "/", sys.argv[2], sys.argv[3], Path(sys.argv[4])
workers = int(sys.argv[sys.argv.index("--workers") + 1]) if "--workers" in sys.argv else 8
ASK, CACHE = W + "ask_model.py", str(out) + ".cache"
PROMPT = """Below are two versions of the same documentation page, A and B. Judge only how the text reads, not which
version has more or better content.

Which version reads MORE like text written by a language model?

Answer with JSON only: {"more_ai_sounding": "A" | "B" | "same", "reason": "<one sentence>"}

===== VERSION A =====
%s

===== VERSION B =====
%s
"""
done = {json.loads(line)["id"] for line in out.read_text().splitlines()} if out.exists() else set()


def one(pair):
    if pair["id"] in done:
        return None
    swap = int(hashlib.sha256(pair["id"].encode()).hexdigest(), 16) % 2 == 1
    a, b = (pair["second"], pair["first"]) if swap else (pair["first"], pair["second"])
    answer = subprocess.run(["python3", ASK, kind, CACHE], input=(PROMPT % (a, b)).encode(), capture_output=True).stdout.decode()
    match = re.search(r'"more_ai_sounding"\s*:\s*"(A|B|same)"', answer)
    if not match:
        return {"id": pair["id"], "failed": True}
    choice = match.group(1)
    picked = "same" if choice == "same" else ("first" if (choice == "A") != swap else "second")
    return {"id": pair["id"], "picked": picked, "swap": swap}


pairs = [json.loads(line) for line in open(pairs_path)]
with out.open("a") as handle, concurrent.futures.ThreadPoolExecutor(workers) as pool:
    for row in pool.map(one, pairs):
        if row:
            handle.write(json.dumps(row) + "\n")
            handle.flush()
print("done")
