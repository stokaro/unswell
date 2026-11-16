#!/usr/bin/env python3
"""Export the measurement of the check of every changed sentence into its record, without any text.

usage: export.py <workspace> <record dir>

copies     one row for each copy of a page: what the comparison without a model found on it
damage     one row for each damaged copy and judge: what the whole check did with it
revisions  one row for each revision: what the whole check did with it under the judge of the other family

The posts of set S are numbered as in the earlier records: in the order of a salted hash whose salt stays in the
workspace. The record holds neither their names nor their words.
"""
import gzip, hashlib, json, re, sys
from pathlib import Path

W, record = Path(sys.argv[1]), Path(sys.argv[2])
G = W / "guided"
MODELS = {"gpt": "gpt-6-astra", "qwen": "private-host/Qwen3.6-35B-A3B"}
load = lambda name: [json.loads(line) for line in open(G / name)] if (G / name).exists() else []
salt = (G / "set-salt.txt").read_text().strip()
posts = load("set.jsonl")
set_id = {r["id"]: f"s-{position + 1:02d}" for position, r in
          enumerate(sorted(posts, key=lambda r: hashlib.sha256((salt + r["id"]).encode()).hexdigest()))}


def unit(row_id, label):
    """The name of a page in the record. Two writers wrote on the same subjects, so a text carries its writer."""
    if label.startswith("gpt"):
        return "g-" + row_id
    if label.startswith("qwen"):
        return "q-" + row_id
    return set_id[row_id]


def label_of(name):
    """The pages that a set of copies or revisions is made of: "setS", "gptF", and so on."""
    return "setS" if name == "blog" else name.rsplit("-", 1)[-1]


words = lambda text: set(re.findall(r"[a-z0-9]+", text.lower()))


def write(name, rows):
    with open(record / "data" / (name + ".jsonl.gz"), "wb") as raw, gzip.GzipFile(fileobj=raw, mode="wb", mtime=0, filename="") as handle:
        for row in rows:
            handle.write((json.dumps(row, sort_keys=True, ensure_ascii=False) + "\n").encode())
    print(name, len(rows))


def kind_and_pages(name):
    """The set of a copy names its kind and its pages, as in "sentence-setS" or "swap-gptF"."""
    kind, label = name.split("-", 1)
    return kind, ("set-s" if label == "setS" else "fresh")


def names(target, suspects):
    """Whether a changed sentence is among the suspects.

    The copies and the command cut a text into sentences in their own ways, so a suspect can be a part of the changed
    sentence or hold it. A suspect names the sentence when the two share at least three words and eight tenths of the
    words of the shorter one.
    """
    target = words(target)
    shared = lambda suspect: len(target & words(suspect))
    return any(shared(suspect) >= 3 and shared(suspect) >= 0.8 * min(len(target), len(words(suspect))) for suspect in suspects)


# ---- the comparison without a model ----
copies = []
for part, audit, sources in (("fitting", "audit-fit17-copies.jsonl", ("copies-dev3.jsonl",)), ("measured", "audit-test17.jsonl", ("copies-test17.jsonl",))):
    source = {(r["set"], r["id"]): r for name in sources for r in load(name)}
    for row in load(audit):
        kind, pages = kind_and_pages(row["set"])
        entry = {"part": part, "pages": pages, "unit": unit(row["id"], label_of(row["set"])), "kind": kind, "compared": row["compared"],
                 "suspects": len(row["suspects"])}
        if kind not in ("join", "contract"):
            entry["named"] = names(source[(row["set"], row["id"])]["damage"], [s[1] for s in row["suspects"]])
        copies.append(entry)
write("copies", copies)


# ---- the whole check ----
def last_attempts(name):
    """The checks of one run: for each pair its last row, and whether a row before it ended without a verdict.

    A pair whose check ended without a verdict was run once more, and the second row follows the first in the file.
    """
    attempts = {}
    for row in load(name):
        attempts.setdefault((row["set"], row["id"]), []).append(row)
    return [(rows[-1], any("error" in earlier for earlier in rows[:-1])) for rows in attempts.values()]


def parts_of(result):
    facts, sentences = result["facts"], result.get("sentences", [])
    kept = sum(f[0] == "kept" for f in facts)
    lost = [s for s in sentences if s[1] != "kept"]
    return {"verdict": "accepted" if result["accepted"] else "rejected", "sound": result["sound"], "named": result["named"],
            "statements": len(facts), "kept": kept, "suspects": len(sentences), "lost": len(lost)}, facts, lost


damage = []
for part, runs in (("fitting", (("gpt", "compare-dev19-damage-gpt.jsonl"), ("qwen", "compare-dev19-damage.jsonl"))),
                   ("measured", (("gpt", "compare-test17-damage-gpt.jsonl"), ("qwen", "compare-test17-damage.jsonl")))):
    for judge, name in runs:
        for result, retried in last_attempts(name):
            kind, pages = kind_and_pages(result["set"])
            entry = {"part": part, "pages": pages, "unit": unit(result["id"], label_of(result["set"])), "kind": kind, "judge": MODELS[judge]}
            if retried:
                entry["retried"] = True
            if "error" in result:
                entry["verdict"] = "no verdict"
            else:
                counted, _, lost = parts_of(result)
                entry.update(counted, reported=names(result["damage"], [s[2] for s in lost]))
            damage.append(entry)
write("damage", damage)

revisions = []
PAGES = {"blog": "set-s", "setS": "set-s", "gptF": "fresh", "qwenF": "fresh"}
for part, runs in (("fitting", (("gpt", "compare-dev19-gpt.jsonl"), ("qwen", "compare-dev19-qwen.jsonl"))),
                   ("read", (("gpt", "compare-test17-gpt.jsonl"), ("qwen", "compare-test17-qwen.jsonl"))),
                   ("fresh", (("gpt", "compare-test17-fresh-gpt.jsonl"), ("qwen", "compare-test17-fresh-qwen.jsonl")))):
    for reviser, name in runs:
        judge = "qwen" if reviser == "gpt" else "gpt"
        for result, retried in last_attempts(name):
            entry = {"part": part, "pages": PAGES[result["set"]], "unit": unit(result["id"], label_of(result["set"])), "reviser": MODELS[reviser],
                     "judge": MODELS[judge]}
            if retried:
                entry["retried"] = True
            if "error" in result:
                entry["verdict"] = "no verdict"
            else:
                counted, facts, lost = parts_of(result)
                entry.update(counted, not_kept={kind: sum(f[0] == kind for f in facts) for kind in ("changed", "missing", "unquoted")},
                             lost_by={reason: sum(s[0] == reason for s in lost) for reason in ("missing", "negation", "reworded")},
                             em_dashes=result["dashes"])
            revisions.append(entry)
write("revisions", revisions)
