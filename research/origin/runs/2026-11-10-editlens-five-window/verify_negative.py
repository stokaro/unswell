#!/usr/bin/env python3
"""Check that the independent verifier rejects corrupted evidence."""

import argparse
import copy
import importlib.util
import json
from pathlib import Path
import tempfile


HERE = Path(__file__).resolve().parent


def verifier():
    spec = importlib.util.spec_from_file_location("five_window_verifier", HERE / "verify.py")
    if spec is None or spec.loader is None:
        raise ValueError("Verifier unavailable")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--data", type=Path, required=True)
    args = parser.parse_args()
    check = verifier()
    scores = json.loads((HERE / "scores.json").read_text())
    result = json.loads((HERE / "result.json").read_text())
    check.verify(args.data, HERE / "scores.json", HERE / "result.json")
    cases = []
    wrong_score = copy.deepcopy(scores)
    cal = check.paper_means(scores, "calibration", "human")
    threshold_paper = sorted(cal, key=lambda paper: cal[paper], reverse=True)[2]
    scored_index = max((index for index, row in enumerate(wrong_score)
                        if row["paper_id"] == threshold_paper),
                       key=lambda index: wrong_score[index]["score"])
    wrong_score[scored_index]["score"] = 0.0
    cases.append(("score", wrong_score, result))
    wrong_identity = copy.deepcopy(scores)
    wrong_identity[0]["text_sha256"] = "0" * 64
    cases.append(("identity", wrong_identity, result))
    cases.append(("missing row", scores[:-1], result))
    wrong_threshold = copy.deepcopy(result)
    wrong_threshold["threshold_strictly_greater"] = 0.0
    cases.append(("threshold", scores, wrong_threshold))
    with tempfile.TemporaryDirectory() as directory:
        root = Path(directory)
        for label, altered_scores, altered_result in cases:
            scores_path, result_path = root / "scores.json", root / "result.json"
            scores_path.write_text(json.dumps(altered_scores))
            result_path.write_text(json.dumps(altered_result))
            try:
                check.verify(args.data, scores_path, result_path)
            except ValueError:
                continue
            raise AssertionError(f"Corruption accepted: {label}")
    print(f"{len(cases)} negative evidence mutations rejected")


if __name__ == "__main__":
    main()
