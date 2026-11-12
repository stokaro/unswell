#!/usr/bin/env python3
"""Export the measurement against human references for the repository, without third-party text.

usage: export_human.py <human directory> <record directory>

Writes coordinates and choices only: for every document its sentence spans,
the spans an editor changed, and each system's finding spans; for every
preference triplet the writers' orders and each judge's order; for every
expert-edited pair each judge's choice. Text is identified by SHA-256.
"""
import collections, gzip, hashlib, json, re, sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))
from common import byte_span, read_jsonl, sentences

ROOT, OUT = Path(sys.argv[1]), Path(sys.argv[2])
(OUT / "data").mkdir(parents=True, exist_ok=True)


def write_gz(path, rows):
    with gzip.GzipFile(path, "wb", mtime=0) as handle:
        for row in rows:
            handle.write((json.dumps(row, ensure_ascii=False, sort_keys=True) + "\n").encode())


systems = sorted(p.name for p in (ROOT / "runs").iterdir() if p.is_dir())
rows = []
for doc in read_jsonl(ROOT / "docs.jsonl"):
    found = {}
    for system in systems:
        path = ROOT / "runs" / system / f"{doc['id']}.json"
        if path.exists():
            found[system] = sorted(json.loads(path.read_text())["spans"])
    row = {"id": doc["id"], "set": doc["set"], "split": doc["split"], "editor": doc["editor"] if doc["set"] != "prs" else "",
           "kind": doc["kind"], "generator": doc.get("generator", ""),
           "text_sha256": hashlib.sha256(doc["text"].encode()).hexdigest(),
           "units": [list(u) for u in (doc.get("units") or sentences(doc["text"]))],
           "reference": sorted(entry["span"] for entry in doc["reference"]), "systems": found}
    if doc["set"] == "prs":
        row.update({"repository": doc["repo"], "pull_request": doc["number"], "path": doc["path"], "base": doc["base"]})
    if doc["split"] == "ceiling":
        row["words"] = [list(byte_span(doc["text"], m.start(), m.end())) for m in re.finditer(r"\S+", doc["text"])]
    rows.append(row)
write_gz(OUT / "data" / "sentences.jsonl.gz", sorted(rows, key=lambda r: r["id"]))

writers = collections.defaultdict(list)
for row in read_jsonl(ROOT / "preferences.jsonl"):
    writers[row["triplet"]].append({"writer": hashlib.sha256(row["writer"].encode()).hexdigest()[:8], "order": row["order"],
                                    "setting": row["setting"]})
judges = collections.defaultdict(dict)
for directory in sorted((ROOT / "judgments" / "triplets").iterdir()):
    for path in directory.glob("*.json"):
        judges[path.stem][directory.name] = json.loads(path.read_text())["order"]
write_gz(OUT / "data" / "preferences.jsonl.gz",
         [{"triplet": triplet, "writers": sorted(entries, key=lambda e: e["writer"]), "judges": judges.get(triplet, {})}
          for triplet, entries in sorted(writers.items())])

pairs = collections.defaultdict(dict)
for directory in sorted((ROOT / "judgments" / "beemo").iterdir()) if (ROOT / "judgments" / "beemo").exists() else []:
    for path in directory.glob("*.json"):
        answer = json.loads(path.read_text())
        pairs[path.stem][directory.name] = {"better": answer["better"], "position": answer["position"]}
write_gz(OUT / "data" / "expert-pairs.jsonl.gz", [{"id": key, "judges": value} for key, value in sorted(pairs.items())])
key_path = ROOT / "study/key.json"
if key_path.exists():
    key = json.loads(key_path.read_text())
    answers = []
    for path in sorted((ROOT / "study/answers/answers").glob("*.json")):
        document = json.loads(path.read_text())
        choices = (document.get("data") or document).get("choices") or {}
        reader = hashlib.sha256(path.stem.encode()).hexdigest()[:8]
        for pair, choice in sorted(choices.items()):
            entry = key.get(pair)
            if entry is None or choice not in ("left", "right", "same"):
                continue
            outcome = "same" if choice == "same" else "revised" if choice == entry["revised"] else "original"
            answers.append({"reader": reader, "pair": pair, "arm": entry["arm"], "reviewer": entry["reviewer"], "outcome": outcome})
    write_gz(OUT / "data" / "readers.jsonl.gz", answers)
    print("reader answers", len(answers))
print("documents", len(rows), collections.Counter((r["set"], r["split"]) for r in rows), "systems", systems,
      "triplets", len(writers), "expert pairs", len(pairs))
