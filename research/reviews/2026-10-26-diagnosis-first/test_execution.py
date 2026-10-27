"""Exercise the runner and collector with a fake CLI, never a model call."""
import contextlib
import copy
import json
from pathlib import Path
import tarfile
import tempfile
import unittest
from unittest import mock

import diagnosis as a
import evaluate as evaluation
import execute
from test_diagnosis import packet, response


def pilot_coverage(completed):
    """Represent all-no-burden valid responses on selected frozen pages."""
    units = {'context/c01.md': 23, 'context/c02.md': 45, 'context/c03.md': 116,
             'context/c04.md': 50, 'context/c05.md': 41, 'context/c06.md': 35}
    assessed = sum(units[page] for page in completed)
    valid_requests = sum((units[page] + a.MAX_UNITS - 1) // a.MAX_UNITS
                         for page in completed)
    return {'requests': 7, 'valid_requests': valid_requests, 'failed_requests': 0,
            'unattempted_requests': 7 - valid_requests,
            'pages': 6, 'complete_pages': len(completed), 'partial_pages': 0,
            'invalid_pages': 0, 'unattempted_pages': 6 - len(completed),
            'incomplete_pages': 6 - len(completed),
            'focus_units': 310, 'assessed_units': assessed,
            'unassessed_units': 310 - assessed,
            'burden_decisions': {'no_burden': assessed, 'diagnose': 0, 'uncertain': 0}}


class ExecutionTests(unittest.TestCase):
    def test_changed_frozen_prompt_rejected_before_any_model_call(self):
        with tempfile.TemporaryDirectory() as temporary:
            packet = Path(temporary)
            (packet / 'prompt.md').write_text('Review the source.')
            a.REFERENCE.freeze(packet)
            (packet / 'prompt.md').write_text('Changed prompt.')
            with mock.patch.object(execute.subprocess, 'Popen') as launch:
                with self.assertRaisesRegex(ValueError, 'Frozen input changed'):
                    execute.verify(packet)
                launch.assert_not_called()

    def test_no_subprocess_without_new_authorization(self):
        with mock.patch.object(execute.subprocess, 'Popen') as launch:
            with self.assertRaisesRegex(ValueError, 'new explicit'):
                execute.execute(Path('/missing'), Path('/unused'))
            launch.assert_not_called()

    def simulate(self, bad_index=None, uncertain=False):
        temporary = tempfile.TemporaryDirectory(); self.addCleanup(temporary.cleanup)
        root = Path(temporary.name); source = root / 'packet'; source.mkdir()
        (source / 'requests').mkdir(); (source / 'prompt.md').write_text('Review the source.')
        (source / 'freeze.json').write_text('{}')
        rows = []
        for r in a.OLD.requests(packet(['Condition holds.'] * 129)):
            path = source / 'requests' / (r['request'] + '.json'); a.write(path, r)
            rows.append({'request': r['request'], 'page': r['page'], 'reference': 'whole/example.md',
                         'request_sha256': a.digest(path.read_bytes())})
        protocol = {'model': 'test-only', 'effort': 'medium', 'total_timeout_seconds': 60,
                    'per_call_timeout_seconds': 5}
        references = (['whole/example.md'] + [f'whole/w{i:02}.md' for i in range(25)] +
                      list(a.PAGES) + [f'confirmation/c{i:02}.md' for i in range(4)])
        manifest = {'requests': rows, 'source_references': references}
        launched = []; output = root / 'run'

        class FakeProcess:
            returncode = 0

            def __init__(self, args, **_kwargs):
                self.args = args; launched.append(args)

            def communicate(self, message, timeout):
                request = json.loads(message.split(b'DOCUMENT JSON:\n', 1)[1])
                result = response(request)
                if uncertain and len(launched) == 1:
                    result['assessments'][0]['burden_decision'] = 'uncertain'
                if len(launched) == bad_index: result['assessments'].pop()
                Path(self.args[self.args.index('--output-last-message') + 1]).write_text(json.dumps(result))
                events = [{'type': 'thread.started'}, {'type': 'turn.started'},
                          {'type': 'item.completed', 'item': {
                              'type': 'agent_message', 'text': json.dumps(result)}},
                          {'type': 'turn.completed', 'usage': {'input_tokens': 1, 'output_tokens': 1}}]
                return '\n'.join(map(json.dumps, events)).encode(), b''

        with contextlib.ExitStack() as stack:
            stack.enter_context(mock.patch.object(execute, 'verify', return_value=(manifest, protocol)))
            stack.enter_context(mock.patch.object(execute.subprocess, 'check_output', return_value='fake-cli'))
            stack.enter_context(mock.patch.object(execute.subprocess, 'Popen', FakeProcess))
            success = execute.execute(source, output, authorized=True)
            collected = evaluation.collect(source, output)
        return source, output, manifest, protocol, launched, success, collected

    def test_valid_multi_request_page_and_tool_flags(self):
        _source, _output, _manifest, _protocol, launched, success, collected = self.simulate()
        self.assertTrue(success); self.assertEqual(len(launched), 3)
        self.assertEqual(collected[2], {'whole/example.md'})
        self.assertEqual(len(collected[3]), 36)
        self.assertEqual(collected[0], {})
        self.assertEqual(collected[4]['assessed_units'], 129)
        self.assertEqual(collected[4]['burden_decisions']['no_burden'], 129)
        self.assertEqual(collected[4]['incomplete_pages'], 0)
        self.assertEqual(collected[4]['invalid_pages'], 0)
        for args in launched:
            self.assertIn('features.shell_tool=false', args)
            self.assertIn('features.multi_agent=false', args)
            self.assertIn('orchestrator.mcp.enabled=false', args)
            self.assertIn('web_search="disabled"', args)

    def test_invalid_window_stops_without_retry_and_page_cannot_pass(self):
        source, output, manifest, protocol, launched, success, collected = self.simulate(bad_index=2)
        self.assertFalse(success); self.assertEqual(len(launched), 2)
        ledger = a.read(output / 'run.json')
        self.assertEqual(len(ledger['unattempted']), 1)
        self.assertEqual(collected[2], set())
        self.assertEqual(collected[4]['assessed_units'], 64)
        self.assertEqual(collected[4]['unassessed_units'], 65)
        self.assertEqual(collected[4]['partial_pages'], 1)
        self.assertEqual(collected[4]['invalid_pages'], 1)
        self.assertEqual(collected[4]['unattempted_pages'], 0)
        ledger['complete'] = True; a.write(output / 'run.json', ledger)
        with mock.patch.object(execute, 'verify', return_value=(manifest, protocol)):
            with self.assertRaisesRegex(ValueError, 'False completion'):
                evaluation.collect(source, output)

    def test_uncertain_unit_is_counted_as_abstention(self):
        _source, _output, _manifest, _protocol, _launched, success, collected = self.simulate(
            uncertain=True)
        self.assertTrue(success)
        self.assertEqual(collected[4]['assessed_units'], 129)
        self.assertEqual(collected[4]['burden_decisions'],
                         {'no_burden': 128, 'diagnose': 0, 'uncertain': 1})

    def test_failed_call_raw_trace_change_rejected_offline(self):
        source, output, manifest, protocol, *_ = self.simulate(bad_index=2)
        failed = output / manifest['requests'][1]['request'] / 'events.jsonl'
        failed.write_text(failed.read_text() + '\n')
        with mock.patch.object(execute, 'verify', return_value=(manifest, protocol)):
            with self.assertRaisesRegex(ValueError, 'Raw call artifact changed'):
                evaluation.collect(source, output)

    def test_changed_execution_flags_rejected_offline(self):
        source, output, manifest, protocol, *_ = self.simulate()
        ledger = a.read(output / 'run.json'); args = ledger['calls'][0]['command']
        args[args.index('features.shell_tool=false')] = 'features.shell_tool=true'
        a.write(output / 'run.json', ledger)
        with mock.patch.object(execute, 'verify', return_value=(manifest, protocol)):
            with self.assertRaisesRegex(ValueError, 'flags changed'): evaluation.collect(source, output)

    def test_other_packet_or_output_path_rejected_offline(self):
        source, output, manifest, protocol, *_ = self.simulate()
        original = a.read(output / 'run.json')
        changes = [('--output-schema',
                    source.parent / 'other-packet' / 'response-schema.json',
                    'another packet'),
                   ('--output-last-message',
                    output.parent / 'other-run' / original['calls'][0]['request'] / 'response.json',
                    'another output directory')]
        with mock.patch.object(execute, 'verify', return_value=(manifest, protocol)):
            for flag, replacement, message in changes:
                with self.subTest(flag=flag):
                    ledger = copy.deepcopy(original)
                    command = ledger['calls'][0]['command']
                    command[command.index(flag) + 1] = str(replacement)
                    a.write(output / 'run.json', ledger)
                    with self.assertRaisesRegex(ValueError, message):
                        evaluation.collect(source, output)

    def test_tool_event_rejected(self):
        events = [{'type': 'thread.started'}, {'type': 'turn.started'},
                  {'type': 'item.completed', 'item': {'type': 'command_execution'}},
                  {'type': 'turn.completed', 'usage': {}}]
        with self.assertRaisesRegex(ValueError, 'tool'):
            execute.cli_compat.trace_usage('\n'.join(map(json.dumps, events)))

    def test_reconnect_error_event_rejected(self):
        events = [{'type': 'thread.started'}, {'type': 'turn.started'},
                  {'type': 'error', 'message': 'invalid peer certificate: UnknownIssuer'},
                  {'type': 'turn.completed', 'usage': {}}]
        with self.assertRaises(ValueError):
            execute.cli_compat.trace_usage('\n'.join(map(json.dumps, events)))

    def test_reference_denominator_remains_123(self):
        events, baseline = evaluation.prior_evaluate.reference()
        self.assertEqual(len(events), 123); self.assertEqual(set(events), set(baseline))

    def test_unavailable_cohorts_keep_all_frozen_page_denominators(self):
        archive = a.ROOT.parents[2] / 'research/reviews/2026-09-26-contextual-evidence/evidence.tar.gz'
        with tarfile.open(archive, 'r:gz') as source:
            manifest = json.load(source.extractfile('packet/manifest.json'))
        page_paths = {row['reference'] for row in manifest['pages']}
        events, _ = evaluation.prior_evaluate.reference()
        review = {'reviewer': 'offline denominator test', 'findings': {},
                  'events': {key: {'status': 'missed' if event['path'] in a.PAGES
                                            else 'unavailable', 'findings': [],
                                   'rationale': 'No model finding is available.'}
                             for key, event in events.items()}}
        result = evaluation.score({}, {}, set(a.PAGES), page_paths, review,
                                  pilot_coverage(set(a.PAGES)))
        self.assertEqual({cohort: result['cohorts'][cohort]['pages']
                          for cohort in ['whole', 'context', 'confirmation']},
                         {'whole': 26, 'context': 6, 'confirmation': 4})
        self.assertEqual({cohort: result['cohorts'][cohort]['valid_pages']
                          for cohort in ['whole', 'context', 'confirmation']},
                         {'whole': 0, 'context': 6, 'confirmation': 0})
        self.assertEqual(sum(row['events'] for row in result['cohorts'].values()), 123)
        self.assertEqual(result['cohorts']['whole']['unavailable_events'] +
                         result['cohorts']['confirmation']['unavailable_events'], 77)
        self.assertIsNone(result['cohorts']['whole']['recall'])
        self.assertIsNone(result['cohorts']['confirmation']['recall'])
        self.assertEqual(result['cohorts']['context']['unavailable_events'], 0)
        self.assertFalse(result['pilot_screen_passed'])
        self.assertFalse(result['product_qualified'])
        self.assertEqual(result['pilot_coverage']['assessed_units'], 310)
        broken_coverage = pilot_coverage(set(a.PAGES))
        broken_coverage['burden_decisions']['uncertain'] = -1
        broken_coverage['burden_decisions']['no_burden'] = 311
        with self.assertRaisesRegex(ValueError, 'Invalid pilot coverage counter'):
            evaluation.score({}, {}, set(a.PAGES), page_paths, review, broken_coverage)
        with self.assertRaisesRegex(ValueError, 'Available pages or reference paths changed'):
            evaluation.score({}, {}, set(a.PAGES), set(a.PAGES), review,
                             pilot_coverage(set(a.PAGES)))
        outside = next(key for key, event in events.items() if event['path'] not in a.PAGES)
        review['events'][outside]['status'] = 'missed'
        with self.assertRaisesRegex(ValueError, 'Unavailable event cannot claim'):
            evaluation.score({}, {}, set(a.PAGES), page_paths, review,
                             pilot_coverage(set(a.PAGES)))
        review['events'][outside]['status'] = 'unavailable'
        selected = next(key for key, event in events.items() if event['path'] in a.PAGES)
        review['events'][selected]['status'] = 'unavailable'
        with self.assertRaisesRegex(ValueError, 'Available event needs'):
            evaluation.score({}, {}, set(a.PAGES), page_paths, review,
                             pilot_coverage(set(a.PAGES)))

    def test_incomplete_pilot_marks_unseen_events_unavailable(self):
        archive = a.ROOT.parents[2] / 'research/reviews/2026-09-26-contextual-evidence/evidence.tar.gz'
        with tarfile.open(archive, 'r:gz') as source:
            manifest = json.load(source.extractfile('packet/manifest.json'))
        page_paths = {row['reference'] for row in manifest['pages']}
        events, _ = evaluation.prior_evaluate.reference()
        completed = set(a.PAGES[:-1])
        review = {'reviewer': 'offline incomplete-pilot test', 'findings': {},
                  'events': {key: {'status': 'missed' if event['path'] in completed
                                            else 'unavailable', 'findings': [],
                                   'rationale': 'No model finding is available.'}
                             for key, event in events.items()}}
        result = evaluation.score({}, {}, completed, page_paths, review,
                                  pilot_coverage(completed))
        context = result['cohorts']['context']
        self.assertEqual(context['valid_pages'], 5)
        self.assertGreater(context['unavailable_events'], 0)
        self.assertEqual(context['available_events'] + context['unavailable_events'], 46)
        self.assertIsNone(context['recall'])
        self.assertFalse(result['pilot_screen_passed'])

    def test_repair_safety_cannot_be_claimed_for_diagnosis_only(self):
        with mock.patch.object(evaluation.prior_evaluate, 'reference', return_value=({}, {})):
            with self.assertRaisesRegex(ValueError, 'cannot claim repair safety'):
                evaluation.score({'f1': {'path': 'context/c01.md'}}, {}, set(),
                                 {'context/c01.md'}, {'findings': {
                    'f1': {'status': 'accepted', 'rationale': 'Source supports it.',
                           'suggestion_safety': {'status': 'safe', 'rationale': 'Invented.'}}}},
                                 pilot_coverage(set()))

    def test_pilot_gate_requires_recall_quality_and_all_pages(self):
        predictions = {f'f{i}': {'path': 'context/c01.md'} for i in range(20)}
        review = {'events': {}, 'findings': {key: {
            'status': 'accepted' if i < 17 else 'uncertain',
            'rationale': 'Reviewed against source.',
            'suggestion_safety': {'status': 'not_applicable',
                                  'rationale': 'No suggestion was emitted.'}}
            for i, key in enumerate(predictions)}}
        page_paths = set(a.PAGES)
        page_paths.update(f'whole/w{i:02}.md' for i in range(26))
        page_paths.update(f'confirmation/c{i:02}.md' for i in range(4))
        result = {'cohorts': {c: {'pages': {'whole': 26, 'context': 6,
                                           'confirmation': 4}[c],
                                 'events': 46 if c == 'context' else 0,
                                 'valid_pages': 6 if c == 'context' else 0,
                                 'full': 37 if c == 'context' else 0,
                                 'findings': 20 if c == 'context' else 0}
                              for c in ['whole', 'context', 'confirmation']}}
        with mock.patch.object(evaluation.prior_evaluate, 'reference', return_value=({}, {})), \
                mock.patch.object(evaluation.prior_evaluate, 'score', return_value=result):
            first = evaluation.score(predictions, {}, set(a.PAGES), page_paths, review,
                                     pilot_coverage(set(a.PAGES)))
            self.assertTrue(first['pilot_screen_passed'])
            self.assertEqual(first['all_valid_request_emissions']['context']['accepted'], 17)
            self.assertEqual(first['all_valid_request_emissions']['context']['uncertain'], 3)
            review['findings']['f0']['status'] = 'uncertain'
            self.assertFalse(evaluation.score(predictions, {}, set(a.PAGES),
                                              page_paths, review,
                                              pilot_coverage(set(a.PAGES)))['pilot_screen_passed'])
            review['findings']['f0']['status'] = 'accepted'
            result['cohorts']['context']['full'] = 36
            self.assertFalse(evaluation.score(predictions, {}, set(a.PAGES),
                                              page_paths, review,
                                              pilot_coverage(set(a.PAGES)))['pilot_screen_passed'])
            result['cohorts']['context']['full'] = 37
            result['cohorts']['context']['valid_pages'] = 5
            self.assertFalse(evaluation.score(predictions, {}, set(a.PAGES),
                                              page_paths, review,
                                              pilot_coverage(set(a.PAGES)))['pilot_screen_passed'])


if __name__ == '__main__':
    unittest.main()
