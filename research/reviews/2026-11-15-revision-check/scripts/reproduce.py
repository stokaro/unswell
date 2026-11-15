#!/usr/bin/env python3
"""Check that a binary reaches the recorded verdicts of the check from the kept answers of the judge.

usage: reproduce.py <unswell binary> <judge kind> <cache dir> <revisions jsonl> <recorded checks jsonl> [--only set,...]

The judge of these checks is a local program, and each of its answers is kept under its prompt. A binary that
sends the same prompts gets the same answers back without a call to the model, and has to reach the same verdict
from them. A binary whose prompt differs in one byte gets a new answer, and its statements then differ.
"""
import json, subprocess, sys, tempfile
from pathlib import Path

binary, kind, cache, revisions, recorded = sys.argv[1:6]
only = sys.argv[sys.argv.index("--only") + 1] if "--only" in sys.argv else ""
with tempfile.TemporaryDirectory() as scratch:
    out = Path(scratch) / "checks.jsonl"
    command = [sys.executable, str(Path(__file__).with_name("compare.py")), binary, kind, cache, str(out), revisions, "--workers", "4"]
    subprocess.run(command + (["--only", only] if only else []), check=True, capture_output=True)
    now = {(r["set"], r["id"]): r for r in map(json.loads, out.read_text().splitlines())}
before = {(r["set"], r["id"]): r for r in map(json.loads, open(recorded))}
same = dict.fromkeys(("verdict", "sound and both answers", "statements and their outcomes", "what was lost", "em dashes"), 0)
for key, row in now.items():
    old = before[key]
    statements = lambda r: [fact[:2] for fact in r.get("facts", [])]
    same["verdict"] += (row.get("accepted"), row.get("reasons")) == (old.get("accepted"), old.get("reasons"))
    same["sound and both answers"] += (row.get("sound"), row.get("named")) == (old.get("sound"), old.get("named"))
    same["statements and their outcomes"] += statements(row) == statements(old)
    same["what was lost"] += row.get("lost") == old.get("lost")
    same["em dashes"] += row.get("dashes") == old.get("dashes")
print(f"{len(now)} pairs; the same " + ", ".join(f"{name} on {count}" for name, count in same.items()))
