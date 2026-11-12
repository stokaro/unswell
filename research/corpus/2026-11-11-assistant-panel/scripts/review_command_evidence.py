#!/usr/bin/env python3
"""Evidence for the review command on generated documents.

usage: review_command_evidence.py <workspace> <rules report> <out json>

Reads the runs under product/: <reviewer>-none holds every bound flag of the
reviewer; <reviewer>-<confirmer> holds the same flags with the confirming
model's verdicts. The reference marks for a run come from the annotator that
is neither the reviewer nor, in the independent runs, the confirming model.
Items are matched by byte overlap. The decision registered on issue 525 uses
the two independent runs pooled.
"""
import collections, json, sys
from pathlib import Path

W, RULES, OUT = Path(sys.argv[1]), Path(sys.argv[2]), Path(sys.argv[3])
RUNS = [("claude", "qwen", "astra", True), ("astra", "qwen", "claude", True),
        ("claude", "astra", "astra", False), ("astra", "claude", "claude", False)]
MODELS = {"claude": "claude-sonnet-5-5", "astra": "gpt-6-astra", "qwen": "private-host/Qwen3.8-27B"}
docs = {}
for line in open(W / "docs.jsonl"):
    doc = json.loads(line)
    docs[doc["doc_id"]] = doc
rule_spans = collections.defaultdict(list)
for finding in json.loads(RULES.read_text())["findings"]:
    span = finding["primary"]["span"]
    rule_spans[Path(finding["primary"]["path"]).stem].append((span["start"], span["end"]))


def marked(doc_id, annotator):
    flags = json.loads((W / "ann" / annotator / f"{doc_id}.json").read_text())["flags"]
    return {f["item"] for f in flags if f["label"] == "needs_revision"}


def overlaps(span, others):
    return any(span[0] < end and start < span[1] for start, end in others)


def measure(reviewer, confirmer, reference):
    counter, verdicts = collections.Counter(), collections.Counter()
    for path in sorted((W / "product" / f"{reviewer}-{confirmer}").glob("*.json")):
        doc = docs[path.stem]
        result = json.loads(path.read_text())["documents"][0]
        first = json.loads((W / "product" / f"{reviewer}-none" / path.name).read_text())["documents"][0]
        kept = [(f["span"]["start"], f["span"]["end"]) for f in result["findings"]]
        dropped = [(f["span"]["start"], f["span"]["end"]) for f in result["unconfirmed"]]
        if sorted(kept + dropped) != sorted((f["span"]["start"], f["span"]["end"]) for f in first["findings"]):
            raise SystemExit(f"{path}: the confirmed run does not hold the reviewer's flags")
        for finding in result["findings"] + result["unconfirmed"]:
            verdicts[(finding.get("confirmation") or {}).get("justified") or "skipped"] += 1
        items = marked(doc["doc_id"], reference)
        segments = [[tuple(s) for s in item["segments"]] for item in doc["items"] if item["id"] in items]
        flat = [segment for parts in segments for segment in parts]
        rules = rule_spans[doc["doc_id"]]
        counter["documents"] += 1
        counter["words"] += len(doc["text"].split())
        counter["flags"] += result["flags"]
        counter["unbound"] += len(result["unbound"])
        counter["bound"] += len(kept) + len(dropped)
        counter["kept"] += len(kept)
        counter["dropped"] += len(dropped)
        counter["marked"] += len(segments)
        counter["marked_touched_before"] += sum(any(overlaps(s, parts) for s in kept + dropped) for parts in segments)
        counter["marked_touched_after"] += sum(any(overlaps(s, parts) for s in kept) for parts in segments)
        counter["marked_touched_by_rules"] += sum(any(overlaps(s, parts) for s in rules) for parts in segments)
        counter["bound_on_marked"] += sum(overlaps(s, flat) for s in kept + dropped)
        counter["kept_on_marked"] += sum(overlaps(s, flat) for s in kept)
        counter["dropped_on_marked"] += sum(overlaps(s, flat) for s in dropped)
        counter["rule_findings"] += len(rules)
        counter["rule_findings_on_marked"] += sum(overlaps(s, flat) for s in rules)
    return counter, verdicts


def ratio(part, whole):
    return part / whole if whole else None


def rounded(values):
    return {key: round(value, 4) if isinstance(value, float) else value for key, value in values.items()}


def derive(c):
    return {"coverage_before": ratio(c["marked_touched_before"], c["marked"]),
            "coverage_after": ratio(c["marked_touched_after"], c["marked"]),
            "coverage_by_rules": ratio(c["marked_touched_by_rules"], c["marked"]),
            "coverage_kept_share": ratio(c["marked_touched_after"], c["marked_touched_before"]),
            "share_bound_on_marked": ratio(c["bound_on_marked"], c["bound"]),
            "share_kept_on_marked": ratio(c["kept_on_marked"], c["kept"]),
            "share_dropped_on_marked": ratio(c["dropped_on_marked"], c["dropped"]),
            "share_rule_findings_on_marked": ratio(c["rule_findings_on_marked"], c["rule_findings"]),
            "bound_per_1000_words": round(1000 * c["bound"] / c["words"], 2),
            "kept_per_1000_words": round(1000 * c["kept"] / c["words"], 2),
            "rule_findings_per_1000_words": round(1000 * c["rule_findings"] / c["words"], 2)}


result = {"set": "long-prose", "selection": "the first 60 generated documents by SHA-256 of the document identifier",
          "models": MODELS, "runs": []}
pooled = collections.Counter()
for reviewer, confirmer, reference, independent in RUNS:
    if not (W / "product" / f"{reviewer}-{confirmer}").exists():
        continue
    counter, verdicts = measure(reviewer, confirmer, reference)
    if independent:
        pooled.update(counter)
    row = {"reviewer": reviewer, "confirmer": confirmer, "reference": reference, "independent_reference": independent,
           **dict(counter), "verdicts": dict(verdicts), **rounded(derive(counter))}
    result["runs"].append(row)
    d = derive(counter)
    print(f"{reviewer} -> {confirmer} (reference {reference}{'' if independent else ', not independent'}): "
          f"documents {counter['documents']}, bound {counter['bound']}, kept {counter['kept']}")
    print(f"   coverage of {counter['marked']} marked items: before {d['coverage_before']:.1%}, after {d['coverage_after']:.1%}, "
          f"rules {d['coverage_by_rules']:.1%}; kept share of coverage {d['coverage_kept_share']:.1%}")
    print(f"   on a marked item: all bound {d['share_bound_on_marked']:.1%}, kept {d['share_kept_on_marked']:.1%}, "
          f"dropped {d['share_dropped_on_marked']:.1%}; rules {d['share_rule_findings_on_marked']:.1%}")
    print(f"   per 1,000 words: bound {d['bound_per_1000_words']}, kept {d['kept_per_1000_words']}, rules {d['rule_findings_per_1000_words']}")
    print(f"   verdicts {dict(verdicts)}")
if pooled:
    d = derive(pooled)
    separation = d["share_kept_on_marked"] - d["share_dropped_on_marked"]
    separates = separation >= 0.10
    keeps = d["coverage_kept_share"] >= 2 / 3
    decision = ("show every bound flag and mark the confirmed ones" if not separates else
                "show confirmed flags only" if keeps else "show both groups, confirmed first")
    result["pooled_independent"] = {**dict(pooled), **rounded(d), "separation": round(separation, 4),
                                    "separates": separates, "keeps_two_thirds_of_coverage": keeps, "decision": decision}
    print(f"pooled independent runs: coverage before {d['coverage_before']:.1%}, after {d['coverage_after']:.1%} "
          f"(kept share {d['coverage_kept_share']:.1%}), rules {d['coverage_by_rules']:.1%}")
    print(f"   kept on a marked item {d['share_kept_on_marked']:.1%}, dropped {d['share_dropped_on_marked']:.1%}, "
          f"separation {separation:.1%} -> {decision}")
OUT.write_text(json.dumps(result, indent=1, sort_keys=True) + "\n")
