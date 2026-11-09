#!/usr/bin/env python3
"""One frozen, offline source-context admission screen for #349."""

from __future__ import annotations

import argparse
import hashlib
import importlib.util
import json
import re
from pathlib import Path

import numpy
import scipy
from scipy.sparse import hstack
import sklearn
from sklearn.feature_extraction.text import TfidfVectorizer
from sklearn.linear_model import LogisticRegression
from sklearn.metrics import roc_auc_score


HERE = Path(__file__).resolve().parent
CONTROL = HERE.parent / "2026-11-09-structured-admission" / "scan-v2.py"
CONTROL_SHA256 = "20a770ce8a8e88dcf2b5ac17408bf4f3006c5179a61bbb7e510eae61c994c063"
CONTROL_RESULT_SHA256 = "df663463350602d3592fc930caa89fa0589a073d3e3e00f4b9983f61cff38765"


def require(condition: bool, message: str) -> None:
    if not condition:
        raise ValueError(message)


def digest(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def load_control():
    require(digest(CONTROL) == CONTROL_SHA256, "control script changed")
    require(digest(CONTROL.with_name("result-v2.json")) == CONTROL_RESULT_SHA256,
            "control result changed")
    spec = importlib.util.spec_from_file_location("frozen_structured_admission", CONTROL)
    require(spec is not None and spec.loader is not None, "control import failed")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def clipped(raw: bytes, limit: int) -> str:
    return raw.decode("utf-8", errors="replace")[:limit]


def source_target_feature(raw: bytes, start: int, end: int) -> str:
    line_start = raw.rfind(b"\n", 0, start) + 1
    line_end = raw.find(b"\n", end)
    if line_end < 0:
        line_end = len(raw)
    previous_end = max(0, line_start - 1)
    previous_start = raw.rfind(b"\n", 0, previous_end) + 1
    next_start = min(len(raw), line_end + 1)
    next_end = raw.find(b"\n", next_start)
    if next_end < 0:
        next_end = len(raw)
    target_line = clipped(raw[line_start:line_end], 400)
    heading_window = raw[max(0, start - 2000):start].decode("utf-8", errors="replace")
    headings = re.findall(r"(?m)^#{1,6}\s+(.+)$", heading_window)
    heading = headings[-1][:120] if headings else ""
    stripped = target_line.lstrip()
    tags = [
        "SHAPE_HEADING" if re.match(r"#{1,6}\s", stripped) else "SHAPE_NOT_HEADING",
        "SHAPE_LIST" if re.match(r"(?:[-*+] |\d+[.)] )", stripped) else "SHAPE_NOT_LIST",
        "SHAPE_TABLE" if stripped.startswith("|") else "SHAPE_NOT_TABLE",
        "SHAPE_QUOTE" if stripped.startswith(">") else "SHAPE_NOT_QUOTE",
        "SHAPE_FENCE" if stripped.startswith(("```", "~~~")) else "SHAPE_NOT_FENCE",
        "SHAPE_HTML" if stripped.startswith("<") else "SHAPE_NOT_HTML",
    ]
    return "\n".join([
        *tags,
        "HEADING " + heading,
        "PREVIOUS " + clipped(raw[previous_start:previous_end], 400),
        "TARGET_LINE " + target_line,
        "NEXT " + clipped(raw[next_start:next_end], 400),
    ])


def source_features(data: dict, aliases: list[str]) -> dict[str, str]:
    originals = data["originals"]["original"]
    sources = {}
    for page, record in originals.items():
        document = record["document"]
        require(document["page"] == page, "original page mismatch")
        raw = document["source"].encode("utf-8")
        require(hashlib.sha256(raw).hexdigest() == document["source_sha256"],
                "original source hash mismatch")
        sources[page] = raw
    require(len(sources) == 36, "source inventory changed")
    features = {}
    for alias in aliases:
        meta = data["metadata"][alias]
        item = data["cores"][meta["stream"]][meta["finding_id"]]
        raw = sources[meta["page"]]
        targets = sorted(item["targets"], key=lambda target:
                         (target["span"]["start"], target["span"]["end"]))
        require(targets, "empty target inventory")
        parts = ["TARGET_COUNT_" + str(min(len(targets), 4))]
        for index, target in enumerate(targets):
            start, end = target["span"]["start"], target["span"]["end"]
            require(0 <= start < end <= len(raw), "target span outside original source")
            require(raw[start:end].decode("utf-8") == target["quote"],
                    "target quote does not match original bytes")
            if index < 3:
                parts.append(source_target_feature(raw, start, end))
        features[alias] = "\n".join(parts)
    return features


def heldout_scores(rows: dict, contexts: dict) -> tuple[dict, dict]:
    whole = {alias: row for alias, row in rows.items() if row["cohort"] == "whole"}
    groups = sorted({row["group"] for row in whole.values()})
    require(len(groups) == 12, "whole repository groups changed")
    scores = {}
    folds = {}
    for excluded in groups:
        train = sorted(a for a, row in whole.items() if row["group"] != excluded)
        test = sorted(a for a, row in whole.items() if row["group"] == excluded)
        candidate_vectorizer = TfidfVectorizer(ngram_range=(1, 2), min_df=2,
                                               max_features=30000, sublinear_tf=True)
        context_vectorizer = TfidfVectorizer(ngram_range=(1, 2), min_df=2,
                                             max_features=30000, sublinear_tf=True)
        train_candidate = candidate_vectorizer.fit_transform(
            [whole[a]["feature"] for a in train])
        test_candidate = candidate_vectorizer.transform(
            [whole[a]["feature"] for a in test])
        train_context = context_vectorizer.fit_transform([contexts[a] for a in train])
        test_context = context_vectorizer.transform([contexts[a] for a in test])
        train_matrix = hstack([train_candidate, train_context], format="csr")
        test_matrix = hstack([test_candidate, test_context], format="csr")
        model = LogisticRegression(C=1.0, class_weight="balanced",
                                   solver="liblinear", max_iter=500, random_state=0)
        model.fit(train_matrix, [whole[a]["accepted"] for a in train])
        probabilities = model.predict_proba(test_matrix)[:, 1]
        scores.update(zip(test, map(float, probabilities), strict=True))
        folds[excluded] = {"train": len(train), "heldout": len(test)}
    require(set(scores) == set(whole), "incomplete held-out scores")
    return scores, folds


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--development-root", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    require(not args.output.exists(), "result already exists")
    control = load_control()
    freeze = json.loads((HERE / "freeze.json").read_text())
    require(freeze["status"] == "frozen_before_exploratory_scan" and
            freeze["protocol_sha256"] == digest(HERE / "PROTOCOL.md") and
            freeze["script_sha256"] == digest(Path(__file__)),
            "protocol or script changed after freeze")
    require((sklearn.__version__, numpy.__version__, scipy.__version__) ==
            ("1.7.2", "2.5.4", "1.18.1"), "dependency versions changed")
    data = control.read_inputs(args.development_root)
    rows, witnesses, _ = control.inventory(data)
    baseline = control.measures(set(rows), rows, witnesses)
    require([(baseline[c]["full"], baseline[c]["accepted"], baseline[c]["delivered"])
             for c in control.COHORTS] ==
            [(48, 1627, 2101), (39, 515, 548), (17, 190, 198)],
            "candidate bank baseline changed")
    contexts = source_features(data, list(rows))
    scores, folds = heldout_scores(rows, contexts)
    whole = {a: row for a, row in rows.items() if row["cohort"] == "whole"}
    operating = control.screen(rows, witnesses, scores)
    fixed = operating["fixed_0_5"]["whole"]
    result = {
        "format": "unswell-349-source-context-admission-exploratory-v1",
        "scope": "exposed_offline_method_screen_only",
        "input_sha256": {key: expected for key, (_, expected) in control.INPUTS.items()},
        "control_script_sha256": CONTROL_SHA256,
        "control_result_sha256": CONTROL_RESULT_SHA256,
        "protocol_sha256": digest(HERE / "PROTOCOL.md"),
        "script_sha256": digest(Path(__file__)),
        "sklearn_version": sklearn.__version__,
        "numpy_version": numpy.__version__,
        "scipy_version": scipy.__version__,
        "folds": folds,
        "baseline": baseline,
        "heldout_auc": roc_auc_score([whole[a]["accepted"] for a in sorted(whole)],
                                      [scores[a] for a in sorted(whole)]),
        **operating,
        "fixed_improves_both": fixed["full"] > 37 and
                               fixed["accepted"] / fixed["delivered"] > 1183 / 1358 and
                               fixed["unresolved_advice"] <= 4,
        "qualified": False,
        "reason": "Exposed candidate bank; no page-disjoint natural transfer, standalone generation, or independent advice safety.",
    }
    args.output.write_text(json.dumps(result, indent=2, sort_keys=True) + "\n")
    print(json.dumps({key: result[key] for key in
                      ("heldout_auc", "fixed_0_5", "optimistic_best_at_85_percent",
                       "optimistic_numeric_feasible", "fixed_improves_both")},
                     sort_keys=True))


if __name__ == "__main__":
    main()
