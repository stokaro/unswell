#!/usr/bin/env python3
"""Make damaged copies of each post, to see whether the check of a revision notices a lost or a changed statement.

usage: damage.py <posts jsonl> <out jsonl> [--split test|dev]

Each post gets up to two copies. In "sentence" one sentence of running prose is removed. In "negation" the word
"not" is removed from one sentence, which turns its statement around. The sentence is picked by a hash of the id of
the post among the sentences of paragraphs that have eight words or more and no digit, code, or link, so that the
checks that need no model do not notice the damage. A row holds the set, the id, the post, the damaged copy, and
the sentence that was removed or changed.
"""
import hashlib, json, re, sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))
import prepare as lib

posts, out = sys.argv[1], Path(sys.argv[2])
split = sys.argv[sys.argv.index("--split") + 1] if "--split" in sys.argv else "test"
SENTENCE = re.compile(r"(?<=[.!?])\s+(?=[A-Z])")


def candidates(text, needs_not):
    """Sentences of paragraphs with their line number, in the order of a hash."""
    found = []
    for number, line in enumerate(text.split("\n")):
        if not line.strip() or lib.STARTS_LINE.match(line):
            continue
        sentences = SENTENCE.split(line)
        if len(sentences) < 2:
            continue
        for position, sentence in enumerate(sentences):
            plain = len(sentence.split()) >= 8 and not re.search(r"[0-9`\[\]()]", sentence)
            if plain and (not needs_not or " not " in sentence):
                found.append((number, position, sentence))
    return found


def damaged(text, key, kind):
    found = candidates(text, kind == "negation")
    if not found:
        return None, None
    found.sort(key=lambda item: hashlib.sha256(f"{kind}:{key}:{item[0]}:{item[1]}".encode()).hexdigest())
    number, position, sentence = found[0]
    lines = text.split("\n")
    sentences = SENTENCE.split(lines[number])
    if kind == "sentence":
        del sentences[position]
    else:
        sentences[position] = sentence.replace(" not ", " ", 1)
    lines[number] = " ".join(sentences)
    return "\n".join(lines), sentence


rows = []
for post in map(json.loads, open(posts)):
    if post["split"] != split:
        continue
    page = "# " + (post.get("title") or "Post") + "\n\n" + post["text"]
    for kind in ("sentence", "negation"):
        copy, sentence = damaged(post["text"], post["id"], kind)
        if copy is None:
            continue
        rows.append({"set": "damage-" + kind, "id": post["id"], "original": page,
                     "revised": "# " + (post.get("title") or "Post") + "\n\n" + copy, "damage": sentence})
out.write_text("".join(json.dumps(row, ensure_ascii=False) + "\n" for row in rows))
print(len(rows), "damaged copies:", {kind: sum(r["set"] == "damage-" + kind for r in rows) for kind in ("sentence", "negation")})
