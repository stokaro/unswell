"""Blackbox corruption tests for the offline whole-page evidence evaluator."""

import gzip
import hashlib
import io
import json
from pathlib import Path
import shutil
import subprocess
import sys
import tarfile
import tempfile
import unittest

TOOL = Path(__file__).with_name('evaluate.py')
RECORD = TOOL.parent.parent


def read(path):
    data = path.read_bytes()
    return json.loads(gzip.decompress(data) if path.suffix == '.gz' else data)


def write(path, value):
    data = (json.dumps(value, ensure_ascii=False, indent=2) + '\n').encode()
    path.write_bytes(gzip.compress(data, mtime=0) if path.suffix == '.gz' else data)


class EvidenceTests(unittest.TestCase):
    def setUp(self):
        self.temporary = tempfile.TemporaryDirectory()
        self.addCleanup(self.temporary.cleanup)
        self.root = Path(self.temporary.name)
        for path in RECORD.iterdir():
            if path.is_file():
                shutil.copyfile(path, self.root / path.name)

    def run_evaluator(self, error=None):
        result = subprocess.run([sys.executable, str(TOOL), str(self.root)], capture_output=True, text=True)
        if error is None:
            self.assertEqual(result.returncode, 0, result.stderr)
            return json.loads(result.stdout)
        self.assertNotEqual(result.returncode, 0)
        self.assertIn(error, result.stderr)

    def reseal(self, name):
        freeze = read(self.root / 'freeze.json')
        freeze['files'][name] = hashlib.sha256((self.root / name).read_bytes()).hexdigest()
        write(self.root / 'freeze.json', freeze)

    def change_report(self, change):
        path = self.root / 'technical.json.gz'
        report = read(path)
        change(report)
        write(path, report)
        runs = read(self.root / 'runs.json')
        runs['technical']['sha256'] = hashlib.sha256(path.read_bytes()).hexdigest()
        write(self.root / 'runs.json', runs)

    def test_frozen_counts_and_separate_anchors(self):
        result = self.run_evaluator()
        for profile in result.values():
            self.assertEqual(profile['ptah-sample']['defects'], 27)
            self.assertEqual(profile['historical-sample']['defects'], 21)
            self.assertEqual(profile['ptah-exposed_anchor']['defects'], 9)
            self.assertEqual(profile['ptah-sample']['detected'], 1)
            self.assertEqual(profile['historical-sample']['detected'], 0)
            self.assertEqual(profile['ptah-exposed_anchor']['detected'], 4)

    def test_frozen_annotation_drift(self):
        path = self.root / 'annotations.json'
        path.write_bytes(path.read_bytes() + b' ')
        self.run_evaluator('Frozen artifact changed')

    def test_source_drift_even_after_archive_reseal(self):
        path = self.root / 'inputs.tar.gz'
        data = io.BytesIO()
        with tarfile.open(path) as source, tarfile.open(fileobj=data, mode='w') as target:
            for info in source:
                content = source.extractfile(info).read()
                if info.name == 'sources/p001.md':
                    content = b'X' + content[1:]
                target.addfile(info, io.BytesIO(content))
        path.write_bytes(gzip.compress(data.getvalue(), mtime=0))
        self.reseal('inputs.tar.gz')
        self.run_evaluator('Source drift')

    def test_invalid_source_span(self):
        path = self.root / 'annotations.json'
        annotations = read(path)
        annotations[0]['defects'][0]['targets'][0]['end'] = 999999
        write(path, annotations)
        self.reseal('annotations.json')
        self.run_evaluator('Invalid source span')

    def test_utf8_boundary(self):
        path = self.root / 'annotations.json'
        annotations = read(path)
        with tarfile.open(self.root / 'inputs.tar.gz') as archive:
            source = archive.extractfile('sources/p001.md').read()
        start = next(i for i, value in enumerate(source) if value >= 192)
        annotations[0]['defects'][0]['targets'][0] = dict(start=start + 1, end=start + 2, quote='x')
        write(path, annotations)
        self.reseal('annotations.json')
        self.run_evaluator('Source span splits UTF-8')

    def test_incomplete_review_coverage(self):
        path = self.root / 'annotations.json'
        annotations = read(path)
        annotations[0]['coverage'].pop()
        write(path, annotations)
        self.reseal('annotations.json')
        self.run_evaluator('Incomplete coverage')

    def test_missing_disposition(self):
        path = self.root / 'dispositions.json'
        rows = read(path)
        rows['technical'].pop()
        write(path, rows)
        self.run_evaluator('Missing finding disposition')

    def test_duplicate_recall_credit(self):
        path = self.root / 'dispositions.json'
        rows = read(path)
        rows['technical'][37]['events'] *= 2
        write(path, rows)
        self.run_evaluator('Duplicate recall credit')

    def test_multiple_findings_count_one_event(self):
        path = self.root / 'technical.json.gz'
        report = read(path)
        finding = dict(report['findings'][37], id='another-independent-diagnostic')
        report['findings'].append(finding)
        write(path, report)
        runs = read(self.root / 'runs.json')
        runs['technical']['sha256'] = hashlib.sha256(path.read_bytes()).hexdigest()
        write(self.root / 'runs.json', runs)
        ledger_path = self.root / 'dispositions.json'
        ledger = read(ledger_path)
        identity = json.dumps(finding, sort_keys=True, separators=(',', ':'), ensure_ascii=False).encode()
        ledger['technical'].append(dict(ledger['technical'][37], index=len(report['findings'])-1,
                                        finding_sha256=hashlib.sha256(identity).hexdigest()))
        write(ledger_path, ledger)
        result = self.run_evaluator()
        self.assertEqual(result['technical']['ptah-sample']['detected'], 1)

    def test_length_overlap_is_not_framing_detection(self):
        path = self.root / 'dispositions.json'
        rows = read(path)
        rows['technical'][26].update(status='actionable', events=['p002-d02'])
        write(path, rows)
        self.run_evaluator('Semantic mismatch')

    def test_unrelated_source_cannot_earn_credit(self):
        path = self.root / 'dispositions.json'
        rows = read(path)
        rows['technical'][37]['events'] = ['p012-d01']
        write(path, rows)
        self.run_evaluator('no source overlap')

    def test_repetition_needs_both_annotated_occurrences(self):
        def remove_occurrence(report):
            report['findings'][37]['related'].pop()
        self.change_report(remove_occurrence)
        path = self.root / 'dispositions.json'
        rows = read(path)
        finding = read(self.root / 'technical.json.gz')['findings'][37]
        identity = json.dumps(finding, sort_keys=True, separators=(',', ':'), ensure_ascii=False).encode()
        rows['technical'][37]['finding_sha256'] = hashlib.sha256(identity).hexdigest()
        write(path, rows)
        self.run_evaluator('no source overlap')

    def test_config_artifact_drift(self):
        (self.root / 'technical.yaml').write_text('version: 1\nextends: [builtin:strict-v1]\n')
        self.run_evaluator('Config artifact drift')

    def test_missing_document(self):
        self.change_report(lambda report: report['documents'].pop())
        self.run_evaluator('Missing/extra reported document')

    def test_incomplete_analysis(self):
        self.change_report(lambda report: report.update(status='incomplete'))
        self.run_evaluator('Incomplete run')

    def test_unexpected_abstention(self):
        self.change_report(lambda report: report['manifest'].update(skipped_rules=['rule-budget-exhausted']))
        self.run_evaluator('Unexpected rule abstention')

    def test_nonfinite_json(self):
        (self.root / 'dispositions.json').write_text('{"technical": NaN}')
        self.run_evaluator('Nonfinite JSON value')

    def test_duplicate_json_key(self):
        (self.root / 'dispositions.json').write_text('{"technical": [], "technical": []}')
        self.run_evaluator('Duplicate JSON key')

    def test_missing_profile(self):
        path = self.root / 'dispositions.json'
        rows = read(path)
        rows.pop('strict')
        write(path, rows)
        self.run_evaluator('Missing profile')


if __name__ == '__main__':
    unittest.main()
