#!/usr/bin/env python3
"""Reduce each human post to a neutral outline of its facts, for a writer that never sees the post.

usage: outline.py <workspace> <human jsonl> <cache dir> <out jsonl> [--workers N]

Run it under the router wrapper: the outline comes from one model for every post.
"""
import concurrent.futures, json, re, subprocess, sys
from pathlib import Path

W, source, cache, out = sys.argv[1].rstrip("/") + "/", sys.argv[2], sys.argv[3], Path(sys.argv[4])
workers = int(sys.argv[sys.argv.index("--workers") + 1]) if "--workers" in sys.argv else 6
PROMPT = """Below is a technical blog post. List what it says as a neutral outline for another writer.

Rules:
- 15 to 30 short bullet points, in the order of the post.
- Facts only: what the product or feature is, what changed, the steps, the commands and identifiers (name them, do not copy code), numbers, versions, names, limits.
- Do not copy sentences or phrases from the post. Do not describe its style or tone.
- No introduction and no conclusion. Output the bullet points only.

Title: %s

===== POST =====
%s
"""
kept = [line for line in out.read_text().splitlines() if not json.loads(line).get("failed")] if out.exists() else []
if out.exists():
    out.write_text("".join(line + "\n" for line in kept))
done = {json.loads(line)["id"] for line in kept}


def one(row):
    if row["id"] in done:
        return None
    body = " ".join(row["text"].split(" ")[:1800])
    answer = subprocess.run(["python3", W + "ask_model.py", "qwen", cache], input=(PROMPT % (row["title"], body)).encode(),
                            capture_output=True).stdout.decode().strip()
    bullets = [line for line in answer.splitlines() if re.match(r"^\s*([-*•]|\d+[.)])\s", line)]
    if len(bullets) < 8:
        return {"id": row["id"], "failed": True}
    return {"id": row["id"], "outline": "\n".join(bullets)}


rows = [json.loads(line) for line in open(source)]
with out.open("a") as handle, concurrent.futures.ThreadPoolExecutor(workers) as pool:
    for row in pool.map(one, rows):
        if row:
            handle.write(json.dumps(row, ensure_ascii=False) + "\n")
            handle.flush()
print("done")
