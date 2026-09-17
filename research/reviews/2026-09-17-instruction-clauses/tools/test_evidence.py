#!/usr/bin/env python3
"""Reject unsupported credit, missing reviews and hidden coverage losses."""
import copy
import sys
import unittest
sys.dont_write_bytecode = True
import summarize as m


class EvidenceTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.sets = m.data_sets()
        cls.pages, _, cls.events = cls.sets['confirmation']
        cls.report = m.audit.read(m.ROOT/'reports/confirmation/after/strict.json.gz')
        cls.rows = m.audit.read(m.ROOT/'dispositions.json')['confirmation']['strict']['after']

    def check(self, report, rows):
        m.metrics.check_rows(report, rows, self.pages, self.events, True)

    def test_complete_review(self):
        self.check(self.report, self.rows)
        m.frozen_inputs()

    def test_missing_retained_finding(self):
        with self.assertRaisesRegex(ValueError, 'Missing or unexpected'):
            self.check(self.report, self.rows[:-1])

    def test_changed_advice_needs_new_review(self):
        report = copy.deepcopy(self.report)
        report['findings'][31]['evidence']['suggestion'] = 'changed'
        self.assertEqual(m.finding_delta(self.report, report),
                         dict(added=[31], removed=[31], changed_at_same_location=1))
        with self.assertRaisesRegex(ValueError, 'review drift'):
            self.check(report, self.rows)

    def test_cross_page_credit(self):
        rows = copy.deepcopy(self.rows)
        rows[3].update(status='actionable', events=['c04-d18'])
        with self.assertRaisesRegex(ValueError, 'Unsupported credit'):
            self.check(self.report, rows)

    def test_incidental_length_overlap(self):
        rows = copy.deepcopy(self.rows)
        rows[20].update(status='actionable', events=['c03-d04'])
        with self.assertRaisesRegex(ValueError, 'Unsupported credit'):
            self.check(self.report, rows)

    def test_controls_cannot_receive_credit(self):
        rows = copy.deepcopy(self.rows)
        rows[1].update(status='actionable', events=['c01-c01'])
        with self.assertRaisesRegex(ValueError, 'Unsupported credit'):
            self.check(self.report, rows)

    def test_reused_source(self):
        sets = dict(self.sets, duplicate=self.sets['confirmation'])
        with self.assertRaisesRegex(ValueError, 'Reused confirmation source'):
            m.metrics.validate_sources(sets)

    def test_source_quote_drift(self):
        with self.assertRaisesRegex(ValueError, 'Source quote differs'):
            m.audit.span(b'original', dict(start=0, end=8, quote='rewritten'))

    def test_partial_does_not_count_as_full(self):
        _, found = m.metrics.page_metrics(self.report, self.rows, self.pages, self.events, True)
        self.assertEqual(found, {'c02-d10', 'c04-d15', 'c04-d18', 'c05-d08'})
        self.assertEqual(self.rows[28]['partial_events'], ['c04-d07'])
        self.assertNotIn('c04-d07', found)

    def test_additional_finding_cannot_increase_recall(self):
        rows = copy.deepcopy(self.rows)
        rows[31]['events'] = ['c04-d01']
        with self.assertRaisesRegex(ValueError, 'cannot receive frozen credit'):
            self.check(self.report, rows)

    def test_additional_finding_requires_edit_and_phase(self):
        for field in ('proposed_edit', 'review_phase'):
            rows = copy.deepcopy(self.rows)
            rows[31].pop(field)
            with self.assertRaisesRegex(ValueError, 'cannot receive frozen credit'):
                self.check(self.report, rows)

    def test_known_budget_abstention(self):
        report = m.audit.read(m.ROOT/'reports/exposed_instruction/after/strict.json.gz')
        m.validate_abstentions(report, 'exposed_instruction')
        with self.assertRaisesRegex(ValueError, 'Abstention evidence drift'):
            m.validate_abstentions(dict(report, abstentions=[]), 'exposed_instruction')
        with self.assertRaisesRegex(ValueError, 'Abstention evidence drift'):
            m.validate_abstentions(report, 'confirmation')

    def test_lost_detections_remain_visible(self):
        pages, _, events = self.sets['exposed_relations']
        evidence = m.audit.read(m.ROOT/'dispositions.json')['exposed_relations']['strict']
        for phase, expected in [('before', True), ('after', False)]:
            report = m.audit.read(m.ROOT/'reports/exposed_relations'/phase/'strict.json.gz')
            _, found = m.metrics.page_metrics(report, evidence[phase], pages, events, True)
            self.assertEqual('c04-d08' in found, expected)
            self.assertEqual('c04-d14' in found, expected)

    def test_removed_capability_controls_have_no_credit(self):
        evidence = m.audit.read(m.ROOT/'dispositions.json')['confirmation']['strict']
        before = m.audit.read(m.ROOT/'reports/confirmation/before/strict.json.gz')
        retained = {m.prior.key(f) for f in self.report['findings']}
        removed = [r for r in evidence['removed'] if m.prior.key(before['findings'][r['index']]) not in retained]
        self.assertEqual(len(removed), 3)
        self.assertTrue(all(r['status'] == 'nonactionable' and not r['events'] for r in removed))


if __name__ == '__main__':
    unittest.main()
