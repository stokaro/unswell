#!/usr/bin/env python3
"""Freeze the diagnosis-first pilot against unchanged source and labels."""
import argparse
from pathlib import Path

import diagnosis as a


def freeze(packet):
    a.REFERENCE.verify_freeze(packet)
    target = a.ROOT / 'evaluation-freeze.json'
    a.require(not target.exists(), 'Refuse to replace an evaluation freeze')
    files = dict(a.read(a.ACCOUNTING / 'evaluation-freeze.json')['files'])
    repository = a.ROOT.parents[2]
    for path in [a.ACCOUNTING / 'evaluation-freeze.json',
                 a.OLD.PRIOR / 'evaluation-freeze.json',
                 a.ACCOUNTING / 'accounting.py',
                 *(p for p in a.ROOT.iterdir() if p.is_file())]:
        files[str(path.relative_to(repository))] = a.digest(path.read_bytes())
    a.write(target, {'model_execution_started': False,
                     'events': 123, 'pilot_events': 46, 'pages': 36, 'pilot_pages': 6,
                     'packet_freeze_sha256': a.digest((packet / 'freeze.json').read_bytes()),
                     'files': dict(sorted(files.items()))})


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('packet', type=Path)
    freeze(parser.parse_args().packet)
