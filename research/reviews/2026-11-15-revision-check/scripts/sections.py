#!/usr/bin/env python3
"""Ask one judge about a page and its revision section by section, in both orders, and count the sections it is sure of.

usage: sections.py <workspace> <model kind> <cache dir> <revisions jsonl> <wording file> <out jsonl> [--limit N] [--workers N]

A page and its revision are cut at their headings. A section is compared when both texts have it under the same
heading, both have 80 words or more, and the two differ. Of a long section the judge reads whole blocks up to about
450 words. A section counts for the revision when the judge names the page in both orders, against it when the judge
names the revision in both orders, and not at all otherwise.
"""
import collections, concurrent.futures, json, re, subprocess, sys
from pathlib import Path

W, kind, cache, revisions, wording, out = sys.argv[1].rstrip("/") + "/", sys.argv[2], sys.argv[3], sys.argv[4], Path(sys.argv[5]).read_text(), Path(sys.argv[6])
option = lambda name, default: sys.argv[sys.argv.index(name) + 1] if name in sys.argv else default
limit, workers = int(option("--limit", "1000")), int(option("--workers", "6"))


def sections(page):
    """Return the heading and the text of each section; text before the first heading has an empty heading."""
    found, heading, lines, fenced = [], "", [], False
    for line in page.split("\n"):
        if line.strip().startswith(("```", "~~~")):
            fenced = not fenced
        if not fenced and re.match(r"^#{1,6}\s", line):
            found.append((heading, "\n".join(lines).strip()))
            heading, lines = line.strip(), []
        else:
            lines.append(line)
    found.append((heading, "\n".join(lines).strip()))
    return [(h, t) for h, t in found if t]


def shortened(text):
    blocks, words = [], 0
    for block in re.split(r"\n\s*\n", text):
        blocks.append(block)
        words += len(block.split())
        if words >= 450:
            break
    return "\n\n".join(blocks) + "\n"


def ask(prompt):
    answer = subprocess.run(["python3", W + "ask_model.py", kind, cache], input=prompt.encode(), capture_output=True).stdout.decode()
    match = re.search(r'"more[a-z_]*"\s*:\s*"(A|B|same)"', answer)
    return match.group(1) if match else None


def one(row):
    before, after = sections(row["original"]), sections(row["revised"])
    votes = []
    if [h for h, _ in before] == [h for h, _ in after]:
        for (heading, page), (_, revised) in zip(before, after):
            if page == revised or min(len(page.split()), len(revised.split())) < 80:
                continue
            page, revised = shortened(page), shortened(revised)
            first, second = ask(wording % (page, revised)), ask(wording % (revised, page))
            votes.append(1 if (first, second) == ("A", "B") else -1 if (first, second) == ("B", "A") else 0)
    return {"id": row["id"], "aligned": [h for h, _ in before] == [h for h, _ in after], "votes": votes}


rows = [r for r in map(json.loads, open(revisions)) if "revised" in r][:limit]
with concurrent.futures.ThreadPoolExecutor(workers) as pool:
    results = list(pool.map(one, rows))
out.write_text("".join(json.dumps(r) + "\n" for r in results))
net = [sum(r["votes"]) for r in results]
print(len(results), "pairs; sections compared", sum(len(r["votes"]) for r in results), "| for the revision", sum(v == 1 for r in results for v in r["votes"]),
      "| against", sum(v == -1 for r in results for v in r["votes"]), "| unsure", sum(v == 0 for r in results for v in r["votes"]))
print("pairs by net votes:", dict(sorted(collections.Counter(net).items())), "| better", sum(n > 0 for n in net), "| worse", sum(n < 0 for n in net), "| no difference", sum(n == 0 for n in net))
