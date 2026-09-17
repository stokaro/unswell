#!/usr/bin/env python3
"""Negative probes for incomplete and unsupported editorial evidence."""
import copy
from pathlib import Path
import shutil
import sys
import tempfile
import unittest

sys.dont_write_bytecode = True
import summarize as m


class EvidenceTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.pages, _, cls.events = m.data_sets()['confirmation']
        cls.report = m.audit.read(m.ROOT/'reports/confirmation/after/strict.json.gz')
        cls.rows = m.audit.read(m.ROOT/'dispositions.json')['confirmation']['strict']['after']

    def test_complete_review(self):
        m.check_rows(self.report, self.rows, self.pages, self.events, True)

    def test_missing_unchanged_finding(self):
        with self.assertRaisesRegex(ValueError, 'Missing or unexpected'):
            m.check_rows(self.report, self.rows[:-1], self.pages, self.events, True)

    def test_finding_drift(self):
        report = copy.deepcopy(self.report)
        report['findings'][0]['message'] += ' changed'
        with self.assertRaisesRegex(ValueError, 'review drift'):
            m.check_rows(report, self.rows, self.pages, self.events, True)

    def test_cross_page_credit(self):
        rows = copy.deepcopy(self.rows)
        rows[35]['events'] = ['c01-d01']
        with self.assertRaisesRegex(ValueError, 'Unsupported credit'):
            m.check_rows(self.report, rows, self.pages, self.events, True)

    def test_incidental_length_overlap(self):
        rows = copy.deepcopy(self.rows)
        rows[2].update(status='actionable', events=['c04-d03'])
        with self.assertRaisesRegex(ValueError, 'Unsupported credit'):
            m.check_rows(self.report, rows, self.pages, self.events, True)

    def test_reused_confirmation_source(self):
        sets = m.data_sets()
        pages, files, events = sets['confirmation']
        sets['duplicated'] = (dict(copied=next(iter(pages.values()))), files, events)
        with self.assertRaisesRegex(ValueError, 'Reused confirmation source'):
            m.validate_sources(sets)

    def test_quoted_source_drift(self):
        with self.assertRaisesRegex(ValueError, 'Source quote differs'):
            m.audit.span(b'original', dict(start=0, end=8, quote='rewritten'))

    def test_probe_cannot_change_its_result_count(self):
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)/'probe'
            shutil.copytree(m.ROOT/'all-rules-probe', root)
            summary = m.audit.read(root/'summary.json')
            summary['sets'][0]['findings'] += 1
            m.audit.write(root/'summary.json', summary)
            with self.assertRaisesRegex(ValueError, 'Probe summary drift'):
                m.validate_probe(root)

    def test_partial_match_does_not_increase_recall(self):
        pages, _, events = m.data_sets()['exposed_repetition']
        report = m.audit.read(m.ROOT/'reports/exposed_repetition/after/technical.json.gz')
        rows = m.audit.read(m.ROOT/'dispositions.json')['exposed_repetition']['technical']['after']
        _, found = m.page_metrics(report, rows, pages, events, False)
        self.assertEqual(found, {'c04-d01', 'c04-d02', 'c06-d01'})
        self.assertTrue(any('c06-d02' in r.get('partial_events', []) for r in rows))


if __name__ == '__main__':
    unittest.main()
