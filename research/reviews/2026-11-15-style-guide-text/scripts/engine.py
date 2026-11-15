#!/usr/bin/env python3
"""Run a binary on prepared texts and keep what the page index counted on each.

usage: engine.py <binary> <out jsonl> <label>=<jsonl>[:field]... [--split dev|test|all]

Each text is checked as one Markdown page. A row holds the set, the page id, the prose words, the index, and
for each signal its count, its sentences, and its rate per 1,000 words. The optional field names the text to
check, "text" by default.
"""
import concurrent.futures, json, subprocess, sys, tempfile
from pathlib import Path

binary, out = sys.argv[1], Path(sys.argv[2])
split = sys.argv[sys.argv.index("--split") + 1] if "--split" in sys.argv else "dev"
specs = [a for a in sys.argv[3:] if "=" in a]


def check(job):
    label, row, field = job
    with tempfile.TemporaryDirectory() as scratch:
        done = subprocess.run([binary, "check", "--stdin", "--filename", "page.md", "--format", "markdown", "--no-gate", "--allow-empty",
                               "--timeout", "2m", "--report", "json:-"], input=row[field].encode(), cwd=scratch, capture_output=True)
    try:
        index = json.loads(done.stdout)["documents"][0].get("ai_sounding") or {}
    except (ValueError, KeyError, IndexError):
        return {"set": label, "id": row["id"], "failed": True}
    return {"set": label, "id": row["id"], "status": index.get("status"), "words": index.get("prose_words"), "index": index.get("index"),
            "level": index.get("level"),
            "runs": {name: [m["sentences"] for m in index.get("marks", []) if m["signal"] == name] for name in ("even-sentences", "subject-openers")},
            "signals": {s["name"]: {"count": s["count"], "sentences": s.get("sentences", 0), "rate": s["per_thousand_words"]} for s in index.get("signals", [])}}


jobs = []
for spec in specs:
    label, path = spec.split("=", 1)
    path, _, field = path.partition(":")
    for row in map(json.loads, open(path)):
        if not row.get("failed") and (split == "all" or row.get("split", split) == split):
            jobs.append((label, row, field or "text"))
with concurrent.futures.ThreadPoolExecutor(6) as pool:
    rows = list(pool.map(check, jobs))
out.write_text("".join(json.dumps(r) + "\n" for r in rows))
print(len(rows), "pages;", sum(1 for r in rows if r.get("failed")), "failed;", sum(1 for r in rows if r.get("status") == "available"), "with an index")
