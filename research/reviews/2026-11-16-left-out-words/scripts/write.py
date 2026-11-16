#!/usr/bin/env python3
"""Have one model write a second post from some outlines, under the other of the two style guides.

usage: write.py <workspace> <human jsonl> <outlines jsonl> <kind> <cache dir> <family> <guides dir> <out jsonl>
       --from N --to M [--workers N]

The first texts were written under style guide A or B by a hash of the id. These are written under the other one,
so they are new texts on the same subjects. The outlines are those of the test split in the order of a hash of the
id, from position N up to M. A row holds the id, the split, the guide, and the cleaned text; a text with too few
words is marked failed and asked again on the next run.
"""
import concurrent.futures, hashlib, json, subprocess, sys
from pathlib import Path

# The lines of running prose are told, and a written text is cleaned, as in the first record.
sys.path.insert(0, str(Path(__file__).resolve().parents[2] / "2026-11-15-revision-check" / "scripts"))
import prepare as lib

W, source, outlines, kind, cache, family, guides, out = sys.argv[1].rstrip("/") + "/", *sys.argv[2:8], Path(sys.argv[8])
option = lambda name, default: sys.argv[sys.argv.index(name) + 1] if name in sys.argv else default
workers, first, last = int(option("--workers", "4")), int(option("--from", "0")), int(option("--to", "15"))
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
short = {(r["id"], "guided") for r in earlier if r.get("failed")}
done = {(r["id"], "guided") for r in earlier if not r.get("failed")}
if out.exists():
    out.write_text("".join(json.dumps(r, ensure_ascii=False) + "\n" for r in earlier if not r.get("failed")))


def guide_of(post):
    return "A" if int(hashlib.sha256(("guide:" + post).encode()).hexdigest()[:8], 16) % 2 == 0 else "B"


def one(job):
    row, variant = job
    guide = "B" if guide_of(row["id"]) == "A" else "A"
    prompt = TASK % (row["title"], STYLE % GUIDE[guide] if guide else "", row["outline"])
    if (row["id"], variant) in short:
        prompt = prompt.replace(LONGER, FIRMER)
    answer = subprocess.run(["python3", W + "ask_model.py", kind, cache], input=prompt.encode(), capture_output=True).stdout.decode()
    text = lib.clean(answer)
    result = {"id": row["id"], "split": "test", "family": family, "guide": guide, "words": lib.words(text)}
    if lib.words(text) < 450:
        return {**result, "failed": True}
    return {**result, "text": text}


outline = {r["id"]: r["outline"] for r in map(json.loads, open(outlines)) if not r.get("failed")}
posts = [{**row, "outline": outline[row["id"]]} for row in map(json.loads, open(source)) if row["split"] == "test" and row["id"] in outline]
posts.sort(key=lambda row: hashlib.sha256(("second:" + row["id"]).encode()).hexdigest())
jobs = [(row, "guided") for row in posts[first:last] if (row["id"], "guided") not in done]
with out.open("a") as handle, concurrent.futures.ThreadPoolExecutor(workers) as pool:
    for row in pool.map(one, jobs):
        handle.write(json.dumps(row, ensure_ascii=False) + "\n")
        handle.flush()
print("done", len(jobs), "jobs")
