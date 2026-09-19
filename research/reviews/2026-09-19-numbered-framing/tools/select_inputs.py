#!/usr/bin/env python3
"""Reproduce the metadata-only selection against the frozen parent revision."""
import hashlib
import json
from pathlib import Path
import subprocess

BASE = '6c5c4e6d860b46105836a377bdf77c1b3b079303'
REPO = Path(__file__).resolve().parents[4]


def read(name):
    return json.loads(subprocess.check_output(['git', 'show', BASE+':'+name], cwd=REPO))


def select():
    frame = read('research/reviews/2026-09-18-structural-wordiness/confirmation/manifest.json')['frame']
    paths = subprocess.check_output(['git', 'ls-tree', '-r', '--name-only', BASE, '--', 'research/reviews'], cwd=REPO, text=True).splitlines()
    used = set()
    for name in paths:
        if name.endswith('/manifest.json'):
            record = read(name)
            if isinstance(record, dict):
                used.update((p.get('repository'), p.get('original_path')) for p in record.get('pages', []) if isinstance(p, dict))
    selected = []
    for cohort in ('ptah', 'historical'):
        pool = [p for p in frame if p['cohort'] == cohort and (p['repository'], p['original_path']) not in used and 250 <= p['words'] <= 1600]
        pool.sort(key=lambda p: hashlib.sha256(('numbered-framing-v1\n'+p['reference']).encode()).hexdigest())
        repos = set()
        for page in pool:
            if cohort == 'historical' and page['repository'] in repos:
                continue
            selected.append(page)
            repos.add(page['repository'])
            if sum(p['cohort'] == cohort for p in selected) == 2:
                break
    if len(selected) != 4:
        raise ValueError('Source pool shortage')
    return selected


if __name__ == '__main__':
    recorded = json.loads((Path(__file__).resolve().parents[1]/'inputs.json').read_text())['pages']
    chosen = select()
    if [p['reference'] for p in chosen] != [p['reference'] for p in recorded]:
        raise ValueError('Selection drift')
    print(json.dumps([p['reference'] for p in chosen], indent=2))
