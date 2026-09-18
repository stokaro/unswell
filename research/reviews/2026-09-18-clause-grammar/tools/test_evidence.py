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
        event = self.events['c03-d01']['targets'][0]
        record['matches'] = [dict(event, page='c03', matches=[dict(reference='https://example.invalid/unreviewed')])]
        record['known_partially_exposed_pages'] = ['c03']
        with self.assertRaisesRegex(ValueError, 'Exposure quote absent'):
            m.validate_exposure(self.sets, record)

    def test_unexpected_abstention_is_rejected(self):
        m.validate_abstentions(self.report, 'confirmation')
        with self.assertRaisesRegex(ValueError, 'Unexpected abstention'):
            m.validate_abstentions(dict(self.report, abstentions=[dict(reason='budget_exhausted')]), 'confirmation')

    def test_frozen_runtime(self):
        record = m.audit.read(m.ROOT/'code-freeze.json')
        m.code_freeze(record)
        record['sha256']['builtin/rhetoric_quality.go'] = '0'*64
        with self.assertRaisesRegex(ValueError, 'Runtime changed'):
            m.code_freeze(record)

    def test_confirmation_does_not_claim_a_gain(self):
        data = m.audit.read(m.ROOT/'dispositions.json')['confirmation']['strict']
        for phase in ('before', 'after'):
            report = m.audit.read(m.ROOT/'reports/confirmation'/phase/'strict.json.gz')
            _, found = m.metrics.page_metrics(report, data[phase], self.pages, self.events, True)
            self.assertEqual(found, {'c03-d11','c05-d08'})

    def test_partial_notice_is_not_full_credit(self):
        row = next(r for r in self.rows if r.get('partial_events'))
        self.assertEqual(row['partial_events'], ['c03-d01'])
        self.assertEqual(row['events'], [])

    def test_additional_repairs_cannot_increase_recall(self):
        rows = copy.deepcopy(self.rows)
        rows[0]['events'] = ['c03-d01']
        with self.assertRaisesRegex(ValueError, 'cannot receive frozen credit'):
            self.check(rows)

    def test_old_losses_are_restored(self):
        pages, _, events = self.sets['exposed_boundaries']
        data = m.audit.read(m.ROOT/'dispositions.json')['exposed_boundaries']['strict']
        for phase, expected in [('before', False), ('after', True)]:
            report = m.audit.read(m.ROOT/'reports/exposed_boundaries'/phase/'strict.json.gz')
            _, found = m.metrics.page_metrics(report, data[phase], pages, events, True)
            self.assertEqual({'c04-d10','c04-d14'}.issubset(found), expected)


if __name__ == '__main__':
    unittest.main()
