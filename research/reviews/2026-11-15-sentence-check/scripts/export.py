#!/usr/bin/env python3
"""Export the measurement of the sentence comparison into its record, without any text.

usage: export.py <workspace> <record dir>

copies     one row for each copy of a page: what the comparison without a model found on it
damage     one row for each damaged copy and judge: what the whole check did with it
revisions  one row for each revision: what the whole check did with it under the judge of the other family
variants   the rule as it grew on the fitting half: each step on the first damaged copies and on ordinary revisions

The posts of set S are numbered as in the earlier records: in the order of a salted hash whose salt stays in the
workspace. The record holds neither their names nor their words. The damaged copies of the fitting half, and its
revisions under the earlier build, were checked by a build that did not yet ask a clearing quote for a word of the
sentence; that rule is applied here to its kept answers, so that those rows say what the final build decides about
a sentence.
"""
import gzip, hashlib, json, re, sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))
from words import content_words

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
    """The pages that a set of copies or revisions is made of: "setS", "gptF", and so on; the first copies are of set S."""
    return "setS" if name.startswith("damage-") or name == "blog" else name.rsplit("-", 1)[-1]


words = lambda text: set(re.findall(r"[a-z0-9]+", text.lower()))


def write(name, rows):
    with open(record / "data" / (name + ".jsonl.gz"), "wb") as raw, gzip.GzipFile(fileobj=raw, mode="wb", mtime=0, filename="") as handle:
        for row in rows:
            handle.write((json.dumps(row, sort_keys=True, ensure_ascii=False) + "\n").encode())
    print(name, len(rows))


def kind_and_pages(name):
    """The set of a copy names its kind and its pages: "damage-sentence" of the first copies, or "sentence-setS"."""
    first, second = name.split("-", 1)
    kind = second if first == "damage" else first
    return kind, ("set-s" if first == "damage" or second == "setS" else "fresh")


def names(target, suspects):
    """Whether a sentence is among the suspects: one of them carries at least eight tenths of its words."""
    target = words(target)
    return any(len(target & words(suspect)) >= 0.8 * len(target) for suspect in suspects)


# ---- the comparison without a model ----
damaged_text = {}
copies = []
for part, audit, sources in (("fitting, first copies", "audit-fit-final-damage.jsonl", ("damage-dev.jsonl",)),
                             ("fitting, second copies", "audit-fit-final-damage.jsonl", ("copies-dev2.jsonl",)),
                             ("measured", "audit-test15.jsonl", ("copies-test15.jsonl",))):
    source = {(r["set"], r["id"]): r for name in sources for r in load(name)}
    for row in load(audit):
        if (row["set"], row["id"]) not in source:
            continue
        kind, pages = kind_and_pages(row["set"])
        entry = {"part": part, "pages": pages, "unit": unit(row["id"], label_of(row["set"])), "kind": kind, "compared": row["compared"],
                 "suspects": len(row["suspects"])}
        if kind in ("sentence", "negation"):
            entry["named"] = names(source[(row["set"], row["id"])]["damage"], [s[1] for s in row["suspects"]])
        copies.append(entry)
write("copies", copies)

# ---- the whole check ----
own = {}
for audit in ("audit-fit-final-damage.jsonl", "audit-fit-final-gpt.jsonl", "audit-fit-final-qwen.jsonl"):
    for row in load(audit):
        for reason, text, _, sentence_words in row["suspects"]:
            own[(row["set"], row["id"], text)] = set(sentence_words)


def final_rule(row):
    """Apply the rule of the final build to the answers of the fitting half: a cleared missing sentence needs a quote with one of its own words."""
    for sentence in row.get("sentences", []):
        key = (row["set"], row["id"], sentence[2])
        if sentence[0] == "missing" and sentence[1] == "kept" and key in own and not (own[key] & content_words(sentence[3])):
            sentence[1] = "unquoted"
    return row


def last_attempts(name):
    """The checks of one run: for each pair its last row, and whether a row before it ended without a verdict.

    A pair whose check ended without a verdict was run once more, and the second row follows the first in the file.
    """
    attempts = {}
    for row in load(name):
        attempts.setdefault((row["set"], row["id"]), []).append(row)
    return [(rows[-1], any("error" in earlier for earlier in rows[:-1])) for rows in attempts.values()]


def parts_of(result):
    """The parts of one verdict. An accepted verdict of the fitting half is read again under the final rule."""
    facts, sentences = result["facts"], result.get("sentences", [])
    kept = sum(f[0] == "kept" for f in facts)
    lost = [s for s in sentences if s[1] != "kept"]
    accepted = result["accepted"] and not lost
    return {"verdict": "accepted" if accepted else "rejected", "sound": result["sound"], "named": result["named"], "statements": len(facts),
            "kept": kept, "suspects": len(sentences), "lost": len(lost)}, facts, lost


damage = []
for part, fitted, runs in (("fitting, first copies", True, (("gpt", "compare-dev15-damage-gpt.jsonl"), ("qwen", "compare-dev15-damage.jsonl"))),
                           ("measured", False, (("gpt", "compare-test15-damage-gpt.jsonl"), ("qwen", "compare-test15-damage.jsonl")))):
    for judge, name in runs:
        for result, retried in last_attempts(name):
            kind, pages = kind_and_pages(result["set"])
            entry = {"part": part, "pages": pages, "unit": unit(result["id"], label_of(result["set"])), "kind": kind, "judge": MODELS[judge]}
            if retried:
                entry["retried"] = True
            if "error" in result:
                entry["verdict"] = "no verdict"
            else:
                counted, _, lost = parts_of(final_rule(result) if fitted else result)
                entry.update(counted, reported=names(result["damage"], [s[2] for s in lost]))
            damage.append(entry)
write("damage", damage)

revisions = []
PAGES = {"blog": "set-s", "setS": "set-s", "gptG": "guided", "qwenG": "guided", "gptF": "fresh", "qwenF": "fresh"}
for part, fitted, runs in (("fitting, earlier build", True, (("gpt", "compare-dev15-gpt-first.jsonl"), ("qwen", "compare-dev15-qwen.jsonl"))),
                           ("fitting", False, (("gpt", "compare-dev16-gpt.jsonl"), ("qwen", "compare-dev16-qwen.jsonl"))),
                           ("read", False, (("gpt", "compare-test15-gpt.jsonl"), ("qwen", "compare-test15-qwen.jsonl"))),
                           ("fresh", False, (("gpt", "compare-test15-fresh-gpt.jsonl"), ("qwen", "compare-test15-fresh-qwen.jsonl")))):
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
                counted, facts, lost = parts_of(final_rule(result) if fitted else result)
                entry.update(counted, not_kept={kind: sum(f[0] == kind for f in facts) for kind in ("changed", "missing", "unquoted")},
                             lost_by={reason: sum(s[0] == reason for s in lost) for reason in ("missing", "negation")}, em_dashes=result["dashes"])
            revisions.append(entry)
write("revisions", revisions)

# ---- the rule as it grew ----
VARIANTS = (("audit1", "all content words of a sentence, anywhere in the revision"),
            ("audit2", "and its own words, near its place"),
            ("audit3", "and words without their endings"),
            ("final", "and paragraphs cut into sentences: the rule of the build"))
variants = []
first = {(r["set"], r["id"]): r for r in load("damage-dev.jsonl")}
for key, name in VARIANTS:
    for row in load(f"audit-fit-{key}-damage.jsonl"):
        if (row["set"], row["id"]) in first:
            variants.append({"variant": name, "kind": kind_and_pages(row["set"])[0], "unit": set_id[row["id"]], "suspects": len(row["suspects"]),
                             "named": names(first[(row["set"], row["id"])]["damage"], [s[1] for s in row["suspects"]])})
    for reviser in ("gpt", "qwen"):
        for row in load(f"audit-fit-{key}-{reviser}.jsonl"):
            variants.append({"variant": name, "kind": "revision", "reviser": MODELS[reviser], "unit": set_id[row["id"]], "suspects": len(row["suspects"])})
write("variants", variants)
