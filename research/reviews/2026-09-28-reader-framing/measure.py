"""Compare explicit offline binaries on four hash-verified public documents."""
import argparse
import hashlib
import json
import pathlib
import subprocess
import time


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def finding_key(finding):
    return json.dumps({key: finding[key] for key in
                      ('rule_id', 'primary', 'related', 'evidence', 'severity', 'gate', 'message')}, sort_keys=True)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--inputs', type=pathlib.Path, required=True)
    parser.add_argument('--before', type=pathlib.Path, required=True)
    parser.add_argument('--after', type=pathlib.Path, required=True)
    parser.add_argument('--output', type=pathlib.Path, required=True)
    args = parser.parse_args()
    here = pathlib.Path(__file__).resolve().parent
    root = here.parents[2]
    manifest = json.loads((here / 'inputs.json').read_text())
    inputs = args.inputs.resolve(strict=True)
    for source in manifest['sources']:
        assert sha(inputs / source['file']) == source['sha256'], source['file']
    output = args.output.resolve()
    output.mkdir(exist_ok=False)
    records, reports = [], {}
    for profile in ('technical', 'strict'):
        for variant, binary in [('before', args.before.resolve(strict=True)), ('after', args.after.resolve(strict=True))]:
            name = variant + '-' + profile
            target = output / (name + '.json')
            command = [str(binary), 'check', '--profile', profile, '--project-root', str(inputs),
                       '--include-source', '--no-gate', '--jobs', '1', '--timeout', '2m',
                       '--report', 'json:' + str(target), *[str(inputs / row['file']) for row in manifest['sources']]]
            start = time.monotonic()
            result = subprocess.run(command, cwd=output, capture_output=True, timeout=150, check=False)
            (output / (name + '.log')).write_bytes(result.stdout + result.stderr)
            assert result.returncode == 0, (name, result.returncode)
            report = json.loads(target.read_text())
            assert report['status'] == 'complete' and report['manifest']['complete']
            assert not report['errors'] and not report.get('abstentions')
            assert len(report['documents']) == len(manifest['sources'])
            reports[variant, profile] = report
            records.append(dict(variant=variant, profile=profile, command=command,
                                seconds=time.monotonic() - start, exit_code=result.returncode,
                                binary_sha256=sha(binary), report_sha256=sha(target),
                                tool_commit=report['manifest']['tool_commit'], findings=len(report['findings'])))
    comparison = {}
    for profile in ('technical', 'strict'):
        before = {finding_key(f): f for f in reports['before', profile]['findings']}
        after = {finding_key(f): f for f in reports['after', profile]['findings']}
        comparison[profile] = dict(added=[after[k] for k in sorted(after.keys() - before.keys())],
                                   removed=[before[k] for k in sorted(before.keys() - after.keys())])
    record = dict(baseline='0f43d6f60daeba4f9eb8b4bf394728490229c108',
                  inputs_sha256=sha(here / 'inputs.json'),
                  candidate_source_hashes={str(p.relative_to(root)): sha(p) for p in sorted((root / 'builtin').glob('*.go'))
                                           if not p.name.endswith('_test.go')},
                  runs=records, comparisons=comparison)
    (output / 'measurement.json').write_text(json.dumps(record, indent=2) + '\n')
    print(json.dumps({p: {k: len(v) for k, v in delta.items()} for p, delta in comparison.items()}))


if __name__ == '__main__':
    main()
