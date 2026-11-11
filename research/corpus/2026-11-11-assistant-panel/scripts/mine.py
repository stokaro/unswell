#!/usr/bin/env python3
"""Mine formulaic phrases from the annotated documents.

usage: mine.py <workspace> <out json>

Repositories are split into mining and development sets by a hash of their
name. A phrase is a word n-gram of one to four tokens. For each phrase the
script counts the items that contain it by label: flagged by both annotators,
by one, or by none. Selection uses the mining repositories only.
"""
import collections, hashlib, json, re, sys
from pathlib import Path

W, OUT = Path(sys.argv[1]), sys.argv[2]
TOKEN = re.compile(r"[A-Za-z]+(?:[-'][A-Za-z]+)*")
MIN_ITEMS, MIN_REPOS = 8, 5
STOP = set("the a an of to in and or for on with by is are be as at this that these those it its from which".split())


def dev_repo(name):
    return int(hashlib.sha256(name.encode()).hexdigest(), 16) % 4 == 0


def load():
    rows = []
    for line in open(W / "docs.jsonl"):
        doc = json.loads(line)
        paths = [W / "ann" / n / f"{doc['doc_id']}.json" for n in ("astra", "claude")]
        if not all(p.exists() for p in paths):
            continue
        astra, claude = [{f["item"]: f for f in json.loads(p.read_text())["flags"]} for p in paths]
        for item in doc["items"]:
            a, c = astra.get(item["id"]), claude.get(item["id"])
            pa = bool(a and a["label"] == "needs_revision")
            pc = bool(c and c["label"] == "needs_revision")
            label = "both" if pa and pc else "one" if pa or pc else "none"
            categories = (a or {}).get("categories", []) + (c or {}).get("categories", [])
            tokens = [t.lower() for t in TOKEN.findall(" ".join(item["text"].split()))]
            rows.append((doc["repository"], doc["origin"], label, tokens, categories))
    return rows


def grams(tokens):
    seen = set()
    for n in (1, 2, 3, 4):
        for i in range(len(tokens) - n + 1):
            gram = tokens[i:i + n]
            if gram[0] in STOP and n == 1:
                continue
            if all(t in STOP for t in gram):
                continue
            seen.add(" ".join(gram))
    return seen


rows = load()
repos = sorted({r[0] for r in rows})
dev = {r for r in repos if dev_repo(r)}
print(f"repositories: mining={len(repos) - len(dev)} development={len(dev)} {sorted(dev)}")
stats = {"mine": collections.defaultdict(lambda: [0, 0, 0]), "dev": collections.defaultdict(lambda: [0, 0, 0])}
repo_hits = collections.defaultdict(set)
category = collections.defaultdict(collections.Counter)
slot = {"both": 0, "one": 1, "none": 2}
for repo, origin, label, tokens, categories in rows:
    part = "dev" if repo in dev else "mine"
    for gram in grams(tokens):
        stats[part][gram][slot[label]] += 1
        if part == "mine" and label != "none":
            repo_hits[gram].add(repo)
            category[gram].update(categories)
candidates = []
for gram, (both, one, none) in stats["mine"].items():
    total = both + one + none
    if total < MIN_ITEMS or len(repo_hits[gram]) < MIN_REPOS:
        continue
    strict, lenient = both / total, (both + one) / total
    if lenient < 0.6:
        continue
    d = stats["dev"].get(gram, [0, 0, 0])
    candidates.append({"phrase": gram, "mine": [both, one, none], "strict": round(strict, 3),
                       "lenient": round(lenient, 3), "repos": len(repo_hits[gram]), "dev": d,
                       "category": category[gram].most_common(1)[0][0] if category[gram] else ""})
candidates.sort(key=lambda c: (-c["lenient"], -sum(c["mine"])))
json.dump({"development_repositories": sorted(dev), "candidates": candidates}, open(W / OUT, "w"), indent=1)
print(f"phrases with at least {MIN_ITEMS} items in {MIN_REPOS}+ repositories and lenient precision >= 0.6: {len(candidates)}")
for floor in (0.9, 0.8, 0.7):
    chosen = [c for c in candidates if c["lenient"] >= floor]
    b = sum(c["dev"][0] for c in chosen); o = sum(c["dev"][1] for c in chosen); n = sum(c["dev"][2] for c in chosen)
    print(f"  lenient >= {floor}: phrases={len(chosen)} development occurrences both/one/none={b}/{o}/{n} "
          f"lenient precision on development={100 * (b + o) / max(1, b + o + n):.1f}% strict={100 * b / max(1, b + o + n):.1f}%")
print("top single words and short phrases by support:")
for c in sorted([c for c in candidates if c["lenient"] >= 0.7 and len(c["phrase"].split()) <= 2], key=lambda c: -sum(c["mine"]))[:60]:
    print(f"   {c['phrase']:28s} mine {c['mine']} lenient {c['lenient']:.2f} strict {c['strict']:.2f} dev {c['dev']} {c['category']}")
