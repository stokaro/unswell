#!/usr/bin/env python3
"""Reject hidden findings, invented event credit and source drift."""
import copy
import sys
import unittest
sys.dont_write_bytecode = True
import summarize as m


class EvidenceTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.sets = m.data_sets()
        cls.pages, cls.files, cls.events = cls.sets['confirmation']
        cls.report = m.audit.read(m.ROOT/'reports/confirmation/after/strict.json.gz')
        cls.rows = m.audit.read(m.ROOT/'dispositions.json')['confirmation']['strict']['after']

    def check(self, rows, report=None):
        m.metrics.check_rows(report or self.report, rows, self.pages, self.events, True)

    def test_complete_review(self):
        self.check(self.rows)
        m.frozen_inputs()

    def test_missing_finding(self):
        with self.assertRaisesRegex(ValueError, 'Missing or unexpected'):
            self.check(self.rows[:-1])

    def test_changed_advice_requires_review(self):
        report = copy.deepcopy(self.report)
        report['findings'][0]['evidence']['suggestion'] = 'changed'
        self.assertEqual(m.finding_delta(self.report, report), dict(added=[0], removed=[0], changed_at_same_location=1))
        with self.assertRaisesRegex(ValueError, 'review drift'):
            self.check(self.rows, report)

    def test_control_cannot_receive_credit(self):
        rows = copy.deepcopy(self.rows)
        rows[0].update(status='actionable', events=['c01-c01'])
        with self.assertRaisesRegex(ValueError, 'Unsupported credit'):
            self.check(rows)

    def test_uncertain_cannot_receive_credit(self):
        rows = copy.deepcopy(self.rows)
        rows[0].update(status='actionable', events=['c02-u01'])
        with self.assertRaisesRegex(ValueError, 'Unsupported credit'):
            self.check(rows)

    def test_source_reuse_rejected(self):
        with self.assertRaisesRegex(ValueError, 'Reused confirmation source'):
            m.metrics.validate_sources(dict(self.sets, duplicate=self.sets['confirmation']))

    def test_source_quote_drift(self):
        with self.assertRaisesRegex(ValueError, 'Source quote differs'):
            m.audit.span(b'original', dict(start=0, end=8, quote='rewritten'))

    def test_prior_source_count(self):
        record = m.audit.read(m.ROOT/'confirmation/source-overlap.json')
        record['prior_references_scanned'] -= 1
        with self.assertRaisesRegex(ValueError, 'Exposure source count drift'):
            m.validate_exposure(self.sets, record)

    def test_false_overlap_rejected(self):
        record = m.audit.read(m.ROOT/'confirmation/source-overlap.json')
        event = self.events['c02-d01']['targets'][0]
        record['matches'] = [dict(event, page='c02', matches=[dict(reference='https://example.invalid/unreviewed')])]
        record['known_partially_exposed_pages'] = ['c02']
        with self.assertRaisesRegex(ValueError, 'Exposure quote absent'):
            m.validate_exposure(self.sets, record)

    def test_known_abstention_is_retained(self):
        report = m.audit.read(m.ROOT/'reports/exposed_instruction/after/strict.json.gz')
        m.validate_abstentions(report, 'exposed_instruction')
        with self.assertRaisesRegex(ValueError, 'Abstention evidence drift'):
            m.validate_abstentions(dict(report, abstentions=[]), 'exposed_instruction')
        with self.assertRaisesRegex(ValueError, 'Abstention evidence drift'):
            m.validate_abstentions(report, 'confirmation')


    def test_confirmation_regression_remains_visible(self):
        data = m.audit.read(m.ROOT/'dispositions.json')['confirmation']['strict']
        for phase, expected in [('before', {'c02-d10','c04-d13'}), ('after', {'c02-d10'})]:
            report = m.audit.read(m.ROOT/'reports/confirmation'/phase/'strict.json.gz')
            _, found = m.metrics.page_metrics(report, data[phase], self.pages, self.events, True)
            self.assertEqual(found, expected)
        self.assertEqual(data['after'][20]['partial_events'], ['c04-d09'])

    def test_lost_named_function_detection(self):
        pages, _, events = self.sets['exposed_clauses']
        data = m.audit.read(m.ROOT/'dispositions.json')['exposed_clauses']['strict']
        for phase, expected in [('before', True), ('after', False)]:
            report = m.audit.read(m.ROOT/'reports/exposed_clauses'/phase/'strict.json.gz')
            _, found = m.metrics.page_metrics(report, data[phase], pages, events, True)
            self.assertEqual('c04-d19' in found, expected)

    def test_additional_repairs_cannot_increase_recall(self):
        rows = copy.deepcopy(self.rows)
        rows[15]['events'] = ['c04-d01']
        with self.assertRaisesRegex(ValueError, 'cannot receive frozen credit'):
            self.check(rows)

    def test_incidental_length_does_not_credit_wordiness(self):
        rows = copy.deepcopy(self.rows)
        rows[0].update(status='actionable', events=['c02-d02'])
        with self.assertRaisesRegex(ValueError, 'Unsupported credit'):
            self.check(rows)


if __name__ == '__main__':
    unittest.main()
