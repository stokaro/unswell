"""Verify that changes to the prereview selection are rejected."""

import json
import shutil
import tempfile
import unittest
from pathlib import Path

from verify_selection import verify


PACKET = Path(__file__).resolve().parent


class SelectionTests(unittest.TestCase):
    def test_frozen_selection(self):
        self.assertEqual(9, len(verify(PACKET)["selected"]))

    def test_changed_search_response(self):
        with tempfile.TemporaryDirectory() as directory:
            copy = Path(directory)
            shutil.copytree(PACKET / "search", copy / "search")
            for name in ("selection.json", "prior-source-repositories.json"):
                shutil.copy2(PACKET / name, copy / name)
            path = copy / "search" / "01.json"
            data = json.loads(path.read_text())
            data["total_count"] += 1
            path.write_text(json.dumps(data))
            with self.assertRaisesRegex(ValueError, "frozen source selection changed"):
                verify(copy)

    def test_dropped_selected_page(self):
        with tempfile.TemporaryDirectory() as directory:
            copy = Path(directory)
            shutil.copytree(PACKET / "search", copy / "search")
            for name in ("selection.json", "prior-source-repositories.json"):
                shutil.copy2(PACKET / name, copy / name)
            path = copy / "selection.json"
            data = json.loads(path.read_text())
            data["selected"].pop()
            path.write_text(json.dumps(data))
            with self.assertRaisesRegex(ValueError, "frozen source selection changed"):
                verify(copy)

    def test_changed_exclusion_snapshot(self):
        with tempfile.TemporaryDirectory() as directory:
            copy = Path(directory)
            shutil.copytree(PACKET / "search", copy / "search")
            for name in ("selection.json", "prior-source-repositories.json"):
                shutil.copy2(PACKET / name, copy / name)
            path = copy / "prior-source-repositories.json"
            data = json.loads(path.read_text())
            data["repositories"].append("example/changed")
            path.write_text(json.dumps(data))
            with self.assertRaisesRegex(ValueError, "frozen source selection changed"):
                verify(copy)


if __name__ == "__main__":
    unittest.main()
