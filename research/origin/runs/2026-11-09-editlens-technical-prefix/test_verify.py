"""Reject altered technical-prefix research evidence."""

from __future__ import annotations

import importlib.util
import json
import shutil
import tempfile
import unittest
from pathlib import Path


HERE = Path(__file__).resolve().parent
SPEC = importlib.util.spec_from_file_location("editlens_technical_verify", HERE / "verify.py")
assert SPEC is not None and SPEC.loader is not None
VERIFY = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(VERIFY)


class VerifyResultTest(unittest.TestCase):
    def setUp(self) -> None:
        self.directory = tempfile.TemporaryDirectory(prefix="editlens-technical-verify-")
        self.root = Path(self.directory.name)
        for name in ("PROTOCOL.md", "run.py", "freeze.json", "scores.json",
                     "pairs.json", "result.json"):
            shutil.copyfile(HERE / name, self.root / name)

    def tearDown(self) -> None:
        self.directory.cleanup()

    def test_clean_evidence_passes(self) -> None:
        VERIFY.verify(self.root)

    def test_changed_pair_fails(self) -> None:
        rows = json.loads((self.root / "pairs.json").read_text())
        rows[0]["delta"] += 0.1
        (self.root / "pairs.json").write_text(json.dumps(rows))
        with self.assertRaisesRegex(ValueError, "Pair score or length mismatch"):
            VERIFY.verify(self.root)

    def test_forged_truncation_count_fails(self) -> None:
        result = json.loads((self.root / "result.json").read_text())
        result["truncated_sources"] = 0
        (self.root / "result.json").write_text(json.dumps(result))
        with self.assertRaisesRegex(ValueError, "Incorrect truncated sources count"):
            VERIFY.verify(self.root)

    def test_forged_direction_pass_fails(self) -> None:
        result = json.loads((self.root / "result.json").read_text())
        result["operations"]["polish"]["direction_pass"] = False
        (self.root / "result.json").write_text(json.dumps(result))
        with self.assertRaisesRegex(ValueError, "Direction gate mismatch"):
            VERIFY.verify(self.root)


if __name__ == "__main__":
    unittest.main()
