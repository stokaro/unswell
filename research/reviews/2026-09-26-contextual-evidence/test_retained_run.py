"""Offline negative tests for the retained experiment, never model calls."""
import io
import json
from pathlib import Path
import tarfile
import tempfile
import unittest

from study import ROOT, digest
from verify_run import verify


@unittest.skipUnless((ROOT / 'evidence.tar.gz').exists(), 'Terminal evidence has not been packaged')
class RetainedRunTests(unittest.TestCase):
    def altered(self, change):
        with tarfile.open(ROOT / 'evidence.tar.gz') as source:
            files = {m.name: source.extractfile(m).read() for m in source.getmembers()}
        change(files)
        # Refresh only the outer packaging checksums, so inner checks must reject the mutation.
        files['archive-manifest.json'] = json.dumps({
            name: digest(raw) for name, raw in files.items() if name != 'archive-manifest.json'
        }).encode()
        with tempfile.TemporaryDirectory() as temporary:
            path = Path(temporary) / 'changed.tar.gz'
            with tarfile.open(path, 'w:gz') as archive:
                for name, raw in files.items():
                    member = tarfile.TarInfo(name); member.size = len(raw)
                    archive.addfile(member, io.BytesIO(raw))
            return verify(path)

    def test_relocated_evidence_reproduces(self):
        result = verify(ROOT / 'evidence.tar.gz')
        self.assertEqual(sum(c['events'] for c in result['cohorts'].values()), 123)
        self.assertFalse(result['product_qualified'])

    def test_enabled_tool_flag_is_rejected(self):
        def change(files):
            ledger = json.loads(files['run/run.json'])
            args = ledger['calls'][1]['command']
            args[args.index('features.shell_tool=false')] = 'features.shell_tool=true'
            files['run/run.json'] = json.dumps(ledger).encode()
        with self.assertRaisesRegex(ValueError, 'Execution flags changed'):
            self.altered(change)

    def test_tool_event_is_rejected(self):
        def change(files):
            key = 'run/page-002/events.jsonl'
            events = [json.loads(line) for line in files[key].splitlines()]
            events.insert(-1, {'type': 'item.completed', 'item': {
                'id': 'unexpected-tool', 'type': 'command_execution', 'command': 'pwd'}})
            files[key] = ('\n'.join(json.dumps(e) for e in events) + '\n').encode()
        with self.assertRaisesRegex(ValueError, 'tool or unexpected item'):
            self.altered(change)

    def test_dropped_frozen_event_is_rejected(self):
        def change(files):
            review = json.loads(files['run/review.json'])
            review['events'].pop(next(iter(review['events'])))
            files['run/review.json'] = json.dumps(review).encode()
        with self.assertRaisesRegex(ValueError, 'Every frozen event'):
            self.altered(change)

    def test_invalid_quote_cannot_be_silently_repaired(self):
        def change(files):
            key = 'run/page-019/response.json'
            response = json.loads(files[key])
            target = response['findings'][2]['targets'][0]
            self.assertTrue(target['quote'].startswith('**Deployment reports**'))
            target['quote'] = target['quote'][2:]
            files[key] = json.dumps(response).encode()
            key = 'run/page-019/events.jsonl'
            events = [json.loads(line) for line in files[key].splitlines()]
            for event in events:
                if event.get('type') == 'item.completed' and event['item']['type'] == 'agent_message':
                    event['item']['text'] = json.dumps(response)
            files[key] = ('\n'.join(json.dumps(e) for e in events) + '\n').encode()
        with self.assertRaisesRegex(ValueError, 'Recorded quotation failure no longer occurs'):
            self.altered(change)

    def test_incomplete_run_cannot_claim_completion(self):
        def change(files):
            ledger = json.loads(files['run/run.json']); ledger['complete'] = True
            files['run/run.json'] = json.dumps(ledger).encode()
        with self.assertRaisesRegex(ValueError, 'Completion ledger differs'):
            self.altered(change)


if __name__ == '__main__':
    unittest.main()
