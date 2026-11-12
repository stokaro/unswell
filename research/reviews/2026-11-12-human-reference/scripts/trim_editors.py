#!/usr/bin/env python3
"""Compare the words that `unswell trim` removes with the words that editors changed.

usage: trim_editors.py <human directory> <split> [--out FILE] [--self-test]
       trim_editors.py <human directory> all --export <record directory>

Registered on issue 525 (comments 6827452069 and 6827499430). For every
reference set and trimming model it reports precision, chance in the trimmed
sentences, chance in the document, and, where the editor's text exists, the
net movement of the word-level edit distance to that text with a random
control. It also reports the same precision for the default rules and for the
review command from the runs that already exist.

With --export it writes two data files for the record, for the development and
test parts, with coordinates and counts only: data/trims.jsonl.gz holds, for
every document and model, each accepted trim with the byte spans of the words
of its sentence, which of them the check never removes, and which the trim
removed, plus the edit distances; data/word-counts.jsonl.gz holds, for every
document, its number of words, how many the editor changed, and the same two
counts under the spans of the earlier systems.
"""
import argparse, collections, gzip, json, random, re, sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))
from common import read_jsonl

WORD = re.compile(r"[^\W_]+(?:['’\-][^\W_]+)*")
KEPT = {"not", "no", "never", "none", "nor", "neither", "without", "cannot", "unless", "except", "only",
        "may", "might", "can", "could", "should", "must", "would", "shall",
        "typically", "usually", "often", "sometimes", "generally", "potentially",
        "possibly", "probably", "likely", "approximately", "roughly", "almost", "nearly"}
MODELS = ("claude", "astra")
SETS = ("beemo", "lamp", "prs")
DRAWS, BOOT, SEED = 20, 2000, 20261112


def removable(word):
    """Mirror the check: negations, modal and hedging words, and numbers stay."""
    return not (word in KEPT or word.endswith(("n't", "n’t")) or any(c.isdigit() for c in word))


def distance(a, b):
    """Word-level edit distance, bit-parallel (Myers 1999, as given by Hyyro)."""
    if not a:
        return len(b)
    if not b:
        return len(a)
    masks = {}
    for index, item in enumerate(a):
        masks[item] = masks.get(item, 0) | (1 << index)
    size = len(a)
    full, last = (1 << size) - 1, 1 << (size - 1)
    plus, minus, score = full, 0, size
    for item in b:
        equal = masks.get(item, 0)
        vertical = equal | minus
        horizontal = (((equal & plus) + plus) ^ plus) | equal
        up = (minus | ~(horizontal | plus)) & full
        down = plus & horizontal
        if up & last:
            score += 1
        elif down & last:
            score -= 1
        up = ((up << 1) | 1) & full
        down = (down << 1) & full
        plus = (down | ~(vertical | up)) & full
        minus = up & vertical
    return score


def plain_distance(a, b):
    row = list(range(len(b) + 1))
    for i, x in enumerate(a, 1):
        previous, row[0] = row[0], i
        for j, y in enumerate(b, 1):
            previous, row[j] = row[j], min(row[j] + 1, row[j - 1] + 1, previous + (x != y))
    return row[-1]


def self_test():
    rng = random.Random(1)
    for _ in range(3000):
        a = [rng.choice("abcde") for _ in range(rng.randint(0, 80))]
        b = [rng.choice("abcde") for _ in range(rng.randint(0, 80))]
        assert distance(a, b) == plain_distance(a, b), (a, b)
    assert removable("very") and not removable("not") and not removable("don't") and not removable("v2")
    words, kept = ["a", "very", "very", "good", "idea"], ["a", "very", "good", "idea"]
    assert removed_positions(words, kept) == ([2], True)
    assert removed_positions(["a", "b"], ["b", "a"])[1] is False
    print("self-test ok")


def removed_positions(original, trimmed):
    """Positions of original words that the trimmed sentence leaves out.

    The trimmed words are a subsequence of the original words, so matching
    them left to right finds an embedding when one exists.
    """
    positions, at = [], 0
    for index, word in enumerate(original):
        if at < len(trimmed) and trimmed[at] == word:
            at += 1
        else:
            positions.append(index)
    return positions, at == len(trimmed)


def words_of(raw):
    """Words of a UTF-8 byte string with byte spans, in lower case."""
    text = raw.decode()
    result, position, last = [], 0, 0
    for match in WORD.finditer(text):
        position += len(text[last:match.start()].encode())
        length = len(match.group().encode())
        result.append((match.group().lower(), position, position + length))
        position, last = position + length, match.end()
    return result


def editor_texts(root):
    texts = {}
    for line in open(root / "raw" / "beemo.jsonl"):
        row = json.loads(line)
        texts[f"beemo-{int(row['id']):04d}"] = row["human_edits"] or ""
    for index, row in enumerate(json.loads((root / "raw" / "lamp" / "LAMP-train-val-test.json").read_text())):
        texts[f"lamp-{index:04d}"] = row["postedit"]
    return texts


def changed_flags(words, reference):
    spans = sorted((r["span"][0], r["span"][1]) for r in reference)
    return [any(a <= start and end <= b for a, b in spans) for _, start, end in words]


def trim_document(doc, run, edited, rng):
    """Per-document counts for one model's accepted trims."""
    raw = doc["text"].encode()
    words = words_of(raw)
    changed = changed_flags(words, doc["reference"])
    row = collections.Counter(words=len(words), changed=sum(changed), trims=len(run["trims"]), rejected=len(run["rejected"]),
                              proposed=run["proposed"], sentences=run["sentences"])
    drops, plans, exported = set(), [], []
    for trim in run["trims"]:
        begin, end = trim["span"]
        assert raw[begin:end].decode() == trim["original"], doc["id"]
        inside = [i for i, (_, a, b) in enumerate(words) if begin <= a and b <= end]
        kept = [w for w, _, _ in words_of(trim["trimmed"].encode())]
        positions, exact = removed_positions([words[i][0] for i in inside], kept)
        if not exact:
            row["inexact"] += 1
            continue
        removed = [inside[p] for p in positions]
        pool = [i for i in inside if removable(words[i][0])]
        if not removed or not pool:
            row["empty"] += 1
            continue
        exported.append({"span": [begin, end], "words": [[words[i][1], words[i][2]] for i in inside],
                         "fixed": [n for n, i in enumerate(inside) if not removable(words[i][0])], "removed": positions})
        row["removed"] += len(removed)
        row["hits"] += sum(changed[i] for i in removed)
        row["within"] += len(removed) * sum(changed[i] for i in pool) / len(pool)
        drops.update(removed)
        plans.append((pool, min(len(removed), len(pool))))
    row["share"] = row["changed"] / row["words"] if row["words"] else 0.0
    row["document"] = row["removed"] * row["share"]
    if edited is not None and row["removed"]:
        target = [w for w, _, _ in words_of(edited.encode())]
        source = [w for w, _, _ in words]
        before = distance(source, target)
        row["moved"] = distance([w for i, w in enumerate(source) if i not in drops], target) - before
        total = 0
        for _ in range(DRAWS):
            random_drops = set()
            for pool, count in plans:
                random_drops.update(rng.sample(pool, count))
            total += distance([w for i, w in enumerate(source) if i not in random_drops], target) - before
        row["control"] = total / DRAWS
        row["measured"] = row["removed"]
        row["before"] = before
    entry = {"id": doc["id"], "sentences": run["sentences"], "proposed": run["proposed"], "rejected": len(run["rejected"]),
             "trims": sorted(exported, key=lambda item: item["span"])}
    if "before" in row:
        entry["distance"] = {"before": row["before"], "after": row["before"] + row["moved"],
                             "control": row["before"] + row["control"]}
    return row, entry


def span_document(doc, spans):
    """Per-document counts for the words under a system's finding spans."""
    words = words_of(doc["text"].encode())
    changed = changed_flags(words, doc["reference"])
    under = [i for i, (_, a, b) in enumerate(words) if any(s <= a and b <= e for s, e in spans)]
    row = collections.Counter(words=len(words), changed=sum(changed), removed=len(under), hits=sum(changed[i] for i in under))
    row["share"] = row["changed"] / row["words"] if row["words"] else 0.0
    row["document"] = row["removed"] * row["share"]
    return row


def ratio(rows, top, bottom):
    total = sum(r[bottom] for r in rows)
    return sum(r[top] for r in rows) / total if total else None


def summarize(rows, with_distance):
    stats = {"precision": ratio(rows, "hits", "removed"), "chance_sentences": ratio(rows, "within", "removed"),
             "chance_document": ratio(rows, "document", "removed"), "pooled_share": ratio(rows, "changed", "words")}
    if with_distance:
        stats["net"] = ratio(rows, "moved", "measured")
        stats["control_net"] = ratio(rows, "control", "measured")
    return stats


def differences(stats, with_distance):
    if stats["precision"] is None:
        return {}
    result = {"precision_minus_sentences": stats["precision"] - stats["chance_sentences"]
              if stats["chance_sentences"] is not None else None,
              "precision_minus_document": stats["precision"] - stats["chance_document"]}
    if with_distance and stats["net"] is not None:
        result["net"] = stats["net"]
        result["net_minus_control"] = stats["net"] - stats["control_net"]
    return {k: v for k, v in result.items() if v is not None}


def intervals(rows, with_distance, rng):
    draws = collections.defaultdict(list)
    for _ in range(BOOT):
        sample = [rows[rng.randrange(len(rows))] for _ in rows]
        for key, value in differences(summarize(sample, with_distance), with_distance).items():
            draws[key].append(value)
    result = {}
    for key, values in draws.items():
        values.sort()
        result[key] = [values[int(0.025 * len(values))], values[min(len(values) - 1, int(0.975 * len(values)))]]
    return result


def report(rows, with_distance, rng, sentences):
    stats = summarize(rows, with_distance)
    entry = {"documents": len(rows), "documents_with_output": sum(1 for r in rows if r["removed"]),
             "removed_words": sum(r["removed"] for r in rows), **stats,
             "differences": differences(stats, with_distance), "intervals": intervals(rows, with_distance, rng)}
    if sentences:
        entry.update(trims=sum(r["trims"] for r in rows), rejected=sum(r["rejected"] for r in rows),
                     proposed=sum(r["proposed"] for r in rows), offered_sentences=sum(r["sentences"] for r in rows),
                     inexact=sum(r["inexact"] for r in rows), empty=sum(r["empty"] for r in rows))
    return entry


def write_gz(path, rows):
    with gzip.GzipFile(path, "wb", mtime=0) as handle:
        for row in rows:
            handle.write((json.dumps(row, ensure_ascii=False, sort_keys=True) + "\n").encode())


def export(root, out):
    """Write the record's data files for the development and test parts."""
    docs = [d for d in read_jsonl(root / "docs.jsonl") if d["split"] in ("dev", "test")]
    edited = editor_texts(root)
    trims, counts = [], []
    for doc in sorted(docs, key=lambda d: d["id"]):
        words = words_of(doc["text"].encode())
        changed = changed_flags(words, doc["reference"])
        systems = {}
        for system in ("rules-default", "review-claude", "review-astra"):
            path = root / "runs" / system / f"{doc['id']}.json"
            if path.exists():
                row = span_document(doc, json.loads(path.read_text())["spans"])
                systems[system] = [row["removed"], row["hits"]]
        counts.append({"id": doc["id"], "words": len(words), "changed": sum(changed), "systems": systems})
        for model in MODELS:
            path = root / "runs" / f"trim-{model}" / f"{doc['id']}.json"
            if not path.exists():
                raise SystemExit(f"no trim run of {model} for {doc['id']}")
            # One generator per document and model, so the control does not depend on the order of the export.
            rng = random.Random(f"{SEED}|{doc['id']}|{model}")
            text = edited.get(doc["id"]) if doc["set"] in ("beemo", "lamp") else None
            trims.append({"model": model, **trim_document(doc, json.loads(path.read_text()), text, rng)[1]})
    (out / "data").mkdir(parents=True, exist_ok=True)
    write_gz(out / "data" / "trims.jsonl.gz", trims)
    write_gz(out / "data" / "word-counts.jsonl.gz", counts)
    print(f"exported {len(trims)} trim rows and {len(counts)} documents")


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("root")
    parser.add_argument("split")
    parser.add_argument("--out")
    parser.add_argument("--export")
    parser.add_argument("--self-test", action="store_true")
    options = parser.parse_args()
    self_test()
    if options.self_test:
        return
    root = Path(options.root).resolve()
    if options.export:
        export(root, Path(options.export).resolve())
        return
    docs = [d for d in read_jsonl(root / "docs.jsonl") if d["split"] == options.split]
    edited = editor_texts(root)
    result = {"split": options.split, "sets": {}}
    for name in SETS:
        selected = [d for d in docs if d["set"] == name]
        with_distance = name in ("beemo", "lamp")
        entry = {"trim": {}, "comparisons": {}}
        for model in MODELS:
            rng = random.Random(f"{SEED}|{name}|{model}")
            rows, missing = [], 0
            for doc in selected:
                path = root / "runs" / f"trim-{model}" / f"{doc['id']}.json"
                if not path.exists():
                    missing += 1
                    continue
                rows.append(trim_document(doc, json.loads(path.read_text()),
                                          edited.get(doc["id"]) if with_distance else None, rng)[0])
            entry["trim"][model] = {**report(rows, with_distance, rng, True), "missing_runs": missing}
        for system in ("rules-default", "review-claude", "review-astra"):
            rng = random.Random(f"{SEED}|{name}|{system}")
            rows = []
            for doc in selected:
                path = root / "runs" / system / f"{doc['id']}.json"
                if path.exists():
                    rows.append(span_document(doc, json.loads(path.read_text())["spans"]))
            entry["comparisons"][system] = report(rows, False, rng, False)
        result["sets"][name] = entry
    text = json.dumps(result, indent=1, sort_keys=True)
    if options.out:
        Path(options.out).write_text(text + "\n")
    for name, entry in result["sets"].items():
        for model, stats in entry["trim"].items():
            line = (f"{name:5} trim-{model:6} docs {stats['documents']:3} with trims {stats['documents_with_output']:3} "
                    f"trims {stats['trims']:4} removed {stats['removed_words']:5} rejected {stats['rejected']:4} "
                    f"missing {stats['missing_runs']} inexact {stats['inexact']}")
            print(line)
            for key in ("precision", "chance_sentences", "chance_document", "pooled_share", "net", "control_net"):
                if stats.get(key) is not None:
                    print(f"      {key:18} {stats[key]:+.3f}")
            for key, value in stats["differences"].items():
                low, high = stats["intervals"][key]
                print(f"      {key:26} {value:+.3f}  [{low:+.3f}, {high:+.3f}]")
        for system, stats in entry["comparisons"].items():
            if stats["precision"] is None:
                continue
            low, high = stats["intervals"]["precision_minus_document"]
            print(f"{name:5} {system:14} words {stats['removed_words']:6} precision {stats['precision']:.3f} "
                  f"chance {stats['chance_document']:.3f} diff {stats['differences']['precision_minus_document']:+.3f} "
                  f"[{low:+.3f}, {high:+.3f}]")


if __name__ == "__main__":
    main()
