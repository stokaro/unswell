"""Contract tests for the offline origin feature experiment."""

from __future__ import annotations

import importlib.util
import tempfile
import unittest
from pathlib import Path


HERE = Path(__file__).resolve().parent
SPEC = importlib.util.spec_from_file_location("ensemble_run", HERE / "run.py")
assert SPEC is not None and SPEC.loader is not None
run = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(run)


class EnsembleContractTest(unittest.TestCase):
    def test_split_preserves_three_source_domains(self):
        humans = {f"{domain}-{index}": {"source_dataset": domain}
                  for domain in run.DOMAINS for index in range(600)}
        roles, digest = run.split_arb(humans)
        self.assertEqual(len(digest), 64)
        for domain in run.DOMAINS:
            values = [roles[f"{domain}-{index}"] for index in range(600)]
            self.assertEqual(values.count("train"), 200)
            self.assertEqual(values.count("calibration"), 100)
            self.assertEqual(values.count("test"), 300)

    def test_repeated_bigram_counts_every_occurrence(self):
        module = run.load_module(run.RICHNESS, "test_richness")
        words = ["alpha", "beta"] * 50
        vector = run.feature_vector(" ".join(words), module, set())
        self.assertIsNotNone(vector)
        self.assertEqual(vector[-1], 1.0)
        self.assertEqual(vector[2], 0.0)
        self.assertEqual(vector[3], 0.45)

    def test_short_document_abstains(self):
        module = run.load_module(run.RICHNESS, "test_short_richness")
        self.assertIsNone(run.feature_vector("word " * 99, module, set()))

    def test_threshold_uses_strict_comparison_at_ties(self):
        rows = [{"ensemble": 1.0} for _ in range(300)]
        result = run.threshold(rows, "ensemble")
        self.assertEqual(result["flagged"], 0)
        self.assertEqual(result["score"], 1.0)

    def test_duplicate_academic_variant_is_rejected(self):
        with tempfile.TemporaryDirectory() as temporary:
            path = Path(temporary) / "docs.jsonl"
            line = ('{"paper_id":"p1","split":"test","variant":"human",'
                    '"text":"Some text","generator":null,"prompt":null}\n')
            path.write_text(line + line)
            with self.assertRaisesRegex(ValueError, "Duplicate academic variant"):
                run.load_academic(path)


if __name__ == "__main__":
    unittest.main()
