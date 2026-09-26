#!/usr/bin/env python3
"""Validate complete paired evidence and source-bound review ledgers offline."""
import copy
import gzip
import hashlib
import json
from pathlib import Path

from replay import ROOT, REFERENCES, sources


def read(path):
    raw = path.read_bytes()
    return json.loads(gzip.decompress(raw) if path.suffix == '.gz' else raw)


def require(value, message):
    if not value:
        raise ValueError(message)


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def key(finding):
    result = copy.deepcopy(finding)
    for field in ['id', 'fingerprint', 'rule_version']:
        result.pop(field, None)
    if result['rule_id'] == 'filler.document-justification':
        result.pop('message', None)
        result['evidence'].pop('suggestion', None)
    return json.dumps(result, sort_keys=True)


def labels(root=ROOT):
    result = {}
    for cohort, reference in REFERENCES.items():
        data = read(reference / 'annotations.json')
        paths = {page['id']: cohort + '/' + page['id'] + Path(page['path']).suffix
                 for page in read(reference / 'manifest.json')['pages']}
        for page in data if isinstance(data, list) else data['pages']:
            for defect in page['defects']:
                result[cohort + '/' + defect['id']] = dict(
                    path=paths[page['page']], targets=defect['targets'], category=defect['category'])
    for page in read(root / 'confirmation-labels.json')['pages']:
        for row in page['judgments']:
            if row['label'] == 'needs_revision':
                result['confirmation/' + row['id']] = dict(path='confirmation/' + page['id'] + '.mdx',
                                                          targets=[{**row['span'], 'quote': row['quote']}],
                                                          category=row['category'])
    return result


def overlap(finding, label):
    return any(loc['path'] == label['path'] and
               any(max(loc['span']['start'], target['start']) < min(loc['span']['end'], target['end'])
                   for target in label['targets']) for loc in [finding['primary'], *finding['related']])


def report(path, expected):
    data = read(path)
    require(data['status'] == 'complete' and data['manifest']['complete'] and not data['errors'], 'Incomplete scan')
    require(len(data['documents']) == len(expected), 'Document count differs')
    require({d['name'] for d in data['documents']} == set(expected), 'Missing document')
    for doc in data['documents']:
        source = expected[doc['name']]
        require(doc['source_hash'] == hashlib.sha256(source).hexdigest(), 'Report source drift')
        require(doc['source'].encode() == source, 'Report source changed')
    for finding in data['findings']:
        for loc in [finding['primary'], *finding['related']]:
            span = loc['span']; source = expected[loc['path']]
            require(0 <= span['start'] < span['end'] <= len(source), 'Invalid range')
            require(source[span['start']:span['end']].decode() == loc['snippet'], 'Invented snippet')
    return data['findings']


def verify(root=ROOT):
    for filename in ['input-freeze.json', 'artifacts.json']:
        for name, sha in read(root / filename)['files'].items():
            require(digest(root / name) == sha, 'Changed artifact: ' + name)
    for label, filename in [('label-freeze.json', 'confirmation-labels.json'),
                            ('additional-freeze.json', 'additional-cases.json')]:
        require(digest(root / filename) == read(root / label)['sha256'], 'Changed frozen labels or cases')
    expected = sources(root)
    all_labels = labels(root)
    review = read(root / 'review.json')
    exposure = read(root / 'exposure-amendment.json')['pages']
    require({p['id'] for p in exposure} == {p['id'] for p in read(root / 'confirmation-inputs.json')['pages']},
            'Incomplete exposure record')
    for page in read(root / 'confirmation-labels.json')['pages']:
        source = expected['confirmation/' + page['id'] + '.mdx']
        require(page['source_sha256'] == hashlib.sha256(source).hexdigest(), 'Label source drift')
        require(page['coverage']['start'] == 0 and page['coverage']['end'] == len(source), 'Incomplete review')
        for row in page['judgments']:
            require(source[row['span']['start']:row['span']['end']].decode() == row['quote'], 'Changed judgment range')
    for label in all_labels.values():
        for target in label['targets']:
            require(expected[label['path']][target['start']:target['end']].decode() == target['quote'], 'Changed event range')
    result = {}
    for profile in ['technical', 'strict']:
        ledger = review['profiles'][profile]
        require(set(ledger['events']) == set(all_labels), 'Incomplete event ledger')
        arms = {arm: report(root / (arm + '-' + profile + '.json.gz'), expected) for arm in ['before', 'after']}
        before_keys = {key(f) for f in arms['before']}
        require(before_keys <= {key(f) for f in arms['after']}, 'Previous finding changed or removed')
        additions = [f for f in arms['after'] if key(f) not in before_keys]
        require(set(ledger['additions']) == {f['id'] for f in additions}, 'Unreviewed addition')
        for row in ledger['additions'].values():
            require(row['disposition'] in ['accepted', 'uncertain', 'rejected'] and row['rationale'], 'Invalid disposition')
        confirmation = [f for f in arms['after'] if f['primary']['path'].startswith('confirmation/')]
        require(set(ledger['confirmation']) == {f['id'] for f in confirmation}, 'Incomplete confirmation dispositions')
        for row in ledger['confirmation'].values():
            require(row['disposition'] in ['accepted', 'uncertain', 'rejected'] and row['rationale'], 'Invalid disposition')
        counts = {cohort: dict(events=sum(name.startswith(cohort + '/') for name in all_labels), before=0, after=0)
                  for cohort in ['whole', 'context', 'confirmation']}
        for name, row in ledger['events'].items():
            require(row['rationale'], 'Missing event rationale')
            for arm in arms:
                status = row[arm]['status']
                require(status in ['full', 'partial', 'missed'], 'Invalid event status')
                ids = row[arm]['findings']
                candidates = [f for f in arms[arm] if f['id'] in ids]
                require(len(candidates) == len(set(ids)), 'Unknown supporting finding')
                if status in ['full', 'partial']:
                    require(candidates and all(overlap(f, all_labels[name]) for f in candidates), 'Unsupported event credit')
                if status == 'full':
                    require(all(f['rule_id'] not in ['syntax.long-sentence', 'format.em-dash-density',
                                                    'syntax.paired-contrast-density'] or
                                f['rule_id'] == 'syntax.long-sentence' and all_labels[name]['category'] == 'needless_complexity'
                                for f in candidates), 'Incidental signal is not full credit')
                    counts[name.split('/')[0]][arm] += 1
        result[profile] = dict(cohorts=counts, before_findings=len(arms['before']), after_findings=len(arms['after']),
                               additions=len(additions), accepted_additions=sum(r['disposition']=='accepted' for r in ledger['additions'].values()),
                               confirmation_findings=len(confirmation), confirmation_dispositions={d:sum(r['disposition']==d for r in ledger['confirmation'].values())for d in ['accepted','uncertain','rejected']})
        strata = {}
        for known in [False, True]:
            ids = {p['id'] for p in exposure if p['prior_review_known'] == known}
            names = [name for name, label in all_labels.items()
                     if label['path'].startswith('confirmation/') and Path(label['path']).stem in ids]
            strata['previously_reviewed' if known else 'no_known_prior_review'] = dict(
                pages=len(ids), events=len(names),
                **{arm: sum(ledger['events'][name][arm]['status'] == 'full' for name in names)
                   for arm in arms})
        result[profile]['confirmation_exposure_strata'] = strata
    return result


if __name__ == '__main__':
    print(json.dumps(verify(), indent=2))
