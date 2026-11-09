"""Check source-bound location accounting before the frozen screen."""

from __future__ import annotations

import unittest

import run as candidate


class SourceBindingTest(unittest.TestCase):
    def test_utf8_content_uses_original_byte_offsets(self) -> None:
        raw = "a—é!".encode("utf-8")
        self.assertEqual(candidate.content_positions(raw, 0, len(raw)),
                         {0, 4, 5})

    def test_empty_target_content_fails(self) -> None:
        with self.assertRaises(ValueError):
            candidate.content_positions(b"---", 0, 3)

    def test_context_stays_inside_eligible_block(self) -> None:
        units = [
            {"text": "One.", "binding": {"block_id": 1, "block_kind": "paragraph"}},
            {"text": "Two.", "binding": {"block_id": 1, "block_kind": "paragraph"}},
            {"text": "Other.", "binding": {"block_id": 2, "block_kind": "paragraph"}},
        ]
        self.assertEqual(candidate.marked_context(units, 1),
                         "PREV One. TARGET Two. NEXT ")

    def test_category_selection_keeps_distinct_units(self) -> None:
        records = [{"page": f"page-{page:03d}", "ordinal": ordinal}
                   for page in range(1, 37) for ordinal in range(3)]
        scores = {name: [0.0] * len(records) for name in candidate.CATEGORIES}
        for page in range(36):
            for name in candidate.CATEGORIES:
                scores[name][page * 3] = 1.0
        selected = candidate.select(records, scores)
        self.assertEqual(len(selected), 108)
        self.assertEqual(selected[:3], [0, 1, 2])

    def test_compound_event_requires_every_target(self) -> None:
        freeze = {name: "pinned" for name in ("main_commit", "main_tree",
                  "archive_sha256", "exporter_sha256", "protocol_sha256",
                  "script_sha256", "ledger_sha256", "frontier_sha256")}
        freeze["dependencies"] = {}
        records = [
            {"page": "page-001", "segments": [{"start": 0, "end": 2}],
             "labels": {name: False for name in candidate.CATEGORIES},
             "ordinal": 0, "text_sha256": "a"},
            {"page": "page-001", "segments": [{"start": 4, "end": 6}],
             "labels": {name: False for name in candidate.CATEGORIES},
             "ordinal": 1, "text_sha256": "b"},
        ]
        targets = {"whole/p001-d01": ("page-001", [{0, 1}, {4, 5}])}
        events = {"whole/p001-d01": {"category": "empty_framing"}}
        scores = {name: [0.1, 0.2] for name in candidate.CATEGORIES}
        partial, _ = candidate.evaluate(records, [0], targets, events, set(),
                                        freeze, scores)
        full, _ = candidate.evaluate(records, [0, 1], targets, events, set(),
                                     freeze, scores)
        self.assertEqual(partial["newly_covered_events"], 0)
        self.assertEqual(full["newly_covered_events"], 1)


if __name__ == "__main__":
    unittest.main()
