#!/usr/bin/env python3
"""Replay the frozen plaintext examples with two explicitly supplied CLI builds."""
import argparse
import gzip
import hashlib
import json
from pathlib import Path
import subprocess
import tarfile
import tempfile

ROOT = Path(__file__).resolve().parent


def sha(data):
    return hashlib.sha256(data).hexdigest()


def read(path):
    return json.loads(path.read_text())


def overlap(left, right):
    return left['start'] < right['end'] and right['start'] < left['end']


def key(finding):
    return (finding['rule_id'], finding['primary']['path'],
            finding['primary']['span']['start'], finding['primary']['span']['end'])


def replay(before, after, output):
    for name, digest in read(ROOT / 'freeze.json').items():
        assert sha((ROOT / name).read_bytes()) == digest, name
    labels = read(ROOT / 'labels.json')
    rows = labels['development'] + [labels['confirmation']]
    sources = {}
    with tarfile.open(ROOT / 'inputs.tar.gz') as archive:
        for row in rows:
            source = archive.extractfile(row['path']).read()
            assert sha(source) == row['sha256']
            sources[Path(row['path']).name] = source
    for group in ('code', 'controls'):
        for span in labels['confirmation'][group]:
            source = sources['whatsnew-2.2.txt']
            assert source[span['start']:span['end']].decode() == span['quote']
    reports = {}
    output.mkdir(parents=True, exist_ok=True)
    with tempfile.TemporaryDirectory(prefix='unswell-plain-code-') as directory:
        work = Path(directory)
        for name, data in sources.items():
            (work / name).write_bytes(data)
        for build, binary in [('before', before), ('after', after)]:
            for profile in ('technical', 'strict'):
                report_path = work / 'report.json'
                command = [str(binary), 'check', '--profile', profile,
                           '--include-source', '--report', 'json:report.json', *sorted(sources)]
                run = subprocess.run(command, cwd=work, capture_output=True, timeout=120, check=False)
                assert run.returncode in (0, 1), run.stderr.decode()
                report = read(report_path)
                assert report['status'] == 'complete' and not report['errors']
                assert report['manifest']['complete']
                for doc in report['documents']:
                    assert doc['source'].encode() == sources[doc['name']]
                reports[build, profile] = report
                payload = (json.dumps(report, indent=2, ensure_ascii=False) + '\n').encode()
                (output / f'{build}-{profile}.json.gz').write_bytes(gzip.compress(payload, mtime=0))
    result = {'scope': labels['scope'], 'reviewer': labels['reviewer'],
              'input_freeze_sha256': sha((ROOT / 'freeze.json').read_bytes()),
              'binary_sha256': {'before': sha(before.read_bytes()), 'after': sha(after.read_bytes())},
              'profiles': {}}
    old_labels = read(ROOT.parent / '2026-09-17-local-repetition/confirmation/annotations.json')
    targets = next(row for row in old_labels if row['page'] == 'c11')['defects']
    for profile in ('technical', 'strict'):
        old, new = reports['before', profile], reports['after', profile]
        assert old['manifest']['config_hash'] == new['manifest']['config_hash']
        assert old['manifest']['ruleset_hash'] == new['manifest']['ruleset_hash']
        old_keys, new_keys = {key(f) for f in old['findings']}, {key(f) for f in new['findings']}
        removed = [f for f in old['findings'] if key(f) not in new_keys]
        added = [f for f in new['findings'] if key(f) not in old_keys]
        documents = []
        for previous, current in zip(old['documents'], new['documents'], strict=True):
            assert previous['name'] == current['name']
            code = [x['span'] for x in current['excluded'] if x['reason'] == 'plain-code']
            other = [x for x in current['excluded'] if x['reason'] != 'plain-code']
            assert other == previous['excluded'], 'Existing exclusions changed'
            assert current['source_hash'] == previous['source_hash']
            if current['name'] == 'whatsnew-2.2.txt':
                expected = [{k: x[k] for k in ('start', 'end')} for x in labels['confirmation']['code'] if x['kind'] == 'c']
                assert code == expected, 'Confirmation excluded unexpected bytes or missed C code'
                for control in labels['confirmation']['controls']:
                    assert not any(overlap(control, span) for span in code)
            if current['name'] == 'whatsnew-2.1.txt':
                for defect in targets:
                    for target in defect['targets']:
                        assert not any(overlap(target, span) for span in code), defect['id']
            documents.append({'path': current['name'], 'source_sha256': sha(sources[current['name']]),
                              'words_before': previous['prose_words'], 'words_after': current['prose_words'],
                              'code_exclusions': code})
        result['profiles'][profile] = {
            'before': len(old['findings']), 'after': len(new['findings']),
            'removed': removed, 'added': added, 'documents': documents,
            'gate_before': old['gate'], 'gate_after': new['gate'],
            'tool_commit_before': old['manifest']['tool_commit'],
            'tool_commit_after': new['manifest']['tool_commit'],
            'frozen_development_defects_preserved': len(targets),
            'confirmation_c_regions': 3, 'confirmation_controls_preserved': 5,
        }
    result['report_sha256'] = {p.name: sha(p.read_bytes()) for p in sorted(output.glob('*.json.gz'))}
    (output / 'summary.json').write_text(json.dumps(result, indent=2, ensure_ascii=False) + '\n')
    for profile, row in result['profiles'].items():
        print(f"{profile}: {row['before']} -> {row['after']}; removed {len(row['removed'])}, added {len(row['added'])}")


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--before', type=Path, required=True)
    parser.add_argument('--after', type=Path, required=True)
    parser.add_argument('--output', type=Path, required=True)
    args = parser.parse_args()
    replay(args.before.resolve(), args.after.resolve(), args.output.resolve())
