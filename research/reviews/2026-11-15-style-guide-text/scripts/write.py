#!/usr/bin/env python3
"""Have one model write a post from each outline, once plainly and once under a style guide.

usage: write.py <workspace> <human jsonl> <outlines jsonl> <kind> <cache dir> <family> <guides dir> <out jsonl>
       [--workers N] [--splits test,dev] [--dev-limit N]

The writer sees the title and the outline, never the human post. The style guide of a post is A or B by a hash of
its id. A row holds the cleaned text; a text with too few words is marked failed and asked again on the next run.
"""
import concurrent.futures, hashlib, json, subprocess, sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))
import prepare as lib

W, source, outlines, kind, cache, family, guides, out = sys.argv[1].rstrip("/") + "/", *sys.argv[2:8], Path(sys.argv[8])
option = lambda name, default: sys.argv[sys.argv.index(name) + 1] if name in sys.argv else default
workers, splits, dev_limit = int(option("--workers", "4")), option("--splits", "test,dev").split(","), int(option("--dev-limit", "100000"))
GUIDE = {name: (Path(guides) / f"style-{name.lower()}.md").read_text() for name in ("A", "B")}
TASK = """Write a post for the engineering blog of the project that the outline below describes.

Title: %s

Use the facts in the outline. Do not add facts that are not in it. Write Markdown: a few section headings, paragraphs, and lists where they fit. Where the outline names a command or code, put a line "[code]" in place of the code block. Write about 700 words. Output the post only, without the title.
%s
===== OUTLINE =====
%s
"""
STYLE = """
Follow this style guide strictly. It comes from another project: apply its rules about wording, sentences, and structure, and ignore what is specific to that project.

===== STYLE GUIDE =====
%s
"""
# A text that came out too short is asked again with a firmer word count; its failed row is dropped.
LONGER = "Write about 700 words."
FIRMER = "Write 800 to 900 words. The post must not be shorter than 750 words."
earlier = [json.loads(line) for line in out.read_text().splitlines()] if out.exists() else []
short = {(r["id"], r["variant"]) for r in earlier if r.get("failed")}
done = {(r["id"], r["variant"]) for r in earlier if not r.get("failed")}
if out.exists():
    out.write_text("".join(json.dumps(r, ensure_ascii=False) + "\n" for r in earlier if not r.get("failed")))


def guide_of(post):
    return "A" if int(hashlib.sha256(("guide:" + post).encode()).hexdigest()[:8], 16) % 2 == 0 else "B"


def one(job):
    row, variant = job
    guide = guide_of(row["id"]) if variant == "guided" else None
    prompt = TASK % (row["title"], STYLE % GUIDE[guide] if guide else "", row["outline"])
    if (row["id"], variant) in short:
        prompt = prompt.replace(LONGER, FIRMER)
    answer = subprocess.run(["python3", W + "ask_model.py", kind, cache], input=prompt.encode(), capture_output=True).stdout.decode()
    text = lib.clean(answer)
    result = {"id": row["id"], "family": family, "variant": variant, "guide": guide, "words": lib.words(text)}
    if lib.words(text) < 450:
        return {**result, "failed": True}
    return {**result, "text": text}


outline = {r["id"]: r["outline"] for r in map(json.loads, open(outlines)) if not r.get("failed")}
posts, dev = [], 0
for row in map(json.loads, open(source)):
    if row["split"] not in splits or row["id"] not in outline:
        continue
    if row["split"] == "dev":
        dev += 1
        if dev > dev_limit:
            continue
    posts.append({**row, "outline": outline[row["id"]]})
jobs = [(row, variant) for row in posts for variant in ("plain", "guided") if (row["id"], variant) not in done]
with out.open("a") as handle, concurrent.futures.ThreadPoolExecutor(workers) as pool:
    for row in pool.map(one, jobs):
        handle.write(json.dumps(row, ensure_ascii=False) + "\n")
        handle.flush()
print("done", len(jobs), "jobs")
