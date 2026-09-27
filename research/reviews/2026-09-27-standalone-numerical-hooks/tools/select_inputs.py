#!/usr/bin/env python3
"""Select by metadata while conservatively excluding prior source identities."""
import hashlib
import json
from pathlib import Path
import subprocess
from urllib.parse import urlparse

BASE = '888d513b1f750184d16bc4ccd0fc055373e0cd0c'
ROOT = Path(__file__).resolve().parents[4]
NAMES = {'manifest.json', 'inputs.json', 'confirmation-inputs.json',
         'selection.json', 'annotations.json', 'source-freeze.json'}


def read(path):
    return json.loads(subprocess.check_output(['git', 'show', BASE + ':' + path], cwd=ROOT))


def select():
    frame = read('research/reviews/2026-09-18-structural-wordiness/confirmation/manifest.json')['frame']
    aliases = {}
    for row in frame:
        identity = (row['repository'], row['original_path'])
        for field in ('original_path', 'input_path'):
            aliases.setdefault(row[field], set()).add(identity)
    used = {}

    def record(value, repository, source):
        if not isinstance(value, str):
            return
        candidates = aliases.get(value, set())
        if value.startswith('https://github.com/'):
            parts = urlparse(value).path.strip('/').split('/')
            if len(parts) >= 5 and parts[2] in ('blob', 'raw'):
                candidates = {('/'.join(parts[:2]), '/'.join(parts[4:]))}
        for identity in candidates:
            if repository is None or identity[0] == repository:
                used.setdefault(identity, set()).add(source)

    def walk(value, repository, source, field=''):
        if isinstance(value, dict):
            repository = value.get('repository', repository)
            if not isinstance(repository, str):
                repository = None
            for key, item in value.items():
                # Sampling frames list possible pages, not prior selections.
                if key not in ('frame', 'universe'):
                    walk(item, repository, source, key)
        elif isinstance(value, list):
            for item in value:
                walk(item, repository, source, field)
        elif field in ('path', 'original_path', 'input_path', 'source_path',
                       'reference', 'url', 'excluded_paths', 'reserved_paths'):
            record(value, repository, source)

    paths = subprocess.check_output(['git', 'ls-tree', '-r', '--name-only', BASE, '--', 'research/reviews'], cwd=ROOT, text=True).splitlines()
    for path in paths:
        if Path(path).name in NAMES:
            walk(read(path), None, path)
    pools = {c: [p for p in frame if p['cohort'] == c and
                 (p['repository'], p['original_path']) not in used and
                 250 <= p['words'] <= 1600] for c in ('ptah', 'historical')}
    selected = []
    for cohort, pool in pools.items():
        pool.sort(key=lambda p: hashlib.sha256(('standalone-numerical-hooks-v1\n' + p['reference']).encode()).hexdigest())
        repos = set()
        for page in pool:
            if cohort == 'historical' and page['repository'] in repos:
                continue
            selected.append(page)
            repos.add(page['repository'])
            if sum(p['cohort'] == cohort for p in selected) == 2:
                break
    return {'base': BASE, 'selection': 'Metadata hash rank; distinct historical repositories; 250..1600 words; all earlier metadata source identities and reserved paths excluded.',
            'eligible_counts': {key: len(value) for key, value in pools.items()},
            'exclusions': [{'repository': k[0], 'original_path': k[1], 'records': sorted(v)} for k, v in sorted(used.items())],
            'pages': selected}


if __name__ == '__main__':
    print(json.dumps(select(), indent=2) + '\n', end='')
