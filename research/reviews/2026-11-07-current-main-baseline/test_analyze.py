"""Negative check for the frozen current-main baseline record."""

import json
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest


HERE = Path(__file__).resolve().parent


class AnalyzeTest(unittest.TestCase):
    def test_rejects_changed_event_ceiling(self):
        with tempfile.TemporaryDirectory() as directory:
            changed = Path(directory) / "result.json"
            result = json.loads((HERE / "result.json").read_text())
            result["profiles"]["technical"]["by_cohort"]["whole"]["possible_full_by_span"] += 1
            changed.write_text(json.dumps(result))
            run = subprocess.run([sys.executable, "-B", str(HERE / "analyze.py"),
                                  "--run", str(HERE), "--output", str(changed), "--check"],
                                 capture_output=True, text=True, check=False)
            self.assertNotEqual(run.returncode, 0)
            self.assertIn("recorded analysis differs from replay", run.stderr)


if __name__ == "__main__":
    unittest.main()
