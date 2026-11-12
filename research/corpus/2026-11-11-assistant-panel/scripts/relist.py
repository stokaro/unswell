#!/usr/bin/env python3
"""Apply the listing criterion again to every reviewed finding of a listed phrase.

usage: relist.py <workspace> <assignment json> <out relisting json> <out assignment json>

ADR 0047 lists a phrase when both reviewers accept at least 80% of at least
five reviewed findings of its word family. A test set becomes development data
once it is opened, so the share is read again over all pages opened so far.

ADR 0050 says when a listed family leaves: its share is below 80% and the upper
end of the 95% Wilson interval is below 80% too. A family at 76% of twenty
findings is one finding away from passing, and a family that leaves cannot come
back, because it produces no more findings. Without the interval the lists
would only shrink with every set.
"""
import collections, json, sys
from pathlib import Path

W = Path(sys.argv[1])
assignment = json.loads(Path(sys.argv[2]).read_text())
REPORTS = ["report-broad-dev.json", "report-broad-dev2.json", "report-final-new.json", "report-real3-r2.json",
           "report-real4-r3.json", "report-real5-r4.json"]
ROUNDS = ["rev", "rev-final", "rev-real3", "rev-real4", "rev-real5"]
RULES = {"hype.vague-praise", "filler.wordy-phrase"}
# Phrases that shipped before the corpus and were never candidates.
BUILTIN = {"due to the fact that": "filler.wordy-phrase", "in order to": "filler.wordy-phrase",
           "at this point in time": "filler.wordy-phrase"}
SUFFIXES = ("ing", "ed", "es", "s", "ly")


def wilson_upper(accepted, reviewed, z=1.96):
    """Return the upper end of the Wilson score interval for a share."""
    if not reviewed:
        return 1.0
    share = accepted / reviewed
    centre = (share + z * z / (2 * reviewed)) / (1 + z * z / reviewed)
    half = z * ((share * (1 - share) / reviewed + z * z / (4 * reviewed * reviewed)) ** 0.5) / (1 + z * z / reviewed)
    return centre + half


def family(phrase, known):
    """Return the shortest listed phrase that this one inflects or extends."""
    for other in sorted(known, key=len):
        if other != phrase and (phrase.startswith(other + " ") or any(phrase == other + s for s in SUFFIXES)
                                or any(phrase == other[:-1] + s for s in SUFFIXES if other.endswith("e"))):
            return other
    return phrase


listed = {phrase: row["rule"] for phrase, row in assignment.items() if row["rule"] in RULES}
listed.update(BUILTIN)
families = {phrase: family(phrase, listed) for phrase in listed}
findings = {}
for name in REPORTS:
    for f in json.loads((W / "engine" / name).read_text())["findings"]:
        if f["rule_id"] in RULES or f["rule_id"] == "policy.banned-phrases":
            findings[f["id"]] = f
votes = collections.defaultdict(dict)
for folder in ROUNDS:
    for reviewer in ("claude", "astra", "qwen"):
        directory = W / folder / reviewer
        for path in directory.glob("*.json") if directory.exists() else []:
            for g in json.loads(path.read_text())["findings"]:
                if g["id"] in findings and g.get("justified"):
                    votes[g["id"]][reviewer] = g["justified"]
sources = {}
table = collections.defaultdict(collections.Counter)
for identifier, vote in votes.items():
    first, second = vote.get("claude"), vote.get("astra") or vote.get("qwen")
    if not first or not second:
        continue
    f = findings[identifier]
    path = f["primary"]["path"]
    source = sources.setdefault(path, (W / "engine" / path).read_bytes())
    spans = f["primary"].get("segments") or [f["primary"]["span"]]
    phrase = " ".join(b" ".join(source[s["start"]:s["end"]] for s in spans).decode("utf-8", "replace").lower().split())
    if phrase not in listed:
        continue
    counts = table[families[phrase]]
    scope = "generated" if path.startswith("v0/") else "real"
    accepted = first == "yes" and second == "yes"
    counts["reviewed"] += 1
    counts["both_accept"] += accepted
    counts[scope + "_reviewed"] += 1
    counts[scope + "_both_accept"] += accepted
rows = []
for name in sorted(set(families.values())):
    counts = table[name]
    members = sorted(p for p, f in families.items() if f == name)
    upper = wilson_upper(counts["both_accept"], counts["reviewed"])
    below = counts["reviewed"] >= 5 and counts["both_accept"] / counts["reviewed"] < 0.8
    rows.append({"family": name, "rule": listed[name], "phrases": members, **{k: counts[k] for k in (
        "reviewed", "both_accept", "real_reviewed", "real_both_accept", "generated_reviewed", "generated_both_accept")},
        "upper_bound": round(upper, 3), "below_criterion": below,
        "decision": "drop" if below and upper < 0.8 else "keep"})
dropped = {p for row in rows if row["decision"] == "drop" for p in row["phrases"]}
for phrase in dropped & set(assignment):
    assignment[phrase] = {**assignment[phrase], "rule": "unassigned", "relisted": "dropped"}
Path(sys.argv[3]).write_text(json.dumps({"criterion": {"minimum_findings": 5, "minimum_both_accept": 0.8,
                                                      "leave_when_upper_bound_below": 0.8}, "families": rows},
                                        indent=1, sort_keys=True) + "\n")
Path(sys.argv[4]).write_text(json.dumps(assignment, indent=1, sort_keys=True) + "\n")
for row in sorted(rows, key=lambda r: (r["rule"], -r["reviewed"])):
    rate = f"{row['both_accept'] / row['reviewed']:.1%}" if row["reviewed"] else "-"
    print(f"{row['rule']:22} {row['family']:28} {row['reviewed']:>4} {rate:>7} upper {row['upper_bound']:.2f} {row['decision']}")
print("dropped phrases:", sorted(dropped))
