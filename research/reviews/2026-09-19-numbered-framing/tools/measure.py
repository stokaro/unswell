#!/usr/bin/env python3
"""Replay pinned complete pages with an explicit offline CLI."""
import argparse
import gzip
import hashlib
import json
from pathlib import Path, PurePosixPath
import subprocess
import tarfile
import tempfile
import time

ROOT = Path(__file__).resolve().parents[1]


def sha(data):
    return hashlib.sha256(data).hexdigest()


def sources():
    freeze = json.loads((ROOT/'input-freeze.json').read_text())
    for name, digest in freeze['files'].items():
        if sha((ROOT/name).read_bytes()) != digest:
            raise ValueError('Frozen input changed: '+name)
    manifest = json.loads((ROOT/'inputs.json').read_text())
    with tarfile.open(ROOT/'inputs.tar.gz') as archive:
        result = {}
        for item in archive:
            path = PurePosixPath(item.name)
            if not item.isfile() or path.is_absolute() or '..' in path.parts or item.name in result:
                raise ValueError('Invalid archive member')
            data = archive.extractfile(item).read()
            if sha(data) != manifest['files'][item.name]:
                raise ValueError('Source hash mismatch')
            result[item.name] = data
    if set(result) != set(manifest['files']):
        raise ValueError('Missing source or notice')
    return result


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--binary', type=Path, required=True)
    parser.add_argument('--output', type=Path, required=True)
    args = parser.parse_args()
    binary = args.binary.resolve(strict=True)
    output = args.output.resolve()
    output.mkdir(parents=True, exist_ok=False)
    records = {}
    with tempfile.TemporaryDirectory(prefix='unswell-counted-') as name:
        work = Path(name)
        for path, data in sources().items():
            target = work/path
            target.parent.mkdir(parents=True, exist_ok=True)
            target.write_bytes(data)
        inputs = ['development/configure-a-provider.md']+sorted(p.relative_to(work).as_posix() for p in (work/'sources').glob('*.md'))
        for profile in ('technical', 'strict'):
            config = f'version: 1\nextends: [builtin:{profile}-v1]\n'.encode()
            (work/'policy.yaml').write_bytes(config)
            target = work/'report.json'
            command = [str(binary), 'check', '--config', 'policy.yaml', '--project-root', '.', '--include-source',
                       '--report', 'json:'+str(target), '--timeout', '5m', *inputs]
            start = time.monotonic()
            run = subprocess.run(command, cwd=work, capture_output=True, timeout=360)
            duration = time.monotonic()-start
            if run.returncode not in (0, 1):
                raise RuntimeError(run.stderr.decode()+run.stdout.decode())
            report = json.loads(target.read_bytes())
            if report['status'] != 'complete' or not report['manifest']['complete'] or report['errors'] or report.get('abstentions'):
                raise ValueError('Incomplete analysis')
            if {d['name'] for d in report['documents']} != set(inputs):
                raise ValueError('Page omitted from scan')
            packed = gzip.compress(target.read_bytes(), mtime=0)
            (output/(profile+'.json.gz')).write_bytes(packed)
            records[profile] = dict(tool_commit=report['manifest']['tool_commit'], binary_sha256=sha(binary.read_bytes()),
                                    report_sha256=sha(packed), config_sha256=sha(config), exit_code=run.returncode,
                                    wall_seconds=duration, documents=len(report['documents']), findings=len(report['findings']))
            print(profile, records[profile], flush=True)
    (output/'runs.json').write_text(json.dumps(records, indent=2)+'\n')


if __name__ == '__main__':
    main()
