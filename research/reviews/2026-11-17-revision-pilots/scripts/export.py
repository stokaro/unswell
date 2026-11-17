#!/usr/bin/env python3
"""Export three pilots of revising a post, and what blind judges said of the revisions, without any text.

usage: export.py <workspace> <record dir>

revisions  one row for each post and way of revising: what two rules of a page reported before and after
answers    one row for each judged pair: which text the judge named
rounds     one row for each judgment of the pilot in rounds
subjects   one row for each judged pair of a human post and a model's text on its subject

The posts of set S are numbered as in the earlier records: in the order of a salted hash whose salt stays in the
workspace. The record holds neither their names nor their words.
"""
import gzip, hashlib, json, sys
from pathlib import Path

W, record = Path(sys.argv[1]), Path(sys.argv[2])
G = W / "guided"
JUDGES = {"gpt": "gpt-6-astra", "qwen27": "private-host/Qwen3.8-27B", "qwen35": "private-host/Qwen3.6-35B-A3B"}
load = lambda name: [json.loads(line) for line in open(G / name)]
salt = (G / "set-salt.txt").read_text().strip()
posts = load("set.jsonl") if (G / "set.jsonl").exists() else load("set.jsonl")
set_id = {r["id"]: f"s-{position + 1:02d}" for position, r in
          enumerate(sorted(posts, key=lambda r: hashlib.sha256((salt + r["id"]).encode()).hexdigest()))}
NAMED = {"model": "post", "human": "human", "same": "same"}


def write(name, rows):
    with open(record / "data" / (name + ".jsonl.gz"), "wb") as raw, gzip.GzipFile(fileobj=raw, mode="wb", mtime=0, filename="") as handle:
        for row in rows:
            handle.write((json.dumps(row, sort_keys=True) + "\n").encode())
    print(name, len(rows))


# The three human posts of each post, numbered in the order of their pair ids.
base = [p for p in load("pairs-dev-set.jsonl") if p["comparison"] == "set"]
number = {}
for unit in {p["unit"] for p in base}:
    for position, pair in enumerate(sorted((p for p in base if p["unit"] == unit), key=lambda p: p["id"]), 1):
        number[pair["id"]] = position

ARMS = {"closing-sentences": ("fix-dev-denial-qwen.jsonl", "pairs-fix3-denial-qwen.jsonl", "denial-qwen",
                              {"gpt": "judge-fix3-denial-qwen-gpt.jsonl", "qwen27": "judge-fix3-denial-qwen-qwen27.jsonl",
                               "qwen35": "judge-fix3-denial-qwen-qwen35.jsonl"}),
        "manner": ("fix-dev-manner-qwen.jsonl", "pairs-fix3-manner-qwen.jsonl", "manner-qwen", {"qwen27": "judge-fix3-manner-qwen-qwen27.jsonl"})}

revisions, answers = [], []
for arm, (fixed, pairs_name, prefix, judges) in ARMS.items():
    for row in load(fixed):
        kept = {"arm": arm, "page": set_id[row["id"]], "revised": "revised" in row}
        if kept["revised"]:
            for side in ("before", "after"):
                kept[side] = {"index": row[side]["index"], "closing_sentences": row[side]["closers"], "register": int(row[side]["register"])}
        revisions.append(kept)
    pairs = {p["id"]: p for p in load(pairs_name)}
    for judge, name in judges.items():
        for row in load(name):
            pair = pairs[row["id"]]
            answers.append({"arm": arm, "page": set_id[pair["unit"]], "pair": number[row["id"][len(prefix) + 1:]], "judge": JUDGES[judge],
                            "named": NAMED[row["picked"]]})
for judge, name in (("gpt", "judge-dev-gpt.jsonl"), ("qwen27", "judge-dev-qwen27.jsonl"), ("qwen35", "judge-dev-set-qwen35.jsonl")):
    known = {p["id"]: p for p in base}
    for row in load(name):
        if row["id"] in known:
            answers.append({"arm": "as-it-is", "page": set_id[known[row["id"]]["unit"]], "pair": number[row["id"]], "judge": JUDGES[judge],
                            "named": NAMED[row["picked"]]})
revisions.sort(key=lambda r: (r["arm"], r["page"]))
answers.sort(key=lambda r: (r["arm"], r["judge"], r["page"], r["pair"]))
write("revisions", revisions)
write("answers", answers)

rounds = []
for row in load("loop-dev-pilot.jsonl"):
    for step in row["rounds"]:
        for pair, named in enumerate(step.get("named", []), 1):
            rounds.append({"page": set_id[row["id"]], "round": step["round"], "pair": pair, "named": named})
rounds.sort(key=lambda r: (r["page"], r["round"], r["pair"]))
write("rounds", rounds)

# Texts that two models wrote under a style guide from the outline of a human post, each against that human post.
subjects = []
same = {p["id"]: p for p in load("pairs-dev.jsonl") if p["comparison"] == "guided"}
for judge, name in (("gpt", "judge-dev-gpt.jsonl"), ("qwen27", "judge-dev-qwen27.jsonl")):
    for row in load(name):
        if row["id"] in same:
            subjects.append({"writer": {"gpt": "first", "qwen": "second"}[same[row["id"]]["family"]],
                             "id": hashlib.sha256(("subject:" + row["id"]).encode()).hexdigest()[:12], "judge": JUDGES[judge],
                             "named": {"model": "model", "human": "human", "same": "same"}[row["picked"]]})
subjects.sort(key=lambda r: (r["judge"], r["writer"], r["id"]))
write("subjects", subjects)
