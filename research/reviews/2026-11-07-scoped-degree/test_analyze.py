"""Negative evidence checks for the scoped-degree replay audit."""

import gzip
import hashlib
import json
from pathlib import Path
import shutil
import tempfile
import unittest

import analyze


class AuditTests(unittest.TestCase):
    def test_changed_added_location_is_rejected(self):
        with tempfile.TemporaryDirectory() as temporary:
            folder = Path(temporary)
            for name in ("technical.json.gz", "strict.json.gz", "runs.json"):
                shutil.copyfile(analyze.HERE / name, folder / name)
            report = analyze.read_json(folder / "technical.json.gz")
            addition = next(finding for finding in report["findings"]
                            if finding["primary"]["path"] == "context/c02.md"
                            and finding["primary"]["span"]["start"] == 1002)
            addition["primary"]["span"]["start"] += 1
            data = gzip.compress((json.dumps(report, separators=(",", ":")) + "\n").encode(), mtime=0)
            (folder / "technical.json.gz").write_bytes(data)
            ledger = analyze.read_json(folder / "runs.json")
            ledger["technical"]["report_sha256"] = hashlib.sha256(data).hexdigest()
            (folder / "runs.json").write_text(json.dumps(ledger), encoding="utf-8")
            with self.assertRaisesRegex(ValueError, "Unexpected added locations"):
                analyze.analyze(folder)


if __name__ == "__main__":
    unittest.main()
