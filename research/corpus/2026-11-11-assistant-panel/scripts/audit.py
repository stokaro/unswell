#!/usr/bin/env python3
"""Third-assistant audit by blind paired comparison with placebo items.

usage: audit.py <workspace> <docs file> <per stratum> <out json>

For every sampled item the auditor sees the original passage and the passage
with the item removed, in random order, without labels. It says which version
serves the reader better and whether the shorter one drops information. An item
is confirmed when the auditor does not prefer the original and reports no lost
information. Placebo items are random items that no annotator flagged.
"""
import concurrent.futures, json, os, random, re, sys, urllib.request
from pathlib import Path

W, DOCS, PER, OUT = Path(sys.argv[1]), sys.argv[2], int(sys.argv[3]), sys.argv[4]
PROMPT = """Two versions of a passage from technical documentation follow. They differ in one place: one version has one extra sentence or fragment.

Decide:
1. Which version serves a reader of this documentation better: "A", "B", or "same".
2. Whether the shorter version drops information that the longer one states. Information means technical facts, conditions, limits, steps, values, names, warnings, and references a reader could act on. Praise, emphasis, announcements of what follows, and restatements of what the passage already says are not information.

Answer with JSON only: {"better": "A" | "B" | "same", "information_lost": true | false, "reason": "<one sentence>"}

=== VERSION A ===
%s

=== VERSION B ===
%s
"""


def ask(prompt):
    body = {"model": os.environ["ROUTER_MODEL"], "temperature": 0, "max_tokens": 2500, "stream": False,
            "messages": [{"role": "user", "content": prompt}]}
    request = urllib.request.Request(os.environ["ROUTER_URL"], data=json.dumps(body).encode(),
                                     headers={"Content-Type": "application/json",
                                              "Authorization": "Bearer " + os.environ["ROUTER_KEY"]})
    for _ in range(4):
        try:
            with urllib.request.urlopen(request, timeout=180) as response:
                text = json.load(response)["choices"][0]["message"]["content"] or ""
            answer = json.loads(text[text.index("{"):text.rindex("}") + 1])
            if answer.get("better") in ("A", "B", "same") and isinstance(answer.get("information_lost"), bool):
                return answer
        except Exception:  # noqa: BLE001 - retried, then recorded as unanswered
            continue
    return None


def passages(doc, index):
    raw = doc["text"].encode()
    items = doc["items"]
    item = items[index]
    low = items[max(0, index - 3)]["start"]
    high = items[min(len(items) - 1, index + 3)]["end"]
    low = raw.rfind(b"\n", 0, low) + 1
    end = raw.find(b"\n", high)
    high = len(raw) if end < 0 else end
    original = raw[low:high]
    edited = bytearray()
    cursor = low
    for start, stop in item["segments"]:
        edited += raw[cursor:start]
        cursor = stop
    edited += raw[cursor:high]
    tidy = lambda b: re.sub(r"[ \t]+\n", "\n", re.sub(r"  +", " ", b.decode("utf-8", "replace"))).strip()
    return tidy(original), tidy(bytes(edited))


random.seed(20261111)
strata = {"consensus": [], "claude_only": [], "astra_only": [], "placebo": []}
docs = {}
for line in open(W / DOCS):
    doc = json.loads(line)
    paths = [W / "ann" / name / f"{doc['doc_id']}.json" for name in ("astra", "claude")]
    if not all(p.exists() for p in paths):
        continue
    docs[doc["doc_id"]] = doc
    astra, claude = [{f["item"]: f for f in json.loads(p.read_text())["flags"]} for p in paths]
    for index, item in enumerate(doc["items"]):
        a, c = astra.get(item["id"]), claude.get(item["id"])
        pa = bool(a and a["label"] == "needs_revision")
        pc = bool(c and c["label"] == "needs_revision")
        categories = sorted(set((a or {}).get("categories", []) + (c or {}).get("categories", [])))
        entry = (doc["doc_id"], index, categories)
        if pa and pc:
            strata["consensus"].append(entry)
        elif pc and not a:
            strata["claude_only"].append(entry)
        elif pa and not c:
            strata["astra_only"].append(entry)
        elif not a and not c and item["kind"] == "sentence":
            strata["placebo"].append(entry)
jobs = []
for name, entries in strata.items():
    for doc_id, index, categories in random.sample(entries, min(PER, len(entries))):
        original, edited = passages(docs[doc_id], index)
        if original == edited or len(original) > 6000:
            continue
        flip = random.random() < 0.5
        prompt = PROMPT % ((edited, original) if flip else (original, edited))
        jobs.append({"stratum": name, "doc": doc_id, "item": docs[doc_id]["items"][index]["id"],
                     "categories": categories, "edited_is": "A" if flip else "B", "prompt": prompt,
                     "text": " ".join(docs[doc_id]["items"][index]["text"].split())})


def work(job):
    answer = ask(job["prompt"])
    job = {k: v for k, v in job.items() if k != "prompt"}
    job["answer"] = answer
    if answer:
        prefers_original = answer["better"] not in ("same", job["edited_is"])
        job["confirmed"] = (not prefers_original) and not answer["information_lost"]
        job["strict"] = answer["better"] == job["edited_is"] and not answer["information_lost"]
    return job


with concurrent.futures.ThreadPoolExecutor(8) as pool:
    results = list(pool.map(work, jobs))
json.dump(results, open(W / OUT, "w"), indent=1, ensure_ascii=False)
print(f"{'stratum':12s} asked answered confirmed strict")
for name in strata:
    rows = [r for r in results if r["stratum"] == name]
    done = [r for r in rows if r["answer"]]
    if done:
        print(f"{name:12s} {len(rows):5d} {len(done):8d} {100 * sum(r['confirmed'] for r in done) / len(done):8.1f}% "
              f"{100 * sum(r['strict'] for r in done) / len(done):5.1f}%")
by_category = {}
for r in results:
    if r["stratum"] == "consensus" and r["answer"]:
        for category in r["categories"]:
            by_category.setdefault(category, []).append(r["confirmed"])
print("consensus confirmed by category:", {k: f"{sum(v)}/{len(v)}" for k, v in sorted(by_category.items())})
