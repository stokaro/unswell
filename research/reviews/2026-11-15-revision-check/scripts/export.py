#!/usr/bin/env python3
"""Export the measurement of the revision check into its record, without any text.

usage: export.py <workspace> <record dir>

published   one answer of one judge on one pair: a post of set S as it is published, against a human text
revisions   one row for each page that a reviser was given: the index and the em dashes before and after, the verdict
            of the check with what it rests on, and the answer of the third model
damage      one row for each damaged copy of a post: how many statements the check did not find
known       one row for each pair with a known difference: what the judge named in each order
fitting     what the fitting half showed: answers of the judges on revisions under several reports, both orders of one
            pair, votes by section, and the index and the em dashes of each revision

The posts of set S are numbered as in the earlier record: in the order of a salted hash whose salt stays in the
workspace. The record holds neither their names nor their words.
"""
import gzip, hashlib, json, sys
from pathlib import Path

W, record = Path(sys.argv[1]), Path(sys.argv[2])
G = W / "guided"
MODELS = {"gpt": "gpt-6-astra", "qwen": "private-host/Qwen3.6-35B-A3B", "qwen35": "private-host/Qwen3.6-35B-A3B", "qwen27": "private-host/Qwen3.8-27B"}
JUDGES = ("gpt", "qwen27", "qwen35")
VOTE = {"model": "post", "human": "human", "same": "same"}
SIDES = {"model": "page", "human": "revision", "same": "same"}
load = lambda name: [json.loads(line) for line in open(G / name)] if (G / name).exists() else []
salt = (G / "set-salt.txt").read_text().strip()
posts = load("set.jsonl")
set_id = {r["id"]: f"s-{position + 1:02d}" for position, r in
          enumerate(sorted(posts, key=lambda r: hashlib.sha256((salt + r["id"]).encode()).hexdigest()))}


def write(name, rows):
    with open(record / "data" / (name + ".jsonl.gz"), "wb") as raw, gzip.GzipFile(fileobj=raw, mode="wb", mtime=0, filename="") as handle:
        for row in rows:
            handle.write((json.dumps(row, sort_keys=True, ensure_ascii=False) + "\n").encode())
    print(name, len(rows))


def against_human(pairs_name, judge_prefix, **labels):
    """Rows for the answers of the three judges on pairs of a post against a human text."""
    pairs = {r["id"]: r for r in load(pairs_name)}
    rows = []
    for label in JUDGES:
        answered = {r["id"]: r["picked"] for r in load(judge_prefix.format(judge=label))}
        for identifier, pair in pairs.items():
            rows.append({**labels, "unit": set_id[pair["unit"]], "against": identifier.rsplit(":", 1)[-1], "judge": MODELS[label],
                         "vote": VOTE.get(answered.get(identifier), "unreadable")})
    return rows


def dashes(row):
    return [row["original"].count("—"), row["revised"].count("—")]


def page_id(label, identifier):
    """The set and the id of a page in the record."""
    if label in ("setS", "blog"):
        return {"set": "set-s", "id": set_id[identifier]}
    return {"set": "guided", "writer": MODELS["gpt" if label == "gptG" else "qwen"], "id": identifier}


# ---- the measured half ----
published = against_human("pairs-test.jsonl", "judge-test-{judge}.jsonl", reviser="none", human="blog") if False else []
record_pairs = {r["id"]: r for r in load("pairs-test.jsonl") if r["comparison"] == "set"}
for label in JUDGES:
    answered = {r["id"]: r["picked"] for r in load(f"judge-test-{label}.jsonl")}
    for identifier, pair in record_pairs.items():
        published.append({"reviser": "none", "human": "blog", "unit": set_id[pair["unit"]], "against": identifier.rsplit(":", 1)[-1],
                          "judge": MODELS[label], "vote": VOTE.get(answered.get(identifier), "unreadable")})
published += against_human("pairs-pub14-asis-styled.jsonl", "judge-pub14-asis-styled-{judge}.jsonl", reviser="none", human="styled")
for reviser in ("gpt", "qwen"):
    for human in ("blog", "styled"):
        published += against_human(f"pairs-pub14-{reviser}-{human}.jsonl", f"judge-pub14-{reviser}-{human}-{{judge}}.jsonl",
                                   reviser=MODELS[reviser], human=human)
write("published", published)

revisions = []
CHECK_JUDGE = {"gpt": "qwen35", "qwen": "gpt"}
for reviser in ("gpt", "qwen"):
    checked = {(r["set"], r["id"]): r for r in load(f"compare-test14-{reviser}.jsonl")}
    third = {r["id"]: r["picked"] for r in load(f"judge-vs-test14-{reviser}-qwen27.jsonl")}
    for row in load(f"fix-test14-{reviser}.jsonl"):
        entry = {"reviser": MODELS[reviser], **page_id(row["set"], row["id"])}
        if "revised" not in row:
            entry["outcome"] = "no mark listed" if row.get("skipped") else "no revision"
            revisions.append(entry)
            continue
        entry.update(outcome="revised", index=[row["before"]["index"], row["after"]["index"]], em_dashes=dashes(row))
        vote = third.get(f"vs14:{reviser}:{row['set']}:{row['id']}")
        entry["third"] = SIDES.get(vote, "unreadable")
        result = checked.get((row["set"], row["id"]))
        if result is not None:
            check = {"judge": MODELS[CHECK_JUDGE[reviser]]}
            if "error" in result:
                check["verdict"] = "no verdict"
            else:
                check.update(verdict="accepted" if result["accepted"] else "rejected", sound=result["sound"], named=result["named"],
                             statements=len(result["facts"]), kept=sum(f[0] == "kept" for f in result["facts"]),
                             lost={kind: len(values) for kind, values in result["lost"].items()}, reasons=len(result["reasons"]))
            entry["check"] = check
        revisions.append(entry)
write("revisions", revisions)

damage = []
copies = {(r["set"], r["id"]): r for r in load("compare-test14-damage.jsonl")}
for row in load("damage-test.jsonl"):
    result = copies.get((row["set"], row["id"]))
    entry = {"unit": set_id[row["id"]], "kind": row["set"].removeprefix("damage-")}
    if result is None or "error" in result:
        entry["verdict"] = "no verdict"
    else:
        entry.update(verdict="accepted" if result["accepted"] else "rejected", statements=len(result["facts"]),
                     not_kept=sum(f[0] != "kept" for f in result["facts"]))
    damage.append(entry)
write("damage", damage)

known = []
for result in load("compare-test14-known.jsonl"):
    entry = {"id": result["id"]}
    entry.update({"sound": "no verdict"} if "sound" not in result else {"sound": result["sound"], "named": result["named"]})
    known.append(entry)
write("known", known)

# ---- the fitting half ----
fitting = []
for variant, reviser, human, pairs_name, prefix in (
        ("as it is", "none", "blog", "pairs-dev-set.jsonl", None),
        ("as it is", "none", "styled", "pairs-dev-styled.jsonl", "judge-dev-styled-{judge}.jsonl"),
        ("report of alpha.13", "gpt", "blog", "pairs-rev13-dev-gpt.jsonl", "judge-rev13-dev-gpt-{judge}.jsonl"),
        ("report of alpha.13", "qwen", "blog", "pairs-rev13-dev-qwen.jsonl", "judge-rev13-dev-qwen-{judge}.jsonl"),
        ("report with sentence lengths", "gpt", "blog", "pairs-lengths-dev.jsonl", "judge-lengths-dev-{judge}.jsonl"),
        ("free rewrite", "gpt", "blog", "pairs-ceiling-dev.jsonl", "judge-ceiling-dev-{judge}.jsonl"),
        ("free rewrite", "gpt", "styled", "pairs-ceiling-dev-styled.jsonl", "judge-ceiling-dev-styled-{judge}.jsonl")):
    labels = {"experiment": "against human text", "variant": variant, "reviser": MODELS.get(reviser, "none"), "human": human}
    if prefix is None:
        pairs = {r["id"]: r for r in load(pairs_name)}
        for label, name in (("gpt", "judge-dev-gpt.jsonl"), ("qwen27", "judge-dev-qwen27.jsonl"), ("qwen35", "judge-dev-set-qwen35.jsonl")):
            answered = {r["id"]: r["picked"] for r in load(name)}
            for identifier, pair in pairs.items():
                fitting.append({**labels, "unit": set_id[pair["unit"]], "against": identifier.rsplit(":", 1)[-1], "judge": MODELS[label],
                                "vote": VOTE.get(answered.get(identifier), "unreadable")})
    else:
        fitting += against_human(pairs_name, prefix, **labels)

VERSUS = (("report of alpha.13", "gpt", "fix-dev13-gpt", JUDGES), ("report of alpha.13", "qwen", "fix-dev13-qwen", JUDGES),
          ("report with sentence lengths", "gpt", "fix-dev-lengths-gpt", JUDGES),
          ("report with the line on em dashes", "gpt", "fix-dev-nodash-gpt", JUDGES), ("report with the line on em dashes", "qwen", "fix-dev-nodash-qwen", JUDGES),
          ("the same, instruction of the record", "gpt", "fix-dev-nodash2-gpt", ("qwen35",)),
          ("the same, instruction of the record", "qwen", "fix-dev-nodash2-qwen", ("qwen27",)))
for variant, reviser, name, judges in VERSUS:
    units = [json.loads(line)["unit"] for line in open(G / f"pairs-vs-{name}.jsonl")]
    for label in judges:
        answered = {r["id"].rsplit(":", 1)[-1]: r["picked"] for r in load(f"judge-vs-{name}-{label}.jsonl")}
        for unit in units:
            fitting.append({"experiment": "page against revision", "variant": variant, "reviser": MODELS[reviser], "unit": set_id[unit],
                            "judge": MODELS[label], "vote": SIDES.get(answered.get(unit), "unreadable")})
    for row in load(name + ".jsonl"):
        if "revised" in row:
            fitting.append({"experiment": "revision", "variant": variant, "reviser": MODELS[reviser], "unit": set_id[row["id"]],
                            "index": [row["before"]["index"], row["after"]["index"]], "em_dashes": dashes(row)})
for row in load("fix-dev-ceiling-gpt.jsonl"):
    if "revised" in row:
        fitting.append({"experiment": "revision", "variant": "free rewrite", "reviser": MODELS["gpt"], "unit": set_id[row["id"]],
                        "index": [row["before"]["index"], row["after"]["index"]], "em_dashes": dashes(row)})

for wording, judge, name in (("wording of the record", "qwen35", "orders-record-q35.jsonl"), ("first wording of the command", "qwen35", "orders-product-q35.jsonl")):
    for row in load(name):
        fitting.append({"experiment": "both orders", "pairs": "revisions of gpt-6-astra", "wording": wording, "judge": MODELS[judge],
                        "unit": set_id[row["id"]], "named": [row["page_first"] or "unreadable", row["revision_first"] or "unreadable"]})
for pairs, judge, name in (("revisions of gpt-6-astra", "qwen35", "compare-dev-nodash-gpt-q35.jsonl"),
                           ("revisions of private-host/Qwen3.6-35B-A3B", "gpt", "compare-dev-nodash-qwen-gpt.jsonl")):
    for row in load(name):
        if "named" in row:
            fitting.append({"experiment": "both orders", "pairs": pairs, "wording": "first wording of the command, in the command", "judge": MODELS[judge],
                            "unit": set_id[row["id"]], "named": row["named"]})
for judge, name in (("qwen35", "orders-control-q35.jsonl"), ("gpt", "orders-control-gpt.jsonl")):
    for row in load(name):
        fitting.append({"experiment": "both orders", "pairs": "a model's text and the human post on its subject", "wording": "wording of the record",
                        "judge": MODELS[judge], "unit": row["id"], "named": [row["page_first"] or "unreadable", row["revision_first"] or "unreadable"]})
for pairs, judge, name in (("revisions of gpt-6-astra", "qwen35", "sections-gpt-by-q35.jsonl"),
                           ("revisions of private-host/Qwen3.6-35B-A3B", "gpt", "sections-qwen-by-gpt.jsonl")):
    for row in load(name):
        fitting.append({"experiment": "sections", "pairs": pairs, "judge": MODELS[judge], "unit": set_id[row["id"]], "votes": row["votes"]})
trial = {(r["set"], r["id"]): r for r in load("compare-dev-damage4-gpt.jsonl")}
for row in load("damage-dev.jsonl"):
    result = trial.get((row["set"], row["id"]))
    if result is not None:
        entry = {"experiment": "damage", "unit": set_id[row["id"]], "kind": row["set"].removeprefix("damage-"), "judge": MODELS["gpt"]}
        entry.update({"verdict": "no verdict"} if "error" in result else {"verdict": "rejected", "not_kept": sum(f[0] != "kept" for f in result["facts"])})
        fitting.append(entry)
write("fitting", fitting)
