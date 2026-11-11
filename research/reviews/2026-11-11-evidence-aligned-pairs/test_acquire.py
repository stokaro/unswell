import json
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch

import acquire
import fetch_sources
import materialize_review


class SelectionTest(unittest.TestCase):
    def make_packet(self, root):
        search = root / "search"
        search.mkdir()
        for name in acquire.QUERIES:
            items = [{"full_name": "Example/Prior", "private": False, "fork": False,
                      "archived": False, "default_branch": "main"}]
            items += [{"full_name": f"Test/{name}{rank}", "private": False,
                       "fork": False, "archived": False, "default_branch": "main"}
                      for rank in range(2, 11)]
            (search / f"{name}.json").write_text(json.dumps({"items": items,
                                                               "incomplete_results": False}))
        excluded = root / "excluded.json"
        excluded.write_text(json.dumps({"repositories": ["example/prior"]}))
        return search, excluded

    @staticmethod
    def fake_api(endpoint, *, fields=()):
        if "/commits/" in endpoint:
            repo = endpoint.split("/")[2]
            result = {"sha": f"commit-{repo}", "commit": {"tree": {"sha": f"tree-{repo}"}}}
        elif "/git/trees/" in endpoint:
            repo = endpoint.rsplit("/", 1)[1].removeprefix("tree-")
            result = {"tree": [{"path": "README.md", "type": "blob",
                                "size": 1000 if repo.endswith("2") else 5000,
                                "sha": f"blob-{repo}"}]}
        else:
            raise AssertionError(f"unexpected API request: {endpoint}")
        return json.dumps(result).encode(), result

    def test_excludes_prior_identity_and_small_readmes_before_selection(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            search, excluded = self.make_packet(root)
            output = root / "selection.json"
            with patch.object(acquire, "get_json", side_effect=self.fake_api):
                acquire.selection(search, excluded, output)
            result = json.loads(output.read_text())
            self.assertEqual(result["status"], "complete")
            self.assertEqual(len(result["selected"]), 24)
            self.assertEqual(len(result["considered"]), 30)
            self.assertEqual([x["reason"] for x in result["considered"][:2]],
                             ["prior_source_or_duplicate", "readme_size_outside_cap"])
            self.assertEqual([x["selection_ordinal"] for x in result["selected"][:8]],
                             list(range(1, 9)))

    def test_incomplete_snapshot_cannot_produce_selection(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            search, excluded = self.make_packet(root)
            snapshot = json.loads((search / "go.json").read_text())
            snapshot["incomplete_results"] = True
            (search / "go.json").write_text(json.dumps(snapshot))
            output = root / "selection.json"
            with self.assertRaisesRegex(SystemExit, "invalid search response"):
                acquire.selection(search, excluded, output)
            self.assertFalse(output.exists())

    def test_source_bytes_must_match_frozen_git_blob(self):
        import base64
        import hashlib

        data = b"# Readme\nExact UTF-8 bytes: \xc3\xa9\n"
        sha = hashlib.sha1(b"blob " + str(len(data)).encode() + b"\0" + data).hexdigest()
        selected = {"readme_blob_sha1": sha, "readme_bytes": len(data)}
        response = {"encoding": "base64", "sha": sha,
                    "content": base64.b64encode(data).decode()}
        self.assertEqual(fetch_sources.verified_blob(response, selected), data)
        with self.assertRaisesRegex(ValueError, "Git blob SHA-1"):
            fetch_sources.verified_blob({**response, "content": base64.b64encode(
                data.replace(b"Exact", b"Other")).decode()}, selected)
        with self.assertRaises(UnicodeDecodeError):
            invalid = b"\xff" * len(data)
            invalid_sha = hashlib.sha1(b"blob " + str(len(invalid)).encode() +
                                       b"\0" + invalid).hexdigest()
            fetch_sources.verified_blob({"encoding": "base64", "sha": invalid_sha,
                                         "content": base64.b64encode(invalid).decode()},
                                        {"readme_blob_sha1": invalid_sha,
                                         "readme_bytes": len(invalid)})

    def test_review_quote_requires_one_exact_utf8_location(self):
        data = "first é second".encode()
        self.assertEqual(materialize_review.locate(data, "é"),
                         {"start": 6, "end": 8, "quote": "é"})
        with self.assertRaisesRegex(ValueError, "missing or repeated"):
            materialize_review.locate(b"same same", "same")
        with self.assertRaisesRegex(ValueError, "missing or repeated"):
            materialize_review.locate(data, "not there")


if __name__ == "__main__":
    unittest.main()
