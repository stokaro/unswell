"""Negative controls for ARB frozen research evidence."""

from __future__ import annotations

import contextlib
import importlib.util
import io
import json
import shutil
import tempfile
import unittest
from pathlib import Path


HERE = Path(__file__).resolve().parent
SPEC = importlib.util.spec_from_file_location("arb_verify", HERE / "verify.py")
if SPEC is None or SPEC.loader is None:
    raise RuntimeError("ARB verifier unavailable")
VERIFY = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(VERIFY)


class EvidenceTest(unittest.TestCase):
    def test_saved_evidence_passes(self):
        with contextlib.redirect_stdout(io.StringIO()):
            VERIFY.verify()

    def test_changed_protocol_fails(self):
        with tempfile.TemporaryDirectory() as temporary:
            copy = Path(temporary) / "arb"
            shutil.copytree(HERE, copy, ignore=shutil.ignore_patterns("__pycache__"))
            path = copy / "PROTOCOL.md"
            path.write_text(path.read_text() + "\nChanged after freeze.\n")
            with self.assertRaisesRegex(ValueError, "Protocol hash mismatch"):
                VERIFY.verify(copy)

    def test_changed_result_fails(self):
        with tempfile.TemporaryDirectory() as temporary:
            copy = Path(temporary) / "arb"
            shutil.copytree(HERE, copy, ignore=shutil.ignore_patterns("__pycache__"))
            path = copy / "result.json"
            result = json.loads(path.read_text())
            result["richness_h2l_feasibility_pass"] = True
            path.write_text(json.dumps(result))
            with self.assertRaisesRegex(ValueError, "Result hash mismatch"):
                VERIFY.verify(copy)


if __name__ == "__main__":
    unittest.main()
