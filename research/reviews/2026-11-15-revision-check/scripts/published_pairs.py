#!/usr/bin/env python3
"""Build the pairs of what a writer who follows the product publishes against human text.

usage: published_pairs.py <name> <posts jsonl> <record pairs jsonl> <styled jsonl> <out prefix> [<revisions jsonl> <compare jsonl>]

For each post of set S of the measured half the published text is its revision when the check accepted it, and the
post as it was otherwise. Without a revisions file the published text is the post. Two files are written:
<out prefix>-blog.jsonl pairs the opening with the three human blog posts of the record, and <out prefix>-styled.jsonl
with three matched articles of the control, picked by a hash. In every row "first" is the post and "second" the human
text; the judge script hides the order.
"""
import hashlib, json, re, sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))
import prepare as lib

name, posts_path, record_pairs, styled_path, prefix = sys.argv[1:6]
revisions_path, compare_path = (sys.argv[6], sys.argv[7]) if len(sys.argv) > 7 else (None, None)
MATCH = 0.732


def paragraph_share(text):
    inside = total = 0
    for line in text.split("\n"):
        count = lib.words(line)
        total += count
        if count and not lib.STARTS_LINE.match(line):
            inside += count
    return inside / total if total else 0.0


def opening(page):
    return lib.excerpt(re.sub(r"\A#\s+[^\n]*\n+", "", lib.clean(page), count=1), 450)


posts = [p for p in map(json.loads, open(posts_path)) if p["split"] == "test"]
published = {p["id"]: lib.excerpt(p["text"], 450) for p in posts}
accepted = 0
if revisions_path:
    verdicts = {r["id"]: r.get("accepted", False) for r in map(json.loads, open(compare_path)) if r["set"] == "setS"}
    for row in map(json.loads, open(revisions_path)):
        if row["set"] == "setS" and "revised" in row and verdicts.get(row["id"]):
            published[row["id"]] = opening(row["revised"])
            accepted += 1
blog = [r for r in map(json.loads, open(record_pairs)) if r["comparison"] == "set"]
articles = [a for a in map(json.loads, open(styled_path)) if a["split"] == "test" and paragraph_share(a["text"]) >= MATCH]
rows_blog = [{**pair, "id": f"{name}:" + pair["id"], "first": published[pair["unit"]]} for pair in blog]
rows_styled = []
for post in posts:
    picked = sorted(articles, key=lambda a: hashlib.sha256(("styled:" + post["id"] + a["id"]).encode()).hexdigest())[:3]
    for article in picked:
        rows_styled.append({"id": f"{name}:styled:{post['id']}:{article['id']}", "comparison": "set", "family": None, "guide": None,
                            "unit": post["id"], "first": published[post["id"]], "second": lib.excerpt(article["text"], 450)})
Path(prefix + "-blog.jsonl").write_text("".join(json.dumps(r, ensure_ascii=False) + "\n" for r in rows_blog))
Path(prefix + "-styled.jsonl").write_text("".join(json.dumps(r, ensure_ascii=False) + "\n" for r in rows_styled))
print(name, "posts", len(posts), "| published as revised", accepted, "| pairs", len(rows_blog), len(rows_styled))
