"""Exercise the runner and collector with a fake CLI, never a model call."""
import contextlib
import copy
import json
from pathlib import Path
import tempfile
import unittest
from unittest import mock

import accounting as a
import evaluate_accounting as evaluation
import execute
from test_accounting import packet, response


class ExecutionTests(unittest.TestCase):
    def test_no_subprocess_without_new_authorization(self):
        with mock.patch.object(execute.subprocess, 'Popen') as launch:
            with self.assertRaisesRegex(ValueError, 'new explicit'):
                execute.execute(Path('/missing'), Path('/unused'))
            launch.assert_not_called()

    def simulate(self, bad_index=None):
        temporary = tempfile.TemporaryDirectory(); self.addCleanup(temporary.cleanup)
        root = Path(temporary.name); source = root / 'packet'; source.mkdir()
        (source / 'requests').mkdir(); (source / 'prompt.md').write_text('Review the source.')
        (source / 'freeze.json').write_text('{}')
        rows = []
        for r in a.requests(packet(['Condition holds.'] * 129)):
            path = source / 'requests' / (r['request'] + '.json'); a.write(path, r)
            rows.append({'request': r['request'], 'page': r['page'], 'reference': 'whole/example.md',
                         'request_sha256': a.digest(path.read_bytes())})
        protocol = {'model': 'test-only', 'effort': 'medium', 'total_timeout_seconds': 60,
                    'per_call_timeout_seconds': 5}
        manifest = {'requests': rows}; launched = []; output = root / 'run'

        class FakeProcess:
            returncode = 0

            def __init__(self, args, **_kwargs):
                self.args = args; launched.append(args)

            def communicate(self, message, timeout):
                request = json.loads(message.split(b'DOCUMENT JSON:\n', 1)[1])
                result = response(request)
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
        self.assertEqual(collected[0], {})
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
        ledger['complete'] = True; a.write(output / 'run.json', ledger)
        with mock.patch.object(execute, 'verify', return_value=(manifest, protocol)):
            with self.assertRaisesRegex(ValueError, 'False completion'):
                evaluation.collect(source, output)

    def test_changed_execution_flags_rejected_offline(self):
        source, output, manifest, protocol, *_ = self.simulate()
        ledger = a.read(output / 'run.json'); args = ledger['calls'][0]['command']
        args[args.index('features.shell_tool=false')] = 'features.shell_tool=true'
        a.write(output / 'run.json', ledger)
        with mock.patch.object(execute, 'verify', return_value=(manifest, protocol)):
            with self.assertRaisesRegex(ValueError, 'flags changed'): evaluation.collect(source, output)

    def test_tool_event_rejected(self):
        events = [{'type': 'thread.started'}, {'type': 'turn.started'},
                  {'type': 'item.completed', 'item': {'type': 'command_execution'}},
                  {'type': 'turn.completed', 'usage': {}}]
        with self.assertRaisesRegex(ValueError, 'tool'):
            execute.cli_compat.trace_usage('\n'.join(map(json.dumps, events)))

    def test_reference_denominator_remains_123(self):
        events, baseline = evaluation.prior_evaluate.reference()
        self.assertEqual(len(events), 123); self.assertEqual(set(events), set(baseline))


if __name__ == '__main__':
    unittest.main()
