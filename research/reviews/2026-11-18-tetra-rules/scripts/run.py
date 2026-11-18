#!/usr/bin/env python3
"""Run the rules of `unswell check` on the original sections of TETRA and keep, for each sentence, what was found.

usage: run.py <import output json> <unswell binary> <rules catalog json> <record directory> [--workers N]

The import output is what `research/reviews/2026-09-30-tetra-import/replay.py` writes after it has verified the
pinned corpus. Only source-eligible files are read. Each section is checked twice, with the default rules and with
every rule of the catalog on. A row of data/sentences.jsonl.gz is one sentence of one editor's file: where it
stands, whether an edit of that editor touches it, whether such an edit belongs to the redundancy aspect of the
corpus, whether a finding of the default rules touches it, and which rules touch it when every rule is on. No text
is written; `same` is a hash of the sentence's words, so that the sentences of two editors of one paper can be
matched.
"""
import concurrent.futures, gzip, hashlib, json, re, subprocess, sys, tempfile
from pathlib import Path

source, binary, catalog, record = sys.argv[1], sys.argv[2], sys.argv[3], Path(sys.argv[4])
workers = int(sys.argv[sys.argv.index("--workers") + 1]) if "--workers" in sys.argv else 6
# The edit types that the corpus groups under its aspect "Redundancy" (aspect_edit-type_map.txt).
REDUNDANCY = {"repetitiveness", "redundancy", "redundant", "conciseness"}
BOUNDARY = re.compile(r'[.!?]+["”’)\]]*(?=\s+["“‘(\[]?[A-Z0-9])|\n+')
rules = [rule["id"] for family in json.load(open(catalog))["families"] for rule in family["rules"]]
EVERY = "version: 1\nextends: [builtin:custom]\nrules:\n" + "".join(f"  {rule}:\n    enabled: true\n" for rule in rules)


def sentences(text):
    """Byte spans of the sentences of a text, by the rule of the record of human references."""
    spans, start = [], 0
    for match in BOUNDARY.finditer(text):
        spans.append((start, match.end()))
        start = match.end()
    spans.append((start, len(text)))
    result = []
    for begin, end in spans:
        piece = text[begin:end]
        stripped = piece.strip()
        if not stripped or not re.search(r"[A-Za-z0-9]", stripped):
            continue
        lead = len(piece) - len(piece.lstrip())
        begin, end = begin + lead, begin + lead + len(stripped)
        result.append((len(text[:begin].encode()), len(text[:end].encode())))
    return result


def overlaps(span, others):
    return any(span[0] < other[1] and other[0] < span[1] for other in others)


def findings(text, config):
    """The findings of one run, each as its byte span and its rule."""
    with tempfile.TemporaryDirectory() as scratch:
        extra = []
        if config:
            Path(scratch, "every.yaml").write_text(config)
            extra = ["--config", "every.yaml"]
        done = subprocess.run([binary, "check", "--stdin", "--filename", "section.txt", "--format", "text", "--no-gate", "--allow-empty",
                               "--timeout", "2m", "--report", "json:-", *extra], input=text.encode(), cwd=scratch, capture_output=True)
    if done.returncode != 0:
        raise SystemExit("unswell check failed: " + done.stderr.decode("utf-8", "replace")[-300:])
    return [(f["primary"]["span"]["start"], f["primary"]["span"]["end"], f["rule_id"]) for f in json.loads(done.stdout).get("findings") or []]


def section_rows(job):
    file, section = job
    text = section["original"]
    edits = [e for e in section["edits"] if not e["unchanged"] and e["original_span"]["start"] < e["original_span"]["end"]]
    every = [(e["original_span"]["start"], e["original_span"]["end"]) for e in edits]
    redundant = [(e["original_span"]["start"], e["original_span"]["end"]) for e in edits if REDUNDANCY & set(e["types"])]
    default, enabled = findings(text, None), findings(text, EVERY)
    raw = text.encode()
    rows = []
    for number, span in enumerate(sentences(text)):
        words = re.findall(r"[A-Za-z0-9]+", raw[span[0]:span[1]].decode("utf-8", "replace").lower())
        rows.append({"paper": file["paper"], "editor": file["editor"], "section": section["name"], "sentence": number, "words": len(words),
                     "same": hashlib.sha256(" ".join(words).encode()).hexdigest()[:12],
                     "edited": overlaps(span, every), "redundancy": overlaps(span, redundant),
                     "default": overlaps(span, default),
                     "every": sorted({found[2] for found in enabled if overlaps(span, [found])})})
    return rows


data = json.load(open(source))
jobs = [(file, section) for file in data["files"] if file["source_eligible"] for section in file["sections"]]
with concurrent.futures.ThreadPoolExecutor(workers) as pool:
    rows = [row for part in pool.map(section_rows, jobs) for row in part]
rows.sort(key=lambda r: (r["paper"], r["editor"], r["section"], r["sentence"]))
out = record / "data" / "sentences.jsonl.gz"
with open(out, "wb") as raw, gzip.GzipFile(fileobj=raw, mode="wb", mtime=0, filename="") as handle:
    handle.write("".join(json.dumps(row, sort_keys=True) + "\n" for row in rows).encode())
print("sections", len(jobs), "sentences", len(rows), "rules", len(rules))
