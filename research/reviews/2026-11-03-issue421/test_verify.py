"""Negative contract tests for the read-only #421 evidence replay."""

import copy
import hashlib
import json
from pathlib import Path
import tempfile
import unittest

import verify


class EvidenceContractTest(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name)
        self.sources = {}
        self.raw = {}
        self.manifest = {'sources': []}
        self.report = {'documents': [], 'findings': []}
        for number in range(1, 37):
            path = f'whole/p{number:03}.md'
            body = (f'page {number:03} ' + 'x' * 200).encode()
            digest = hashlib.sha256(body).hexdigest()
            source_path = self.root / 'inputs' / path
            source_path.parent.mkdir(parents=True, exist_ok=True)
            source_path.write_bytes(body)
            source = {'path': 'archive/inputs/' + path, 'source_sha256': digest,
                      'bytes': len(body), 'repository': 'example/repo'}
            self.manifest['sources'].append(source)
            self.report['documents'].append({'name': 'inputs/' + path,
                                             'source_hash': digest,
                                             'bytes': len(body)})
            self.sources[path] = source
            self.raw[path] = body

    def test_source_drift_and_missing_page_rejected(self):
        sources, raw = verify.source_identity(self.report, self.manifest,
                                               self.root)
        self.assertEqual(len(sources), len(raw))
        path = self.root / 'inputs/whole/p001.md'
        path.write_bytes(b'changed')
        with self.assertRaisesRegex(ValueError, 'Source hash'):
            verify.source_identity(self.report, self.manifest, self.root)
        path.write_bytes(self.raw['whole/p001.md'])
        missing = copy.deepcopy(self.manifest)
        missing['sources'].pop()
        with self.assertRaisesRegex(ValueError, 'Missing source page'):
            verify.source_identity(self.report, missing, self.root)

    def test_hidden_published_delivery_rejected(self):
        path = 'whole/p001.md'
        published = {'group': 'original', 'profile': 'technical', 'pages': []}
        for source_path, source in self.sources.items():
            published['pages'].append({'path': source_path,
                                       'sourceSha256': source['source_sha256'],
                                       'report': {'findings': []}})
        for number in range(154):
            message = f'message {number}'
            self.report['findings'].append({
                'id': f'f{number}', 'rule_id': 'syntax.long-sentence',
                'primary': {'path': 'inputs/' + path,
                            'span': {'start': 0, 'end': 4}},
                'message': message})
            published['pages'][0]['report']['findings'].append({
                'ruleId': 'syntax.long-sentence', 'start': 0, 'end': 4,
                'message': message})
        self.assertEqual(verify.compare_published(self.report, published,
                                                  self.sources, self.raw), 154)
        published['pages'][0]['report']['findings'].pop()
        with self.assertRaisesRegex(ValueError, 'Hidden, missing'):
            verify.compare_published(self.report, published,
                                     self.sources, self.raw)

    def event_fixture(self):
        events = {}
        for part, count in verify.DENOMINATORS.items():
            for number in range(count):
                event_id = f'{part}/event-{number:03}'
                events[event_id] = {'original_event': {
                    'path': f'{part}/p001.md',
                    'targets': [{'start': 0, 'end': 1, 'quote': 'x'}]},
                    'after': {'status': 'missed', 'finding_ids': []}}
        sources = {f'{part}/p001.md': {'source_sha256': 'x'}
                   for part in verify.DENOMINATORS}
        raw = {path: b'x' for path in sources}
        findings = {'f1': {'primary': {'path': 'inputs/whole/p001.md'}}}
        return events, findings, sources, raw

    def test_denominator_unsupported_full_and_incompatible_join_rejected(self):
        events, findings, sources, raw = self.event_fixture()
        verify.event_impact(events, findings, set(), sources, raw)
        missing = copy.deepcopy(events)
        missing.pop(next(iter(missing)))
        with self.assertRaisesRegex(ValueError, 'denominator'):
            verify.event_impact(missing, findings, set(), sources, raw)
        unsupported = copy.deepcopy(events)
        unsupported['whole/event-000']['after']['status'] = 'full'
        with self.assertRaisesRegex(ValueError, 'lacks witness'):
            verify.event_impact(unsupported, findings, set(), sources, raw)
        wrong_page = copy.deepcopy(events)
        wrong_page['whole/event-000']['original_event']['path'] = 'whole/unknown.md'
        with self.assertRaisesRegex(ValueError, 'page or cohort'):
            verify.event_impact(wrong_page, findings, set(), sources, raw)

    def test_long_grade_and_identity_change_rejected(self):
        findings = {}
        rows = []
        for number in range(154):
            identifier = f'f{number}'
            rule = 'syntax.long-sentence' if number < 97 else 'other.rule'
            finding = {'id': identifier, 'rule_id': rule,
                       'primary': {'path': 'inputs/whole/p001.md',
                                   'span': {'start': 0, 'end': 1}}}
            findings[identifier] = finding
            if number < 97:
                rows.append({'target': {'ID': identifier,
                                        'Path': 'inputs/whole/p001.md',
                                        'Span': {'start': 0, 'end': 1},
                                        'Judgment': 'accepted' if number < 52
                                        else 'uncertain'}})
        report = {'findings': list(findings.values())}
        observations = {'complete': True, 'rows': rows}
        verify.long_sentence_rows(report, observations)
        observations['rows'][0]['target']['Judgment'] = 'rejected'
        with self.assertRaisesRegex(ValueError, 'identity or grade'):
            verify.long_sentence_rows(report, observations)


if __name__ == '__main__':
    unittest.main()
