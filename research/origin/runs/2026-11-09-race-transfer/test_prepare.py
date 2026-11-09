"""Integrity tests for the source-matched RACE input selector."""

from __future__ import annotations

import importlib.util
from pathlib import Path
import tempfile
import unittest


SCRIPT = Path(__file__).with_name("prepare.py")
SPEC = importlib.util.spec_from_file_location("race_prepare", SCRIPT)
assert SPEC is not None and SPEC.loader is not None
MODULE = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(MODULE)
PARSE_SCRIPT = Path(__file__).with_name("parse.py")
PARSE_SPEC = importlib.util.spec_from_file_location("race_parse", PARSE_SCRIPT)
assert PARSE_SPEC is not None and PARSE_SPEC.loader is not None
PARSE = importlib.util.module_from_spec(PARSE_SPEC)
PARSE_SPEC.loader.exec_module(PARSE)
INFER_SCRIPT = Path(__file__).with_name("infer.py")
INFER_SPEC = importlib.util.spec_from_file_location("race_infer", INFER_SCRIPT)
assert INFER_SPEC is not None and INFER_SPEC.loader is not None
INFER = importlib.util.module_from_spec(INFER_SPEC)
INFER_SPEC.loader.exec_module(INFER)


class PrepareTests(unittest.TestCase):
    def test_rejects_wrong_source_hash(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            source = Path(temporary) / "arb.parquet"
            source.write_bytes(b"tampered")
            with self.assertRaisesRegex(ValueError, "ARB byte count changed"):
                MODULE.verify_arb(source)

    def test_selects_test_sources_only(self) -> None:
        humans = {}
        roles = {}
        for domain in MODULE.DOMAINS:
            for index in range(301):
                pair = f"{domain}-{index:03}"
                humans[pair] = {"source_dataset": domain}
                roles[pair] = "test" if index < 300 else "calibration"
        selected = MODULE.select_pairs(humans, roles)
        self.assertEqual(len(selected), 60)
        self.assertTrue(all(roles[pair] == "test" for pair in selected))
        self.assertEqual({pair.split("-")[0] for pair in selected},
                         set(MODULE.DOMAINS))

    def test_parser_rejects_changed_input_hash(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            path = Path(temporary) / "input.jsonl"
            path.write_text("{}\n", encoding="utf-8")
            with self.assertRaisesRegex(ValueError, "SHA-256 mismatch"):
                PARSE.require_hash(path, PARSE.INPUT_SHA256)

    def test_inference_rejects_changed_checkpoint_hash(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            path = Path(temporary) / "best_model.pth"
            path.write_bytes(b"changed")
            with self.assertRaisesRegex(ValueError, "SHA-256 mismatch"):
                INFER.require_hash(path, INFER.CHECKPOINT_SHA256)


if __name__ == "__main__":
    unittest.main()
