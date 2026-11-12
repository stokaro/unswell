#!/usr/bin/env python3
"""Bring LAMP and Beemo to one shape: text, split, and the spans an editor changed.

usage: prepare.py <LAMP Writing_Alignment directory> <out directory>

LAMP gives the edited span of every edit. Beemo gives the edited text only, so
its reference is the run of words that the expert removed or replaced, found
with a word-level alignment. Words that the expert only added have no span in
the original and are not counted.
"""
import ast, collections, difflib, hashlib, json, re, sys, time, urllib.request
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))
from common import byte_span, write_jsonl

LAMP, OUT = Path(sys.argv[1]), Path(sys.argv[2])
RAW = OUT / "raw"
RAW.mkdir(parents=True, exist_ok=True)
SPLITS = {"validation": "dev", "test": "test"}


def locate(text, edits):
    """Find each edit's original text in the paragraph, left to right per text."""
    spans, missing, used = [], 0, collections.defaultdict(int)
    for edit in edits:
        original = (edit.get("originalText") or "").strip()
        if not original:
            missing += 1
            continue
        at = text.find(original, used[original])
        if at < 0:
            at = text.find(original)
        if at < 0:
            missing += 1
            continue
        used[original] = at + len(original)
        category = re.sub(r"\s*\(.*\)$", "", edit.get("categorization", "")).replace("/ ", "/").strip()
        spans.append({"span": list(byte_span(text, at, at + len(original))), "category": category})
    return spans, missing


rows, stats = [], collections.Counter()
for index, row in enumerate(json.loads((LAMP / "LAMP/LAMP-train-val-test.json").read_text())):
    split = SPLITS.get(row.get("data-split"))
    if split is None:
        continue
    spans, missing = locate(row["preedit"], row["fine_grained_edits"])
    stats[f"lamp {split} edits located"] += len(spans)
    stats[f"lamp {split} edits not located"] += missing
    rows.append({"id": f"lamp-{index:04d}", "set": "lamp", "split": split, "format": "text", "text": row["preedit"],
                 "editor": row.get("editor") or row["id"].split("_")[0], "kind": row["type"], "generator": row["source"],
                 "reference": spans})
for name in ("W3", "W12", "W16"):
    for index, row in enumerate(json.loads((LAMP / f"LAMP/reannotation/{name}_edits.json").read_text())):
        edits = row["fine_grained_edits"]
        if isinstance(edits, str):
            edits = ast.literal_eval(edits)
        spans, missing = locate(row["preedit"], edits)
        stats["lamp ceiling edits located"] += len(spans)
        stats["lamp ceiling edits not located"] += missing
        rows.append({"id": f"lamp-re-{name}-{index:02d}", "set": "lamp", "split": "ceiling", "format": "text",
                     "text": row["preedit"], "editor": name, "kind": row["type"], "generator": row["source"],
                     "reference": spans})

beemo = RAW / "beemo.jsonl"
if not beemo.exists():
    fetched = []
    for offset in range(0, 2200, 100):
        url = ("https://datasets-server.huggingface.co/rows?dataset=toloka/beemo&config=default&split=train"
               f"&offset={offset}&length=100")
        for attempt in range(5):
            try:
                with urllib.request.urlopen(url, timeout=60) as response:
                    page = json.load(response)
                break
            except Exception:  # noqa: BLE001
                time.sleep(5 * (attempt + 1))
        else:
            raise SystemExit(f"cannot fetch {url}")
        fetched += [entry["row"] for entry in page["rows"]]
        if len(page["rows"]) < 100:
            break
    write_jsonl(beemo, [{key: row[key] for key in ("id", "category", "model", "model_output", "human_edits")} for row in fetched])
candidates = []
for line in open(beemo):
    row = json.loads(line)
    text, edited = row["model_output"] or "", row["human_edits"] or ""
    if not text.strip() or not edited.strip() or text.strip() == edited.strip():
        stats["beemo unchanged or empty"] += 1
        continue
    tokens = [(m.group(), m.start(), m.end()) for m in re.finditer(r"\S+", text)]
    other = re.findall(r"\S+", edited)
    matcher = difflib.SequenceMatcher(a=[t[0] for t in tokens], b=other, autojunk=False)
    spans = []
    for tag, a0, a1, _, _ in matcher.get_opcodes():
        if tag in ("replace", "delete") and a1 > a0:
            spans.append({"span": list(byte_span(text, tokens[a0][1], tokens[a1 - 1][2])), "category": tag})
    if not spans:
        stats["beemo additions only"] += 1
        continue
    candidates.append({"id": f"beemo-{int(row['id']):04d}", "set": "beemo", "format": "markdown", "text": text,
                       "editor": "expert", "kind": row["category"], "generator": row["model"], "reference": spans})
candidates.sort(key=lambda r: hashlib.sha256(r["id"].encode()).hexdigest())
for index, row in enumerate(candidates[:450]):
    row["split"] = "dev" if index < 150 else "test"
    rows.append(row)
stats["beemo eligible"] = len(candidates)
write_jsonl(OUT / "docs.jsonl", rows)
for key, value in sorted(stats.items()):
    print(key, value)
print(collections.Counter((r["set"], r["split"]) for r in rows))
