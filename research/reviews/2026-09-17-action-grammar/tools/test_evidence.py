#!/usr/bin/env python3
"""Check source exposure, honest credit and failure-path evidence."""
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

    def check(self, rows, report=None):
        m.metrics.check_rows(report or self.report, rows, self.pages, self.events, True)

    def test_complete_review(self):
        self.check(self.rows)
        m.frozen_inputs()

    def test_missing_retained_finding(self):
        with self.assertRaisesRegex(ValueError, 'Missing or unexpected'):
            self.check(self.rows[:-1])

    def test_changed_advice_requires_review(self):
        report = copy.deepcopy(self.report)
        report['findings'][21]['evidence']['suggestion'] = 'changed'
        self.assertEqual(m.finding_delta(self.report, report),
                         dict(added=[21], removed=[21], changed_at_same_location=1))
        with self.assertRaisesRegex(ValueError, 'review drift'):
            self.check(self.rows, report)

    def test_length_overlap_does_not_credit_metaphor(self):
        rows = copy.deepcopy(self.rows)
        rows[2].update(status='actionable', events=['c02-d06'])
        with self.assertRaisesRegex(ValueError, 'Unsupported credit'):
            self.check(rows)

    def test_controls_do_not_credit(self):
        rows = copy.deepcopy(self.rows)
        rows[15].update(status='actionable', events=['c04-c01'])
        with self.assertRaisesRegex(ValueError, 'Unsupported credit'):
            self.check(rows)

    def test_additional_does_not_credit(self):
        rows = copy.deepcopy(self.rows)
        rows[24]['events'] = ['c04-d01']
        with self.assertRaisesRegex(ValueError, 'cannot receive frozen credit'):
            self.check(rows)

    def test_additional_requires_edit_and_phase(self):
        for field in ('proposed_edit', 'review_phase'):
            rows = copy.deepcopy(self.rows)
            rows[24].pop(field)
            with self.assertRaisesRegex(ValueError, 'cannot receive frozen credit'):
                self.check(rows)

    def test_two_occurrences_cover_two_distinct_events(self):
        self.assertEqual(self.rows[33]['events'], ['c04-d25', 'c04-d26'])
        _, found = m.metrics.page_metrics(self.report, self.rows, self.pages, self.events, True)
        self.assertEqual(found, {'c02-d12', 'c02-d13', 'c04-d09', 'c04-d25', 'c04-d26', 'c04-d28'})

    def test_partial_is_not_full(self):
        pages, _, events = self.sets['exposed_purpose']
        report = m.audit.read(m.ROOT/'reports/exposed_purpose/after/strict.json.gz')
        rows = m.audit.read(m.ROOT/'dispositions.json')['exposed_purpose']['strict']['after']
        _, found = m.metrics.page_metrics(report, rows, pages, events, True)
        self.assertEqual(rows[18]['partial_events'], ['c06-d02'])
        self.assertNotIn('c06-d02', found)

    def test_reused_source_rejected(self):
        with self.assertRaisesRegex(ValueError, 'Reused confirmation source'):
            m.metrics.validate_sources(dict(self.sets, duplicate=self.sets['confirmation']))

    def test_partial_exposure_cannot_be_hidden(self):
        record = m.audit.read(m.ROOT/'confirmation/source-overlap.json')
        record['known_partially_exposed_pages'] = []
        with self.assertRaisesRegex(ValueError, 'Hidden partial exposure'):
            m.validate_exposure(self.sets, record)

    def test_false_prior_overlap_rejected(self):
        record = m.audit.read(m.ROOT/'confirmation/source-overlap.json')
        record['matches'][0]['matches'][0]['reference'] = 'https://example.invalid/unreviewed'
        with self.assertRaisesRegex(ValueError, 'Exposure quote absent'):
            m.validate_exposure(self.sets, record)

    def test_exposure_subsets_keep_all_denominators(self):
        result = m.audit.read(m.ROOT/'summary.json')['confirmation']['technical']
        parts = m.exposure_metrics(result)
        self.assertEqual(parts['without_known_overlap']['after'], dict(pages=5, defects=24, detected=2))
        self.assertEqual(parts['known_partial_exposure']['after'], dict(pages=1, defects=33, detected=4))

    def test_known_abstention_is_not_hidden(self):
        report = m.audit.read(m.ROOT/'reports/exposed_instruction/after/strict.json.gz')
        m.validate_abstentions(report, 'exposed_instruction')
        with self.assertRaisesRegex(ValueError, 'Abstention evidence drift'):
            m.validate_abstentions(dict(report, abstentions=[]), 'exposed_instruction')
        with self.assertRaisesRegex(ValueError, 'Abstention evidence drift'):
            m.validate_abstentions(report, 'confirmation')


if __name__ == '__main__':
    unittest.main()
