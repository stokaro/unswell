#!/usr/bin/env python3
"""Write the data of the record of the fourth form of a closing denial from the kept runs.

usage: export.py <workspace> <record directory>

Reads what two builds of syntax.closing-denial-density report on the fitting parts, the measured parts, and the sets
without parts. Writes data/pages.jsonl.gz. No text, path, repository, or title is written. A post of set S has the
number it has in the earlier records; every other page has a hash of its set and id, the same as in the record of the
two page habits.
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

posts = [json.loads(line) for line in open(G / "set.jsonl")]
salt = (G / "set-salt.txt").read_text().strip()
set_id = {r["id"]: f"s-{position + 1:02d}" for position, r in
          enumerate(sorted(posts, key=lambda r: hashlib.sha256((salt + r["id"]).encode()).hexdigest()))}


def opaque(*parts):
    return hashlib.sha256("\x00".join(parts).encode()).hexdigest()[:12]


pages = []
for name in ("habit-forms-dev4.jsonl", "habit-forms-test1.jsonl", "habit-forms-none1.jsonl"):
    for row in map(json.loads, open(G / name)):
        if "error" in row:
            raise SystemExit(f"a page without a report in {name}")
        corpus = CORPUS[row["corpus"]]
        # "three" and "four" are the places that the rule lists with its limits as they ship, 0 when it is quiet.
        # "counted" is every closing sentence that the rule takes for a denial, with no limit on the page.
        pages.append({"corpus": corpus, "part": PART[row["split"]], "id": set_id[row["id"]] if corpus == "set-s" else opaque(corpus, row["id"]),
                      "three": row["old"], "four": row["new"], "counted_three": len(row["old_sentences"]),
                      "counted_four": len(row["new_sentences"]),
                      "kept": sum(1 for sentence in row["old_sentences"] if sentence in row["new_sentences"])})
pages.sort(key=lambda r: (r["corpus"], r["part"], r["id"]))
data = "".join(json.dumps(row, sort_keys=True) + "\n" for row in pages).encode()
with open(record / "data" / "pages.jsonl.gz", "wb") as raw, gzip.GzipFile(fileobj=raw, mode="wb", mtime=0, filename="") as out:
    out.write(data)
print("pages", len(pages))
