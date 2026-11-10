"""Negative checks for the published paired-model result."""

import json
from pathlib import Path
import shutil
import tempfile
import unittest

from verify_model_result import RUN, verify


class VerifyModelResultTests(unittest.TestCase):
    def test_published_result_reproduces(self):
        summary = verify()
        self.assertEqual(summary["calls"], 18)
        self.assertEqual(summary["diagnoses"], {"atomic": 134, "freeform": 137})
        self.assertFalse(summary["advance"])

    def test_changed_raw_response_is_rejected(self):
        with tempfile.TemporaryDirectory() as directory:
            copied = Path(directory) / "run"
            shutil.copytree(RUN, copied)
            response = copied / "page-001-freeform" / "response.json"
            value = json.loads(response.read_text(encoding="utf-8"))
            value["page_id"] = "changed"
            response.write_text(json.dumps(value), encoding="utf-8")
            with self.assertRaisesRegex(ValueError, "mismatched trace"):
                verify(copied)

    def test_changed_grade_is_rejected(self):
        with tempfile.TemporaryDirectory() as directory:
            copied = Path(directory) / "run"
            shutil.copytree(RUN, copied)
            result = copied / "result.json"
            value = json.loads(result.read_text(encoding="utf-8"))
            value["advance"] = True
            result.write_text(json.dumps(value), encoding="utf-8")
            with self.assertRaisesRegex(ValueError, "reviewed result changed"):
                verify(copied)


if __name__ == "__main__":
    unittest.main()
