#!/usr/bin/env python3
"""Build docs-real.jsonl from the corpus tool's candidates for the real pages."""
import collections, hashlib, json, sys
from pathlib import Path

W = Path(sys.argv[1])
MIN_FRAGMENT_WORDS = 6
candidates = json.loads((W / "real-candidates.json").read_text())
manifest = {s["sha256"]: s for s in json.loads((W / "real-manifest.json").read_text())["sources"]}
by_sha = collections.defaultdict(list)
for wrapper in candidates["units"]:
    by_sha[wrapper["unit"]["source"]["sha256"]].append(wrapper)
docs = []
for digest, wrappers in by_sha.items():
    source = manifest[digest]
    data = (W / "real" / source["path"]).read_bytes()
    assert hashlib.sha256(data).hexdigest() == digest
    spans, seen = [], set()
    for wrapper in wrappers:
        unit = wrapper["unit"]
        segments = [(s["start"], s["end"]) for s in unit["source"]["segments"]]
        key = (unit["kind"], tuple(segments))
        if key in seen:
            continue
        seen.add(key)
        spans.append({"kind": unit["kind"], "role": unit.get("role", ""), "segments": segments,
                      "start": segments[0][0], "end": segments[-1][1], "text": unit["text"],
                      "words": wrapper["words"]})
    sentences = [s for s in spans if s["kind"] == "sentence"]
    chosen = list(sentences)
    for span in spans:
        if span["kind"] == "paragraph":
            if not any(span["start"] <= s["start"] and s["end"] <= span["end"] for s in sentences):
                chosen.append(span)
        elif span["kind"] == "fragment" and span["words"] >= MIN_FRAGMENT_WORDS:
            chosen.append(span)
    chosen.sort(key=lambda s: (s["start"], s["end"]))
    for number, span in enumerate(chosen, 1):
        span["id"] = f"i{number:03d}"
    docs.append({"doc_id": "real-" + source["id"], "origin": "real", "family": "", "model": "",
                 "operation": "", "prompt": "", "task_id": "", "repository": source["repository"],
                 "path": source["path"], "partition": "final_test", "sha256": digest,
                 "reference": source["reference"], "text": data.decode("utf-8"), "items": chosen,
                 "units_found": len(wrappers)})
docs.sort(key=lambda d: d["doc_id"])
with open(W / "docs-real.jsonl", "w") as handle:
    for doc in docs:
        handle.write(json.dumps(doc, ensure_ascii=False) + "\n")
items = sum(len(d["items"]) for d in docs)
print(f"real documents={len(docs)} repositories={len({d['repository'] for d in docs})} items={items} "
      f"words={sum(len(d['text'].split()) for d in docs)} largest document items={max(len(d['items']) for d in docs)}")
