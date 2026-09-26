import copy
import json
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch

import run
import study


class EvidenceContractTests(unittest.TestCase):
    def setUp(self):
        text = 'A café is described here. The same words. The same words. `DROP TABLE`'
        data = text.encode()
        self.packet = {'page': 'p', 'source': text, 'source_sha256': study.digest(data),
                       'units': [{'id': 'u0', 'span': {'start': 0, 'end': len(data)}}],
                       'protected': [{'span': {'start': data.index(b'`'), 'end': len(data)},
                                      'reason': 'code'}]}
        self.response = {'page': 'p', 'coverage': [{'unit': 'u0', 'status': 'reviewed'}],
                         'findings': [{'id': 'f1', 'category': 'needless_repetition',
                                       'targets': [{'unit': 'u0', 'quote': 'The same words.', 'occurrence': 2}],
                                       'diagnostic': 'Repeated assertion.',
                                       'why_revision_needed': 'The two statements are identical.',
                                       'suggestion': 'Remove the second statement.',
                                       'meaning_preservation': 'Retain the original statement.'}]}

    def test_unicode_and_repeated_occurrence(self):
        result = study.validate(self.packet, self.response)
        target = result['findings'][0]['targets'][0]
        raw = self.packet['source'].encode()
        self.assertEqual(target['span']['start'], raw.rindex(b'The same words.'))
        self.assertEqual(raw[target['span']['start']:target['span']['end']].decode(), target['quote'])

    def test_source_drift(self):
        self.packet['source'] += ' changed'
        with self.assertRaises(ValueError): study.validate(self.packet, self.response)

    def test_coverage_is_complete(self):
        for coverage in [[], [{'unit': 'invented', 'status': 'reviewed'}],
                         self.response['coverage'] * 2]:
            with self.subTest(coverage=coverage):
                response = copy.deepcopy(self.response); response['coverage'] = coverage
                with self.assertRaises(ValueError): study.validate(self.packet, response)

    def test_protected_target_is_rejected(self):
        self.response['findings'][0]['targets'] = [{'unit': 'u0', 'quote': 'DROP TABLE', 'occurrence': 1}]
        with self.assertRaises(ValueError): study.validate(self.packet, self.response)

    def test_unknown_unit_and_fabricated_quote(self):
        for key, value in [('unit', 'wrong'), ('quote', 'fabricated'), ('occurrence', 3), ('occurrence', True)]:
            with self.subTest(key=key, value=value):
                response = copy.deepcopy(self.response); response['findings'][0]['targets'][0][key] = value
                with self.assertRaises(ValueError): study.validate(self.packet, response)

    def test_missing_explanation_is_not_a_diagnostic(self):
        self.response['findings'][0]['why_revision_needed'] = ''
        with self.assertRaises(ValueError): study.validate(self.packet, self.response)

    def test_unexpected_fields_and_duplicate_ids(self):
        response = copy.deepcopy(self.response); response['score'] = 0.99
        with self.assertRaises(ValueError): study.validate(self.packet, response)
        self.response['findings'] *= 2
        with self.assertRaises(ValueError): study.validate(self.packet, self.response)

    def test_empty_findings_are_valid_but_not_missing_coverage(self):
        self.response['findings'] = []
        self.assertEqual(study.validate(self.packet, self.response)['findings'], [])

    def test_freeze_rejects_changed_and_added_files(self):
        with tempfile.TemporaryDirectory() as name:
            root = Path(name); (root / 'a').write_text('one'); study.freeze(root)
            study.verify_freeze(root)
            (root / 'a').write_text('two')
            with self.assertRaises(ValueError): study.verify_freeze(root)
            (root / 'a').write_text('one'); (root / 'extra').write_text('x')
            with self.assertRaises(ValueError): study.verify_freeze(root)

    def test_execution_requires_explicit_authorization(self):
        with patch('run.subprocess.Popen') as launch:
            with self.assertRaises(ValueError): run.execute(Path('/missing'), Path('/missing-output'))
            launch.assert_not_called()

    def test_trace_rejects_tools_failures_and_missing_completion(self):
        complete = {'type': 'turn.completed', 'usage': {'input_tokens': 1, 'output_tokens': 2}}
        self.assertEqual(run.trace_usage(json.dumps(complete)), complete['usage'])
        for bad in [{'type': 'turn.failed'}, {'type': 'function_call'},
                    {'type': 'item.completed', 'item': {'type': 'command_execution'}}]:
            with self.subTest(event=bad):
                with self.assertRaises(ValueError): run.trace_usage(json.dumps(bad) + '\n' + json.dumps(complete))
        for raw in ['', json.dumps(complete) + '\n' + json.dumps(complete)]:
            with self.assertRaises(ValueError): run.trace_usage(raw)

    def test_command_uses_existing_login_without_home_override(self):
        args = run.command(Path('/packet'), Path('/output'), Path('/empty'),
                           {'model': 'gpt-6-astra', 'effort': 'medium'})
        self.assertIn('--ignore-user-config', args)
        self.assertIn('features.shell_tool=false', args)
        self.assertIn('project_doc_max_bytes=0', args)
        self.assertIn('features.skip_host_skill_discovery=true', args)
        self.assertNotIn('--dangerously-bypass-approvals-and-sandbox', args)
        self.assertFalse(any('CODEX_HOME' in arg for arg in args))


if __name__ == '__main__':
    unittest.main()
