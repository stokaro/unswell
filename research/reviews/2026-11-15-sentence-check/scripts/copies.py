#!/usr/bin/env python3
"""Make damaged copies and harmless copies of pages, for the sentence comparison of the check of a revision.

usage: copies.py <out jsonl> --salt <text> <label>=<pages jsonl>... [--split test|dev] [--avoid <damage jsonl>]

A damaged copy has one sentence of running prose removed ("sentence"), or the word "not" removed from one sentence
("negation"). A harmless copy has two sentences of one paragraph joined with ", and" ("join"), or its negations
written as contractions ("contract"): nothing that the page says is gone. The sentence is picked by a hash of the
salt and the id among the sentences of paragraphs that have eight words or more and no digit, code, or link. With
--avoid, a sentence that an earlier file damaged in the same page is not picked again.

A row holds the set (the kind of copy and the label of the pages), the id, the page, the copy, and the sentence
that was removed, changed, or joined.
"""
import hashlib, json, re, sys
from pathlib import Path

# The lines of running prose are told as in the earlier record.
sys.path.insert(0, str(Path(__file__).resolve().parents[2] / "2026-11-15-revision-check" / "scripts"))
import prepare as lib

out = Path(sys.argv[1])
option = lambda name, default: sys.argv[sys.argv.index(name) + 1] if name in sys.argv else default
salt, split, avoid_path = option("--salt", ""), option("--split", "test"), option("--avoid", "")
assert salt, "a salt is required"
SENTENCE = re.compile(r"(?<=[.!?])\s+(?=[A-Z])")
CONTRACTIONS = ((" do not ", " don't "), (" does not ", " doesn't "), (" did not ", " didn't "), (" is not ", " isn't "),
                (" are not ", " aren't "), (" was not ", " wasn't "), (" cannot ", " can't "), (" will not ", " won't "))
avoided = {}
if avoid_path:
    for row in map(json.loads, open(avoid_path)):
        avoided.setdefault(row["id"], set()).add(row["damage"])
plain = lambda sentence: len(sentence.split()) >= 8 and not re.search(r"[0-9`\[\]()]", sentence)


def lines_of(text):
    """Lines of running prose with their sentences."""
    for number, line in enumerate(text.split("\n")):
        if line.strip() and not lib.STARTS_LINE.match(line):
            sentences = SENTENCE.split(line)
            if len(sentences) >= 2:
                yield number, sentences


def picked(found, kind, key):
    return min(found, key=lambda item: hashlib.sha256(f"{salt}:{kind}:{key}:{item[0]}:{item[1]}".encode()).hexdigest()) if found else None


def rewritten(text, number, sentences):
    lines = text.split("\n")
    lines[number] = " ".join(sentences)
    return "\n".join(lines)


def damaged(text, key, kind):
    found = [(number, position, sentences) for number, sentences in lines_of(text) for position, sentence in enumerate(sentences)
             if plain(sentence) and (kind == "sentence" or " not " in sentence) and sentence not in avoided.get(key, ())]
    choice = picked(found, kind, key)
    if choice is None:
        return None, None
    number, position, sentences = choice
    sentence = sentences[position]
    changed = sentences[:position] + ([] if kind == "sentence" else [sentence.replace(" not ", " ", 1)]) + sentences[position + 1:]
    return rewritten(text, number, changed), sentence


def joined(text, key):
    found = [(number, position, sentences) for number, sentences in lines_of(text) for position in range(len(sentences) - 1)
             if plain(sentences[position]) and plain(sentences[position + 1]) and sentences[position].endswith(".")]
    choice = picked(found, "join", key)
    if choice is None:
        return None, None
    number, position, sentences = choice
    first, second = sentences[position], sentences[position + 1]
    both = first[:-1] + ", and " + second[0].lower() + second[1:]
    return rewritten(text, number, sentences[:position] + [both] + sentences[position + 2:]), first + " " + second


def contracted(text):
    lines, changed = text.split("\n"), 0
    for number, line in enumerate(lines):
        if not line.strip() or lib.STARTS_LINE.match(line) or "`" in line:
            continue
        for long, short in CONTRACTIONS:
            changed += line.count(long)
            line = line.replace(long, short)
        lines[number] = line
    return ("\n".join(lines), f"{changed} contractions") if changed else (None, None)


rows = []
for spec in [argument for argument in sys.argv[2:] if "=" in argument and not argument.startswith("--")]:
    label, path = spec.split("=", 1)
    for post in map(json.loads, open(path)):
        if post.get("split", split) != split:
            continue
        title = "# " + (post.get("title") or "Post") + "\n\n"
        copies = {"sentence": damaged(post["text"], post["id"], "sentence"), "negation": damaged(post["text"], post["id"], "negation"),
                  "join": joined(post["text"], post["id"]), "contract": contracted(post["text"])}
        for kind, (copy, sentence) in copies.items():
            if copy is not None:
                rows.append({"set": f"{kind}-{label}", "id": post["id"], "original": title + post["text"], "revised": title + copy,
                             "damage": sentence})
out.write_text("".join(json.dumps(row, ensure_ascii=False) + "\n" for row in rows))
counts = {}
for row in rows:
    counts[row["set"]] = counts.get(row["set"], 0) + 1
print(len(rows), "copies:", counts)
