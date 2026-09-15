"""Collect engine-owned applicability in bounded batches without model calls."""
import argparse
import hashlib
import json
import pathlib
import subprocess

parser = argparse.ArgumentParser()
parser.add_argument('--study', type=pathlib.Path, required=True)
parser.add_argument('--binary', type=pathlib.Path, required=True)
parser.add_argument('--cohort', choices=['historical', 'controlled'], required=True)
parser.add_argument('--family', choices=['luna', 'haiku', 'qwen'])
args = parser.parse_args()
d = args.study.resolve()
analysis = d / 'analysis'
sources = []
for p in sorted((analysis / 'acquisition' / args.cohort / 'shards').glob('*.json')):
    sources.extend(json.loads(p.read_text())['sources'])
if args.family:
    assert args.cohort == 'controlled'
    sources = [s for s in sources if s['path'].startswith('generated/2026-09-15-long-' + args.family + '/')]
assert sources
sources.sort(key=lambda s: (s['repository'], s['path']))
classes = json.loads((d / 'rule-classes.json').read_text())
outdir = analysis / 'features'
outdir.mkdir(exist_ok=True)
for offset in range(0, len(sources), 8):
    batch = sources[offset:offset+8]
    out = outdir / (args.cohort + ('-' + args.family if args.family else '') + '-' + str(offset // 8).zfill(3) + '.json')
    expected = {args.cohort + '/' + s['repository'].replace('/', '__') + '/' + s['path']: s['sha256'] for s in batch}
    if out.exists():
        result = json.loads(out.read_text())
        assert result['manifest']['complete']
        assert {s['name']: s['source_hash'] for s in result['documents']} == expected
        print('reused', out.name, len(batch), flush=True)
        continue
    command = [str(args.binary.resolve()), 'check', '--project-root', str(analysis / 'work'), '--allow-config-outside-root',
               '--config', str(d / 'policy.yaml'), '--no-gate', '--jobs', '1', '--timeout', '30m', '--report', 'json:' + str(out),
               '--feature', 'prose-words', '--feature', 'prose-sentences']
    for rule in classes['rules']:
        command += ['--feature', 'activation/' + rule['rule_id']]
    for path, digest in expected.items():
        source_path = analysis / 'work' / path
        assert hashlib.sha256(source_path.read_bytes()).hexdigest() == digest
        command.append(str(source_path))
    with out.with_suffix('.log').open('wb') as log:
        subprocess.run(command, stdout=log, stderr=subprocess.STDOUT, check=True)
    result = json.loads(out.read_text())
    assert result['manifest']['complete']
    assert {s['name']: s['source_hash'] for s in result['documents']} == expected
    print('collected', out.name, len(batch), flush=True)
print(args.cohort, len(sources), 'feature documents collected')
