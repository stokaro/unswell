"""Reject invalid credit and changes concealed by stable source spans."""
import copy
from pathlib import Path
import sys
import unittest

sys.dont_write_bytecode = True
sys.path.insert(0, str(Path(__file__).resolve().parent))
import compare as c


class EvidenceTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.pages, _, cls.events = c.confirmation(c.ROOT)
        cls.report = c.audit.read(c.ROOT/'reports/confirmation/after/technical.json.gz')
        cls.row = c.audit.read(c.ROOT/'dispositions.json')['confirmation']['technical']['after_credits'][0]

    def test_frozen_records_validate(self):
        result = c.evaluate()
        self.assertEqual(len(result['development']['technical']['delta']['added']), 7)
        self.assertEqual(result['confirmation']['technical']['delta'], {'added': [], 'removed': []})

    def test_reject_uncertain_credit(self):
        row = copy.deepcopy(self.row)
        row['events'] = ['c01-u01']
        with self.assertRaisesRegex(ValueError, 'Uncertain/control credit'):
            c.validate_claim(self.report['findings'][1],row,self.pages,self.events)

    def test_reject_different_page_credit(self):
        row = copy.deepcopy(self.row)
        row['events'] = ['c05-d01']
        with self.assertRaisesRegex(ValueError, 'Unsupported credit'):
            c.validate_claim(self.report['findings'][1],row,self.pages,self.events)

    def test_length_does_not_prove_framing(self):
        finding = copy.deepcopy(self.report['findings'][1])
        finding['rule_id'] = 'syntax.long-sentence'
        row = dict(self.row,finding_sha256=c.audit.sha(c.audit.encode(finding)))
        with self.assertRaisesRegex(ValueError, 'Unsupported credit'):
            c.validate_claim(finding,row,self.pages,self.events)

    def test_stable_location_cannot_hide_gate_change(self):
        before = {'findings': [self.report['findings'][1]]}
        after = copy.deepcopy(before)
        after['findings'][0]['gate'] = 'forbid'
        with self.assertRaisesRegex(ValueError, 'Retained diagnostic changed'):
            c.delta(before,after)

    def test_reject_stale_finding_hash(self):
        row = dict(self.row,finding_sha256='0'*64)
        with self.assertRaisesRegex(ValueError,'Finding review drift'):
            c.validate_claim(self.report['findings'][1],row,self.pages,self.events)


if __name__ == '__main__':
    unittest.main()
