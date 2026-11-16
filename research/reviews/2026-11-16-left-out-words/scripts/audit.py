#!/usr/bin/env python3
"""Read what `unswell compare` finds without a model: the sentences of a page that its revision may have lost.

usage: audit.py <unswell binary> <out jsonl> <revisions jsonl>... [--only set,...]

A row of the output holds the set and the id, the suspects, and, for a damaged copy, whether the damaged sentence
is among them.
"""
import concurrent.futures, json, re, subprocess, sys, tempfile
from pathlib import Path

binary, out = sys.argv[1], Path(sys.argv[2])
only = set(sys.argv[sys.argv.index("--only") + 1].split(",")) if "--only" in sys.argv else set()
paths = [a for a in sys.argv[3:] if not a.startswith("--") and a not in only and a != ",".join(sorted(only))]
FAKE = str(Path(__file__).with_name("fake_judge.py"))
words = lambda text: set(re.findall(r"[a-z0-9]+", text.lower()))


def audit(row):
    with tempfile.TemporaryDirectory() as scratch:
        page, revised, report = Path(scratch) / "page.md", Path(scratch) / "revised.md", Path(scratch) / "report.json"
        page.write_text(row["original"])
        revised.write_text(row["revised"])
        subprocess.run([binary, "compare", str(page), str(revised), "--quiet", "--report", "json:" + str(report), "--command", sys.executable,
                        "--command-arg", FAKE], cwd=scratch, capture_output=True)
        result = json.loads(report.read_text())
    # The first trial build listed the suspects at the top of its report.
    suspects = (result.get("sentences") or {}).get("suspects") or result.get("suspects") or []
    kept = {"set": row["set"], "id": row["id"], "compared": (result.get("sentences") or {}).get("compared"),
            "suspects": [[s["reason"], s["text"], s.get("revised", s.get("passage", s.get("counterpart", ""))), s.get("words", []),
                          s.get("cut", ""), s.get("outcome", "")] for s in suspects]}
    if "damage" in row:
        target = words(row["damage"])
        kept["caught"] = any(len(target & words(s["text"])) >= 0.8 * len(target) for s in suspects)
    return kept


rows = [row for path in paths for row in map(json.loads, open(path)) if "revised" in row and (not only or row["set"] in only)]
with concurrent.futures.ThreadPoolExecutor(8) as pool, out.open("w") as handle:
    results = list(pool.map(audit, rows))
    for row in results:
        handle.write(json.dumps(row, ensure_ascii=False) + "\n")
by = {}
for row in results:
    entry = by.setdefault(row["set"], {"pairs": 0, "with a suspect": 0, "missing": 0, "negation": 0, "reworded": 0, "cut": 0, "caught": 0,
                                       "damaged": 0, "gone without a question": 0})
    entry["pairs"] += 1
    entry["with a suspect"] += bool(row["suspects"])
    for suspect in row["suspects"]:
        entry[suspect[0]] += 1
        entry["gone without a question"] += suspect[2] == ""
    if "caught" in row:
        entry["damaged"] += 1
        entry["caught"] += row["caught"]
for name, entry in sorted(by.items()):
    print(name, entry)
