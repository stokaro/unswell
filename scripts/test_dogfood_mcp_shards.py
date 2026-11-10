"""Check that MCP dogfood shard accounting rejects missing and repeated sources."""

import unittest

from scripts.dogfood_mcp_shards import validate_coverage


def batch(*identities):
    return {"result": {"documents": [
        {"name": name, "source_hash": digest} for name, digest in identities
    ]}}


class CoverageTests(unittest.TestCase):
    def test_complete_shards(self):
        validate_coverage([("a.md", "a"), ("b.md", "b")],
                          [batch(("a.md", "a")), batch(("b.md", "b"))])

    def test_missing_source_is_rejected(self):
        with self.assertRaisesRegex(ValueError, "omit, duplicate, or change"):
            validate_coverage([("a.md", "a"), ("b.md", "b")],
                              [batch(("a.md", "a"))])

    def test_repeated_source_is_rejected(self):
        with self.assertRaisesRegex(ValueError, "omit, duplicate, or change"):
            validate_coverage([("a.md", "a"), ("b.md", "b")],
                              [batch(("a.md", "a")), batch(("a.md", "a"))])

    def test_changed_hash_is_rejected(self):
        with self.assertRaisesRegex(ValueError, "omit, duplicate, or change"):
            validate_coverage([("a.md", "a"), ("b.md", "b")],
                              [batch(("a.md", "a")), batch(("b.md", "wrong"))])


if __name__ == "__main__":
    unittest.main()
