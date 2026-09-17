#!/usr/bin/env python3
"""Check that applicability parity cannot hide changed findings or observations."""
import copy
import sys
import unittest
sys.dont_write_bytecode=True
import budget_parity as m


class EvidenceTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        root=m.ROOT/'reports/exposed_instruction'
        cls.before=m.audit.read(root/'before/technical.json.gz')
        cls.after=m.audit.read(root/'after/technical.json.gz')

    def check(self,before,after):
        return m.compare(before,after,'exposed_instruction',False)

    def test_only_expected_abstention_is_removed(self):
        self.assertEqual(self.check(self.before,self.after),1)

    def test_missing_baseline_abstention(self):
        b=copy.deepcopy(self.before);b['abstentions']=[]
        with self.assertRaisesRegex(ValueError,'Unexpected before'):self.check(b,self.after)

    def test_persistent_abstention(self):
        a=copy.deepcopy(self.after);a['abstentions']=[m.ABSTENTION]
        with self.assertRaisesRegex(ValueError,'After abstention'):self.check(self.before,a)

    def test_inconsistent_applicability(self):
        a=copy.deepcopy(self.after);a['manifest']['abstained_rules']=['repetition.repeated-claim']
        with self.assertRaisesRegex(ValueError,'After applicability'):self.check(self.before,a)

    def test_changed_findings(self):
        a=copy.deepcopy(self.after);a['findings'].pop()
        with self.assertRaisesRegex(ValueError,'Report semantics'):self.check(self.before,a)

    def test_changed_gate(self):
        a=copy.deepcopy(self.after);a['gate']['passed']=False
        with self.assertRaisesRegex(ValueError,'Report semantics'):self.check(self.before,a)

    def test_changed_source(self):
        a=copy.deepcopy(self.after);a['documents'][0]['source']+='changed'
        with self.assertRaisesRegex(ValueError,'Report semantics'):self.check(self.before,a)

    def test_changed_observation(self):
        root=m.ROOT/'activations/confirmation'
        b=m.audit.read(root/'before/technical.json.gz');a=m.audit.read(root/'after/technical.json.gz')
        m.compare(b,a,'confirmation',True)
        a['features']['sources'][0]['units'][0]['values'][0]['reason']='changed'
        with self.assertRaisesRegex(ValueError,'Report semantics'):m.compare(b,a,'confirmation',True)


if __name__=='__main__':unittest.main()
