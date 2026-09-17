#!/usr/bin/env python3
"""Reject invalid recall claims and post-freeze semantic changes."""
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
        report['findings'][5]['evidence']['suggestion'] = 'changed'
        self.assertEqual(m.finding_delta(self.report, report),
                         dict(added=[5], removed=[5], changed_at_same_location=1))
        with self.assertRaisesRegex(ValueError, 'review drift'):
            self.check(report, self.rows)

    def test_cross_page_credit(self):
        rows = copy.deepcopy(self.rows)
        rows[0].update(status='actionable', events=['c03-d01'])
        with self.assertRaisesRegex(ValueError, 'Unsupported credit'):
            self.check(self.report, rows)

    def test_incidental_length_overlap(self):
        rows = copy.deepcopy(self.rows)
        rows[7].update(status='actionable', events=['c04-d01'])
        with self.assertRaisesRegex(ValueError, 'Unsupported credit'):
            self.check(self.report, rows)

    def test_controls_cannot_receive_credit(self):
        rows = copy.deepcopy(self.rows)
        rows[2].update(status='actionable', events=['c02-c02'])
        with self.assertRaisesRegex(ValueError, 'Unsupported credit'):
            self.check(self.report, rows)

    def test_reused_source(self):
        sets = dict(self.sets, duplicate=self.sets['confirmation'])
        with self.assertRaisesRegex(ValueError, 'Reused confirmation source'):
            m.metrics.validate_sources(sets)

    def test_source_quote_drift(self):
        with self.assertRaisesRegex(ValueError, 'Source quote differs'):
            m.audit.span(b'original', dict(start=0, end=8, quote='rewritten'))

    def test_circular_reason_partial_to_full(self):
        pages, _, events = self.sets['exposed_repetition']
        evidence = m.audit.read(m.ROOT/'dispositions.json')['exposed_repetition']['strict']
        for phase, expected in [('before', False), ('after', True)]:
            report = m.audit.read(m.ROOT/'reports/exposed_repetition'/phase/'strict.json.gz')
            _, found = m.metrics.page_metrics(report, evidence[phase], pages, events, False)
            self.assertEqual('c06-d02' in found, expected)

    def test_additional_finding_cannot_increase_recall(self):
        rows = copy.deepcopy(self.rows)
        rows[13]['events'] = ['c05-d01']
        with self.assertRaisesRegex(ValueError, 'cannot receive frozen credit'):
            self.check(self.report, rows)

    def test_additional_finding_requires_edit_and_phase(self):
        for field in ('proposed_edit', 'review_phase'):
            rows = copy.deepcopy(self.rows)
            rows[13].pop(field)
            with self.assertRaisesRegex(ValueError, 'cannot receive frozen credit'):
                self.check(self.report, rows)

    def test_known_budget_abstention(self):
        report = m.audit.read(m.ROOT/'reports/exposed_instruction/after/strict.json.gz')
        m.validate_abstentions(report, 'exposed_instruction')
        with self.assertRaisesRegex(ValueError, 'Abstention evidence drift'):
            m.validate_abstentions(dict(report, abstentions=[]), 'exposed_instruction')
        with self.assertRaisesRegex(ValueError, 'Abstention evidence drift'):
            m.validate_abstentions(report, 'confirmation')

    def test_prototype_semantic_drift(self):
        prototype = m.audit.read(m.ROOT/'prototype-reports/confirmation/after/strict.json.gz')
        m.validate_prototype(prototype, self.report, 'confirmation')
        prototype['findings'][0]['message'] += ' changed'
        with self.assertRaisesRegex(ValueError, 'Prototype semantic drift'):
            m.validate_prototype(prototype, self.report, 'confirmation')

    def test_prototype_abstention_must_be_retained(self):
        prototype = m.audit.read(m.ROOT/'prototype-reports/exposed_scope/after/strict.json.gz')
        final = m.audit.read(m.ROOT/'reports/exposed_scope/after/strict.json.gz')
        m.validate_prototype(prototype, final, 'exposed_scope')
        prototype['abstentions'] = []
        with self.assertRaisesRegex(ValueError, 'Prototype abstention drift'):
            m.validate_prototype(prototype, final, 'exposed_scope')


if __name__ == '__main__':
    unittest.main()
