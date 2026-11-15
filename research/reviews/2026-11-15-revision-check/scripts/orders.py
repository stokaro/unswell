#!/usr/bin/env python3
"""Ask one judge about a page and its revision in both orders, with a given wording, and count what it names.

usage: orders.py <workspace> <model kind> <cache dir> <revisions jsonl> <wording file> <out jsonl> [--limit N] [--workers N]
                        [--raw]

The wording file holds the question; %s stands twice, for text A and for text B. The texts are the openings of the page
and of its revision as the record prepares them, or the first blocks of the source up to about 450 words with --raw.
"""
import collections, concurrent.futures, hashlib, json, re, subprocess, sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))
import prepare as lib

W, kind, cache, revisions, wording, out = sys.argv[1].rstrip("/") + "/", sys.argv[2], sys.argv[3], sys.argv[4], Path(sys.argv[5]).read_text(), Path(sys.argv[6])
option = lambda name, default: sys.argv[sys.argv.index(name) + 1] if name in sys.argv else default
limit, workers, raw = int(option("--limit", "1000")), int(option("--workers", "6")), "--raw" in sys.argv


def opening(page):
    if raw:
        blocks, words = [], 0
        for block in re.split(r"\n\s*\n", page.strip()):
            blocks.append(block)
            words += len(block.split())
            if words >= 450:
                break
        return "\n\n".join(blocks) + "\n"
    return lib.excerpt(re.sub(r"\A#\s+[^\n]*\n+", "", lib.clean(page), count=1), 450)


def ask(prompt):
    answer = subprocess.run(["python3", W + "ask_model.py", kind, cache], input=prompt.encode(), capture_output=True).stdout.decode()
    match = re.search(r'"more[a-z_]*"\s*:\s*"(A|B|same)"', answer)
    return match.group(1) if match else None


def one(row):
    page, revised = opening(row["original"]), opening(row["revised"])
    first, second = ask(wording % (page, revised)), ask(wording % (revised, page))
    named = lambda letter, page_is_a: None if letter is None else "same" if letter == "same" else "page" if (letter == "A") == page_is_a else "revision"
    return {"id": row["id"], "page_first": named(first, True), "revision_first": named(second, False)}


rows = [r for r in map(json.loads, open(revisions)) if "revised" in r][:limit]
with concurrent.futures.ThreadPoolExecutor(workers) as pool:
    results = list(pool.map(one, rows))
out.write_text("".join(json.dumps(r) + "\n" for r in results))
count = collections.Counter((r["page_first"], r["revision_first"]) for r in results)
print(len(results), "pairs:", {" / ".join(map(str, key)): value for key, value in count.most_common()})
print("page in both orders:", count[("page", "page")], "| revision in both:", count[("revision", "revision")],
      "| second text both times:", count[("revision", "page")], "| first text both times:", count[("page", "revision")])
