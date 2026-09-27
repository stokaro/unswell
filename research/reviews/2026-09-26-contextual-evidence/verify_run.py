#!/usr/bin/env python3
"""Recompute a retained experiment offline, including its original CLI traces."""
import argparse
import json
from pathlib import Path
import tarfile
import tempfile

from report_run import report
from study import digest, read, require


def verify(archive):
    with tempfile.TemporaryDirectory(prefix='unswell-contextual-replay-') as temporary:
        root = Path(temporary)
        with tarfile.open(archive) as source:
            members = source.getmembers()
            require(len(members) <= 512 and sum(m.size for m in members) <= 64 << 20,
                    'Evidence archive exceeds bounds')
            names = [m.name for m in members]
            require(len(names) == len(set(names)), 'Duplicate archive member')
            for member in members:
                name = Path(member.name)
                require(member.isfile() and not name.is_absolute() and '..' not in name.parts,
                        'Invalid evidence member')
                path = root / name; path.parent.mkdir(parents=True, exist_ok=True)
                path.write_bytes(source.extractfile(member).read())
        manifest = read(root / 'archive-manifest.json')
        require(set(manifest) == set(names) - {'archive-manifest.json'}, 'Archive manifest differs')
        for name, expected in manifest.items():
            require(digest((root / name).read_bytes()) == expected, 'Changed archive member: ' + name)
        result = report(root / 'packet', root / 'run', root / 'run' / 'review.json')
        require(result == read(root / 'summary.json'), 'Saved summary differs from retained evidence')
        return result


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('archive', type=Path)
    args = parser.parse_args()
    print(json.dumps(verify(args.archive), indent=2))
