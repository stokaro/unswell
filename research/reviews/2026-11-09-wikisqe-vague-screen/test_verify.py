"""Negative checks for the saved WikiSQE transfer assessment."""

import gzip
import json
from pathlib import Path
import tempfile
import unittest

import verify


HERE = Path(__file__).resolve().parent


class VerifyScreenTest(unittest.TestCase):
    def copied_run(self, directory):
        target = Path(directory)
        for name in ("scores.json.gz", "top-candidates.json", "result.json"):
            (target / name).write_bytes((HERE / name).read_bytes())
        return target

    def test_rejects_changed_candidate_ranking(self):
        with tempfile.TemporaryDirectory() as directory:
            target = self.copied_run(directory)
            scored = json.loads(gzip.decompress((target / "scores.json.gz").read_bytes()))
            scored["selected_indices"][1] = scored["selected_indices"][0]
            with (target / "scores.json.gz").open("wb") as file:
                with gzip.GzipFile(fileobj=file, mode="wb", mtime=0) as compressed:
                    compressed.write(json.dumps(scored, separators=(",", ":")).encode())
            result = json.loads((target / "result.json").read_text())
            result["scores_sha256"] = verify.digest(target / "scores.json.gz")
            (target / "result.json").write_text(json.dumps(result))
            with self.assertRaisesRegex(ValueError, "Invalid selected indices"):
                verify.verify(target)

    def test_rejects_inflated_event_ceiling(self):
        with tempfile.TemporaryDirectory() as directory:
            target = self.copied_run(directory)
            result = json.loads((target / "result.json").read_text())
            result["new_full_event_span_ceiling"] += 1
            (target / "result.json").write_text(json.dumps(result))
            with self.assertRaisesRegex(ValueError, "Frozen-event assessment changed"):
                verify.verify(target)


if __name__ == "__main__":
    unittest.main()
