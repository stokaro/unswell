#!/usr/bin/env python3
"""Negative integrity checks for the frozen APE assessment."""

import gzip
import hashlib
import importlib.util
import json
from pathlib import Path
import shutil
import tempfile
import unittest


HERE = Path(__file__).resolve().parent
SPEC = importlib.util.spec_from_file_location("ape_assess", HERE / "assess.py")
ASSESS = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(ASSESS)


class AssessmentTest(unittest.TestCase):
    def test_frozen_result(self):
        result = ASSESS.assess(HERE)
        self.assertEqual(result["new_full_event_span_ceiling"], 5)
        self.assertFalse(result["admission_possible_by_span"])
        self.assertEqual(result, json.loads((HERE / "assessment.json").read_text()))

    def test_changed_candidate_artifact_rejected(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            for name in ("run.json", "scores.json.gz", "top-candidates.json"):
                shutil.copyfile(HERE / name, root / name)
            with (root / "top-candidates.json").open("a") as stream:
                stream.write(" ")
            with self.assertRaisesRegex(ValueError, "Run artifact changed"):
                ASSESS.assess(root)

    def test_changed_ranking_rejected_even_with_updated_hash(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            for name in ("run.json", "scores.json.gz", "top-candidates.json"):
                shutil.copyfile(HERE / name, root / name)
            scores = json.loads(gzip.decompress((root / "scores.json.gz").read_bytes()))
            indices = scores["selected_indices"]
            indices[0], indices[1] = indices[1], indices[0]
            (root / "scores.json.gz").write_bytes(gzip.compress(
                json.dumps(scores, separators=(",", ":"), ensure_ascii=False).encode(), mtime=0))
            run = json.loads((root / "run.json").read_text())
            run["scores_sha256"] = hashlib.sha256(
                (root / "scores.json.gz").read_bytes()).hexdigest()
            (root / "run.json").write_text(json.dumps(run))
            with self.assertRaisesRegex(ValueError, "Candidate ranking changed"):
                ASSESS.assess(root)


if __name__ == "__main__":
    unittest.main()
