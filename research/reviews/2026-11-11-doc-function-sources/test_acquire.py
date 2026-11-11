import json
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch

import acquire
import fetch_sources


class AcquisitionTest(unittest.TestCase):
    @staticmethod
    def packet(root):
        search = root / "search"
        search.mkdir()
        for topic in acquire.QUERIES:
            items = [{"full_name": "Prior/Seen", "default_branch": "main"}]
            items += [{"full_name": f"Test/{topic}-{rank}", "default_branch": "main"}
                      for rank in range(2, 10)]
            (search / f"{topic}.json").write_text(json.dumps({
                "items": items, "incomplete_results": False}))
        excluded = root / "exclusions.json"
        excluded.write_text(json.dumps({"repositories": ["prior/seen"]}))
        return search, excluded

    @staticmethod
    def api(endpoint, *, fields=()):
        repo = endpoint.split("/")[2]
        if "/commits/" in endpoint:
            result = {"sha": f"commit-{repo}",
                      "commit": {"tree": {"sha": f"tree-{repo}"}}}
        elif "/git/trees/" in endpoint:
            rank = int(repo.rsplit("-", 1)[1])
            function = "guides" if rank % 2 == 0 else "reference"
            result = {"truncated": False, "tree": [
                {"type": "blob", "path": f"docs/{function}/long.md",
                 "size": 9000, "sha": f"blob-{repo}"},
                {"type": "blob", "path": f"docs/{function}/short.md",
                 "size": 8000, "sha": "excluded-small"},
            ]}
        else:
            raise AssertionError(endpoint)
        return json.dumps(result).encode(), result

    def test_path_function_and_size_are_metadata_only(self):
        self.assertEqual(acquire.file_functions({"type": "blob",
                                                "path": "site/docs/guides/setup.md",
                                                "size": 9000}), {"guide"})
        self.assertEqual(acquire.file_functions({"type": "blob",
                                                "path": "docs/api/reference.mdx",
                                                "size": 9000}), {"concept_reference"})
        self.assertEqual(acquire.file_functions({"type": "blob",
                                                "path": "docs/guides/README.md",
                                                "size": 9000}), set())
        self.assertEqual(acquire.file_functions({"type": "blob",
                                                "path": "docs/guides/setup.md",
                                                "size": 8191}), set())

    def test_fixed_selection_excludes_prior_and_balances_functions(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            search, excluded = self.packet(root)
            output = root / "selection.json"
            with patch.object(acquire, "get_json", side_effect=self.api):
                acquire.select(search, excluded, output)
            selected = json.loads(output.read_text())
            self.assertEqual(selected["status"], "complete")
            self.assertEqual(len(selected["selected"]), 32)
            self.assertEqual(len(selected["considered"]), 36)
            for topic in acquire.QUERIES:
                for function in ("guide", "concept_reference"):
                    matching = [x for x in selected["selected"]
                                if x["topic"] == topic and x["function"] == function]
                    self.assertEqual([x["selection_ordinal"] for x in matching],
                                     [1, 2, 3, 4])
            self.assertEqual(selected["considered"][0]["reason"],
                             "prior_source_or_duplicate")

    def test_incomplete_snapshot_stops_without_selection(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            search, excluded = self.packet(root)
            snapshot = json.loads((search / "database.json").read_text())
            snapshot["incomplete_results"] = True
            (search / "database.json").write_text(json.dumps(snapshot))
            output = root / "selection.json"
            with self.assertRaisesRegex(ValueError, "invalid search snapshot"):
                acquire.select(search, excluded, output)
            self.assertFalse(output.exists())

    def test_fetched_bytes_must_match_frozen_blob(self):
        import base64
        import hashlib

        data = "A document with é and a preserved byte span.\n".encode()
        sha = hashlib.sha1(b"blob " + str(len(data)).encode() + b"\0" + data).hexdigest()
        selected = {"blob_sha1": sha, "bytes": len(data)}
        response = {"encoding": "base64", "sha": sha,
                    "content": base64.b64encode(data).decode()}
        self.assertEqual(fetch_sources.verified_blob(response, selected), data)
        with self.assertRaisesRegex(ValueError, "Git blob SHA-1"):
            fetch_sources.verified_blob({**response, "content": base64.b64encode(
                data.replace(b"document", b"argument")).decode()}, selected)


if __name__ == "__main__":
    unittest.main()
