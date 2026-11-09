#!/usr/bin/env python3
"""Exploratory, repository-held-out structured candidate screen for #349."""

from __future__ import annotations

import argparse
import hashlib
import json
from pathlib import Path

import sklearn
from sklearn.feature_extraction.text import TfidfVectorizer
from sklearn.linear_model import LogisticRegression
from sklearn.metrics import roc_auc_score


HERE = Path(__file__).resolve().parent
INPUTS = {
    "evaluation": ("full-separated-judgment-source-necessity-astra-medium-v1-evaluation/evaluation.json",
                   "820a25b21c5c3d13b9a3819f9560e2e4ab37ae519368131d294b6ea784fd7f51"),
    "metadata": ("recovered-admission-bank-v1/alias-metadata.json",
                 "c1df0e533b90d18458d3782b8464b0380297cf439b70a59fd3de52e8dcec6f06"),
    "cores": ("recovered-admission-bank-v1/cores.json",
              "2eff16e368786905cf2d49be3eae7589d70597aa0f0627b329f43d02b8317927"),
    "reviews": ("recovered-admission-bank-v1/reviews.json",
                "d57beb6ee3213ab16c286dd6a0a0bd7ed5adf24f1c67cd40f71d50458ab85ba2"),
    "originals": ("joint-component-sol-high-evaluation/originals.json",
                  "4f0051badc4b16dfe088d88cae69bd0dda89af2e9f62b4bed317309c4db65082"),
}
COHORTS = ("whole", "context", "confirmation")


def require(condition: bool, message: str) -> None:
    if not condition:
        raise ValueError(message)


def digest(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def read_inputs(root: Path) -> dict:
    result = {}
    for key, (name, expected) in INPUTS.items():
        path = root / name
        require(path.is_file() and digest(path) == expected, "input mismatch: " + key)
        result[key] = json.loads(path.read_text())
    return result


def feature(item: dict, stream: str) -> str:
    targets = " ".join(target.get("quote", "") for target in item["targets"])[:2000]
    fields = [
        "CATEGORY_" + item.get("category", "unknown"),
        "OPERATOR_" + item.get("operator", "unknown"),
        "STREAM_" + stream,
        "TARGET " + targets,
        "DIAGNOSTIC " + item["diagnostic"],
        "NEED " + item.get("why_revision_needed", ""),
        "BURDEN " + item.get("wording_burden", ""),
        "RETENTION " + item.get("retention_case", ""),
        "PRESERVATION " + item.get("meaning_preservation", ""),
    ]
    return "\n".join(fields)


def inventory(data: dict) -> tuple[dict, dict, dict]:
    evaluation, metadata = data["evaluation"], data["metadata"]
    cores, reviews, originals = data["cores"], data["reviews"], data["originals"]
    page_group = {}
    for group, numbers in originals["groups"].items():
        for number in numbers:
            page = f"page-{number:03d}"
            require(page not in page_group, "duplicate page")
            page_group[page] = group
    require(len(page_group) == 36 and len(set(page_group.values())) == 13,
            "repository inventory changed")
    selected = evaluation["all_selected_actual_aliases"]
    require(len(selected) == 2847 and len(set(selected)) == 2847,
            "selected alias inventory changed")
    rows = {}
    for alias in selected:
        meta = metadata[alias]
        item = cores[meta["stream"]][meta["finding_id"]]
        judgment = reviews[meta["stream"]]["judgments"][meta["finding_id"]]
        cohort = originals["paths"][meta["page"]].split("/")[0]
        require(cohort in COHORTS and judgment["status"] in
                ("accepted", "uncertain", "rejected"), "candidate status changed")
        rows[alias] = {
            "cohort": cohort,
            "group": page_group[meta["page"]],
            "accepted": judgment["status"] == "accepted",
            "feature": feature(item, meta["stream"]),
            "unresolved_advice": judgment.get("suggestion_safety", {}).get("status") == "uncertain",
        }
    counts = {cohort: sum(row["cohort"] == cohort for row in rows.values())
              for cohort in COHORTS}
    require(counts == {"whole": 2101, "context": 548, "confirmation": 198},
            "cohort inventory changed")
    witnesses = evaluation["complete_witnesses"]
    require({cohort: sum(key.startswith(cohort + "/") for key in witnesses)
            for cohort in COHORTS} == {"whole": 57, "context": 46, "confirmation": 20},
            "frozen event inventory changed")
    return rows, witnesses, page_group


def measures(kept: set[str], rows: dict, witnesses: dict) -> dict:
    result = {}
    for cohort in COHORTS:
        aliases = {a for a in kept if rows[a]["cohort"] == cohort}
        events = {key: item for key, item in witnesses.items()
                  if key.startswith(cohort + "/")}
        full = sum(any(set(group) <= aliases for group in item.get("full_groups", []))
                   for item in events.values())
        accepted = sum(rows[a]["accepted"] for a in aliases)
        result[cohort] = {
            "full": full, "events": len(events),
            "accepted": accepted, "delivered": len(aliases),
            "fraction": accepted / len(aliases) if aliases else 0.0,
            "unresolved_advice": sum(rows[a]["unresolved_advice"] for a in aliases),
        }
    return result


def heldout_scores(rows: dict) -> tuple[dict, dict]:
    whole = {alias: row for alias, row in rows.items() if row["cohort"] == "whole"}
    groups = sorted({row["group"] for row in whole.values()})
    require(len(groups) == 13, "whole repository groups changed")
    scores = {}
    fold_sizes = {}
    for excluded in groups:
        train = sorted(a for a, row in whole.items() if row["group"] != excluded)
        test = sorted(a for a, row in whole.items() if row["group"] == excluded)
        vectorizer = TfidfVectorizer(ngram_range=(1, 2), min_df=2,
                                     max_features=30000, sublinear_tf=True)
        train_matrix = vectorizer.fit_transform([whole[a]["feature"] for a in train])
        test_matrix = vectorizer.transform([whole[a]["feature"] for a in test])
        model = LogisticRegression(C=1.0, class_weight="balanced",
                                   solver="liblinear", max_iter=500, random_state=0)
        model.fit(train_matrix, [whole[a]["accepted"] for a in train])
        probabilities = model.predict_proba(test_matrix)[:, 1]
        scores.update(zip(test, map(float, probabilities), strict=True))
        fold_sizes[excluded] = {"train": len(train), "heldout": len(test)}
    require(set(scores) == set(whole), "incomplete held-out scores")
    return scores, fold_sizes


def screen(rows: dict, witnesses: dict, scores: dict) -> dict:
    unchanged = {a for a, row in rows.items() if row["cohort"] != "whole"}
    fixed = measures(unchanged | {a for a, value in scores.items() if value >= 0.5},
                     rows, witnesses)
    ranked = sorted(((value, alias) for alias, value in scores.items()), reverse=True)
    kept = set(unchanged)
    best = None
    feasible = None
    for value, alias in ranked:
        kept.add(alias)
        outcome = measures(kept, rows, witnesses)
        whole = outcome["whole"]
        if whole["fraction"] >= 0.85:
            candidate = {"threshold": value, "whole": whole}
            if best is None or (whole["full"], whole["delivered"]) > \
                    (best["whole"]["full"], best["whole"]["delivered"]):
                best = candidate
            if whole["full"] >= 46 and feasible is None:
                feasible = candidate
    return {"fixed_0_5": fixed, "optimistic_best_at_85_percent": best,
            "optimistic_first_numeric_pass": feasible,
            "optimistic_numeric_feasible": feasible is not None}


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--development-root", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    require(not args.output.exists(), "result already exists")
    data = read_inputs(args.development_root)
    rows, witnesses, _ = inventory(data)
    baseline = measures(set(rows), rows, witnesses)
    require([(baseline[c]["full"], baseline[c]["accepted"], baseline[c]["delivered"])
             for c in COHORTS] == [(48, 1627, 2101), (39, 515, 548), (17, 190, 198)],
            "baseline changed")
    scores, folds = heldout_scores(rows)
    whole = {a: row for a, row in rows.items() if row["cohort"] == "whole"}
    result = {
        "format": "unswell-349-structured-admission-exploratory-v1",
        "scope": "exposed_offline_method_screen_only",
        "input_sha256": {key: expected for key, (_, expected) in INPUTS.items()},
        "protocol_sha256": digest(HERE / "PROTOCOL.md"),
        "script_sha256": digest(Path(__file__)),
        "sklearn_version": sklearn.__version__,
        "folds": folds,
        "baseline": baseline,
        "heldout_auc": roc_auc_score([whole[a]["accepted"] for a in sorted(whole)],
                                      [scores[a] for a in sorted(whole)]),
        **screen(rows, witnesses, scores),
        "qualified": False,
        "reason": "Exposed candidates and pooled held-out threshold envelope are not new-page confirmation; advice safety and standalone generation are not established.",
    }
    args.output.write_text(json.dumps(result, indent=2, sort_keys=True) + "\n")
    print(json.dumps({key: result[key] for key in
                      ("heldout_auc", "fixed_0_5", "optimistic_best_at_85_percent",
                       "optimistic_numeric_feasible")}, sort_keys=True))


if __name__ == "__main__":
    main()
