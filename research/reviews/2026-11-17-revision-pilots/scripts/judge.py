#!/usr/bin/env python3
"""Ask one model which of two openings of blog posts reads more as written by a language model.

usage: judge.py <workspace> <pairs jsonl> <model kind> <cache dir> <out jsonl> [--workers N]

A pair row holds "id", "first" (the model's text) and "second" (the human text). The two are shown as A and B in an
order fixed by the id, with no hint of which is which. The answer is A, B, or same, with one sentence of reason.
"""
import concurrent.futures, hashlib, json, re, subprocess, sys
from pathlib import Path

W, pairs_path, kind, cache, out = sys.argv[1].rstrip("/") + "/", sys.argv[2], sys.argv[3], sys.argv[4], Path(sys.argv[5])
workers = int(sys.argv[sys.argv.index("--workers") + 1]) if "--workers" in sys.argv else 8
PROMPT = """Below are the openings of two technical blog posts, A and B. Judge only how the text reads, not the subject and not which post has more or better content.

Which post reads MORE like text written by a language model?

Answer with JSON only: {"more_ai_sounding": "A" | "B" | "same", "reason": "<one sentence>"}

===== POST A =====
%s

===== POST B =====
%s
"""
done = {json.loads(line)["id"] for line in out.read_text().splitlines()} if out.exists() else set()


def one(pair):
    if pair["id"] in done:
        return None
    swap = int(hashlib.sha256(pair["id"].encode()).hexdigest(), 16) % 2 == 1
    a, b = (pair["second"], pair["first"]) if swap else (pair["first"], pair["second"])
    answer = subprocess.run(["python3", W + "ask_model.py", kind, cache], input=(PROMPT % (a, b)).encode(), capture_output=True).stdout.decode()
    match = re.search(r'"more_ai_sounding"\s*:\s*"(A|B|same)"', answer)
    if not match:
        return None
    choice = match.group(1)
    reason = re.search(r'"reason"\s*:\s*"((?:[^"\\]|\\.)*)"', answer)
    picked = "same" if choice == "same" else ("model" if (choice == "A") != swap else "human")
    return {"id": pair["id"], "picked": picked, "swap": swap, "reason": reason.group(1) if reason else ""}


pairs = [json.loads(line) for line in open(pairs_path)]
with out.open("a") as handle, concurrent.futures.ThreadPoolExecutor(workers) as pool:
    for row in pool.map(one, pairs):
        if row:
            handle.write(json.dumps(row, ensure_ascii=False) + "\n")
            handle.flush()
print("done")
