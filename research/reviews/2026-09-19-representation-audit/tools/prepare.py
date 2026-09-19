#!/usr/bin/env python3
"""Bind exposed whole-page annotations to their original archived source bytes."""
import argparse
import gzip
import hashlib
import json
from pathlib import Path
import tarfile

ROOT = Path(__file__).resolve().parents[4]
REVIEWS = ROOT / 'research/reviews'
EXCLUDED = {'2026-09-17-framing-recall', '2026-09-17-repetition-scope'}


def digest(data):
    return hashlib.sha256(data).hexdigest()


def checked_targets(review, raw):
    for kind in ('defects', 'uncertain', 'controls'):
        for event in review.get(kind, []):
            for target in event['targets']:
                start, end = target['start'], target['end']
                if not 0 <= start < end <= len(raw):
                    raise ValueError('Invalid annotation range')
                if raw[start:end].decode('utf-8') != target['quote']:
                    raise ValueError('Annotation quote drift')


def prepare():
    base = REVIEWS / '2026-09-17-full-page-recall'
    directories = [base] + [p.parent for p in sorted(REVIEWS.glob('*/confirmation/annotations.json'))
                           if p.parent.parent.name not in EXCLUDED]
    pages, files, seen = [], {}, set()
    for directory in directories:
        manifest = json.loads((directory/'manifest.json').read_text())
        annotations = json.loads((directory/'annotations.json').read_text())
        if isinstance(annotations, dict):
            annotations = annotations['pages']
        labels = {a['page']: a for a in annotations}
        for name in ('manifest.json', 'annotations.json', 'inputs.tar.gz'):
            path = directory/name
            files[str(path.relative_to(ROOT))] = digest(path.read_bytes())
        with tarfile.open(directory/'inputs.tar.gz') as archive:
            for page in manifest['pages']:
                identity = (page['repository'], page['original_path'])
                if identity in seen:
                    raise ValueError('Repeated source identity: '+str(identity))
                seen.add(identity)
                raw = archive.extractfile(page['path']).read()
                review = labels[page['id']]
                if digest(raw) != page['sha256'] or review['source_sha256'] != page['sha256']:
                    raise ValueError('Source identity drift')
                checked_targets(review, raw)
                identifier = directory.parent.name if directory.name == 'confirmation' else directory.name
                identifier += '/'+page['id']
                pages.append(dict(id=identifier, repository=page['repository'],
                                  original_path=page['original_path'], reference=page['reference'],
                                  cohort=page['cohort'], format=page['format'], sha256=page['sha256'],
                                  text=raw.decode('utf-8'), review=review))
    return dict(version=1, basis='exposed-assistant-review', input_files=files,
                excluded_narrow_reviews=sorted(EXCLUDED), pages=pages)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--output', type=Path, required=True)
    args = parser.parse_args()
    if args.output.exists():
        raise ValueError('Output must be new')
    dataset = prepare()
    payload = json.dumps(dataset, ensure_ascii=False, sort_keys=True).encode()+b'\n'
    args.output.write_bytes(gzip.compress(payload, mtime=0))
    print(json.dumps(dict(pages=len(dataset['pages']), sha256=digest(args.output.read_bytes()),
                          defects=sum(len(p['review']['defects']) for p in dataset['pages']))))


if __name__ == '__main__':
    main()
