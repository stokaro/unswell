"""Negative contract tests for source-bound model responses and packet identity."""

import json
from pathlib import Path
import tempfile
import unittest
from unittest import mock

import run


class PilotRunTest(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.rows = run.source_rows()

    def valid_generation(self):
        row = self.rows[0]
        item = run.read(run.ROOT / "reference.json")["events"][0]
        return row, {"page_id": row["page_id"],
                     "source_sha256": row["source_sha256"],
                     "findings": [{"unit_ordinal": item["unit_ordinal"],
                                   "quote": item["quote"],
                                   "category": item["category"],
                                   "diagnosis": "The sentence repeats a sequence without adding a step.",
                                   "basis": "The configuration paragraph already establishes the file."}]}

    def test_valid_source_binding(self):
        row, response = self.valid_generation()
        bound = run.bound_generation(response, row)
        self.assertEqual(len(bound["findings"]), 1)
        self.assertEqual(bound["findings"][0]["span"],
                         run.read(run.ROOT / "reference.json")["events"][0]["span"])

    def test_protected_markdown_boundary_rejected(self):
        row = next(p for p in self.rows if p["page_id"] == "eza.md")
        response = {"page_id": row["page_id"], "source_sha256": row["source_sha256"],
                    "findings": [{"unit_ordinal": 5, "quote": "small**, **fast",
                                  "category": "wordiness", "diagnosis": "Test diagnosis",
                                  "basis": "Test source context"}]}
        with self.assertRaisesRegex(ValueError, "protected span"):
            run.bound_generation(response, row)

    def test_source_identity_rejected(self):
        row, response = self.valid_generation()
        response["source_sha256"] = "0" * 64
        with self.assertRaisesRegex(ValueError, "source binding"):
            run.bound_generation(response, row)

    def test_admission_identity_rejected(self):
        row, response = self.valid_generation()
        candidate = run.bound_generation(response, row)["findings"][0]
        admission = {"candidate_id": "wrong", "page_id": candidate["page_id"],
                     "source_sha256": candidate["source_sha256"],
                     "decision": "admit", "basis": "Source-specific case"}
        with self.assertRaisesRegex(ValueError, "candidate_id"):
            run.bound_admission(admission, candidate)

    def test_packet_mutation_rejected(self):
        with tempfile.TemporaryDirectory() as directory:
            packet = Path(directory) / "packet"
            with mock.patch.object(run, "PACKET", packet):
                run.prepare()
                self.assertEqual(run.verify()["pages"], 4)
                path = packet / "requests" / "nushell.md.txt"
                path.write_bytes(path.read_bytes() + b"altered")
                with self.assertRaisesRegex(ValueError, "packet freeze"):
                    run.verify()


if __name__ == "__main__":
    unittest.main()
