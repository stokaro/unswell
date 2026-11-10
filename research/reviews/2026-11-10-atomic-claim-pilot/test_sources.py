"""Verify that source integrity failures stop the acquisition packet."""

import json
import shutil
import tempfile
import unittest
from pathlib import Path

from verify_sources import verify


PACKET = Path(__file__).resolve().parent


def copy_packet(destination: Path):
    shutil.copytree(PACKET / "search", destination / "search")
    shutil.copytree(PACKET / "sources", destination / "sources")
    for name in ("selection.json", "prior-source-repositories.json"):
        shutil.copy2(PACKET / name, destination / name)


class SourceTests(unittest.TestCase):
    def test_pinned_sources(self):
        self.assertEqual(9, len(verify(PACKET)["sources"]))

    def test_changed_source_byte(self):
        with tempfile.TemporaryDirectory() as directory:
            copy = Path(directory)
            copy_packet(copy)
            path = copy / "sources" / "page-001.source"
            path.write_bytes(path.read_bytes() + b"x")
            with self.assertRaisesRegex(ValueError, "source bytes changed"):
                verify(copy)

    def test_changed_blob_identity(self):
        with tempfile.TemporaryDirectory() as directory:
            copy = Path(directory)
            copy_packet(copy)
            path = copy / "sources" / "inventory.json"
            data = json.loads(path.read_text())
            data["sources"][0]["api_blob_sha1"] = "0" * 40
            path.write_text(json.dumps(data))
            with self.assertRaisesRegex(ValueError, "source bytes changed"):
                verify(copy)

    def test_missing_page(self):
        with tempfile.TemporaryDirectory() as directory:
            copy = Path(directory)
            copy_packet(copy)
            (copy / "sources" / "page-009.source").unlink()
            with self.assertRaisesRegex(ValueError, "source file set changed"):
                verify(copy)

    def test_extra_page(self):
        with tempfile.TemporaryDirectory() as directory:
            copy = Path(directory)
            copy_packet(copy)
            (copy / "sources" / "page-010.source").write_text("unexpected")
            with self.assertRaisesRegex(ValueError, "source file set changed"):
                verify(copy)


if __name__ == "__main__":
    unittest.main()
