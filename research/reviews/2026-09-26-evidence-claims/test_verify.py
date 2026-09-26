"""Negative probes for the evidence boundary, not editorial-label generation."""
import hashlib
import json
from pathlib import Path
import shutil
import tempfile
import unittest

import verify


class EvidenceTests(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name) / 'record'
        shutil.copytree(verify.ROOT, self.root)

    def update_review(self, change):
        path = self.root / 'review.json'
        data = json.loads(path.read_text())
        change(data)
        path.write_text(json.dumps(data))
        manifest_path = self.root / 'artifacts.json'
        manifest = json.loads(manifest_path.read_text())
        manifest['files']['review.json'] = hashlib.sha256(path.read_bytes()).hexdigest()
        manifest_path.write_text(json.dumps(manifest))

    def test_complete_record(self):
        result = verify.verify(self.root)
        self.assertEqual(result['technical']['development']['matched'], {'before': 13, 'after': 15})

    def test_changed_source_rejected(self):
        with (self.root / 'confirmation-inputs.json').open('a') as output:
            output.write(' ')
        with self.assertRaisesRegex(ValueError, 'Changed frozen file'):
            verify.verify(self.root)

    def test_omitted_event_rejected(self):
        self.update_review(lambda data: data['events'].pop('p024-d01'))
        with self.assertRaisesRegex(ValueError, 'Incomplete event ledger'):
            verify.verify(self.root)

    def test_invented_credit_rejected(self):
        self.update_review(lambda data: data['events']['p001-d01'].update(before='full', rule_id='filler.unscoped-assurance'))
        with self.assertRaisesRegex(ValueError, 'Unbacked full match'):
            verify.verify(self.root)

    def test_missing_addition_review_rejected(self):
        self.update_review(lambda data: data['additions'].pop())
        with self.assertRaisesRegex(ValueError, 'Unreviewed additional finding'):
            verify.verify(self.root)

    def test_missing_confirmation_review_rejected(self):
        self.update_review(lambda data: data['confirmation']['strict'].pop())
        with self.assertRaisesRegex(ValueError, 'Missing confirmation disposition'):
            verify.verify(self.root)


if __name__ == '__main__':
    unittest.main()
