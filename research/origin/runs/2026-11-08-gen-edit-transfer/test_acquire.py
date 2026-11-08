"""Negative and identity tests for the GEN acquisition boundary."""

from __future__ import annotations

import gzip
import hashlib
import importlib.util
import io
import json
import tempfile
import unittest
from pathlib import Path

MODULE = Path(__file__).with_name("acquire.py")
spec = importlib.util.spec_from_file_location("gen_acquire", MODULE)
assert spec and spec.loader
acquire = importlib.util.module_from_spec(spec)
spec.loader.exec_module(acquire)


class AcquisitionTests(unittest.TestCase):
    def setUp(self) -> None:
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name)
        self.human = self.root / "human.jsonl"
        with self.human.open("w") as stream:
            for domain_index, domain in enumerate(acquire.DOMAINS):
                for local_id in range(1000):
                    row = {"source": domain, "prompt_id": domain_index * 1000 + local_id,
                           "label": 0, "text_type": "human",
                           "text": f"Original source {domain_index} {local_id}."}
                    stream.write(json.dumps(row) + "\n")
        self.human_sha = hashlib.sha256(self.human.read_bytes()).hexdigest()
        self.base = {"source": acquire.DOMAINS[0], "prompt_id": 0, "seed_id": 0,
                     "model": "test-model", "edit_category": "clarity",
                     "edit_prompt": "Clarify this text.", "target_ratio": 30,
                     "edit_ratio": 42.0, "label": 2, "text_type": "ai_edited",
                     "text": "Revised source.", "seed_text": "Original source 0 0."}

    def run_rows(self, rows: list[dict], size_delta: int = 0) -> dict:
        data = b"".join(json.dumps(row).encode() + b"\n" for row in rows)
        return acquire.acquire(io.BytesIO(data), self.human,
                               self.root / "selected.jsonl.gz",
                               self.root / "manifest.json",
                               expected_bytes=len(data) + size_delta,
                               expected_human_sha=self.human_sha)

    def test_selection_split_and_source_identity(self) -> None:
        other_ratio = dict(self.base, target_ratio=50)
        result = self.run_rows([self.base, other_ratio])
        self.assertEqual(result["input_rows"], 2)
        self.assertEqual(result["selected_rows"], 1)
        self.assertEqual(result["source_mismatches_by_domain"], {})
        with gzip.open(self.root / "selected.jsonl.gz", "rt") as stream:
            self.assertEqual([json.loads(line)["target_ratio"] for line in stream], [30])
        _, roles, digest = acquire.human_groups(self.human, self.human_sha)
        self.assertEqual(len(digest), 64)
        for domain in acquire.DOMAINS:
            self.assertEqual(sum(role == "calibration" for (source, _), role in roles.items()
                                 if source == domain), 500)

    def test_changed_human_input_is_rejected(self) -> None:
        self.human.write_bytes(self.human.read_bytes().replace(b"Original source 0 0.",
                                                       b"Altered source 0 0.", 1))
        with self.assertRaisesRegex(ValueError, "Human input SHA-256 changed"):
            self.run_rows([self.base])

    def test_bad_label_is_rejected(self) -> None:
        with self.assertRaisesRegex(ValueError, "Invalid edited text or label"):
            self.run_rows([dict(self.base, label=1)])

    def test_source_mismatch_is_excluded(self) -> None:
        other = dict(self.base, prompt_id=1, seed_id=1, seed_text="Not the source")
        result = self.run_rows([self.base, other])
        self.assertEqual(result["source_mismatches_by_domain"], {"arxiv": 1})
        self.assertEqual(result["selected_rows"], 1)

    def test_seed_id_from_another_domain_is_rejected(self) -> None:
        with self.assertRaisesRegex(ValueError, "Invalid edited source group"):
            self.run_rows([dict(self.base, seed_id=200)])

    def test_truncated_stream_is_rejected(self) -> None:
        with self.assertRaisesRegex(ValueError, "Edited input size changed"):
            self.run_rows([self.base], size_delta=1)

    def test_duplicate_condition_is_reported(self) -> None:
        result = self.run_rows([self.base, dict(self.base, text="Another revision.")])
        self.assertEqual(result["duplicate_selected_condition_keys"], 1)


if __name__ == "__main__":
    unittest.main()
