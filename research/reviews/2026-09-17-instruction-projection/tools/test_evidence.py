#!/usr/bin/env python3
"""Check rejection of invalid evidence and preservation of partial results."""
import copy
import sys
import unittest
sys.dont_write_bytecode = True
import summarize as m


class EvidenceTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.pages, _, cls.events = m.data_sets()['confirmation']
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
        delta = m.finding_delta(self.report, report)
        self.assertEqual(delta, dict(added=[5], removed=[5], changed_at_same_location=1))
        with self.assertRaisesRegex(ValueError, 'review drift'):
            self.check(report, self.rows)

    def test_cross_page_credit(self):
        rows = copy.deepcopy(self.rows)
        rows[5].update(status='actionable', events=['c02-d01'])
        with self.assertRaisesRegex(ValueError, 'Unsupported credit'):
            self.check(self.report, rows)

    def test_incidental_length_overlap(self):
        rows = copy.deepcopy(self.rows)
        rows[9].update(status='actionable', events=['c02-d07'])
        with self.assertRaisesRegex(ValueError, 'Unsupported credit'):
            self.check(self.report, rows)

    def test_controls_cannot_receive_credit(self):
        rows = copy.deepcopy(self.rows)
        rows[5].update(status='actionable', events=['c01-c01'])
        with self.assertRaisesRegex(ValueError, 'Unsupported credit'):
            self.check(self.report, rows)

    def test_reused_source(self):
        sets = m.data_sets()
        sets['duplicate'] = sets['confirmation']
        with self.assertRaisesRegex(ValueError, 'Reused confirmation source'):
            m.metrics.validate_sources(sets)

    def test_source_quote_drift(self):
        with self.assertRaisesRegex(ValueError, 'Source quote differs'):
            m.audit.span(b'original', dict(start=0, end=8, quote='rewritten'))

    def test_partial_is_not_full_recall(self):
        _, found = m.metrics.page_metrics(self.report, self.rows, self.pages, self.events, True)
        self.assertTrue(any('c01-d03' in r.get('partial_events', []) for r in self.rows))
        self.assertNotIn('c01-d03', found)

    def test_related_method_changes_partial_to_full(self):
        pages, _, events = m.data_sets()['exposed_relations']
        evidence = m.audit.read(m.ROOT/'dispositions.json')['exposed_relations']['strict']
        for phase, expected in [('before', False), ('after', True)]:
            report = m.audit.read(m.ROOT/'reports/exposed_relations'/phase/'strict.json.gz')
            _, found = m.metrics.page_metrics(report, evidence[phase], pages, events, True)
            self.assertEqual('c06-d05' in found, expected)

    def additional_case(self):
        pages, _, events = m.data_sets()['development']
        report = m.audit.read(m.ROOT/'reports/development/after/strict.json.gz')
        rows = m.audit.read(m.ROOT/'dispositions.json')['development']['strict']['after']
        row = next(r for r in rows if r['status'] == 'additional_actionable')
        return pages, events, report, rows, row

    def test_additional_finding_cannot_increase_recall(self):
        pages, events, report, rows, row = self.additional_case()
        row['events'] = [next(iter(events))]
        with self.assertRaisesRegex(ValueError, 'cannot receive frozen credit'):
            m.metrics.check_rows(report, rows, pages, events, True)

    def test_additional_finding_requires_edit_and_phase(self):
        for field in ('proposed_edit', 'review_phase'):
            pages, events, report, rows, row = self.additional_case()
            row.pop(field)
            with self.assertRaisesRegex(ValueError, 'cannot receive frozen credit'):
                m.metrics.check_rows(report, rows, pages, events, True)

    def test_known_budget_abstention(self):
        report = m.audit.read(m.ROOT/'reports/exposed_instruction/after/strict.json.gz')
        m.validate_abstentions(report, 'exposed_instruction')
        with self.assertRaisesRegex(ValueError, 'Abstention evidence drift'):
            m.validate_abstentions(dict(report, abstentions=[]), 'exposed_instruction')
        with self.assertRaisesRegex(ValueError, 'Abstention evidence drift'):
            m.validate_abstentions(report, 'confirmation')


if __name__ == '__main__':
    unittest.main()
