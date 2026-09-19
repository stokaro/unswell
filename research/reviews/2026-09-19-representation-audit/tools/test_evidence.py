#!/usr/bin/env python3
"""Reject misleading representation counts and mismatched source evidence."""
import copy
import unittest

from evaluate import classify, evaluate
from prepare import checked_targets
from supervision import label_unit


class RepresentationTests(unittest.TestCase):
    def test_unmarked_and_partially_clean_units_are_not_negatives(self):
        self.assertEqual(label_unit({1, 2}, [], [], [])[0], None)
        self.assertEqual(label_unit({1, 2}, [], [], [{1}])[0], None)
        self.assertEqual(label_unit({1, 2}, [], [], [{1, 2}])[0], 0)

    def test_uncertain_or_partial_defect_overlap_cannot_be_clean(self):
        self.assertEqual(label_unit({1, 2}, [{1, 8}], [], [{1, 2}])[0], None)
        self.assertEqual(label_unit({1, 2}, [], [{1}], [{1, 2}])[0], None)
        self.assertEqual(label_unit({1, 2}, [{1}], [], [{2}])[0], 1)

    def test_protected_gap_requires_two_pieces(self):
        target = {0, 1, 8, 9}
        value = classify(target, {0: target}, [{0, 1}, {8, 9}], [])
        self.assertEqual(value['representation'], 'protected_gap_split')
        self.assertTrue(value['one_block'])
        self.assertFalse(value['one_piece'])

    def test_multiple_locations_count_once(self):
        value = classify({1, 8}, {0: {1}, 1: {8}}, [{1}, {8}], [{1}, {8}])
        self.assertEqual(value['representation'], 'multiple_blocks')
        self.assertEqual(value['blocks'], [0, 1])

    def test_missing_prose_and_token_loss_are_distinct(self):
        self.assertEqual(classify({8}, {0: {1}}, [{1}], [])['representation'], 'no_eligible_prose')
        self.assertEqual(classify({1, 8}, {0: {1, 8}}, [{1}], [])['representation'], 'prepared_token_loss')

    def test_list_fragment_is_not_a_sentence_or_missing_text(self):
        value = classify({1}, {3: {1, 2}}, [{1, 2}], [])
        self.assertEqual(value['representation'], 'one_prepared_piece')
        self.assertFalse(value['one_sentence'])

    def test_quotes_use_utf8_bytes(self):
        review = {'defects': [{'targets': [{'start': 0, 'end': 5, 'quote': 'café'}]}]}
        checked_targets(review, 'café is valid'.encode())
        wrong = copy.deepcopy(review)
        wrong['defects'][0]['targets'][0]['end'] = 4
        with self.assertRaises((ValueError, UnicodeError)):
            checked_targets(wrong, 'café is valid'.encode())

    def test_incomplete_run_cannot_produce_measurements(self):
        with self.assertRaisesRegex(ValueError, 'Incomplete analysis'):
            evaluate({}, {'status': 'partial', 'manifest': {'complete': False}, 'errors': []})

    def test_missing_document_is_not_an_empty_negative(self):
        report = {'status': 'complete', 'manifest': {'complete': True, 'skipped_rules': []},
                  'errors': [], 'documents': []}
        with self.assertRaisesRegex(ValueError, 'Document identity mismatch'):
            evaluate({'pages': [{'id': 'expected', 'sha256': 'x'}]}, report)


if __name__ == '__main__':
    unittest.main()
