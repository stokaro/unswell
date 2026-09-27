import json
from pathlib import Path
import shutil
import tempfile
import unittest

from study import ROOT, digest
from verify_confirmation import verify


class ConfirmationIntegrityTests(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory(); self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name) / 'confirmation'
        shutil.copytree(ROOT / 'confirmation', self.root)

    def mutate_review(self, change, rebind=False):
        path = self.root / 'annotations.json'; value = json.loads(path.read_text())
        change(value); path.write_text(json.dumps(value))
        if rebind:
            # Forge consistent hashes in this test copy to exercise semantic validation too.
            initial = self.root / 'annotations-initial.json'
            value = json.loads(initial.read_text()); change(value); initial.write_text(json.dumps(value))
            amendment = self.root / 'review-amendment.json'; value = json.loads(amendment.read_text())
            value['current_annotations_sha256'] = digest(path.read_bytes())
            value['initial_annotations_sha256'] = digest(initial.read_bytes())
            amendment.write_text(json.dumps(value))

    def test_complete_review_retains_uncertain_and_clean_examples(self):
        result = verify(self.root)
        self.assertEqual(result['pages'], 3)
        self.assertEqual(result['judgments'], {'defect': 14, 'acceptable': 12, 'uncertain': 7})
        self.assertFalse(result['model_run'])

    def test_changed_source_fails(self):
        with (self.root / 'sources/p01.md').open('a') as output: output.write('\nchanged\n')
        with self.assertRaisesRegex(ValueError, 'changed'): verify(self.root)

    def test_changed_judgment_fails_freeze(self):
        self.mutate_review(lambda r: r['judgments'].pop())
        with self.assertRaisesRegex(ValueError, 'Frozen review changed'): verify(self.root)

    def test_wrong_quote_fails_even_without_freeze(self):
        self.mutate_review(lambda r: r['judgments'][0]['targets'][0].update(quote='Invented quote.'), rebind=True)
        with self.assertRaisesRegex(ValueError, 'Quotation'): verify(self.root, check_freeze=False)

    def test_missing_section_fails_even_without_freeze(self):
        self.mutate_review(lambda r: r['pages'][0]['sections'].pop(), rebind=True)
        with self.assertRaisesRegex(ValueError, 'Unreviewed end'): verify(self.root, check_freeze=False)


if __name__ == '__main__':
    unittest.main()
