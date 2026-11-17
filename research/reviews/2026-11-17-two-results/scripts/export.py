#!/usr/bin/env python3
"""Export the measurement of the two results of the check of a revision into its record, without any text.

usage: export.py <workspace> <record dir>

copies     one row for each copy of a measured page and judge: what the whole check did with it
revisions  one row for each revision, judge, and run: what the whole check did with it
fitting    one row for each pair of the fitting half: the revisions in two runs, and the harmless copies
same       one row for each revision of one reviser under the judge of its own family, as a control

The posts of set S are numbered as in the earlier records: in the order of a salted hash whose salt stays in the
workspace. The record holds neither their names nor their words.
"""
import gzip, hashlib, json, sys
from pathlib import Path

W, record = Path(sys.argv[1]), Path(sys.argv[2])
G = W / "guided"
MODELS = {"gpt": "gpt-6-astra", "qwen": "private-host/Qwen3.6-35B-A3B"}
# The judge of a revision is of the other family than its reviser.
OTHER = {"gpt": "qwen", "qwen": "gpt"}
SOUND = ("the judge does not find the revision less like a model's text than the page",
         "the judge finds the revision more like a model's text than the page")
load = lambda name: [json.loads(line) for line in open(G / name)]
salt = (G / "set-salt.txt").read_text().strip()
posts = load("set.jsonl")
set_id = {r["id"]: f"s-{position + 1:02d}" for position, r in
          enumerate(sorted(posts, key=lambda r: hashlib.sha256((salt + r["id"]).encode()).hexdigest()))}


def unit(row_id, label, second=False):
    """The name of a page in the record. Two writers wrote on the same subjects, so a text carries its writer.

    The texts that were written for the record of the left-out words are second texts on subjects of earlier ones.
    They carry a 2, as they do there.
    """
    if label.startswith("gpt"):
        return ("g2-" if second else "g-") + row_id
    if label.startswith("qwen"):
        return ("q2-" if second else "q-") + row_id
    return set_id[row_id]


def write(name, rows):
    with open(record / "data" / (name + ".jsonl.gz"), "wb") as raw, gzip.GzipFile(fileobj=raw, mode="wb", mtime=0, filename="") as handle:
        for row in rows:
            handle.write((json.dumps(row, sort_keys=True, ensure_ascii=False) + "\n").encode())
    print(name, len(rows))


def last_attempts(name):
    """The last row of each pair, and whether an earlier attempt had ended without a verdict.

    A pair whose check ended without a verdict was run once more, and the second row follows the first in the file.
    """
    attempts = {}
    for row in load(name):
        attempts.setdefault((row["set"], row["id"]), []).append(row)
    return [(rows[-1], any("error" in earlier for earlier in rows[:-1])) for rows in attempts.values()]


def failed(reasons):
    """The checks of "nothing lost" that a pair failed, by the reasons that the command gave."""
    kinds = set()
    for reason in reasons:
        if reason in SOUND:
            continue
        kinds.add("statements" if "statement" in reason else "sentences" if "sentence" in reason
                  else "losses" if "loses" in reason else "dashes" if "em dash" in reason else "other")
    assert "other" not in kinds, reasons
    return sorted(kinds)


def result(row, retried):
    """What the record keeps of one check."""
    kept = {"retried": retried, "verdict": "error" not in row}
    if not kept["verdict"]:
        return kept
    lost = failed(row["reasons"]) if "lost_reasons" not in row else failed(row["lost_reasons"])
    if "nothing_lost" in row:
        # A build with two results says it itself, and with all required the reasons say the same.
        assert row["nothing_lost"] == (not lost) and row["require"] == "all", row["id"]
        assert failed(row["reasons"]) == lost, row["id"]
    kept.update(nothing_lost=not lost, failed=lost, sound=row["sound"], accepted=row["accepted"])
    assert kept["accepted"] == (kept["nothing_lost"] and kept["sound"] == "better"), row["id"]
    return kept


def pages_of(label):
    return "set-s" if label in ("setS", "blog") else "fresh"


# ---- the copies of the measured part ----
copies = []
for judge, name in (("gpt", "compare-test21-copies-gpt.jsonl"), ("qwen", "compare-test21-copies.jsonl")):
    for row, retried in last_attempts(name):
        kind, label = row["set"].split("-", 1)
        copies.append({"judge": MODELS[judge], "pages": pages_of(label), "kind": kind, "page": unit(row["id"], label, second=True),
                       **result(row, retried)})
copies.sort(key=lambda r: (r["judge"], r["pages"], r["kind"], r["page"]))
write("copies", copies)

# ---- the revisions of the measured part, twice ----
revisions = []
for judge in ("gpt", "qwen"):
    reviser = OTHER[judge]
    for run, letter in ((1, "a"), (2, "b")):
        for name in (f"compare-test21{letter}-{reviser}.jsonl", f"compare-test21{letter}-fresh-{reviser}.jsonl"):
            for row, retried in last_attempts(name):
                kept = result(row, retried)
                if kept["verdict"]:
                    kept.update(statements=len(row["facts"]), kept=sum(1 for fact in row["facts"] if fact[0] == "kept"))
                revisions.append({"judge": MODELS[judge], "reviser": MODELS[reviser], "pages": pages_of(row["set"]),
                                  "page": unit(row["id"], row["set"]), "run": run, **kept})
revisions.sort(key=lambda r: (r["judge"], r["pages"], r["page"], r["run"]))
write("revisions", revisions)

# ---- the fitting half ----
fitting = []
for judge in ("gpt", "qwen"):
    reviser = OTHER[judge]
    for run, name in ((1, f"compare-fit20-{reviser}.jsonl"), (2, f"compare-split1-{reviser}.jsonl")):
        for row, retried in last_attempts(name):
            fitting.append({"what": "revision", "judge": MODELS[judge], "build": "earlier", "kind": "revision", "page": unit(row["id"], row["set"]),
                            "run": run, **result(row, retried)})
for judge, build, name in (("gpt", "earlier", "compare-split1-harmless-gpt.jsonl"), ("qwen", "earlier", "compare-split1-harmless.jsonl"),
                           ("qwen", "record", "compare-split2-harmless.jsonl")):
    for row, retried in last_attempts(name):
        kind, label = row["set"].split("-", 1)
        fitting.append({"what": "harmless", "judge": MODELS[judge], "build": build, "kind": kind, "page": unit(row["id"], label), "run": 1,
                        **result(row, retried)})
fitting.sort(key=lambda r: (r["what"], r["judge"], r["build"], r["kind"], r["page"], r["run"]))
write("fitting", fitting)

# ---- a control: the revisions of one reviser under the judge of its own family ----
# The two judges found nothing lost in very different numbers of revisions, and each had checked the revisions of
# the other reviser. Here one judge checks the revisions that the other judge checked.
same = []
for name in ("compare-test21s-qwen.jsonl", "compare-test21s-fresh-qwen.jsonl"):
    for row, retried in last_attempts(name):
        same.append({"judge": MODELS["qwen"], "reviser": MODELS["qwen"], "pages": pages_of(row["set"]), "page": unit(row["id"], row["set"]),
                     **result(row, retried)})
same.sort(key=lambda r: (r["pages"], r["page"]))
write("same", same)
