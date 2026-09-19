#!/usr/bin/env python3
"""Reject missing evidence and credit assigned to a neighboring sentence."""
import copy
import gzip
import json
import sys
import unittest
sys.dont_write_bytecode = True
import verify as v


class EvidenceTests(unittest.TestCase):
    def setUp(self):
        self.before = json.loads(gzip.decompress((v.ROOT/'reports/before/strict.json.gz').read_bytes()))
        self.after = json.loads(gzip.decompress((v.ROOT/'reports/after/strict.json.gz').read_bytes()))
        self.review = json.loads((v.ROOT/'source-review.json').read_text())
        self.new = next(f for f in self.after['findings'] if f['rule_id'] == v.RULE)

    def test_complete_record(self):
        self.assertEqual(v.evaluate()['strict']['confirmation_detected'], 0)

    def test_missing_related_heading(self):
        self.new['related'] = []
        with self.assertRaisesRegex(ValueError, 'paired source'):
            v.check_pair(self.before, self.after, self.review)

    def test_neighboring_sentence_is_not_credit(self):
        self.new['primary']['span']['start'] += 60
        with self.assertRaisesRegex(ValueError, 'Wrong source'):
            v.check_pair(self.before, self.after, self.review)

    def test_unreviewed_extra_warning(self):
        self.after['findings'].append(copy.deepcopy(self.new))
        with self.assertRaisesRegex(ValueError, 'exactly one'):
            v.check_pair(self.before, self.after, self.review)

    def test_existing_rule_drift(self):
        self.before['findings'][0]['message'] = 'Changed diagnostic'
        with self.assertRaisesRegex(ValueError, 'existing diagnostic'):
            v.check_pair(self.before, self.after, self.review)

    def test_incomplete_run(self):
        self.after['status'] = 'partial'
        with self.assertRaisesRegex(ValueError, 'Incomplete'):
            v.check_pair(self.before, self.after, self.review)


if __name__ == '__main__':
    unittest.main()
