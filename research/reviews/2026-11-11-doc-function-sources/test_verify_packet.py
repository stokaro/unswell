"""Negative tests for the frozen source-only packet verifier."""

import json
from pathlib import Path
import shutil
import tempfile
import unittest

import materialize_review
import verify_packet


class PacketVerificationTest(unittest.TestCase):
    def test_published_packet_passes(self):
        self.assertEqual(verify_packet.verify()["positives"], 8)

    def test_changed_original_span_is_rejected(self):
        with tempfile.TemporaryDirectory() as directory:
            copy = Path(directory) / "packet"
            shutil.copytree(verify_packet.HERE, copy)
            path = copy / "source-only-review.json"
            review = json.loads(path.read_text())
            review["labels"][0]["span"]["end"] += 1
            path.write_text(json.dumps(review))
            with self.assertRaisesRegex(ValueError, "source span differs"):
                verify_packet.verify(copy)

    def test_missing_or_ambiguous_quote_is_rejected_by_materializer(self):
        with self.assertRaisesRegex(ValueError, "quote missing or repeated"):
            materialize_review.locate(b"source", "absent")
        with self.assertRaisesRegex(ValueError, "quote missing or repeated"):
            materialize_review.locate(b"repeated repeated", "repeated")


if __name__ == "__main__":
    unittest.main()
