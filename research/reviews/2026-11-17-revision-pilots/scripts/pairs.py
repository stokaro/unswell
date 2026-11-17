#!/usr/bin/env python3
"""Pair the opening of each revised fitting post with the three human posts it was compared with as it is.

usage: pairs.py <name> <revisions jsonl> <out pairs jsonl>

The pairs are those of pairs-dev-set.jsonl with the post replaced by its revision; a post without a revision is
left out. In every row "first" is the revised post and "second" the human post; the judge script hides the order.
"""
import json, re, sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))
import guided_lib as lib

name, revisions, out = sys.argv[1], sys.argv[2], Path(sys.argv[3])
G = Path(__file__).resolve().parents[2] / "guided"
opening = lambda page: lib.excerpt(re.sub(r"\A#\s+[^\n]*\n+", "", lib.clean(page), count=1), 450)
revised = {r["id"]: opening(r["revised"]) for r in map(json.loads, open(revisions)) if "revised" in r}
rows = [{**pair, "id": f"{name}:" + pair["id"], "first": revised[pair["unit"]]}
        for pair in map(json.loads, open(G / "pairs-dev-set.jsonl")) if pair["comparison"] == "set" and pair["unit"] in revised]
out.write_text("".join(json.dumps(r, ensure_ascii=False) + "\n" for r in rows))
print(name, "revised posts", len(revised), "| pairs", len(rows))
