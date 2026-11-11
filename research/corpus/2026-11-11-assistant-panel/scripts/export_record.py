#!/usr/bin/env python3
"""Export the corpus record for the repository, without third-party text.

usage: export_record.py <workspace> <out dir>

Writes manifests, item coordinates, annotation labels, finding reviews with
their spans, audit outcomes, phrase statistics and decisions, the engine
configurations, and a summary. Page text, quotes, and reviewer
prose are left out. A page is fetched again from its pinned commit when needed.
"""
import collections, gzip, json, shutil, sys
from pathlib import Path

W, OUT = Path(sys.argv[1]), Path(sys.argv[2])
SETS = [("long-prose", "docs.jsonl", "development"), ("real-dev-1", "docs-realdev.jsonl", "development"),
        ("real-dev-2", "docs-realdev2.jsonl", "development"), ("real-test-1a", "docs-real.jsonl", "final test 1"),
        ("real-test-1b", "docs-real2.jsonl", "final test 1"), ("real-test-2", "docs-real3.jsonl", "final test 2"),
        ("real-test-3", "docs-real4.jsonl", "final test 3")]
MANIFESTS = {"real-dev-1": "realdev", "real-dev-2": "realdev2", "real-test-1a": "real", "real-test-1b": "real2",
             "real-test-2": "real3", "real-test-3": "real4"}
ROUNDS = [("development", "rev", ["report-broad-dev.json", "report-broad-dev2.json"]),
          ("final-test-1", "rev-final", ["report-final-new.json"]),
          ("final-test-2", "rev-real3", ["report-real3-r2.json"]),
          ("final-test-3", "rev-real4", ["report-real4-r3.json"]),
          ("final-test-3-removed", "rev-real4-removed", ["report-real4-removed.json"])]
AUDITS = [("development", "audit-spans-dev.json"), ("final-test-1", "audit-spans-final.json")]


def write_gz(path, rows):
    path.parent.mkdir(parents=True, exist_ok=True)
    with gzip.GzipFile(path, "wb", mtime=0) as handle:
        for row in rows:
            handle.write((json.dumps(row, ensure_ascii=False, sort_keys=True) + "\n").encode())


PANEL = {"astra": {"model": "gpt-6-astra", "effort": "medium", "roles": ["annotator", "reviewer"]},
         "claude": {"model": "claude-sonnet-5-5", "effort": "default", "roles": ["annotator", "reviewer"]},
         "qwen": {"model": "private-host/Qwen3.8-27B", "effort": "default", "roles": ["auditor", "reviewer"]}}
summary = {"panel": PANEL, "sets": [], "rounds": []}
for name, docs_file, role in SETS:
    items, labels = [], []
    documents = repositories = item_count = 0
    seen = set()
    both = collections.Counter()
    for line in open(W / docs_file):
        doc = json.loads(line)
        documents += 1
        seen.add(doc["repository"])
        item_count += len(doc["items"])
        items.append({"doc": doc["doc_id"], "repository": doc["repository"], "path": doc["path"],
                      "sha256": doc["sha256"], "reference": doc.get("reference", ""), "origin": doc["origin"],
                      "family": doc.get("family", ""), "operation": doc.get("operation", ""),
                      "items": [{"id": i["id"], "kind": i["kind"], "segments": i["segments"]} for i in doc["items"]]})
        marks = {}
        for annotator in ("astra", "claude"):
            path = W / "ann" / annotator / f"{doc['doc_id']}.json"
            if not path.exists():
                continue
            answer = json.loads(path.read_text())
            flags = [{"item": f["item"], "label": f["label"], "categories": f["categories"]} for f in answer["flags"]]
            labels.append({"doc": doc["doc_id"], "annotator": annotator, "model": answer["model"],
                           "effort": answer["effort"], "rubric": answer["rubric"],
                           "prompt_sha256": answer["prompt_sha256"], "flags": flags,
                           "unbound_flags": len(answer.get("dropped", []))})
            marks[annotator] = {f["item"] for f in flags if f["label"] == "needs_revision"}
        if len(marks) == 2:
            both["documents"] += 1
            both["items"] += len(doc["items"])
            both["astra"] += len(marks["astra"])
            both["claude"] += len(marks["claude"])
            both["both"] += len(marks["astra"] & marks["claude"])
    write_gz(OUT / "items" / f"{name}.jsonl.gz", items)
    if labels:
        write_gz(OUT / "labels" / f"{name}.jsonl.gz", labels)
    summary["sets"].append({"set": name, "role": role, "documents": documents, "repositories": len(seen),
                            "items": item_count, "annotated": dict(both)})
    prefix = MANIFESTS.get(name)
    if prefix:
        (OUT / "manifests").mkdir(parents=True, exist_ok=True)
        manifest = json.loads((W / f"{prefix}-manifest.json").read_text())
        (OUT / "manifests" / f"{name}.json").write_text(json.dumps(manifest, indent=1, ensure_ascii=False) + "\n")
        log = json.loads((W / f"{prefix}-selection-log.json").read_text())
        (OUT / "manifests" / f"{name}-selection.json").write_text(json.dumps(log, indent=1, ensure_ascii=False) + "\n")

rename = {}
for name, docs_file, _ in SETS:
    folder = {"long-prose": "v0"}.get(name, MANIFESTS.get(name))
    rename[folder] = name
for round_name, folder, reports in ROUNDS:
    located = {}
    for report in reports:
        for finding in json.loads((W / "engine" / report).read_text())["findings"]:
            span = finding["primary"]["span"]
            located[finding["id"]] = {"severity": finding["severity"], "span": [span["start"], span["end"]]}
    findings = {}
    for reviewer in ("claude", "astra", "qwen"):
        for path in sorted((W / folder / reviewer).glob("*.json")) if (W / folder / reviewer).exists() else []:
            record = json.loads(path.read_text())
            source_set, document = record["path"].split("/", 1)
            for finding in record["findings"]:
                row = findings.setdefault(finding["id"], {"id": finding["id"], "set": rename.get(source_set, source_set),
                                                          "doc": document.removesuffix(".md"), "rule": finding["rule"],
                                                          "group": finding["group"], "phrase": finding["phrase"],
                                                          "reviews": {}, **located[finding["id"]]})
                row["reviews"][reviewer] = finding.get("justified")
    rows = sorted(findings.values(), key=lambda r: (r["set"], r["doc"], r["id"]))
    write_gz(OUT / "reviews" / f"{round_name}.jsonl.gz", rows)
    table = collections.defaultdict(collections.Counter)
    for row in rows:
        first, second = row["reviews"].get("claude"), row["reviews"].get("astra") or row["reviews"].get("qwen")
        if not first or not second:
            continue
        c = table[row["rule"]]
        c["reviewed"] += 1
        c["both_accept"] += first == "yes" and second == "yes"
        c["any_accept"] += first == "yes" or second == "yes"
        c["both_reject"] += first == "no" and second == "no"
    summary["rounds"].append({"round": round_name, "findings": len(rows),
                              "rules": {rule: dict(c) for rule, c in sorted(table.items())}})

(OUT / "audit").mkdir(parents=True, exist_ok=True)
for name, source in AUDITS:
    outcomes = [{"kind": r["kind"], "rule": r["rule"], "phrase": r["phrase"], "edited_is": r["edited_is"],
                 "better": (r["answer"] or {}).get("better"),
                 "information_lost": (r["answer"] or {}).get("information_lost"), "confirmed": r.get("confirmed")}
                for r in json.loads((W / source).read_text())]
    (OUT / "audit" / f"{name}.json").write_text(json.dumps(outcomes, indent=1) + "\n")
items_audit = [{"stratum": r["stratum"], "doc": r["doc"], "item": r["item"], "categories": r["categories"],
                "edited_is": r["edited_is"], "better": (r["answer"] or {}).get("better"),
                "information_lost": (r["answer"] or {}).get("information_lost"), "confirmed": r.get("confirmed"),
                "strict": r.get("strict")} for r in json.loads((W / "audit-v0.json").read_text())]
(OUT / "audit" / "long-prose-items.json").write_text(json.dumps(items_audit, indent=1) + "\n")
(OUT / "phrases").mkdir(parents=True, exist_ok=True)
shutil.copy(W / "lexicon-candidates.json", OUT / "phrases" / "candidates.json")
decisions = json.loads((W / "engine/release.decisions.json").read_text())
(OUT / "phrases" / "decisions.json").write_text(json.dumps(decisions, indent=1, ensure_ascii=False) + "\n")
mined = json.loads((W / "lexicon-v0.json").read_text())
for row in mined["phrases"]:
    row.pop("unflagged_examples", None)
(OUT / "phrases" / "mined.json").write_text(json.dumps(mined, indent=1, ensure_ascii=False) + "\n")
for name, source in (("excess.json", "excess-v0.json"), ("assignment.json", "engine/release.assignment.json"),
                     ("rule-statements.json", "rule-statements.json")):
    data = json.loads((W / source).read_text())
    (OUT / "phrases" / name).write_text(json.dumps(data, indent=1, ensure_ascii=False) + "\n")
(OUT / "rounds").mkdir(parents=True, exist_ok=True)
for name, source in (("round-3-development.json", "engine/round3-dev.json"),
                     ("round-3-test.json", "engine/round3-test.json"),
                     ("round-3-length-bands.json", "engine/round3-evidence.json")):
    shutil.copy(W / source, OUT / "rounds" / name)
(OUT / "configs").mkdir(parents=True, exist_ok=True)
for name in ("broad.yaml", "release.yaml"):
    shutil.copy(W / "engine" / name, OUT / "configs" / name)
for name, prefix in MANIFESTS.items():
    failures = json.loads((W / f"{prefix}-parse-failures.json").read_text())
    (OUT / "manifests" / f"{name}-parse-failures.json").write_text(json.dumps(failures, indent=1) + "\n")
(OUT / "summary.json").write_text(json.dumps(summary, indent=1, sort_keys=True) + "\n")
size = sum(p.stat().st_size for p in OUT.rglob("*") if p.is_file())
print(f"exported {sum(1 for p in OUT.rglob('*') if p.is_file())} files, {size / 1e6:.1f} MB")
for entry in summary["sets"]:
    print("  ", entry["set"], entry["documents"], "documents", entry["items"], "items", entry["annotated"])
