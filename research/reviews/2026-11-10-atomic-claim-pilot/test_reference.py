"""Check that a frozen source-only reference cannot silently drift."""

import json
import shutil
import tempfile
import unittest
from pathlib import Path

from verify_reference import verify


PACKET = Path(__file__).resolve().parent


def copy_packet(destination: Path):
    shutil.copytree(PACKET / "search", destination / "search")
    shutil.copytree(PACKET / "sources", destination / "sources")
    for name in ("selection.json", "prior-source-repositories.json", "input.json",
                 "inventory.json", "reference_draft.json", "reference.json"):
        shutil.copy2(PACKET / name, destination / name)


class ReferenceTests(unittest.TestCase):
    def test_frozen_reference(self):
        record = verify(PACKET)
        self.assertEqual((15, 21), (len(record["events"]), len(record["controls"])))

    def test_changed_label(self):
        with tempfile.TemporaryDirectory() as directory:
            copy = Path(directory)
            copy_packet(copy)
            path = copy / "reference.json"
            record = json.loads(path.read_text())
            record["events"][0]["purpose"] = "vague_claims"
            path.write_text(json.dumps(record))
            with self.assertRaisesRegex(ValueError, "frozen source-only reference changed"):
                verify(copy)

    def test_changed_inventory(self):
        with tempfile.TemporaryDirectory() as directory:
            copy = Path(directory)
            copy_packet(copy)
            path = copy / "inventory.json"
            path.write_bytes(path.read_bytes() + b" ")
            with self.assertRaisesRegex(ValueError, "frozen source-only reference changed"):
                verify(copy)

    def test_changed_source(self):
        with tempfile.TemporaryDirectory() as directory:
            copy = Path(directory)
            copy_packet(copy)
            path = copy / "sources" / "page-001.source"
            path.write_bytes(path.read_bytes() + b"x")
            with self.assertRaisesRegex(ValueError, "source bytes changed"):
                verify(copy)

    def test_input_text_does_not_match_source(self):
        with tempfile.TemporaryDirectory() as directory:
            copy = Path(directory)
            copy_packet(copy)
            path = copy / "input.json"
            record = json.loads(path.read_text())
            record["pages"][0]["text"] += "x"
            path.write_text(json.dumps(record))
            with self.assertRaisesRegex(ValueError, "wrong source inventory"):
                verify(copy)


if __name__ == "__main__":
    unittest.main()
