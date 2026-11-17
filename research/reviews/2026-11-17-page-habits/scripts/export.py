#!/usr/bin/env python3
"""Write the data of the record of the two page habits from the kept runs.

usage: export.py <workspace> <record directory>

Reads the load of both rules on the fitting parts, the measured parts, and the corpora without parts, and the
pairs of the reference of commits that remove AI-sounding wording. Writes data/pages.jsonl.gz and
data/pairs.jsonl.gz. No text, path, repository, or title is written. A post of set S has the number it has in the
earlier records; every other page has a hash of its corpus and id.
"""
import gzip, hashlib, json, sys
from pathlib import Path

W, record = Path(sys.argv[1]), Path(sys.argv[2])
G = W / "guided"
CORPUS = {
    "posts of set S": "set-s", "blog posts of people": "human-blog", "articles of people under a style guide": "human-styled",
    "documentation pages of people": "human-docs", "model texts, guided, first writer": "model-guided-1",
    "model texts, guided, second writer": "model-guided-2", "model texts, plain, first writer": "model-plain-1",
    "model texts, plain, second writer": "model-plain-2", "fresh model texts, guided, first writer": "fresh-1",
    "fresh model texts, guided, second writer": "fresh-2", "pages of public repositories today": "public-today",
    "pages of the panel corpus that models wrote or polished": "panel",
}
PART = {"dev": "fitting", "test": "measured", "none": "whole"}
DENIAL, REGISTER = "syntax.closing-denial-density", "syntax.uniform-register"

posts = [json.loads(line) for line in open(G / "set.jsonl")]
salt = (G / "set-salt.txt").read_text().strip()
set_id = {r["id"]: f"s-{position + 1:02d}" for position, r in
          enumerate(sorted(posts, key=lambda r: hashlib.sha256((salt + r["id"]).encode()).hexdigest()))}


def opaque(*parts):
    return hashlib.sha256("\x00".join(parts).encode()).hexdigest()[:12]


def write(name, rows):
    data = "".join(json.dumps(row, sort_keys=True) + "\n" for row in rows).encode()
    with open(record / "data" / (name + ".jsonl.gz"), "wb") as raw, gzip.GzipFile(fileobj=raw, mode="wb", mtime=0, filename="") as out:
        out.write(data)
    print(name, len(rows))


pages = []
for name in ("habit-load-dev2.jsonl", "habit-load-test1.jsonl", "habit-load-none2.jsonl"):
    for row in map(json.loads, open(G / name)):
        if "error" in row:
            raise SystemExit(f"a page without a report in {name}")
        corpus = CORPUS[row["corpus"]]
        pages.append({"corpus": corpus, "part": PART[row["split"]], "id": set_id[row["id"]] if corpus == "set-s" else opaque(corpus, row["id"]),
                      "words": row["words"], "denial": row[DENIAL], "register": row[REGISTER]})
pages.sort(key=lambda r: (r["corpus"], r["part"], r["id"]))
write("pages", pages)

pairs = []
for part, name in (("development", "habit-deslop-dev2.jsonl"), ("test", "habit-deslop-test.jsonl")):
    for row in map(json.loads, open(G / name)):
        if row.get("skip") or row.get("assistant_trailer"):
            continue
        reported = lambda side, rule: int(row[side]["counts"].get(rule, 0) > 0)
        pairs.append({"part": part, "id": opaque("pair", row["id"]), "denial": [reported("before", DENIAL), reported("after", DENIAL)],
                      "register": [reported("before", REGISTER), reported("after", REGISTER)]})
pairs.sort(key=lambda r: (r["part"], r["id"]))
write("pairs", pairs)
