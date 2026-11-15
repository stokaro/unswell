#!/usr/bin/env python3
"""Export the measurement of text written under a style guide into its record, without any text.

usage: export.py <workspace> <record dir> <unswell binary> <earlier binary>

pages        every page of both halves with what the build counted on it and the index of the earlier release
judgments    one row for each judge's answer on the measured half
revisions    the index of each revised page before and after
earlier      the pages of the earlier record of the index, read by the build

The posts of the style-guide set are numbered in the order of a salted hash; the salt stays in the workspace, and
the record holds neither their names nor their word counts.
"""
import gzip, hashlib, json, subprocess, sys, tempfile, concurrent.futures
from pathlib import Path

W, record, binary, earlier = Path(sys.argv[1]), Path(sys.argv[2]), sys.argv[3], sys.argv[4]
G = W / "guided"
WRITERS = {"gpt": "gpt-6-astra", "qwen": "private-host/Qwen3.6-35B-A3B"}
JUDGES = {"gpt": "gpt-6-astra", "qwen27": "private-host/Qwen3.8-27B", "qwen35": "private-host/Qwen3.6-35B-A3B"}
load = lambda path: [json.loads(line) for line in open(path)]
salt_path = G / "set-salt.txt"
if not salt_path.exists():
    salt_path.write_text(hashlib.sha256(str(sorted(p.name for p in G.iterdir())).encode() + open("/dev/urandom", "rb").read(16)).hexdigest())
salt = salt_path.read_text().strip()


def check(job):
    which, text = job
    with tempfile.TemporaryDirectory() as scratch:
        done = subprocess.run([which, "check", "--stdin", "--filename", "page.md", "--format", "markdown", "--no-gate", "--allow-empty",
                               "--timeout", "2m", "--report", "json:-"], input=text.encode(), cwd=scratch, capture_output=True)
    return json.loads(done.stdout)["documents"][0]["ai_sounding"]


def write(name, rows):
    with gzip.GzipFile(record / "data" / (name + ".jsonl.gz"), "wb", mtime=0) as handle:
        for row in rows:
            handle.write((json.dumps(row, sort_keys=True, ensure_ascii=False) + "\n").encode())
    print(name, len(rows))


half = lambda split: "measured" if split == "test" else "fitting"
human = load(G / "human.jsonl")
readme = load(G / "readme-human.jsonl")
styled = load(G / "set.jsonl")
order = sorted(styled, key=lambda r: hashlib.sha256((salt + r["id"]).encode()).hexdigest())
set_id = {r["id"]: f"s-{position + 1:02d}" for position, r in enumerate(order)}
human_by = {r["id"]: r for r in human}
pages = []
for r in human:
    pages.append(({"set": "human-blog", "half": half(r["split"]), "id": r["id"], "repository": r["repo"], "path": r["path"], "commit": r["commit"]}, r["text"], True))
for r in readme:
    pages.append(({"set": "readme", "half": half(r["split"]), "id": r["id"]}, r["raw"], True))
for r in styled:
    pages.append(({"set": "set-s", "half": half(r["split"]), "id": set_id[r["id"]]}, r["text"], False))
for family, writer in WRITERS.items():
    for r in load(G / f"written-{family}.jsonl"):
        if not r.get("failed"):
            pages.append(({"set": r["variant"] if r["variant"] == "guided" else "plain", "writer": writer, "guide": r["guide"],
                           "half": half(human_by[r["id"]]["split"]), "id": r["id"]}, r["text"], True))
with concurrent.futures.ThreadPoolExecutor(6) as pool:
    now = list(pool.map(check, [(binary, text) for _, text, _ in pages]))
    before = list(pool.map(check, [(earlier, text) for _, text, _ in pages]))
out = []
for (entry, _, detailed), index, old in zip(pages, now, before):
    assert index["status"] == "available" and old["status"] == "available", entry
    entry.update(index=index["index"], marks=index["readings"]["marks"], rhythm=index["readings"]["rhythm"], earlier=old["index"])
    if detailed:
        entry.update(words=index["prose_words"], signals={s["name"]: [s["count"], s.get("sentences", 0)] for s in index["signals"]})
    out.append(entry)
write("pages", sorted(out, key=lambda e: (e["set"], e.get("writer") or "", e["half"], e["id"])))

# The judges' answers on the measured half. A unit of the style-guide set is its post; of the other comparisons, the pair.
judgments = []
pairs = {r["id"]: r for r in load(G / "pairs-test.jsonl")}
for label, judge in JUDGES.items():
    answered = {r["id"]: r["picked"] for r in load(G / f"judge-test-{label}.jsonl")}
    for identifier, pair in pairs.items():
        parts = identifier.split(":")
        entry = {"comparison": pair["comparison"], "judge": judge, "vote": answered.get(identifier, "unreadable")}
        if pair["comparison"] == "set":
            entry.update(unit=set_id[parts[1]], against=parts[2])
        elif pair["comparison"] == "cross":
            entry.update(unit=parts[2], writer=WRITERS[parts[1]], against=parts[3])
        else:
            entry.update(unit=parts[2], writer=WRITERS[parts[1]], guide=pair["guide"])
        judgments.append(entry)
revisions = []
for family, reviser in WRITERS.items():
    path = G / f"fix-test-{family}.jsonl"
    if not path.exists():
        continue
    unit_of = lambda r: (set_id[r["id"]] if r["set"] == "setS" else r["id"])
    kind = {"setS": ("set-s", None), "gptG": ("guided", WRITERS["gpt"]), "qwenG": ("guided", WRITERS["qwen"])}
    for r in load(path):
        entry = {"reviser": reviser, "set": kind[r["set"]][0], "id": unit_of(r)}
        if kind[r["set"]][1]:
            entry["writer"] = kind[r["set"]][1]
        if r.get("skipped"):
            entry["outcome"] = "no mark listed"
        elif r.get("failed"):
            entry["outcome"] = "no revision"
        else:
            entry.update(outcome="revised", before=[r["before"]["index"], r["before"]["readings"]["marks"], r["before"]["readings"]["rhythm"]],
                         after=[r["after"]["index"], r["after"]["readings"]["marks"], r["after"]["readings"]["rhythm"]],
                         em_dashes_added=r["revised"].count("—") - r["original"].count("—"))
            if r["set"] != "setS":
                entry["words"] = [r["before"]["words"], r["after"]["words"]]
        revisions.append(entry)
    for label, judge in JUDGES.items():
        answers = G / f"judge-fix-test-{family}-{label}.jsonl"
        if not answers.exists():
            continue
        answered = {r["id"]: r["picked"] for r in load(answers)}
        for r in load(path):
            if "revised" not in r:
                continue
            entry = {"comparison": "revision", "reviser": reviser, "judge": judge, "unit": unit_of(r), "set": kind[r["set"]][0],
                     "vote": {"model": "page", "human": "revision", "same": "same"}.get(answered.get(f"fix-{family}:{r['set']}:{r['id']}"), "unreadable")}
            if kind[r["set"]][1]:
                entry["writer"] = kind[r["set"]][1]
            judgments.append(entry)
write("judgments", sorted(judgments, key=lambda e: json.dumps(e, sort_keys=True)))
write("revisions", sorted(revisions, key=lambda e: json.dumps(e, sort_keys=True)))

# The earlier record's pages, read by the build.
reading = load(G / "old-reading-final.jsonl")
write("earlier", [{"group": r["group"], "id": r["id"], "side": r["side"], "outside_main": r["outside_main"], "index": r["index"],
                   "marks": r["readings"]["marks"], "rhythm": r["readings"]["rhythm"]} for r in reading])
