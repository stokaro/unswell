#!/usr/bin/env python3
"""Prove the independent verifier rejects changed research evidence."""

import argparse
import copy
import json
import tempfile
from pathlib import Path

from verify import HERE, verify


def rejected(data: Path, scores: list[dict], result: dict, directory: Path) -> bool:
    scores_file = directory / "scores.json"
    result_file = directory / "result.json"
    scores_file.write_text(json.dumps(scores) + "\n")
    result_file.write_text(json.dumps(result) + "\n")
    try:
        verify(data, scores_file, result_file)
    except ValueError:
        return True
    return False


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--data", type=Path, required=True)
    args = parser.parse_args()
    original_scores = json.loads((HERE / "scores.json").read_text())
    original_result = json.loads((HERE / "result.json").read_text())
    with tempfile.TemporaryDirectory() as name:
        directory = Path(name)
        assert verify(args.data, HERE / "scores.json", HERE / "result.json")

        scores = copy.deepcopy(original_scores)
        scores[0]["score"] = 0.999
        assert rejected(args.data, scores, original_result, directory), "Changed score passed"

        scores = copy.deepcopy(original_scores)
        scores[1]["text_sha256"] = "0" * 64
        assert rejected(args.data, scores, original_result, directory), "Changed source passed"

        scores = copy.deepcopy(original_scores)
        del scores[-1]
        assert rejected(args.data, scores, original_result, directory), "Missing row passed"

        result = copy.deepcopy(original_result)
        result["threshold_strictly_greater"] = 0.1
        assert rejected(args.data, original_scores, result, directory), "Changed threshold passed"
    print("4 negative evidence mutations rejected")


if __name__ == "__main__":
    main()
