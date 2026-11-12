#!/usr/bin/env python3
"""Trim what the review flagged, trim random other sentences as a control, and keep only pure trims.

usage: effect_trim.py <corpus workspace> <unswell binary> <reviewer> [--pages N] [--workers N]

The reviewer's bound flags are read from product/<reviewer>-none. The same
model rewrites: the reviewed arm gets each flagged sentence with its defect and
reason, the control arm gets as many random unflagged sentences and is told
only to improve the wording. A revision replaces the bytes of its sentence and
nothing else. Each revised sentence is checked for changed numbers, inline
code, and links.
"""
import argparse, collections, concurrent.futures, hashlib, json, random, re, subprocess, sys, tempfile
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))
from common import overlaps, write_jsonl

STATEMENTS = {
    "empty_framing": "The opening only announces the statement and adds no information.",
    "needless_repetition": "The same claim is stated again with no new distinction.",
    "wordiness": "A longer phrase where a shorter one states the same thing.",
    "unjustified_intensifiers": "Praise or emphasis that adds no usable information.",
    "vague_claims": "A claim with no scope or measurement.",
    "formulaic_transitions": "A transition that expresses no relationship.",
    "needless_complexity": "Nesting that conveys no additional condition.",
}
RULES = """You may only delete words and adjust capitalization and punctuation. Do not add, replace, or reorder words.
Keep every negation, number, condition, name, identifier, link, and piece of code. Keep the sentence itself: if deleting
words cannot improve it, return it unchanged. Change nothing outside the listed sentences.

Answer with JSON only: {"revisions": [{"id": "<id>", "text": "<the trimmed sentence>"}]} with one entry per listed sentence.
"""
REVIEWED = """You trim sentences of one technical document. Below is the whole document for context, then a list of
sentences with the defect a reviewer found in each.

Remove words from each listed sentence so that its defect is gone or smaller.
""" + RULES
CONTROL = """You trim sentences of one technical document. Below is the whole document for context, then a list of
sentences.

Remove words that carry no information from each listed sentence.
""" + RULES
PROTECTED = {"not", "no", "never", "none", "nor", "neither", "without", "cannot", "unless", "except", "only",
             "may", "might", "can", "could", "should", "must", "would", "shall",
             "typically", "usually", "often", "sometimes", "generally", "potentially", "possibly", "probably", "likely",
             "approximately", "roughly", "almost", "nearly"}
STRIP = ".,;:!?()[]\"'“”‘’*_"


def normal(text):
    """Words of a sentence, lower-cased and without adjacent punctuation; empty pieces dropped."""
    return [word for word in (piece.strip(STRIP).lower() for piece in text.split()) if word]


def trim_problem(original, revised):
    """Return why a revision is not a pure trim of the original, or an empty string."""
    before, after = normal(original), normal(revised)
    if after == before:
        return "unchanged"
    position, removed = 0, []
    for word in before:
        if position < len(after) and after[position] == word:
            position += 1
        else:
            removed.append(word)
    if position != len(after):
        return "adds, replaces, or reorders words"
    if 2 * len(after) < len(before):
        return "removes more than half of the words"
    for word in removed:
        if word in PROTECTED or word.endswith("n't") or word.endswith("n’t") or any(ch.isdigit() for ch in word):
            return "removes a negation, a number, or a modal word"
        if "`" in word or "http" in word or "](" in word:
            return "removes code or a link"
    return ""


parser = argparse.ArgumentParser()
parser.add_argument("workspace")
parser.add_argument("binary")
parser.add_argument("reviewer", choices=["claude", "astra"])
parser.add_argument("--pages", type=int, default=60)
parser.add_argument("--workers", type=int, default=4)
options = parser.parse_args()
W = Path(options.workspace).resolve()
HUMAN = W / "human"
ASK, CACHE = str(W / "ask_model.py"), str(HUMAN / "cache")
out = HUMAN / "effect-trim" / options.reviewer
out.mkdir(parents=True, exist_ok=True)
documents = [json.loads(line) for line in open(W / "docs.jsonl")]
documents = [d for d in documents if d["origin"] == "generated"]
documents = sorted(documents, key=lambda d: hashlib.sha256(d["doc_id"].encode()).hexdigest())[:options.pages]


def units(doc):
    """Return the prose sentences of the document as the engine sees them: (start, end, words)."""
    with tempfile.TemporaryDirectory() as scratch:
        done = subprocess.run([options.binary, "check", "--stdin", "--filename", doc["doc_id"] + ".md", "--format", "markdown",
                               "--no-gate", "--allow-empty", "--prepared-feature", "prose-words", "--prepared-kind", "sentence",
                               "--report", "json:-"], input=doc["text"].encode(), cwd=scratch, capture_output=True)
    if done.returncode != 0:
        raise ValueError(done.stderr.decode("utf-8", "replace")[-300:])
    sources = (json.loads(done.stdout).get("prepared_features") or {}).get("sources") or []
    result = []
    for unit in sources[0]["units"] if sources else []:
        if unit["binding"]["kind"] != "sentence":
            continue
        segments = unit["binding"]["segments"]
        words = next((v.get("number") for v in unit.get("values", []) if v.get("id") == "prose-words"), 0) or 0
        result.append((segments[0]["start"], segments[-1]["end"], words))
    return sorted(set(result))


def ask(prompt):
    last = None
    for _ in range(3):
        with tempfile.TemporaryDirectory() as scratch:
            done = subprocess.run(["python3", ASK, options.reviewer, CACHE], input=prompt.encode(), cwd=scratch, capture_output=True)
        if done.returncode != 0:
            last = done.stderr.decode("utf-8", "replace")[-200:]
            continue
        answer = done.stdout.decode()
        try:
            return json.loads(answer[answer.index("{"):answer.rindex("}") + 1])["revisions"]
        except (ValueError, KeyError) as error:
            last = str(error)
    raise ValueError(last)


def protected(text):
    """Numbers, inline code, and link targets of a sentence, as a sorted list."""
    found = re.findall(r"`[^`]*`", text) + re.findall(r"\]\([^)]*\)", text) + re.findall(r"https?://[^\s)>\]]+", text)
    plain = re.sub(r"`[^`]*`|\]\([^)]*\)|https?://[^\s)>\]]+", " ", text)
    return sorted(found + re.findall(r"\d+(?:[.,]\d+)*", plain))


def revise(raw, listed, instructions):
    """Ask for trimmed versions of the listed sentences and apply the ones that pass the check."""
    lines = [instructions, "\n===== DOCUMENT =====\n", raw.decode(), "\n===== SENTENCES =====\n"]
    for entry in listed:
        lines.append(f"\n[{entry['id']}] {entry['text']}\n")
        if entry.get("defect"):
            lines.append(f"defect: {entry['defect']}\nreviewer's reason: {entry['reason']}\n")
    answers = {str(r.get("id", "")).strip("[] "): str(r.get("text", "")) for r in ask("".join(lines))}
    revised, stats = raw, collections.Counter(listed=len(listed))
    for entry in sorted(listed, key=lambda e: -e["span"][0]):
        if entry["id"] not in answers:
            stats["not returned"] += 1
            continue
        text = " ".join(answers[entry["id"]].split()) if "\n" not in entry["text"] else answers[entry["id"]].strip()
        problem = trim_problem(entry["text"], text)
        if problem:
            stats[problem] += 1
            continue
        stats["applied"] += 1
        start, end = entry["span"]
        revised = revised[:start] + text.encode() + revised[end:]
    return revised, dict(stats)


def one(doc):
    target = out / f"{doc['doc_id']}.json"
    if target.exists():
        return json.loads(target.read_text())
    raw = doc["text"].encode()
    report = json.loads((W / "product" / f"{options.reviewer}-none" / f"{doc['doc_id']}.json").read_text())["documents"][0]
    sentences = units(doc)
    flagged, taken = [], set()
    for finding in report["findings"]:
        span = (finding["span"]["start"], finding["span"]["end"])
        unit = next(((s, e) for s, e, _ in sentences if s <= span[0] and span[1] <= e), None)
        if unit is None or unit in taken:
            continue
        taken.add(unit)
        defect = " Or: ".join(STATEMENTS[c] for c in finding["categories"] if c in STATEMENTS)
        flagged.append({"id": f"s{len(flagged) + 1:03d}", "span": list(unit), "text": raw[unit[0]:unit[1]].decode(),
                        "defect": defect, "reason": finding["reason"]})
    spans = [(f["span"]["start"], f["span"]["end"]) for f in report["findings"]]
    free = [(s, e) for s, e, words in sentences if words >= 6 and not overlaps((s, e), spans)]
    random.Random(hashlib.sha256(doc["doc_id"].encode()).hexdigest()).shuffle(free)
    control = [{"id": f"s{i + 1:03d}", "span": list(unit), "text": raw[unit[0]:unit[1]].decode()}
               for i, unit in enumerate(sorted(free[:len(flagged)]))]
    result = {"id": doc["doc_id"], "findings": len(report["findings"]), "flagged_sentences": len(flagged),
              "control_sentences": len(control), "raw": raw.decode()}
    if flagged:
        reviewed, result["reviewed_stats"] = revise(raw, flagged, REVIEWED)
        result["reviewed"] = reviewed.decode()
    if control:
        controlled, result["control_stats"] = revise(raw, control, CONTROL)
        result["control"] = controlled.decode()
    target.write_text(json.dumps(result, ensure_ascii=False))
    return result


results, failed = [], 0
with concurrent.futures.ThreadPoolExecutor(options.workers) as pool:
    futures = {pool.submit(one, doc): doc for doc in documents}
    for future in concurrent.futures.as_completed(futures):
        try:
            results.append(future.result())
        except Exception as error:  # noqa: BLE001
            failed += 1
            print(futures[future]["doc_id"], "FAILED", str(error)[:200], flush=True)
results.sort(key=lambda r: r["id"])
for first, second in (("raw", "reviewed"), ("raw", "control"), ("reviewed", "control")):
    rows = [{"id": r["id"], "texts": {first: r[first], second: r[second]}} for r in results
            if first in r and second in r and r[first] != r[second]]
    write_jsonl(HUMAN / "pairs" / f"trim-{options.reviewer}-{first}-{second}.jsonl", rows)
    print(f"pairs {first}/{second}: {len(rows)}")
totals = {}
for arm in ("reviewed_stats", "control_stats"):
    counter = collections.Counter()
    for r in results:
        counter.update(r.get(arm, {}))
    totals[arm] = dict(counter)
print(f"{options.reviewer}: documents {len(results)}, failed {failed}, totals {totals}")
