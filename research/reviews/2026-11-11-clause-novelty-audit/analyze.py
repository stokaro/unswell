"""Exposed-only, fixed lexical clause-relation diagnostic for issue 349.

No fitting, threshold selection, or product decision is performed here.
"""
import collections
import gzip
import hashlib
import json
import re
from pathlib import Path

BASE = Path(__file__).resolve().parent
REPORT = BASE.parent / "2026-11-07-current-main-baseline/technical.json.gz"
REVIEW = BASE.parent / "2026-11-07-current-main-editorial-review/findings.json"
REPORT_SHA256 = "d45eaaaaf601cf66ee54ca649937c463f650b192d0a13f9de6807b4831e23aa1"
REVIEW_SHA256 = "7f1575602b2a7838e99fcd451d31b79654b331df3120e32c1e6d9e5154447f08"
SPLIT = re.compile(r"[,;:]|\b(?:but|while|whereas|because|and|or|so)\b", re.I)
WORDS = re.compile(r"[A-Za-z][A-Za-z0-9_]*")
STOP = frozenset("a an and are as at be because been but by can for from if in is it may of on or so that the their them these this to was were when where while whereas which with would you".split())


def content(segment):
    return {w.lower() for w in WORDS.findall(segment) if len(w) >= 3 and w.lower() not in STOP}


def novelty(snippet):
    parts = [content(part) for part in SPLIT.split(snippet)]
    parts = [part for part in parts if len(part) >= 3]
    if len(parts) < 2:
        return None
    overlaps = [len(a & b) / len(a | b) for a, b in zip(parts, parts[1:])]
    return 1 - sum(overlaps) / len(overlaps)


def auc(rows, score):
    positive = [score(x) for x in rows if x["status"] == "actionable"]
    negative = [score(x) for x in rows if x["status"] != "actionable"]
    if not positive or not negative:
        return None
    return sum((p > n) + 0.5 * (p == n) for p in positive for n in negative) / (len(positive) * len(negative))


def main():
    if hashlib.sha256(REPORT.read_bytes()).hexdigest() != REPORT_SHA256:
        raise ValueError("The frozen technical report changed")
    if hashlib.sha256(REVIEW.read_bytes()).hexdigest() != REVIEW_SHA256:
        raise ValueError("The frozen editorial review changed")
    review = {x["id"]: x for x in json.loads(REVIEW.read_text())}
    report = json.loads(gzip.decompress(REPORT.read_bytes()))
    rows = []
    for finding in report["findings"]:
        if finding["rule_id"] != "syntax.long-sentence" or not finding["primary"]["path"].startswith("whole/"):
            continue
        decision = review[finding["id"]]
        if "technical" not in decision["profiles"]:
            continue
        metrics = finding["evidence"]["metrics"]
        words = next(m["value"] for m in metrics if m["name"] == "length")
        rows.append({"path": decision["path"], "status": decision["status"], "novelty": novelty(finding["primary"]["snippet"]), "words": words})
    eligible = [x for x in rows if x["novelty"] is not None]
    if len(rows) != 84 or len(eligible) != 84:
        raise ValueError("The frozen whole-page finding inventory changed")
    groups = collections.defaultdict(list)
    for row in eligible:
        groups[row["path"]].append(row)
    grouped = {path: {"count": len(group), "auc": round(auc(group, lambda x: x["novelty"]), 4)} for path, group in groups.items() if auc(group, lambda x: x["novelty"]) is not None}
    out = {
        "format": "unswell-clause-novelty-audit-v1",
        "report_base_commit": "5dd76191740c3d2f003b6ec90a9b381ccf1b772a",
        "exploratory_only": True,
        "report_sha256": hashlib.sha256(REPORT.read_bytes()).hexdigest(),
        "review_sha256": hashlib.sha256(REVIEW.read_bytes()).hexdigest(),
        "total": len(rows),
        "eligible": len(eligible),
        "eligible_actionable": sum(x["status"] == "actionable" for x in eligible),
        "eligible_nonaccepted": sum(x["status"] != "actionable" for x in eligible),
        "novelty_auc": round(auc(eligible, lambda x: x["novelty"]), 4),
        "word_count_auc_same_rows": round(auc(eligible, lambda x: x["words"]), 4),
        "groups_with_both_labels": grouped,
    }
    print(json.dumps(out, indent=2, sort_keys=True))


if __name__ == "__main__":
    main()
