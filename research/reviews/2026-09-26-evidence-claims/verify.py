#!/usr/bin/env python3
"""Verify complete paired reports and source-bound editorial judgments offline."""
import gzip
import hashlib
import json
from pathlib import Path
import tarfile

ROOT = Path(__file__).resolve().parent
REFERENCE = ROOT.parent / '2026-09-17-full-page-recall'


def read(path):
    data = path.read_bytes()
    return json.loads(gzip.decompress(data) if path.suffix == '.gz' else data)


def require(value, message):
    if not value:
        raise ValueError(message)


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def key(finding):
    return json.dumps({k: v for k, v in finding.items()
                       if k not in {'id', 'fingerprint', 'rule_version'}}, sort_keys=True)


def locations(finding):
    return [finding['primary'], *finding['related']]


def overlaps(finding, path, targets):
    return any(loc['path'] == path and
               any(max(loc['span']['start'], t['start']) < min(loc['span']['end'], t['end'])
                   for t in targets) for loc in locations(finding))


def report(path, sources):
    data = read(path)
    require(data['status'] == 'complete' and data['manifest']['complete'] and not data['errors'], 'Incomplete scan')
    require({d['name'] for d in data['documents']} == set(sources), 'Missing source')
    for doc in data['documents']:
        require(doc['source'].encode() == sources[doc['name']], 'Changed report source')
    for finding in data['findings']:
        for loc in locations(finding):
            span = loc['span']
            require(0 <= span['start'] < span['end'] <= len(sources[loc['path']]), 'Invalid span')
            require(sources[loc['path']][span['start']:span['end']].decode() == loc['snippet'], 'Invented snippet')
    return data['findings']


def verify(root=ROOT):
    for name in ['freeze.json', 'confirmation-freeze.json', 'artifacts.json']:
        for file, expected in read(root / name)['files'].items():
            require(digest(root / file) == expected, 'Changed frozen file: ' + file)
    original = read(REFERENCE / 'freeze.json')
    for file, expected in original['files'].items():
        require(digest(REFERENCE / file) == expected, 'Changed original reference')
    manifest = read(REFERENCE / 'manifest.json')
    pages = {p['id']: p for p in manifest['pages']}
    with tarfile.open(REFERENCE / 'inputs.tar.gz') as archive:
        sources = {p['path']: archive.extractfile(p['path']).read() for p in pages.values()}
    labels = {d['id']: (p['page'], d) for p in read(REFERENCE / 'annotations.json') for d in p['defects']}
    review = read(root / 'review.json')
    require(set(review['events']) == set(labels), 'Incomplete event ledger')
    confirmation = read(root / 'confirmation-inputs.json')['pages']
    confirm_sources = {p['id'] + '.md': p['source'].encode() for p in confirmation}
    confirm_labels = read(root / 'confirmation-labels.json')['reviews']
    require({p['id'] for p in confirm_labels} == {p['id'] for p in confirmation}, 'Missing page review')
    for page in confirm_labels:
        source = confirm_sources[page['id'] + '.md']
        require(hashlib.sha256(source).hexdigest() == page['sha256'], 'Changed confirmation source')
        require(page['reviewed_ranges'][0]['start'] == 0 and page['reviewed_ranges'][-1]['end'] == len(source), 'Incomplete review')
        for defect in page['defects']:
            require(source[defect['start']:defect['end']].decode() == defect['quote'], 'Changed label span')
    summary = {}
    for profile in ['technical', 'strict']:
        before = report(root / f'development-before-{profile}.json.gz', sources)
        after = report(root / f'development-after-{profile}.json.gz', sources)
        old = {key(f) for f in before}
        new = [f for f in after if key(f) not in old]
        require(old <= {key(f) for f in after}, 'Previous finding changed or removed')
        require(len(new) == len(review['additions']), 'Unreviewed additional finding')
        for finding in new:
            matches = [r for r in review['additions'] if r['path'] == finding['primary']['path'] and
                       r['snippet'] == finding['primary']['snippet'] and r['rule_id'] == finding['rule_id']]
            require(len(matches) == 1 and matches[0]['disposition'] == 'accepted', 'Missing disposition')
        counts = {'before': 0, 'after': 0}
        cohorts = {arm: {} for arm in counts}
        for event, row in review['events'].items():
            pid, label = labels[event]
            cohort = 'anchor' if pages[pid]['selection'] == 'exposed_anchor' else pages[pid]['cohort']
            for arm, findings in [('before', before), ('after', after)]:
                matched = row[arm] == 'full'
                require(row[arm] in ['full', 'partial', 'missed'], 'Unknown match status')
                if matched:
                    require(any(f['rule_id'] == row['rule_id'] and overlaps(f, pages[pid]['path'], label['targets']) for f in findings), 'Unbacked full match')
                counts[arm] += matched
                bucket = cohorts[arm].setdefault(cohort, {'matched': 0, 'total': 0})
                bucket['matched'] += matched
                bucket['total'] += 1
        cb = report(root / f'confirmation-before-{profile}.json.gz', confirm_sources)
        ca = report(root / f'confirmation-after-{profile}.json.gz', confirm_sources)
        require([key(f) for f in cb] == [key(f) for f in ca], 'Confirmation findings changed')
        dispositions = review['confirmation'][profile]
        require(len(dispositions) == len(ca), 'Missing confirmation disposition')
        for f, row in zip(ca, dispositions):
            require(row['rule_id'] == f['rule_id'] and row['path'] == f['primary']['path'] and row['snippet'] == f['primary']['snippet'], 'Confirmation review mismatch')
        summary[profile] = dict(development=dict(pages=len(pages), total_events=len(labels), matched=counts,
                                                cohorts=cohorts, findings_before=len(before), findings_after=len(after), additions=len(new), removed=0),
                                confirmation=dict(pages=len(confirmation), events=sum(len(p['defects']) for p in confirm_labels),
                                                  full_matches=0, added=0, findings=len(ca),
                                                  accepted=sum(r['disposition']=='accepted' for r in dispositions),
                                                  uncertain=sum(r['disposition']=='uncertain' for r in dispositions),
                                                  rejected=sum(r['disposition']=='rejected' for r in dispositions)))
    return summary


if __name__ == '__main__':
    print(json.dumps(verify(), indent=2))
