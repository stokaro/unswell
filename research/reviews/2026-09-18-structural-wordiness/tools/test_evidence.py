#!/usr/bin/env python3
"""Reject invented detection credit, missing review and changed frozen inputs."""
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
        m.evaluate()

    def test_missing_finding(self):
        with self.assertRaisesRegex(ValueError, 'Missing or unexpected'):
            self.check(self.rows[:-1])

    def test_changed_advice_requires_review(self):
        report = copy.deepcopy(self.report)
        report['findings'][0]['evidence']['suggestion'] = 'changed'
        self.assertEqual(m.finding_delta(self.report, report), dict(added=[0], removed=[0], changed_at_same_location=1))
        with self.assertRaisesRegex(ValueError, 'review drift'):
            self.check(self.rows, report)

    def test_control_and_uncertainty_cannot_receive_credit(self):
        for kind in ('controls', 'uncertain'):
            rows = copy.deepcopy(self.rows)
            eid = next(eid for eid, e in self.events.items() if e['kind'] == kind)
            rows[0].update(status='actionable', events=[eid])
            with self.subTest(kind=kind), self.assertRaisesRegex(ValueError, 'Unsupported credit'):
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
        event = next(e for e in self.events.values() if e['context'] and len(e['context'][0]['quote'].split()) >= 12)
        record['matches'] = [dict(event['context'][0], page=event['page'], matches=[dict(reference='https://example.invalid/unreviewed')])]
        record['known_partially_exposed_pages'] = [event['page']]
        with self.assertRaisesRegex(ValueError, 'Exposure quote absent'):
            m.validate_exposure(self.sets, record)

    def test_unexpected_abstention(self):
        m.validate_abstentions(self.report, 'confirmation')
        with self.assertRaisesRegex(ValueError, 'Unexpected abstention'):
            m.validate_abstentions(dict(self.report, abstentions=[dict(reason='budget_exhausted')]), 'confirmation')

    def test_frozen_runtime(self):
        record = m.audit.read(m.ROOT/'code-freeze.json')
        m.code_freeze(record)
        record['sha256']['builtin/structural_support.go'] = '0'*64
        with self.assertRaisesRegex(ValueError, 'Runtime changed'):
            m.code_freeze(record)

    def test_partial_repairs_do_not_become_full_credit(self):
        _, found = m.metrics.page_metrics(self.report, self.rows, self.pages, self.events, True)
        self.assertEqual(found, {'c08-d04', 'c08-d05'})
        partial = {eid for r in self.rows for eid in r.get('partial_events', [])}
        self.assertEqual(partial, {'c01-d05', 'c07-d01', 'c09-d08', 'c09-d09'})
        self.assertFalse(partial & found)

    def test_additional_repair_cannot_increase_recall(self):
        rows = copy.deepcopy(self.rows)
        next(r for r in rows if r['status'] == 'additional_actionable')['events'] = ['c08-d05']
        with self.assertRaisesRegex(ValueError, 'cannot receive frozen credit'):
            self.check(rows)

    def test_case_folded_operator_is_not_editorial_credit(self):
        rows = [r for r in self.rows if self.report['findings'][r['index']]['primary']['snippet'] == 'AND']
        self.assertEqual(len(rows), 3)
        self.assertTrue(all(r['status'] == 'nonactionable' and not r['events'] for r in rows))

    def test_rejected_cleft_is_absent_from_frozen_candidate(self):
        record = m.audit.read(m.ROOT/'code-freeze.json')
        source = m.audit.archive_sources(m.ROOT/'runtime-inputs.tar.gz')
        self.assertNotIn('builtin/structural_cleft.go', record['sha256'])
        self.assertNotIn(b'syntax.subject-cleft', source['builtin/structural_catalog.go'])


if __name__ == '__main__':
    unittest.main()
