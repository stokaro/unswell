#!/usr/bin/env python3
"""Test a list of phrases that chat models favor on expert-edited answers the record has not used.

usage: vocabulary.py <human directory> prepare
       vocabulary.py <human directory> run <part> <phrases json> <unswell binary> [--workers N]
       vocabulary.py <human directory> select <phrases json> <out json>
       vocabulary.py <human directory> confirm <phrases json>
       vocabulary.py <human directory> export <phrases json> <record directory>
       vocabulary.py <human directory> load <phrases json> <unswell binary> <out json> [--workers N]

Registered on issue 525 (comment 6832853530). The fresh answers are the
eligible Beemo rows beyond the first 450 in hash order: 451 to 1,300 are the
selection part and the rest the confirmation part. A phrase is dropped on the
selection part when it has at least four findings and the expert removed or
replaced every flagged word in fewer than half of them.
"""
import collections, concurrent.futures, difflib, gzip, hashlib, json, math, random, re, subprocess, sys, tempfile
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))
from common import byte_span, read_jsonl, write_jsonl
from trim_editors import changed_flags, words_of

ROOT = Path(sys.argv[1]).resolve()
USED, SELECTION_END = 450, 1300
MIN_FINDINGS, KEEP_SHARE = 4, 0.5
BOOT, SEED = 2000, 20261112
PARTS = ("selection", "confirmation")


def prepare():
    """Write docs-fresh.jsonl in the shape of docs.jsonl, with the same reference as prepare.py gives."""
    candidates = []
    for line in open(ROOT / "raw" / "beemo.jsonl"):
        row = json.loads(line)
        text, edited = row["model_output"] or "", row["human_edits"] or ""
        if not text.strip() or not edited.strip() or text.strip() == edited.strip():
            continue
        tokens = [(m.group(), m.start(), m.end()) for m in re.finditer(r"\S+", text)]
        other = re.findall(r"\S+", edited)
        matcher = difflib.SequenceMatcher(a=[t[0] for t in tokens], b=other, autojunk=False)
        spans = [{"span": list(byte_span(text, tokens[a0][1], tokens[a1 - 1][2])), "category": tag}
                 for tag, a0, a1, _, _ in matcher.get_opcodes() if tag in ("replace", "delete") and a1 > a0]
        if spans:
            candidates.append({"id": f"beemo-{int(row['id']):04d}", "set": "beemo", "format": "markdown", "text": text,
                               "editor": "expert", "kind": row["category"], "generator": row["model"], "reference": spans})
    candidates.sort(key=lambda r: hashlib.sha256(r["id"].encode()).hexdigest())
    used = {d["id"] for d in read_jsonl(ROOT / "docs.jsonl") if d["set"] == "beemo"}
    assert {c["id"] for c in candidates[:USED]} == used, "the first 450 are not the answers of the record"
    fresh = candidates[USED:]
    for index, row in enumerate(fresh, USED + 1):
        row["split"] = "selection" if index <= SELECTION_END else "confirmation"
    write_jsonl(ROOT / "docs-fresh.jsonl", fresh)
    print(collections.Counter(row["split"] for row in fresh))


def fresh(part):
    return [d for d in read_jsonl(ROOT / "docs-fresh.jsonl") if d["split"] == part]


def normal(text):
    return " ".join(text.lower().replace("’", "'").split())


def run(part, phrases_path, binary, workers):
    phrases = json.loads(Path(phrases_path).read_text())
    policy = ("version: 1\nextends: [builtin:custom]\nrules:\n  policy.banned-phrases:\n    enabled: true\n"
              "    parameters:\n      phrases: " + json.dumps(phrases) + "\n      positions: [any]\n")
    out = ROOT / "runs" / f"vocabulary-{part}"
    out.mkdir(parents=True, exist_ok=True)

    def one(doc):
        target = out / f"{doc['id']}.json"
        if target.exists():
            return True
        with tempfile.TemporaryDirectory() as scratch:
            Path(scratch, "policy.yaml").write_text(policy)
            done = subprocess.run([binary, "check", "--stdin", "--filename", doc["id"] + ".md", "--format", "markdown",
                                   "--config", "policy.yaml", "--no-gate", "--allow-empty", "--report", "json:-"],
                                  input=doc["text"].encode(), cwd=scratch, capture_output=True)
        if done.returncode != 0:
            print(doc["id"], "FAILED", done.stderr.decode("utf-8", "replace")[-200:], flush=True)
            return False
        raw = doc["text"].encode()
        found = []
        for finding in json.loads(done.stdout).get("findings") or []:
            span = finding["primary"]["span"]
            found.append({"span": [span["start"], span["end"]], "phrase": normal(raw[span["start"]:span["end"]].decode())})
        target.write_text(json.dumps({"findings": found}))
        return True

    docs = fresh(part)
    with concurrent.futures.ThreadPoolExecutor(workers) as pool:
        done = list(pool.map(one, docs))
    print(f"vocabulary-{part}: {len(docs)} documents, {done.count(False)} failed")


def findings(part, phrases):
    """Yield, per document, its word flags and the findings of listed phrases with the indexes of their words."""
    listed = {normal(p) for p in phrases}
    for doc in fresh(part):
        words = words_of(doc["text"].encode())
        changed = changed_flags(words, doc["reference"])
        found = []
        for finding in json.loads((ROOT / "runs" / f"vocabulary-{part}" / f"{doc['id']}.json").read_text())["findings"]:
            if finding["phrase"] not in listed:
                continue
            begin, end = finding["span"]
            inside = [i for i, (_, a, b) in enumerate(words) if begin <= a and b <= end]
            if inside:
                found.append((finding["phrase"], finding["span"], inside))
        yield doc, words, changed, found


def wilson(yes, total, z=1.96):
    if not total:
        return [0.0, 0.0]
    share = yes / total
    centre = (share + z * z / (2 * total)) / (1 + z * z / total)
    half = z * (share * (1 - share) / total + z * z / (4 * total * total)) ** 0.5 / (1 + z * z / total)
    return [centre - half, centre + half]


def select(phrases_path, out_path):
    phrases = json.loads(Path(phrases_path).read_text())
    count, whole = collections.Counter(), collections.Counter()
    for _, _, changed, found in findings("selection", phrases):
        for phrase, _, inside in found:
            count[phrase] += 1
            whole[phrase] += all(changed[i] for i in inside)
    dropped = sorted(p for p in count if count[p] >= MIN_FINDINGS and whole[p] < KEEP_SHARE * count[p])
    kept = [p for p in phrases if normal(p) not in dropped]
    Path(out_path).write_text(json.dumps(kept, indent=1) + "\n")
    print(f"{len(phrases)} phrases, {sum(count.values())} findings of {len(count)} phrases on the selection part")
    for phrase, number in count.most_common():
        mark = "DROPPED" if phrase in dropped else ""
        print(f"  {number:4} findings, every flagged word changed in {whole[phrase]:4}  {phrase}  {mark}")
    print(f"dropped {len(dropped)}: {dropped}; kept {len(kept)}")


def rows_for(part, phrases):
    """Per-document sums for the word-level comparison, and per-finding outcomes."""
    table, outcomes = [], []
    for doc, words, changed, found in findings(part, phrases):
        under = sorted({i for _, _, inside in found for i in inside})
        share = sum(changed) / len(words) if words else 0.0
        table.append((len(under), sum(changed[i] for i in under), len(under) * share))
        outcomes += [all(changed[i] for i in inside) for _, _, inside in found]
    return table, outcomes


def confirm(phrases_path):
    phrases = json.loads(Path(phrases_path).read_text())
    for part in PARTS:
        table, outcomes = rows_for(part, phrases)
        under, hits, chance = (sum(column) for column in zip(*table))
        rng = random.Random(f"{SEED}|vocabulary|{part}")
        draws = []
        for _ in range(BOOT):
            a = b = c = 0.0
            for _ in table:
                row = table[rng.randrange(len(table))]
                a += row[0]
                b += row[1]
                c += row[2]
            if a:
                draws.append((b - c) / a)
        draws.sort()
        low, high = draws[int(0.025 * len(draws))], draws[int(0.975 * len(draws))]
        yes, total = sum(outcomes), len(outcomes)
        interval = wilson(yes, total)
        print(f"{part}: documents {len(table)}, with a finding {sum(1 for r in table if r[0])}, findings {total}, words under findings {int(under)}")
        print(f"   words the expert changed {hits / under:.3f}, chance {chance / under:.3f}, "
              f"difference {(hits - chance) / under:+.3f} [{low:+.3f}, {high:+.3f}]")
        print(f"   findings with every flagged word changed {yes} of {total} = {yes / total:.3f} "
              f"[{interval[0]:.3f}, {interval[1]:.3f}]")


def export(phrases_path, record):
    """Write the findings on both parts for the record: coordinates, counts, and the index of each phrase.

    The selection part carries the findings of the whole list, so that the record can recount which phrases
    left it. The confirmation part ran with the kept phrases only.
    """
    full = json.loads((ROOT / "vocabulary-full.json").read_text())
    kept = json.loads(Path(phrases_path).read_text())
    index = {normal(phrase): number for number, phrase in enumerate(full)}
    rows = []
    for part, phrases in (("selection", full), ("confirmation", kept)):
        for doc, words, changed, found in findings(part, phrases):
            rows.append({"id": doc["id"], "part": part, "text_sha256": hashlib.sha256(doc["text"].encode()).hexdigest(),
                         "words": len(words), "changed": sum(changed),
                         "reference": sorted(entry["span"] for entry in doc["reference"]),
                         "findings": [{"span": span, "phrase": index[phrase], "words": [[words[i][1], words[i][2]] for i in inside]}
                                      for phrase, span, inside in sorted(found, key=lambda item: (item[1], index[item[0]]))]})
    path = Path(record) / "data" / "vocabulary.jsonl.gz"
    with gzip.GzipFile(path, "wb", mtime=0) as handle:
        for row in sorted(rows, key=lambda r: r["id"]):
            handle.write((json.dumps(row, sort_keys=True) + "\n").encode())
    dropped = [phrase for phrase in full if phrase not in kept]
    with gzip.GzipFile(Path(record) / "data" / "vocabulary-phrases.jsonl.gz", "wb", mtime=0) as handle:
        handle.write((json.dumps({"phrases": full, "dropped": dropped}, sort_keys=True) + "\n").encode())
    print(f"exported {len(rows)} documents to {path}")


def load(phrases_path, binary, out_path, workers):
    """Count the findings of the list on the pages of public repositories of the assistant-panel corpus."""
    import glob
    phrases = json.loads(Path(phrases_path).read_text())
    policy = ("version: 1\nextends: [builtin:custom]\nrules:\n  policy.banned-phrases:\n    enabled: true\n"
              "    parameters:\n      phrases: " + json.dumps(phrases) + "\n      positions: [any]\n")
    pages = []
    for path in sorted(glob.glob(str(ROOT.parent / "docs-real*.jsonl"))):
        pages += [(row["doc_id"], row["text"]) for row in read_jsonl(path)]

    def one(page):
        name, text = page
        with tempfile.TemporaryDirectory() as scratch:
            Path(scratch, "policy.yaml").write_text(policy)
            done = subprocess.run([binary, "check", "--stdin", "--filename", name + ".md", "--format", "markdown",
                                   "--config", "policy.yaml", "--no-gate", "--allow-empty", "--report", "json:-"],
                                  input=text.encode(), cwd=scratch, capture_output=True)
        if done.returncode != 0:
            return None
        return len(json.loads(done.stdout).get("findings") or []), len(words_of(text.encode()))

    with concurrent.futures.ThreadPoolExecutor(workers) as pool:
        results = list(pool.map(one, pages))
    done = [r for r in results if r is not None]
    summary = {"pages": len(pages), "failed": results.count(None), "pages_with_finding": sum(1 for f, _ in done if f),
               "findings": sum(f for f, _ in done), "words": sum(w for _, w in done)}
    Path(out_path).write_text(json.dumps(summary, indent=1, sort_keys=True) + "\n")
    print(summary, f"findings per 1,000 words {1000 * summary['findings'] / summary['words']:.2f}")


if __name__ == "__main__":
    command = sys.argv[2]
    if command == "prepare":
        prepare()
    elif command == "run":
        workers = int(sys.argv[sys.argv.index("--workers") + 1]) if "--workers" in sys.argv else 6
        run(sys.argv[3], sys.argv[4], sys.argv[5], workers)
    elif command == "select":
        select(sys.argv[3], sys.argv[4])
    elif command == "confirm":
        confirm(sys.argv[3])
    elif command == "load":
        workers = int(sys.argv[sys.argv.index("--workers") + 1]) if "--workers" in sys.argv else 6
        load(sys.argv[3], sys.argv[4], sys.argv[5], workers)
    elif command == "export":
        export(sys.argv[3], sys.argv[4])
    else:
        raise SystemExit(__doc__)
