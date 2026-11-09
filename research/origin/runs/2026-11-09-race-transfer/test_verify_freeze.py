"""Negative control for the frozen transfer packet verifier."""

from __future__ import annotations

import importlib.util
from pathlib import Path
import tempfile
import unittest


SCRIPT = Path(__file__).with_name("verify_freeze.py")
SPEC = importlib.util.spec_from_file_location("race_freeze", SCRIPT)
assert SPEC is not None and SPEC.loader is not None
FREEZE = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(FREEZE)


class FreezeTests(unittest.TestCase):
    def test_rejects_changed_file(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            path = Path(temporary) / "input.jsonl"
            path.write_bytes(b"original")
            expected = FREEZE.digest(path)
            path.write_bytes(b"changed")
            with self.assertRaisesRegex(ValueError, "Frozen file changed"):
                FREEZE.verify_file(path, expected)


if __name__ == "__main__":
    unittest.main()
