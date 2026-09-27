#!/usr/bin/env python3
"""Verify the separately frozen prospective Ptah review without running a detector."""
import argparse
import hashlib
from pathlib import Path

from study import ROOT, digest, read, require


def verify(root, check_freeze=True):
    if check_freeze:
        frozen = read(root / 'review-freeze.json')
        for name, expected in frozen.items():
            require(not Path(name).is_absolute() and '..' not in Path(name).parts,
                    'Unsafe frozen path')
            require(digest((root / name).read_bytes()) == expected, 'Frozen review changed: ' + name)
    source_freeze = read(root / 'source-freeze.json')
    for name, expected in source_freeze.items():
        require(not Path(name).is_absolute() and '..' not in Path(name).parts, 'Unsafe source path')
        require(digest((root / name).read_bytes()) == expected, 'Source changed: ' + name)
    selection = read(root / 'selection.json'); review = read(root / 'annotations.json')
    require(len(selection['pages']) == 3, 'Confirmation page set changed')
    require(review['source_freeze_sha256'] == digest((root / 'source-freeze.json').read_bytes()),
            'Review source identity changed')
    require(review['model_outputs_seen'] is False and review['detector_outputs_seen'] is False,
            'This artifact must remain the pre-output review')
    amendment = read(root / 'review-amendment.json')
    initial = read(root / 'annotations-initial.json')
    require(amendment['initial_annotations_sha256'] == digest((root / 'annotations-initial.json').read_bytes())
            and amendment['current_annotations_sha256'] == digest((root / 'annotations.json').read_bytes()),
            'Amendment identity differs')
    prior = {r['id']: r for r in initial['judgments']}
    current = {r['id']: r for r in review['judgments']}
    require(prior.keys() == current.keys() and initial['pages'] == review['pages'],
            'Amendment removed a judgment or changed its source')
    changed = {key for key in prior if prior[key] != current[key]}
    require(changed == {row['id'] for row in amendment['changes']}, 'Undocumented amendment')
    for row in amendment['changes']:
        before, after = prior[row['id']], current[row['id']]
        require(before['status'] == row['from'] and after['status'] == row['to'] and
                before['targets'] == after['targets'] and before['page'] == after['page'] and
                row['reason'] == after['rationale'], 'Amendment differs from the preserved judgment')
    require(bool(review['reviewer'].strip()), 'Reviewer missing')
    pages = {p['page']: p for p in review['pages']}
    require(len(pages) == len(review['pages']) == len(selection['pages']) and
            set(pages) == {1, 2, 3}, 'Missing or repeated page review')
    sources = {}; bodies = {}
    for i, selected in enumerate(selection['pages'], 1):
        page = pages[i]; name = selected['source_file']
        require(page['id'] == selected['id'] and page['source_file'] == name, 'Page identity changed')
        require(name in source_freeze, 'Unfrozen source')
        data = (root / name).read_bytes(); data.decode('utf-8')
        require(digest(data) == page['source_sha256'] and len(data) == selected['bytes'],
                'Source hash or size differs')
        blob = hashlib.sha1(b'blob ' + str(len(data)).encode() + b'\0' + data).hexdigest()
        require(blob == selected['git_blob'], 'Pinned Git source differs')
        require(data.startswith(b'---\n') and b'\n---\n' in data[4:], 'Expected front matter')
        start = data.index(b'\n---\n', 4) + 5; bodies[i] = start
        for section in page['sections']:
            span = section['span']
            require(section['status'] == 'reviewed' and span['start'] == start and
                    start < span['end'] <= len(data), 'Incomplete body coverage')
            data[span['start']:span['end']].decode('utf-8')
            start = span['end']
        require(start == len(data), 'Unreviewed end of document')
        sources[i] = data
    seen = set(); counts = {'defect': 0, 'acceptable': 0, 'uncertain': 0}
    for row in review['judgments']:
        require(row['id'] not in seen and row['page'] in pages, 'Duplicate or unknown judgment')
        seen.add(row['id'])
        require(row['status'] in counts, 'Unknown judgment status'); counts[row['status']] += 1
        require(row['rationale'].strip() and row['meaning_to_preserve'].strip(), 'Missing reasoning')
        require(row['status'] != 'defect' or bool(row['proposed_repair']), 'Defect has no proposed repair')
        require(bool(row['targets']), 'Judgment has no source target')
        for target in row['targets']:
            span = target['span']; data = sources[row['page']]
            require(bodies[row['page']] <= span['start'] < span['end'] <= len(data),
                    'Target outside reviewed prose body')
            require(data[span['start']:span['end']].decode('utf-8') == target['quote'],
                    'Quotation and source range differ')
    return {'pages': len(pages), 'judgments': counts, 'detector_run': False, 'model_run': False,
            'claim': 'Source consistency only; editorial judgments are attributed assistant review.'}


if __name__ == '__main__':
    import json
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('root', nargs='?', type=Path, default=ROOT / 'confirmation')
    print(json.dumps(verify(parser.parse_args().root), indent=2))
