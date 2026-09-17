#!/usr/bin/env python3
"""Reject missing findings, unsupported credit and hidden abstention evidence."""
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
        m.previous.check_rows(report, rows, self.pages, self.events, True)

    def test_complete_review(self):
        self.check(self.report, self.rows)

    def test_missing_unchanged_finding(self):
        with self.assertRaisesRegex(ValueError, 'Missing or unexpected'):
            self.check(self.report, self.rows[:-1])

    def test_finding_drift(self):
        report = copy.deepcopy(self.report)
        report['findings'][0]['message'] += ' changed'
        with self.assertRaisesRegex(ValueError, 'review drift'):
            self.check(report, self.rows)

    def test_cross_page_credit(self):
        rows = copy.deepcopy(self.rows)
        rows[132]['events'] = ['c02-d01']
        with self.assertRaisesRegex(ValueError, 'Unsupported credit'):
            self.check(self.report, rows)

    def test_incidental_length_overlap(self):
        rows = copy.deepcopy(self.rows)
        rows[7].update(status='actionable', events=['c03-d03'])
        with self.assertRaisesRegex(ValueError, 'Unsupported credit'):
            self.check(self.report, rows)

    def test_acceptable_control_cannot_receive_credit(self):
        rows = copy.deepcopy(self.rows)
        rows[5].update(status='actionable', events=['c03-c01'])
        with self.assertRaisesRegex(ValueError, 'Unsupported credit'):
            self.check(self.report, rows)

    def test_reused_confirmation_source(self):
        sets = m.data_sets()
        pages, files, events = sets['confirmation']
        sets['duplicated'] = (dict(copied=next(iter(pages.values()))), files, events)
        with self.assertRaisesRegex(ValueError, 'Reused confirmation source'):
            m.previous.validate_sources(sets)

    def test_quoted_source_drift(self):
        with self.assertRaisesRegex(ValueError, 'Source quote differs'):
            m.audit.span(b'original', dict(start=0, end=8, quote='rewritten'))

    def test_partial_match_does_not_increase_recall(self):
        report = m.audit.read(m.ROOT/'reports/confirmation/before/strict.json.gz')
        rows = m.audit.read(m.ROOT/'dispositions.json')['confirmation']['strict']['before']
        _, found = m.previous.page_metrics(report, rows, self.pages, self.events, True)
        self.assertEqual(found, {'c05-d02', 'c06-d05', 'c08-d14', 'c09-d06'})
        self.assertTrue(any('c09-d09' in r.get('partial_events', []) for r in rows))
        self.assertNotIn('c09-d09', found)

    def test_known_abstention_is_preserved(self):
        m.validate_abstentions(self.report, 'confirmation')
        for changed in ([], self.report['abstentions'] * 2):
            report = dict(self.report, abstentions=changed)
            with self.assertRaisesRegex(ValueError, 'Abstention evidence drift'):
                m.validate_abstentions(report, 'confirmation')

    def test_unexpected_abstention_cannot_hide_in_old_set(self):
        with self.assertRaisesRegex(ValueError, 'Abstention evidence drift'):
            m.validate_abstentions(self.report, 'development')

    def test_importance_semantics_identifies_framing_not_authorship(self):
        finding = self.report['findings'][121]
        self.assertTrue(m.previous.credits_match(finding, self.events['c08-d14'], self.pages))
        unrelated = dict(self.events['c08-d14'], category='needless_repetition')
        self.assertFalse(m.previous.credits_match(finding, unrelated, self.pages))


if __name__ == '__main__':
    unittest.main()
