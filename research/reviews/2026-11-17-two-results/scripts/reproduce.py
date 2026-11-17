#!/usr/bin/env python3
"""Check that a binary reaches the recorded results of the check from the kept answers of the judge, offline.

usage: reproduce.py <unswell binary> <kind> <cache dir> <revisions jsonl> <recorded checks jsonl> [--only set,...] [--workers N]

The judge of these checks is a local program, and each of its answers is kept under its prompt. offline_judge.py
gives those answers back and fails on a prompt that it has no answer for, so this replay asks no model. A pair whose
replay ends without a verdict is counted apart: the binary sent a prompt that the recorded run did not send, or the
recorded run had asked that prompt twice and kept the second answer.
"""
import concurrent.futures, json, subprocess, sys, tempfile
from pathlib import Path

binary, kind, cache, revisions, recorded = sys.argv[1:6]
option = lambda name, default: sys.argv[sys.argv.index(name) + 1] if name in sys.argv else default
only, workers = set(option("--only", "").split(",")) - {""}, int(option("--workers", "6"))
JUDGE = ["--command", sys.executable, "--command-arg", str(Path(__file__).with_name("offline_judge.py")), "--command-arg", kind,
         "--command-arg", cache]


def replay(row):
    with tempfile.TemporaryDirectory() as scratch:
        page, revised, report = Path(scratch) / "page.md", Path(scratch) / "revised.md", Path(scratch) / "report.json"
        page.write_text(row["original"])
        revised.write_text(row["revised"])
        subprocess.run([binary, "compare", str(page), str(revised), "--quiet", "--report", "json:" + str(report), "--attempts", "1", *JUDGE],
                       cwd=scratch, capture_output=True)
        result = json.loads(report.read_text()) if report.exists() else {"error": "no report"}
    if result.get("error") or "verdict" not in result:
        return (row["set"], row["id"]), None
    verdict = result["verdict"]
    return (row["set"], row["id"]), {
        "verdict": (verdict["accepted"], verdict["reasons"]), "nothing lost": (verdict["kept"]["passed"], verdict["kept"]["reasons"]),
        "sound": (result["sound"]["verdict"], [result["sound"]["page_first"]["named"], result["sound"]["revision_first"]["named"]]),
        "statements": [[f["outcome"], f["text"]] for f in result["facts"]["statements"]],
        "sentences": [[s["reason"], s.get("outcome", ""), s["text"]] for s in result["sentences"]["suspects"]],
        "losses": result["lost"], "em dashes": [result["em_dashes"]["page"], result["em_dashes"]["revision"]]}


def kept(row):
    return {"verdict": (row["accepted"], row["reasons"]), "nothing lost": (row["nothing_lost"], row["lost_reasons"]),
            "sound": (row["sound"], row["named"]), "statements": [fact[:2] for fact in row["facts"]],
            "sentences": [sentence[:3] for sentence in row["sentences"]], "losses": row["lost"], "em dashes": row["dashes"]}


before = {}
for row in map(json.loads, open(recorded)):
    if "error" not in row:
        before[(row["set"], row["id"])] = kept(row)
jobs = [row for row in map(json.loads, open(revisions)) if "revised" in row and (not only or row["set"] in only) and (row["set"], row["id"]) in before]
with concurrent.futures.ThreadPoolExecutor(workers) as pool:
    now = dict(pool.map(replay, jobs))
replayed = {key: value for key, value in now.items() if value is not None}
parts = ("verdict", "nothing lost", "sound", "statements", "sentences", "losses", "em dashes")
same = {part: sum(1 for key, value in replayed.items() if json.dumps(value[part]) == json.dumps(before[key][part])) for part in parts}
print(f"{len(jobs)} pairs; replayed {len(replayed)}; without a kept answer {len(jobs) - len(replayed)}; the same "
      + ", ".join(f"{part} on {count}" for part, count in same.items()))
