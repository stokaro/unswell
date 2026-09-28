"""Compare complete pages with two explicitly supplied offline CLI binaries."""
import argparse
import gzip
import hashlib
import json
import platform
import shutil
import subprocess
import time
from pathlib import Path


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def write(path, value):
    path.write_text(json.dumps(value, ensure_ascii=False, indent=2) + '\n', encoding='utf-8')


def key(finding):
    # Versions, fingerprints, and policy identities intentionally change when a
    # rule changes. Compare its actual locations and diagnostic evidence.
    return json.dumps({k: finding[k] for k in ('rule_id', 'primary', 'related', 'evidence')}, sort_keys=True)


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--before', type=Path, required=True)
    ap.add_argument('--after', type=Path, required=True)
    ap.add_argument('--output', type=Path, required=True)
    args = ap.parse_args()
    here = Path(__file__).resolve().parent
    root = here.parents[2]
    frozen = here / 'confirmation'
    for name, digest in json.loads((frozen / 'freeze.json').read_text(encoding='utf-8')).items():
        assert sha(frozen / name) == digest, name
    manifest = json.loads((frozen / 'manifest.json').read_text(encoding='utf-8'))
    fixtures = json.loads((root / 'testdata/definition-benefit/controls.json').read_text(encoding='utf-8'))
    output = args.output.resolve()
    output.mkdir(exist_ok=False)
    inputs = output / 'inputs'
    shutil.copytree(frozen / 'inputs', inputs)
    (inputs / 'exposed-glossary.mdx').write_text(fixtures['cases'][0]['text'], encoding='utf-8')
    record = {'scope': 'One exposed positive page and three frozen negative complete pages; assistant review, not default qualification.',
              'baseline_revision': 'c18213636bcc1bcd5b46769a56c068adb66f0b94',
              'host': {'system': platform.system(), 'machine': platform.machine()},
              'confirmation_freeze_sha256': sha(frozen / 'freeze.json'),
              'candidate_source_hashes': {str(p.relative_to(root)): sha(p) for p in sorted((root / 'builtin').glob('*.go')) if not p.name.endswith('_test.go')},
              'runs': [], 'comparisons': {}}
    reports = {}
    for name, binary in [('before', args.before.resolve()), ('after', args.after.resolve())]:
        for profile in ('technical', 'strict'):
            target = output / (name + '-' + profile + '.json')
            command = [str(binary), 'check', '--profile', profile, '--project-root', str(output),
                       '--include-source', '--no-gate', '--jobs', '1', '--report', 'json:' + str(target), str(inputs)]
            started = time.monotonic()
            proc = subprocess.run(command, cwd=output, capture_output=True, timeout=120, check=False)
            elapsed = time.monotonic() - started
            assert proc.returncode == 0, (command, proc.returncode, proc.stderr.decode())
            report = json.loads(target.read_text(encoding='utf-8'))
            assert not report['errors'] and not report.get('abstentions')
            assert len(report['documents']) == len(manifest['pages']) + 1
            reports[name, profile] = report
            record['runs'].append({'variant': name, 'profile': profile, 'binary_sha256': sha(binary),
                                   'command': command, 'exit_code': proc.returncode, 'seconds': elapsed,
                                   'report_sha256': sha(target), 'findings': len(report['findings'])})
            with open(str(target) + '.gz', 'wb') as raw:
                with gzip.GzipFile(filename='', mode='wb', fileobj=raw, mtime=0) as stream:
                    stream.write(target.read_bytes())
    for profile in ('technical', 'strict'):
        before = {key(f): f for f in reports['before', profile]['findings']}
        after = {key(f): f for f in reports['after', profile]['findings']}
        added = [after[k] for k in sorted(after.keys() - before.keys())]
        removed = [before[k] for k in sorted(before.keys() - after.keys())]
        record['comparisons'][profile] = {'added': added, 'removed': removed,
            'confirmation_target_findings': sum(f['primary']['path'].endswith(tuple(p['input'].split('/')[-1] for p in manifest['pages']))
                and any(m['name'] == 'definition-benefit-restatements' for m in f['evidence']['metrics'])
                for f in reports['after', profile]['findings'])}
    write(output / 'measurement.json', record)
    print(json.dumps({p: {'added': len(r['added']), 'removed': len(r['removed']),
                         'confirmation_target_findings': r['confirmation_target_findings']}
                      for p, r in record['comparisons'].items()}, indent=2))


if __name__ == '__main__':
    main()
