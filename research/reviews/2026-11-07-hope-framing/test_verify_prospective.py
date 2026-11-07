"""Negative checks for the optional external-source verifier."""

import copy
import json
import os
from pathlib import Path
import shutil
import tempfile
import unittest

import verify_prospective as verify


@unittest.skipUnless(
    os.environ.get("UNSWELL_HOPE_SOURCES") and os.environ.get("UNSWELL_HOPE_REPORTS"),
    "pinned external source and report archives are required",
)
class ProspectiveVerifierTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.sources = Path(os.environ["UNSWELL_HOPE_SOURCES"])
        cls.reports = Path(os.environ["UNSWELL_HOPE_REPORTS"])
        cls.selection = verify.read_json(verify.HERE / "prospective-selection.json")
        cls.reference = verify.read_json(verify.HERE / "prospective-reference.json")

    def test_source_byte_change_rejected(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            for row in self.selection["sources"]:
                shutil.copyfile(self.sources / f"{row['id']}.md", root / f"{row['id']}.md")
            path = root / "q2-r6.md"
            data = bytearray(path.read_bytes())
            data[2500] ^= 1
            path.write_bytes(data)
            with self.assertRaises(AssertionError):
                verify.source_bytes(self.selection, root)

    def test_report_byte_change_rejected(self):
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / "candidate-technical.json"
            data = (self.reports / path.name).read_bytes()
            path.write_bytes(data + b" ")
            sources = verify.source_bytes(self.selection, self.sources)
            expected = verify.digest(data)
            with self.assertRaises(AssertionError):
                verify.verify_report(path, expected, sources)

    def test_changed_old_finding_rejected(self):
        main = json.loads((self.reports / "main-technical.json").read_text())
        candidate = json.loads((self.reports / "candidate-technical.json").read_text())
        candidate["findings"][0]["message"] += " altered"
        with self.assertRaises(AssertionError):
            verify.compare(main, candidate, self.reference)

    def test_promoted_control_rejected(self):
        main = json.loads((self.reports / "main-technical.json").read_text())
        candidate = json.loads((self.reports / "candidate-technical.json").read_text())
        reference = copy.deepcopy(self.reference)
        next(row for row in reference["decisions"] if row["event_id"] == "h06")[
            "decision"
        ] = "definite_actionable"
        with self.assertRaises(AssertionError):
            verify.compare(main, candidate, reference)


if __name__ == "__main__":
    unittest.main()
