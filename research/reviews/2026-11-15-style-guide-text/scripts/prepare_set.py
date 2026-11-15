#!/usr/bin/env python3
"""Prepare the posts of a blog checkout the same way as the human posts.

usage: prepare_set.py <posts dir> <out jsonl> <prefix>

Each post is <posts dir>/<slug>/index.md. The slug decides the half of the post.
"""
import json, sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))
import prepare as lib

# The half of a post follows from a hash of a prefix and its name. The prefix names the source and is not kept here.
SET_PREFIX = sys.argv[3] if len(sys.argv) > 3 else ""

rows = []
for post in sorted(Path(sys.argv[1]).glob("*/index.md")):
    raw = post.read_text()
    if "\ndraft: true" in raw.split("\n---", 2)[1] if raw.startswith("---") else False:
        continue
    text = lib.clean(raw)
    if lib.words(text) < 500:
        continue
    slug = post.parent.name
    rows.append({"id": slug, "title": lib.front_matter_title(raw), "words": lib.words(text),
                 "split": lib.bucket(SET_PREFIX + slug), "text": text})
Path(sys.argv[2]).write_text("".join(json.dumps(row, ensure_ascii=False) + "\n" for row in rows))
print(len(rows), "posts;", sum(r["split"] == "test" for r in rows), "measured,", sum(r["split"] == "dev" for r in rows), "for fitting")
