"""Reject changed frozen evidence while allowing a later CI recipe change."""

import unittest

from verify_historical import compare_frozen


class HistoricalFreezeTests(unittest.TestCase):
    def test_changed_current_makefile_uses_published_blob(self):
        compare_frozen({"Makefile": "original", "source": "source-hash"},
                       {"Makefile": "new", "source": "source-hash"}, "original")

    def test_changed_historical_makefile_is_rejected(self):
        with self.assertRaisesRegex(ValueError, "Makefile"):
            compare_frozen({"Makefile": "original"}, {"Makefile": "new"}, "wrong")

    def test_changed_source_is_rejected(self):
        with self.assertRaisesRegex(ValueError, "source"):
            compare_frozen({"Makefile": "original", "source": "source-hash"},
                           {"Makefile": "new", "source": "wrong"}, "original")

    def test_added_source_is_rejected(self):
        with self.assertRaisesRegex(ValueError, "file set"):
            compare_frozen({"Makefile": "original"},
                           {"Makefile": "new", "extra": "extra"}, "original")


if __name__ == "__main__":
    unittest.main()
