#!/usr/bin/env python3
"""Write the data of the control for the record: pages that people wrote under a style guide, as the build reads them.

usage: control_export.py <binary> <out jsonl.gz> <set S pages jsonl> <people pages jsonl>...

Each page of the measured half is checked as one Markdown page. A row of the output holds a page of people: where
it is from, how much of its prose stands in paragraphs, what each signal counted, the two readings, and its
sentences in running paragraphs as [stretch, words, plain]. The output holds no text.

The sentences are rebuilt from the build's result: its blocks, its sentences, and the spans that it protects. A
block is a paragraph unless its line starts as a heading, a list item, a table, or a quote, or it is indented under
a list item, or it is the one line that the preparation puts in the place of a code block. A stretch is a row of paragraphs with nothing but blank lines between them. A sentence is plain by the
rule of the build: five to twenty words and none of ? ! ( ) ; : and the em dash among what a reader sees of it.

The posts of set S are not written to the output. Their figures are printed, to be kept beside the record.
"""
import concurrent.futures, gzip, html, json, re, statistics, subprocess, sys, tempfile
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))
import prepare as lib

ARTICLES = {"GoogleChrome/web.dev", "GoogleChrome/developer.chrome.com", "GoogleCloudPlatform/community", "istio/istio.io"}
MARKER = re.compile(r"^\s*(?:[-*+]|\d+[.)])\s+")
BREAKS = "?!();:—"
MATCH = 0.732
SIGNALS = ("em-dash", "emoji-in-heading", "emoji-at-list-item", "arrow", "stock-word", "bold-label-in-list", "plain-sentence-run",
           "subject-opener-run")


def paragraph_share(text):
    inside = total = 0
    for line in text.split("\n"):
        count = lib.words(line)
        total += count
        if count and not lib.STARTS_LINE.match(line):
            inside += count
    return inside / total if total else 0.0


def kind(raw, start, state):
    """Name the kind of the block that starts here. An indented block belongs to the list item above it."""
    line_start = raw.rfind(b"\n", 0, start) + 1
    line_end = raw.find(b"\n", start)
    line = raw[line_start:line_end if line_end >= 0 else len(raw)].decode("utf-8", "replace")
    indent = len(line) - len(line.lstrip(" \t"))
    if MARKER.match(line):
        state["list"] = True
        return "list"
    if state.get("list") and indent >= 2:
        return "list"
    state["list"] = False
    if line.strip() == "[code]":
        return "code"
    if re.match(r"^\s*#{1,6}\s", line):
        return "heading"
    if line.lstrip().startswith("|"):
        return "table"
    if line.lstrip().startswith(">"):
        return "quote"
    return "paragraph"


def visible(raw, start, end, excluded):
    """The words of a sentence as a reader sees them: without what the build protects, and without link and entity syntax."""
    kept, at = [], start
    for a, b in excluded:
        if b <= start or a >= end:
            continue
        kept.append(raw[at:max(at, a)])
        at = max(at, b)
    kept.append(raw[at:end])
    piece = b" ".join(kept).decode("utf-8", "replace")
    piece = re.sub(r"!?\[([^\]]*)\]\([^)]*\)", r"\1", piece, flags=re.S)
    piece = re.sub(r"\]\([^)]*\)", "", piece, flags=re.S)
    return html.unescape(piece)


def rebuild(text, result):
    raw = text.encode()
    excluded = sorted((e["span"]["start"], e["span"]["end"]) for e in result["documents"][0].get("excluded", []))
    units = result["assessments"]
    blocks = sorted((a for a in units if a["scope"] == "paragraph"), key=lambda a: a["span"]["start"])
    sentences = sorted((a for a in units if a["scope"] == "sentence"), key=lambda a: a["span"]["start"])
    out, stretch, end, state = [], 0, -1, {}
    for block in blocks:
        start, stop = block["span"]["start"], block["span"]["end"]
        if kind(raw, start, state) != "paragraph" or block["words"] == 0:
            end = -1
            continue
        if end < 0 or start < end or raw[end:start].strip():
            stretch += 1
        end = stop
        for s in sentences:
            if start <= s["span"]["start"] < stop:
                piece = visible(raw, s["span"]["start"], s["span"]["end"], excluded)
                out.append([stretch, s["words"], int(5 <= s["words"] <= 20 and not any(m in piece for m in BREAKS))])
    return out


def in_runs(sentences):
    """Count the plain sentences that stand in runs of two or more inside one stretch."""
    total, run, last = 0, 0, None
    for stretch, _, plain in sentences + [[None, 0, 0]]:
        if plain and stretch == last:
            run += 1
        else:
            total += run if run >= 2 else 0
            run = 1 if plain else 0
        last = stretch
    return total


def by_chance(sentences):
    """The expected count of plain sentences in runs when the plain ones are placed at random among the sentences."""
    n, k = len(sentences), sum(s[2] for s in sentences)
    if n < 2 or k < 2:
        return 0.0
    one = (k - 1) / (n - 1)
    two = 1 - ((n - k) / (n - 1)) * ((n - k - 1) / (n - 2)) if n > 2 else one
    expected = 0.0
    for i, (stretch, _, _) in enumerate(sentences):
        neighbors = (i > 0 and sentences[i - 1][0] == stretch) + (i + 1 < n and sentences[i + 1][0] == stretch)
        expected += (k / n) * (one if neighbors == 1 else two if neighbors == 2 else 0.0)
    return expected


def check(job):
    binary, row = job
    with tempfile.TemporaryDirectory() as scratch:
        done = subprocess.run([binary, "check", "--stdin", "--filename", "page.md", "--format", "markdown", "--no-gate", "--allow-empty",
                               "--timeout", "2m", "--report", "json:-"], input=row["text"].encode(), cwd=scratch, capture_output=True)
    result = json.loads(done.stdout)
    index = result["documents"][0]["ai_sounding"]
    signals = {s["name"]: [s["count"], s.get("sentences", 0)] for s in index["signals"]}
    return {"row": row, "words": index["prose_words"], "signals": {name: signals[name] for name in SIGNALS}, "index": index["index"],
            "marks": index["readings"]["marks"], "rhythm": index["readings"]["rhythm"], "sentences": rebuild(row["text"], result)}


def figures(page):
    s = page["sentences"]
    n = len(s)
    return {"n": n, "words": statistics.mean(x[1] for x in s), "long": sum(x[1] > 20 for x in s) / n, "plain": sum(x[2] for x in s) / n,
            "runs": in_runs(s), "chance": by_chance(s), "per_thousand": 1000 * n / page["words"]}


def med(values):
    return statistics.median(values)


def auc(positive, negative):
    return sum((p > q) + 0.5 * (p == q) for p in positive for q in negative) / (len(positive) * len(negative))


def main():
    binary, out = sys.argv[1], Path(sys.argv[2])
    rows, seen = [], set()
    for path in sys.argv[3:]:
        for row in map(json.loads, open(path)):
            if row["split"] == "test" and row["id"] not in seen:
                seen.add(row["id"])
                rows.append(row)
    with concurrent.futures.ThreadPoolExecutor(6) as pool:
        pages = list(pool.map(check, [(binary, row) for row in rows]))
    people = [p for p in pages if "repo" in p["row"]]
    set_s = [p for p in pages if "repo" not in p["row"]]
    lines = []
    for p in sorted(people, key=lambda p: (p["row"]["repo"] not in ARTICLES, p["row"]["repo"], p["row"]["id"])):
        row = p["row"]
        share = paragraph_share(row["text"])
        assert abs(share - MATCH) > 0.0001, row["id"]
        lines.append(json.dumps({"id": row["id"], "kind": "articles" if row["repo"] in ARTICLES else "documentation", "repo": row["repo"],
                                 "path": row["path"], "commit": row["commit"], "guide": row["guide"], "paragraphs": round(share, 4),
                                 "words": p["words"], "signals": p["signals"], "index": p["index"], "marks": p["marks"], "rhythm": p["rhythm"],
                                 "sentences": p["sentences"]}, sort_keys=True, separators=(",", ":")))
    with open(out, "wb") as raw, gzip.GzipFile(fileobj=raw, mode="wb", mtime=0, filename="") as handle:
        handle.write(("\n".join(lines) + "\n").encode())
    print(len(lines), "pages of people written;", out.stat().st_size, "bytes")
    equal = sum(in_runs(p["sentences"]) == p["signals"]["plain-sentence-run"][1] for p in pages)
    print("rebuilt runs equal the build's on", equal, "of", len(pages), "pages")
    matched = [p for p in people if paragraph_share(p["row"]["text"]) >= MATCH]
    f_s = [figures(p) for p in set_s if len(p["sentences"]) >= 10]
    f_m = [figures(p) for p in matched if len(p["sentences"]) >= 10]
    plain = sorted(f["plain"] for f in f_s)
    print("set S, measured half, not in the output:", json.dumps({
        "posts": len(f_s), "words_per_sentence": round(med([f["words"] for f in f_s]), 1), "over_20_words": round(med([f["long"] for f in f_s]), 3),
        "plain": round(med(plain), 3), "plain_lower_quartile": round(plain[len(plain) // 4], 3),
        "in_runs": round(med([f["runs"] / f["n"] for f in f_s]), 3), "by_chance": round(med([f["chance"] / f["n"] for f in f_s]), 3),
        "runs_to_chance": round(sum(f["runs"] for f in f_s) / sum(f["chance"] for f in f_s), 2),
        "paragraph_sentences_per_thousand_words": round(med([f["per_thousand"] for f in f_s]), 1),
        "auc_words_per_sentence": round(auc([-f["words"] for f in f_s], [-f["words"] for f in f_m]), 3),
        "auc_plain": round(auc([f["plain"] for f in f_s], [f["plain"] for f in f_m]), 3)}))


main()
