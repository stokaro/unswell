#!/usr/bin/env python3
"""Bind a newly prepared packet to its evaluator and unchanged reference labels."""
import argparse
from pathlib import Path

import accounting as a


def freeze(packet):
    a.REFERENCE.verify_freeze(packet)
    target = a.ROOT / 'evaluation-freeze.json'
    a.require(not target.exists(), 'Refuse to replace an evaluation freeze')
    files = dict(a.read(a.PRIOR / 'evaluation-freeze.json')['files'])
    repository = a.ROOT.parents[2]
    for path in [a.PRIOR / 'cli_compat.py', *(p for p in a.ROOT.iterdir() if p.is_file())]:
        files[str(path.relative_to(repository))] = a.digest(path.read_bytes())
    a.write(target, {'model_execution_started': False, 'events': 123, 'pages': 36,
                     'packet_freeze_sha256': a.digest((packet / 'freeze.json').read_bytes()),
                     'files': dict(sorted(files.items()))})


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('packet', type=Path)
    freeze(parser.parse_args().packet)
