"""Negative checks for the archived hope-framing evidence contract."""

import gzip
import hashlib
import json
from pathlib import Path
import shutil
import sys
import tempfile
import unittest


sys.path.insert(0, str(Path(__file__).resolve().parent))
import analyze  # noqa: E402


class HopeFramingEvidenceTests(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.folder = Path(self.temp.name)
        for name in ("runs.json", "review.json", "technical.json.gz", "strict.json.gz"):
            shutil.copy2(analyze.HERE / name, self.folder / name)

    def write_report(self, profile, report):
        payload = gzip.compress((json.dumps(report, separators=(",", ":")) + "\n").encode(), mtime=0)
        (self.folder / f"{profile}.json.gz").write_bytes(payload)
        runs = analyze.read(self.folder / "runs.json")
        runs[profile]["report_sha256"] = hashlib.sha256(payload).hexdigest()
        (self.folder / "runs.json").write_text(json.dumps(runs))

    def test_expected_finding_cannot_disappear(self):
        report = analyze.read(self.folder / "technical.json.gz")
        report["findings"] = [f for f in report["findings"]
                              if analyze.location(f) != ("whole/p017.md", 15250, 15329)]
        self.write_report("technical", report)
        with self.assertRaisesRegex(ValueError, "Added locations differ"):
            analyze.analyze(self.folder)

    def test_existing_finding_cannot_change_silently(self):
        report = analyze.read(self.folder / "technical.json.gz")
        report["findings"][0]["message"] = "changed advice"
        self.write_report("technical", report)
        with self.assertRaisesRegex(ValueError, "Previous finding content changed"):
            analyze.analyze(self.folder)

    def test_source_hash_cannot_drift(self):
        report = analyze.read(self.folder / "technical.json.gz")
        report["documents"][0]["source"] += "\n"
        self.write_report("technical", report)
        with self.assertRaisesRegex(ValueError, "Source hash drift"):
            analyze.analyze(self.folder)

    def test_review_cannot_promote_unaccepted_warning(self):
        review = analyze.read(self.folder / "review.json")
        review["decisions"][0]["status"] = "uncertain"
        (self.folder / "review.json").write_text(json.dumps(review))
        with self.assertRaisesRegex(ValueError, "Invalid editorial decision"):
            analyze.analyze(self.folder)


if __name__ == "__main__":
    unittest.main()
