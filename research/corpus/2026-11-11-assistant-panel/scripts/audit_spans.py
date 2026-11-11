#!/usr/bin/env python3
"""Third-assistant audit of phrase findings by blind paired comparison.

usage: audit_spans.py <workspace> <report json> <assignment json> <sample size> <out json> [rule,rule...]

For each sampled finding the auditor compares the original sentence context with
a version where the flagged phrase is removed, or replaced by a plain word for
inflated verbs. Placebo items remove a random ordinary word from an unflagged
sentence position. The order of the two versions is random.
"""
import json, os, random, re, sys, concurrent.futures, urllib.request
from pathlib import Path

W = Path(sys.argv[1])
REPORT, ASSIGN, SIZE, OUT = sys.argv[2], sys.argv[3], int(sys.argv[4]), sys.argv[5]
RULES = set(sys.argv[6].split(",")) if len(sys.argv) > 6 else None
PLAIN = {"leverage": "use", "leverages": "uses", "leveraging": "using", "leveraged": "used",
         "utilize": "use", "utilizes": "uses", "utilizing": "using", "facilitate": "help", "facilitates": "helps",
         "facilitating": "helping", "harness": "use", "harnesses": "uses", "harnessing": "using",
         "showcase": "show", "showcases": "shows", "showcasing": "showing", "enhance": "improve",
         "enhances": "improves", "enhancing": "improving", "streamline": "simplify", "streamlines": "simplifies",
         "empower": "let", "empowers": "lets", "in order to": "to"}
PROMPT = """Two versions of a short passage from technical documentation follow. They differ in one word or phrase.

Decide:
1. Which version is better for a reader of this documentation: "A", "B", or "same".
2. Whether the shorter or plainer version drops information. Information means technical facts, conditions, limits, steps, values, names, and warnings a reader could act on. Praise, emphasis, and stock phrases are not information. An ungrammatical result counts as worse, not as lost information.

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
        except Exception:  # noqa: BLE001
            continue
    return None


def window(raw, start, end):
    low = raw.rfind(b"\n\n", 0, start)
    low = 0 if low < 0 else low + 2
    high = raw.find(b"\n\n", end)
    high = len(raw) if high < 0 else high
    if high - low > 1500:
        low, high = max(low, start - 600), min(high, end + 600)
    return low, high


def tidy(text):
    text = re.sub(r"[ \t]{2,}", " ", text)
    text = re.sub(r" ([,.;:!?])", r"\1", text)
    text = re.sub(r"\b(a) ([aeiouAEIOU])", r"an \2", text)
    text = re.sub(r"\b(an) ([^aeiouAEIOU\W])", r"a \2", text)
    return text.strip()


random.seed(20261112)
assignment = json.loads(Path(ASSIGN).read_text())
report = json.loads(Path(REPORT).read_text())
jobs = []
flagged = {}
for finding in report["findings"]:
    flagged.setdefault(finding["primary"]["path"], []).append(finding["primary"]["span"])
candidates = [f for f in report["findings"] if (RULES is None or f["rule_id"] in RULES)]
random.shuffle(candidates)
for finding in candidates:
    if len(jobs) >= SIZE:
        break
    path = finding["primary"]["path"]
    raw = (W / "engine" / path).read_bytes()
    span = finding["primary"]["span"]
    phrase = raw[span["start"]:span["end"]].decode("utf-8", "replace")
    if len(phrase) > 60 or "\n" in phrase:
        continue
    low, high = window(raw, span["start"], span["end"])
    original = raw[low:high].decode("utf-8", "replace")
    replacement = PLAIN.get(phrase.lower(), "")
    if replacement and phrase[0].isupper():
        replacement = replacement.capitalize()
    before = raw[low:span["start"]].decode("utf-8", "replace")
    after = raw[span["end"]:high].decode("utf-8", "replace")
    if not replacement and (not before.strip() or before.rstrip()[-1:] in ".!?:") and phrase[:1].isupper():
        # The phrase opened a sentence: the next word takes the capital.
        stripped = after.lstrip(" ,")
        after = " " * (1 if before and not before.endswith((" ", "\n")) else 0) + stripped[:1].upper() + stripped[1:]
    edited = before + replacement + after
    jobs.append({"kind": "finding", "rule": finding["rule_id"], "phrase": phrase.lower(), "path": path,
                 "original": tidy(original), "edited": tidy(edited)})
placebo_target = max(1, len(jobs) // 3)
paths = list(flagged)
tries = 0
while sum(j["kind"] == "placebo" for j in jobs) < placebo_target and tries < 20000:
    tries += 1
    path = random.choice(paths)
    raw = (W / "engine" / path).read_bytes()
    words = [m for m in re.finditer(rb"(?<![\w`/.#-])[a-z]{5,12}(?![\w`/.-])", raw)]
    if not words:
        continue
    match = random.choice(words)
    if any(s["start"] <= match.start() < s["end"] for s in flagged[path]):
        continue
    low, high = window(raw, match.start(), match.end())
    chunk = raw[low:high]
    if b"```" in chunk or b"|" in chunk or len(chunk) < 80:
        continue
    original = chunk.decode("utf-8", "replace")
    edited = raw[low:match.start()].decode("utf-8", "replace") + raw[match.end():high].decode("utf-8", "replace")
    jobs.append({"kind": "placebo", "rule": "placebo", "phrase": match.group(0).decode(), "path": path,
                 "original": tidy(original), "edited": tidy(edited)})


def work(job):
    flip = random.random() < 0.5
    answer = ask(PROMPT % ((job["edited"], job["original"]) if flip else (job["original"], job["edited"])))
    job["edited_is"] = "A" if flip else "B"
    job["answer"] = answer
    if answer:
        prefers_original = answer["better"] not in ("same", job["edited_is"])
        job["confirmed"] = (not prefers_original) and not answer["information_lost"]
    return job


with concurrent.futures.ThreadPoolExecutor(8) as pool:
    results = list(pool.map(work, jobs))
json.dump(results, open(W / OUT, "w"), indent=1, ensure_ascii=False)
for kind in ("finding", "placebo"):
    done = [r for r in results if r["kind"] == kind and r["answer"]]
    if done:
        print(f"{kind}: answered={len(done)} confirmed={100 * sum(r['confirmed'] for r in done) / len(done):.1f}%")
by_rule = {}
for r in results:
    if r["kind"] == "finding" and r["answer"]:
        by_rule.setdefault(r["rule"], []).append(r["confirmed"])
print({k: f"{sum(v)}/{len(v)}" for k, v in by_rule.items()})
