#!/usr/bin/env python3
"""Prepare a set of real pages: complete the manifest, extract units, build the document file.

usage: prep_real.py <workspace> <prefix> <partition>
Reads <prefix>-manifest.json and the pages under <prefix>/. Writes docs-<prefix>.jsonl.
Pages the engine cannot parse are removed and listed in <prefix>-parse-failures.json.
"""
import collections, hashlib, json, re, subprocess, sys
from pathlib import Path

W, PREFIX, PARTITION = Path(sys.argv[1]), sys.argv[2], sys.argv[3]
ROOT = W / PREFIX
manifest_path = W / f"{PREFIX}-manifest.json"
manifest = json.loads(manifest_path.read_text())
licenses = {}
for source in manifest["sources"]:
    repo = source["repository"]
    if "rights" not in source:
        if repo not in licenses:
            out = subprocess.run(["gh", "api", f"repos/{repo}", "--jq", '.license.spdx_id // "NOASSERTION"'],
                                 capture_output=True, text=True).stdout.strip()
            licenses[repo] = out or "NOASSERTION"
        source["rights"] = {"license": licenses[repo],
                            "evidence": "License identifier reported by the hosting service for the repository at selection time. The page is annotated locally and is not redistributed.",
                            "allowed_uses": ["annotation"]}
    folder = source["path"].split("/")[0]
    commit = re.search(r"/blob/([0-9a-f]{40})/", source["reference"]).group(1)
    notice = (f"Source notice\n\nRepository: {repo}\nCommit: {commit}\n"
              f"License identifier reported by the hosting service: {source['rights']['license']}\n"
              "The page is used locally for annotation and is not redistributed.\n").encode()
    (ROOT / folder / "SOURCE-NOTICE.txt").write_bytes(notice)
    source["notices"] = [{"path": f"{folder}/SOURCE-NOTICE.txt", "sha256": hashlib.sha256(notice).hexdigest(), "bytes": len(notice)}]
    source["partition"] = PARTITION
failures = []
corpus = str(W / "bin/corpus")
CHUNK = 60  # the corpus tool bounds the size of one extraction
units = []
kept_sources = []
all_sources = manifest["sources"]
for offset in range(0, len(all_sources), CHUNK):
    part = dict(manifest)
    part["sources"] = list(all_sources[offset:offset + CHUNK])
    while part["sources"]:
        payload = json.dumps(part).encode()
        plan = subprocess.run([corpus, "plan"], input=payload, capture_output=True)
        if plan.returncode:
            sys.exit("plan failed: " + plan.stderr.decode()[:400])
        extract = subprocess.run([corpus, "extract", "--root", str(ROOT)], input=plan.stdout, capture_output=True)
        if extract.returncode == 0:
            units += json.loads(extract.stdout)["units"]
            kept_sources += part["sources"]
            break
        hit = re.search(r"extract (d\d{6}): (.*)", extract.stderr.decode())
        if not hit:
            sys.exit("extract failed: " + extract.stderr.decode()[:400])
        source = [x for x in part["sources"] if x["id"] == hit.group(1)][0]
        failures.append({"document": source["document"], "reference": source["reference"], "error": hit.group(2).strip()})
        part["sources"] = [x for x in part["sources"] if x["id"] != hit.group(1)]
manifest["sources"] = kept_sources
manifest_path.write_text(json.dumps(manifest, indent=1, ensure_ascii=False))
(W / f"{PREFIX}-parse-failures.json").write_text(json.dumps(failures, indent=1))
candidates = {"units": units}
by_sha = collections.defaultdict(list)
for wrapper in candidates["units"]:
    by_sha[wrapper["unit"]["source"]["sha256"]].append(wrapper)
sources = {s["sha256"]: s for s in manifest["sources"]}
docs = []
for digest, wrappers in by_sha.items():
    source = sources[digest]
    data = (ROOT / source["path"]).read_bytes()
    spans, seen = [], set()
    for wrapper in wrappers:
        unit = wrapper["unit"]
        segments = [(s["start"], s["end"]) for s in unit["source"]["segments"]]
        key = (unit["kind"], tuple(segments))
        if key in seen:
            continue
        seen.add(key)
        spans.append({"kind": unit["kind"], "role": unit.get("role", ""), "segments": segments,
                      "start": segments[0][0], "end": segments[-1][1], "text": unit["text"], "words": wrapper["words"]})
    sentences = [s for s in spans if s["kind"] == "sentence"]
    chosen = list(sentences)
    for span in spans:
        if span["kind"] == "paragraph":
            if not any(span["start"] <= s["start"] and s["end"] <= span["end"] for s in sentences):
                chosen.append(span)
        elif span["kind"] == "fragment" and span["words"] >= 6:
            chosen.append(span)
    chosen.sort(key=lambda s: (s["start"], s["end"]))
    for number, span in enumerate(chosen, 1):
        span["id"] = f"i{number:03d}"
    docs.append({"doc_id": f"{PREFIX}-{source['id']}", "origin": "real", "family": "", "model": "", "operation": "",
                 "prompt": "", "task_id": "", "repository": source["repository"], "path": source["path"],
                 "partition": PARTITION, "sha256": digest, "reference": source["reference"],
                 "text": data.decode("utf-8"), "items": chosen, "units_found": len(wrappers)})
docs.sort(key=lambda d: d["doc_id"])
with open(W / f"docs-{PREFIX}.jsonl", "w") as handle:
    for doc in docs:
        handle.write(json.dumps(doc, ensure_ascii=False) + "\n")
print(f"{PREFIX}: documents={len(docs)} repositories={len({d['repository'] for d in docs})} "
      f"items={sum(len(d['items']) for d in docs)} parse failures={len(failures)}")
for failure in failures:
    print("   parse failure:", failure["document"], "|", failure["error"][:80])
