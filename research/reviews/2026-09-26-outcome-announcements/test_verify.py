"""Negative probes reject evidence loss and unsupported recall credit."""
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

    def change_review(self, change):
        path = self.root / 'review.json'
        data = json.loads(path.read_text())
        change(data['profiles']['technical'])
        path.write_text(json.dumps(data))
        manifest = self.root / 'artifacts.json'
        values = json.loads(manifest.read_text())
        values['files']['review.json'] = hashlib.sha256(path.read_bytes()).hexdigest()
        manifest.write_text(json.dumps(values))

    def test_complete_record(self):
        result = verify.verify(self.root)['technical']
        self.assertEqual(result['cohorts']['context'], {'events': 46, 'before': 4, 'after': 5})
        self.assertEqual(result['cohorts']['confirmation'], {'events': 20, 'before': 1, 'after': 1})
        self.assertEqual(result['confirmation_exposure_strata']['no_known_prior_review'],
                         {'pages': 2, 'events': 8, 'before': 0, 'after': 0})

    def test_changed_source(self):
        path = self.root / 'confirmation-inputs.json'
        path.write_text(path.read_text() + ' ')
        with self.assertRaisesRegex(ValueError, 'Changed artifact'):
            verify.verify(self.root)

    def test_omitted_event(self):
        self.change_review(lambda ledger: ledger['events'].pop('context/c04-d11'))
        with self.assertRaisesRegex(ValueError, 'Incomplete event ledger'):
            verify.verify(self.root)

    def test_unbacked_credit(self):
        self.change_review(lambda ledger: ledger['events']['confirmation/c01-r01']['after'].update(status='full'))
        with self.assertRaisesRegex(ValueError, 'Unsupported event credit'):
            verify.verify(self.root)

    def test_incidental_length_is_not_rhetorical_credit(self):
        self.change_review(lambda ledger: ledger['events']['confirmation/c02-r03']['after'].update(
            status='full', findings=['a48989b32dcc8ab6']))
        with self.assertRaisesRegex(ValueError, 'Incidental signal'):
            verify.verify(self.root)

    def test_missing_addition_review(self):
        self.change_review(lambda ledger: ledger['additions'].clear())
        with self.assertRaisesRegex(ValueError, 'Unreviewed addition'):
            verify.verify(self.root)

    def test_missing_confirmation_review(self):
        self.change_review(lambda ledger: ledger['confirmation'].popitem())
        with self.assertRaisesRegex(ValueError, 'Incomplete confirmation dispositions'):
            verify.verify(self.root)


if __name__ == '__main__':
    unittest.main()
