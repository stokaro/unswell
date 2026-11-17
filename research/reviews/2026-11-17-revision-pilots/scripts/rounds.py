#!/usr/bin/env python3
"""Revise a post in rounds by what a blind judge says about it, until the judge no longer names it.

usage: rounds.py <workspace> <judge kind> <judge cache> <reviser kind> <reviser cache> <out jsonl> [--limit N] [--rounds N]
                      [--workers N]

A pilot. In each round the judge compares the opening of the post with the openings of two posts that people wrote,
as text A and text B in an order fixed by the pair, with no hint of which is which. When the judge names the post in
either pair, the reviser gets the page and what the judge said about how it reads, and revises the page. The rounds
end when the judge names the post in neither pair, or after the last round. The third human post of each post is
not used here, so that a later comparison has a pair that these rounds have not seen.

A row holds the post, its text after the rounds, and for each round what the judge named in each pair.
"""
import concurrent.futures, hashlib, json, re, subprocess, sys
from pathlib import Path

W, judge, judge_cache, reviser, reviser_cache, out = (sys.argv[1].rstrip("/") + "/", sys.argv[2], sys.argv[3], sys.argv[4], sys.argv[5],
                                                      Path(sys.argv[6]))
option = lambda name, default: sys.argv[sys.argv.index(name) + 1] if name in sys.argv else default
limit, rounds, workers = int(option("--limit", "6")), int(option("--rounds", "4")), int(option("--workers", "3"))
sys.path.insert(0, str(Path(__file__).parent))
import guided_lib as lib

G = Path(W) / "guided"
JUDGE = """Below are the openings of two technical blog posts, A and B. Judge only how the text reads, not the subject and not which post has more or better content.

Which post reads MORE like text written by a language model?

Answer with JSON only: {"more_ai_sounding": "A" | "B" | "same", "reason": "<one sentence>"}

===== POST A =====
%s

===== POST B =====
%s
"""
REVISE = ("Below is a Markdown page. A reader compared its opening with posts that people wrote, without knowing which is which, and "
          "took the page for text written by a language model. The reader's remarks follow; in each remark the page is the text with "
          "the letter that is given. Revise the whole page so that these remarks no longer hold for it. Keep every heading, list, "
          "table, link, and code block as it is, keep every fact and every number, add no fact, and add no em dash. Answer with the "
          "complete revised page and nothing else, with no code fence around it.\n\n")
opening = lambda page: lib.excerpt(re.sub(r"\A#\s+[^\n]*\n+", "", lib.clean(page), count=1), 450)


def ask(kind, cache, prompt):
    return subprocess.run(["python3", W + "ask_model.py", kind, cache], input=prompt.encode(), capture_output=True).stdout.decode()


def judged(pair_id, text, human, round_number):
    """What the judge names for one pair, and its remark with the letter of the page."""
    swap = int(hashlib.sha256(f"{pair_id}:{round_number}".encode()).hexdigest(), 16) % 2 == 1
    a, b = (human, text) if swap else (text, human)
    answer = ask(judge, judge_cache, JUDGE % (a, b))
    match = re.search(r'"more_ai_sounding"\s*:\s*"(A|B|same)"', answer)
    if not match:
        return "none", ""
    reason = re.search(r'"reason"\s*:\s*"((?:[^"\\]|\\.)*)"', answer)
    choice = match.group(1)
    picked = "same" if choice == "same" else ("post" if (choice == "A") != swap else "human")
    return picked, f"(the page is text {'B' if swap else 'A'}) " + (reason.group(1) if reason else "")


def one(job):
    post, pairs = job
    page = "# " + (post.get("title") or "Post") + "\n\n" + post["text"]
    history = []
    for number in range(rounds + 1):
        results = [judged(pair["id"], opening(page), pair["second"], number) for pair in pairs]
        history.append({"round": number, "named": [picked for picked, _ in results]})
        if all(picked != "post" for picked, _ in results) or number == rounds:
            break
        remarks = "\n".join(f"- {remark}" for picked, remark in results if picked == "post")
        answer = ask(reviser, reviser_cache, REVISE + "===== REMARKS =====\n" + remarks + "\n\n===== PAGE =====\n" + page)
        if len(answer.split()) < 0.5 * len(page.split()):
            history.append({"round": number, "failed": True})
            break
        page = answer
    return {"set": "blog", "id": post["id"], "original": "# " + (post.get("title") or "Post") + "\n\n" + post["text"], "revised": page,
            "rounds": history}


by_unit = {}
for pair in map(json.loads, open(G / "pairs-dev-set.jsonl")):
    if pair["comparison"] == "set":
        by_unit.setdefault(pair["unit"], []).append(pair)
done = {r["id"] for r in map(json.loads, out.read_text().splitlines())} if out.exists() else set()
posts = [r for r in map(json.loads, open(G / "set.jsonl")) if r["split"] == "dev" and r["id"] in by_unit and r["id"] not in done]
posts.sort(key=lambda r: hashlib.sha256(("loop:" + r["id"]).encode()).hexdigest())
jobs = [(post, sorted(by_unit[post["id"]], key=lambda p: p["id"])[:2]) for post in posts[:limit]]
with out.open("a") as handle, concurrent.futures.ThreadPoolExecutor(workers) as pool:
    for row in pool.map(one, jobs):
        handle.write(json.dumps(row, ensure_ascii=False) + "\n")
        handle.flush()
print("done", len(jobs))
