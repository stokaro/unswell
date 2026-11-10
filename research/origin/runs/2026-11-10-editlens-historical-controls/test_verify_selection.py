"""Reject damaged historical source-frame selection evidence."""

from __future__ import annotations

import importlib.util
import json
from pathlib import Path
import shutil
import tempfile
import unittest
from unittest import mock


HERE = Path(__file__).resolve().parent
SPEC = importlib.util.spec_from_file_location("selection_verifier", HERE / "verify_selection.py")
assert SPEC is not None and SPEC.loader is not None
selection = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(selection)
FILES = ("inventory.json", "blob_audit.json", "eligibility.json", "frame.json")


class SelectionVerifierTests(unittest.TestCase):
    def setUp(self) -> None:
        self.temporary = tempfile.TemporaryDirectory()
        self.addCleanup(self.temporary.cleanup)
        self.root = Path(self.temporary.name)
        for name in FILES:
            shutil.copyfile(HERE / name, self.root / name)

    def verify(self):
        with mock.patch.object(selection, "HERE", self.root):
            return selection.verify()

    def change(self, name: str, mutate) -> None:
        path = self.root / name
        payload = json.loads(path.read_text())
        mutate(payload)
        path.write_text(json.dumps(payload, sort_keys=True, indent=2) + "\n")

    def test_complete_source_frame(self) -> None:
        result = self.verify()
        self.assertEqual((result["candidates"], result["selected"],
                          result["selected_windows"]), (506, 462, 3735))

    def test_missing_selected_document(self) -> None:
        self.change("frame.json", lambda payload: payload["selected"].pop())
        with self.assertRaisesRegex(ValueError, "Incomplete or duplicate source decision"):
            self.verify()

    def test_changed_blob_audit(self) -> None:
        self.change("blob_audit.json",
                    lambda payload: payload["rows"][0].update(bytes=1))
        with self.assertRaisesRegex(ValueError, "Blob audit digest changed"):
            self.verify()

    def test_forged_exclusion_reason(self) -> None:
        self.change("frame.json",
                    lambda payload: payload["excluded"][0].update(reason="after_scoring"))
        with self.assertRaisesRegex(ValueError, "Unknown exclusion reason"):
            self.verify()


if __name__ == "__main__":
    unittest.main()
