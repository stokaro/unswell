"""Negative controls for source-bound packet verification."""

import copy
import hashlib
from pathlib import Path
import sys
import tempfile
import unittest

HERE = Path(__file__).resolve().parent
sys.path.insert(0, str(HERE))
import verify_packet


class VerifyPacketTest(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.sources = Path(self.temp.name)
        data = b"Evidence says value is seven.\n"
        (self.sources / "page-001.source").write_bytes(data)
        sha = hashlib.sha256(data).hexdigest()
        blob = hashlib.sha1(b"blob " + str(len(data)).encode() + b"\0" + data).hexdigest()
        identity = {"repo": "example/project", "path": "README.md",
                    "commit": "a" * 40, "url": "https://example.test/source"}
        self.selection = {"selected": [dict(number=i, search_blob_sha1=blob,
                                            **identity) for i in range(1, 21)]}
        self.inventory = {"records": [dict(number=i, status="fetched",
                                             source_sha256=sha, source_blob_sha1=blob,
                                             source_file="page-001.source")
                                      for i in range(1, 21)]}
        labels = ["entailment"] * 6 + ["neutral"] * 6 + ["contradiction"] * 4
        self.packet = {
            "format": "unswell-349-explicit-proposition-nli-v1",
            "source_inventory_sha256": "f" * 64,
            "cases": [dict(id=f"case-{i}", source_number=1, source_sha256=sha,
                           evidence_spans=[{"start": 0, "end": len(data) - 1,
                                            "quote": data[:-1].decode()}],
                           premise=data[:-1].decode(), hypothesis="value is seven",
                           expected_relation=label, **identity)
                      for i, label in enumerate(labels)]}

    def check(self, packet):
        verify_packet.verify(packet, self.selection, self.inventory,
                             self.sources, "f" * 64)

    def test_valid(self):
        self.check(self.packet)

    def test_changed_quote(self):
        packet = copy.deepcopy(self.packet)
        packet["cases"][0]["evidence_spans"][0]["quote"] += " extra"
        with self.assertRaisesRegex(ValueError, "premise is not the bound excerpts"):
            self.check(packet)

    def test_changed_source_hash(self):
        packet = copy.deepcopy(self.packet)
        packet["cases"][0]["source_sha256"] = "0" * 64
        with self.assertRaisesRegex(ValueError, "source identity"):
            self.check(packet)

    def test_missing_case(self):
        packet = copy.deepcopy(self.packet)
        packet["cases"].pop()
        with self.assertRaisesRegex(ValueError, "selection or case count"):
            self.check(packet)

    def test_changed_source_bytes(self):
        (self.sources / "page-001.source").write_text("Evidence says value is eight.\n")
        with self.assertRaisesRegex(ValueError, "source bytes changed"):
            self.check(self.packet)


if __name__ == "__main__":
    unittest.main()
