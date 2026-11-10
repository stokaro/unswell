"""Negative checks for the #349 source and reference freeze."""

import json
from pathlib import Path
import shutil
import tempfile
import unittest

import verify_source as target


class VerifySourceTest(unittest.TestCase):
    def copy_packet(self, directory):
        root = Path(directory)
        for name in (*target.FILES, "freeze.json"):
            path = root / name
            path.parent.mkdir(parents=True, exist_ok=True)
            shutil.copy2(target.ROOT / name, path)
        return root

    def refresh_freeze(self, root):
        (root / "freeze.json").write_text(json.dumps({
            name: target.sha((root / name).read_bytes()) for name in target.FILES
        }, sort_keys=True, indent=2) + "\n")

    def test_original(self):
        self.assertEqual(target.verify()["events"], 4)

    def test_source_mutation_rejected_even_with_refreshed_freeze(self):
        with tempfile.TemporaryDirectory() as directory:
            root = self.copy_packet(directory)
            path = root / "sources" / "nushell.source"
            path.write_bytes(path.read_bytes().replace(b"A new type of shell.",
                                                      b"A strange type of shell.", 1))
            self.refresh_freeze(root)
            with self.assertRaisesRegex(ValueError, "Git blob"):
                target.verify(root)

    def test_changed_review_span_rejected_even_with_refreshed_freeze(self):
        with tempfile.TemporaryDirectory() as directory:
            root = self.copy_packet(directory)
            path = root / "reference.json"
            data = json.loads(path.read_text())
            data["events"][0]["span"]["start"] += 1
            path.write_text(json.dumps(data, ensure_ascii=False, sort_keys=True, indent=2) + "\n")
            self.refresh_freeze(root)
            with self.assertRaisesRegex(ValueError, "review quote"):
                target.verify(root)

    def test_protected_target_rejected_even_with_refreshed_freeze(self):
        with tempfile.TemporaryDirectory() as directory:
            root = self.copy_packet(directory)
            path = root / "reference.json"
            data = json.loads(path.read_text())
            source = (root / "sources" / "nushell.source").read_bytes()
            quote = b"ls | where type == \"dir\" | table"
            self.assertEqual(source.count(quote), 1)
            item = data["events"][0]
            item["quote"] = quote.decode()
            item["span"] = {"start": source.index(quote), "end": source.index(quote) + len(quote)}
            path.write_text(json.dumps(data, ensure_ascii=False, sort_keys=True, indent=2) + "\n")
            self.refresh_freeze(root)
            with self.assertRaisesRegex(ValueError, "review eligible binding"):
                target.verify(root)


if __name__ == "__main__":
    unittest.main()
