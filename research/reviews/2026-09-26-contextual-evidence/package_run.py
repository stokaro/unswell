#!/usr/bin/env python3
"""Package a terminal, reviewed run without issuing model requests."""
import argparse
import gzip
import io
import json
from pathlib import Path
import tarfile

from report_run import report
from study import ROOT, digest, read, require


def build(packet, output, destination):
    ledger = read(output / 'run.json')
    require(all(c['status'] != 'started' for c in ledger['calls']), 'Run is still active')
    require(ledger['complete'] or bool(ledger.get('stop_reason')), 'Run has no terminal result')
    result = report(packet, output, output / 'review.json')
    files = {}
    for path in sorted(packet.rglob('*')):
        if path.is_file() and '__pycache__' not in path.parts:
            files['packet/' + path.relative_to(packet).as_posix()] = path.read_bytes()
    for name in ['run.json', 'run-original.json', 'cli-compat-amendment.json',
                 'authorization.json', 'cli_compat.py', 'review.json', 'invalid-response-analysis.json']:
        files['run/' + name] = (output / name).read_bytes()
    for call in ledger['calls']:
        for name in ['events.jsonl', 'stderr.txt', 'response.json', 'bound.json']:
            path = output / call['page'] / name
            if path.exists():
                files['run/' + call['page'] + '/' + name] = path.read_bytes()
    files['summary.json'] = (json.dumps(result, indent=2) + '\n').encode()
    references = [
        ('whole', ROOT.parent / '2026-09-17-full-page-recall' / 'inputs.tar.gz'),
        ('context', ROOT.parent / '2026-09-17-proposition-repetition' / 'confirmation' / 'inputs.tar.gz'),
    ]
    for cohort, archive in references:
        with tarfile.open(archive) as source:
            for member in source.getmembers():
                if member.isfile() and member.name.startswith('notices/'):
                    require('..' not in Path(member.name).parts, 'Invalid notice path')
                    files[cohort + '/' + member.name] = source.extractfile(member).read()
    files['confirmation/notices/ptah/LICENSE'] = (
        ROOT.parent / '2026-09-26-outcome-announcements' / 'LICENSE.ptah').read_bytes()
    manifest = {name: digest(raw) for name, raw in sorted(files.items())}
    files['archive-manifest.json'] = (json.dumps(manifest, indent=2) + '\n').encode()
    require(not destination.exists(), 'Refuse to overwrite an evidence archive')
    with destination.open('xb') as target:
        with gzip.GzipFile(filename='', mode='wb', fileobj=target, mtime=0) as zipped:
            with tarfile.open(fileobj=zipped, mode='w') as archive:
                for name, raw in sorted(files.items()):
                    member = tarfile.TarInfo(name)
                    member.size = len(raw); member.mode = 0o644
                    archive.addfile(member, io.BytesIO(raw))
    return {'file': destination.name, 'sha256': digest(destination.read_bytes()),
            'members': len(files), 'bytes': destination.stat().st_size}


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('packet', type=Path); parser.add_argument('output', type=Path)
    parser.add_argument('destination', type=Path)
    args = parser.parse_args()
    print(json.dumps(build(args.packet, args.output, args.destination), indent=2))
