"""Negative evidence probes for the frozen control-output verifier."""

import json
import shutil
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path


HERE = Path(__file__).resolve().parent
FILES = ("controls.json", "freeze-v2.json", "run-v2.py", "result-v2.json",
         "reports-v2.json.gz", "verify.py")


class VerifyTest(unittest.TestCase):
    def test_rejects_changed_source_quote(self):
        with tempfile.TemporaryDirectory() as directory:
            target = Path(directory)
            for name in FILES:
                shutil.copyfile(HERE / name, target / name)
            result_path = target / "result-v2.json"
            result = json.loads(result_path.read_text())
            result["cases"][3]["findings"][0]["quote"] = "wrong quote"
            result_path.write_text(json.dumps(result))
            run = subprocess.run([sys.executable, "-B", str(target / "verify.py")],
                                 capture_output=True, text=True, check=False)
            self.assertNotEqual(run.returncode, 0)
            self.assertIn("finding quote differs from original bytes", run.stderr)


if __name__ == "__main__":
    unittest.main()
