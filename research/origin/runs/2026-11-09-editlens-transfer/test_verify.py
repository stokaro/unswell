"""Reject changed EditLens research evidence."""

from __future__ import annotations

import importlib.util
import json
import shutil
import tempfile
import unittest
from pathlib import Path


HERE = Path(__file__).resolve().parent
SPEC = importlib.util.spec_from_file_location("editlens_verify", HERE / "verify.py")
assert SPEC is not None and SPEC.loader is not None
VERIFY = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(VERIFY)


class VerifyResultTest(unittest.TestCase):
    def setUp(self) -> None:
        self.directory = tempfile.TemporaryDirectory(prefix="editlens-verify-")
        self.root = Path(self.directory.name)
        for name in ("PROTOCOL.md", "run.py", "freeze.json", "scores.json", "result.json"):
            shutil.copyfile(HERE / name, self.root / name)

    def tearDown(self) -> None:
        self.directory.cleanup()

    def test_clean_evidence_passes(self) -> None:
        VERIFY.verify(self.root)

    def test_changed_score_fails(self) -> None:
        rows = json.loads((self.root / "scores.json").read_text())
        target = next(row for row in rows if row["role"] == "test" and
                      row["regime"] == "h2l" and row["score"] is not None)
        target["score"] = 1 - target["score"]
        (self.root / "scores.json").write_text(json.dumps(rows))
        with self.assertRaises(ValueError):
            VERIFY.verify(self.root)

    def test_forged_pass_fails(self) -> None:
        result = json.loads((self.root / "result.json").read_text())
        result["feasibility_pass"] = False
        (self.root / "result.json").write_text(json.dumps(result))
        with self.assertRaisesRegex(ValueError, "Feasibility decision mismatch"):
            VERIFY.verify(self.root)

    def test_changed_protocol_fails(self) -> None:
        with (self.root / "PROTOCOL.md").open("a") as stream:
            stream.write("\nChanged criterion.\n")
        with self.assertRaisesRegex(ValueError, "Protocol changed"):
            VERIFY.verify(self.root)


if __name__ == "__main__":
    unittest.main()
