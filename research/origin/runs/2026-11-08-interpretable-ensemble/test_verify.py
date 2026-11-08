"""Negative tests for the saved origin-ensemble evidence gate."""

from __future__ import annotations

import importlib.util
import json
import shutil
import tempfile
import unittest
from pathlib import Path
from unittest import mock


HERE = Path(__file__).resolve().parent
SPEC = importlib.util.spec_from_file_location("ensemble_verify", HERE / "verify.py")
assert SPEC is not None and SPEC.loader is not None
verifier = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(verifier)


class SavedEvidenceTest(unittest.TestCase):
    def staged(self, directory: Path) -> None:
        for name in ("freeze.json", "result.json", "PROTOCOL.md", "run.py",
                     "posthoc.py", "posthoc-result.json"):
            shutil.copyfile(HERE / name, directory / name)

    def test_saved_evidence_passes(self):
        verifier.verify()

    def test_changed_protocol_is_rejected(self):
        with tempfile.TemporaryDirectory() as temporary:
            directory = Path(temporary)
            self.staged(directory)
            (directory / "PROTOCOL.md").write_text("Changed protocol\n")
            with self.assertRaisesRegex(ValueError, "protocol hash mismatch"):
                verifier.verify(directory)

    def test_changed_result_is_rejected(self):
        with tempfile.TemporaryDirectory() as temporary:
            directory = Path(temporary)
            self.staged(directory)
            (directory / "result.json").write_text("{}\n")
            with self.assertRaisesRegex(ValueError, "Result hash mismatch"):
                verifier.verify(directory)

    def test_changed_post_result_audit_is_rejected(self):
        with tempfile.TemporaryDirectory() as temporary:
            directory = Path(temporary)
            self.staged(directory)
            (directory / "posthoc-result.json").write_text("{}\n")
            with self.assertRaisesRegex(ValueError, "Post-result audit hash mismatch"):
                verifier.verify(directory)

    def test_changed_decision_is_rejected_even_with_new_result_hash(self):
        with tempfile.TemporaryDirectory() as temporary:
            directory = Path(temporary)
            self.staged(directory)
            path = directory / "result.json"
            result = json.loads(path.read_text())
            result["advance_pass"] = True
            path.write_text(json.dumps(result, indent=2, sort_keys=True) + "\n")
            with mock.patch.object(verifier, "RESULT_SHA", verifier.sha(path)):
                with self.assertRaisesRegex(ValueError, "Frozen decision mismatch"):
                    verifier.verify(directory)

    def test_changed_denominator_is_rejected_even_with_new_result_hash(self):
        with tempfile.TemporaryDirectory() as temporary:
            directory = Path(temporary)
            self.staged(directory)
            path = directory / "result.json"
            result = json.loads(path.read_text())
            result["scores"]["ensemble"]["arb"]["regimes"]["h2l"]["raw_pairs"] = 3599
            path.write_text(json.dumps(result, indent=2, sort_keys=True) + "\n")
            with mock.patch.object(verifier, "RESULT_SHA", verifier.sha(path)):
                with self.assertRaisesRegex(ValueError, "Invalid paired score accounting"):
                    verifier.verify(directory)


if __name__ == "__main__":
    unittest.main()
