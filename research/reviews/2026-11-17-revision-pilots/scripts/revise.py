#!/usr/bin/env python3
"""Have a model revise each post of the fitting half, in one of two ways that were not tried before.

usage: revise.py <workspace> <unswell binary> <model kind> <cache dir> <out jsonl> <arm: denial|manner> [--workers N] [--limit N]
                      [--split dev|test]

denial   The model gets the page and a report: the lines of `unswell check` about the page index, and then the
         two findings of a page as a whole with what the text report does not print yet, the closing sentences
         themselves. It is asked to resolve what the report points at and to change nothing else.
manner   The model gets the page and the openings of two blog posts that people wrote, and is asked to say what
         the page says in the manner of those two: how their sentences vary and how they speak to the reader. Every
         fact, number, link, heading, list, table, and code block stays.

A row holds the page, the revision, what the model was shown, the index of both, and for both whether each of the
two rules reports and how many closing sentences the first lists.
"""
import concurrent.futures, hashlib, json, subprocess, sys, tempfile
from pathlib import Path

W, binary, kind, cache, out, arm = sys.argv[1].rstrip("/") + "/", sys.argv[2], sys.argv[3], sys.argv[4], Path(sys.argv[5]), sys.argv[6]
option = lambda name, default: sys.argv[sys.argv.index(name) + 1] if name in sys.argv else default
workers, limit, split = int(option("--workers", "2")), int(option("--limit", "100000")), option("--split", "dev")
G = Path(W) / "guided"
DENIAL, REGISTER = "syntax.closing-denial-density", "syntax.uniform-register"
KEEP = ("Keep every heading, list, table, link, and code block as it is, keep every fact and every number, add no fact, "
        "and add no em dash. Answer with the complete revised page and nothing else, with no code fence around it.\n\n")
INSTRUCTIONS = {
    "denial": "Below is a Markdown page and the report of a prose linter about it. Revise the page so that what the report points at "
              "is resolved. Change nothing else. " + KEEP,
    "manner": "Below is a Markdown page, and before it the openings of two blog posts that people wrote. Say what the page says in the "
              "manner of those two posts: how the length and the build of their sentences vary, how they speak to the reader, and how "
              "they move from one point to the next. Do not copy their words or their subjects. " + KEEP,
}


def check(text):
    """The index of a page, the lines of the report about it, the closing sentences that the first rule lists, and
    whether the second rule reports the page."""
    with tempfile.TemporaryDirectory() as scratch:
        done = subprocess.run([binary, "check", "--stdin", "--filename", "page.md", "--format", "markdown", "--no-gate", "--allow-empty",
                               "--timeout", "2m", "--report", "json:result.json", "--report", "text:-"], input=text.encode(), cwd=scratch,
                              capture_output=True)
        result = json.loads((Path(scratch) / "result.json").read_text())
    index = result["documents"][0].get("ai_sounding") or {}
    lines = done.stdout.decode().splitlines()
    start = next((i for i, l in enumerate(lines) if l.startswith("AI-sounding index:")), None)
    report = [l for l in lines[start:] if not l.startswith(("The index is the higher", "Revision probability"))] if start is not None else []
    source = text.encode()
    closers, register = [], False
    for finding in result.get("findings") or []:
        if finding["rule_id"] == DENIAL:
            for place in [finding["primary"]] + (finding.get("related") or []):
                closers.append(" ".join(source[place["span"]["start"]:place["span"]["end"]].decode().split()))
        if finding["rule_id"] == REGISTER:
            register = True
    return index, "\n".join(report), closers, register


def habits(closers, register):
    parts = []
    if closers:
        parts.append(f"{len(closers)} paragraphs end by saying what something is not or does not do. One such sentence is an honest "
                     "limit; this many are a habit. Keep at most two that a reader would miss. End the other paragraphs on what the "
                     "thing is or does, or move the limit beside the claim that it limits. The closing sentences:\n"
                     + "\n".join(f'  - "{sentence}"' for sentence in closers))
    if register:
        parts.append("The page never leaves one formal register: no contraction, no aside in parentheses, no question, and no sentence "
                     "that opens with And, But, or So. Say one or two things the way you would to a colleague.")
    return "\n".join(parts)


people = [r for r in map(json.loads, open(G / "human.jsonl")) if r["split"] == "dev"]


def exemplars(post_id):
    """Two posts of people from the fitting half, picked by a hash of the post, as openings of about 500 words."""
    picked = sorted(people, key=lambda r: hashlib.sha256(("manner:" + post_id + r["id"]).encode()).hexdigest())[:2]
    return "\n\n".join(f"===== POST OF A PERSON {number} =====\n" + " ".join(r["text"].split(" ")[:500]) for number, r in enumerate(picked, 1))


def one(row):
    page = "# " + (row.get("title") or "Post") + "\n\n" + row["text"]
    before, report, closers, register = check(page)
    if before.get("status") != "available":
        return {"set": "blog", "id": row["id"], "skipped": True}
    if arm == "denial":
        shown = "\n".join(part for part in (report, habits(closers, register)) if part)
        prompt = INSTRUCTIONS[arm] + "===== REPORT =====\n" + shown + "\n\n===== PAGE =====\n" + page
    else:
        shown = exemplars(row["id"])
        prompt = INSTRUCTIONS[arm] + shown + "\n\n===== PAGE =====\n" + page
    answer = subprocess.run(["python3", W + "ask_model.py", kind, cache], input=prompt.encode(), capture_output=True).stdout.decode()
    if len(answer.split()) < 0.5 * len(page.split()):
        return {"set": "blog", "id": row["id"], "failed": True}
    after, _, closers_after, register_after = check(answer)
    keep = lambda i: {"index": i.get("index"), "level": i.get("level"), "readings": i.get("readings")}
    return {"set": "blog", "id": row["id"], "original": page, "revised": answer, "shown": shown if arm == "denial" else "two posts of people",
            "before": {**keep(before), "closers": len(closers), "register": register},
            "after": {**keep(after), "closers": len(closers_after), "register": register_after}}


done = {r["id"] for r in map(json.loads, out.read_text().splitlines())} if out.exists() else set()
rows = [r for r in map(json.loads, open(G / "set.jsonl")) if r["split"] == split and r["id"] not in done]
rows.sort(key=lambda r: hashlib.sha256(("pick:" + r["id"]).encode()).hexdigest())
with out.open("a") as handle, concurrent.futures.ThreadPoolExecutor(workers) as pool:
    for row in pool.map(one, rows[:limit]):
        handle.write(json.dumps(row, ensure_ascii=False) + "\n")
        handle.flush()
print("done", len(rows[:limit]))
