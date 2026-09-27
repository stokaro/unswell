#!/usr/bin/env python3
"""Evaluate reviewed diagnostics; span overlap alone never supplies a judgment."""
import argparse
import json
from pathlib import Path
import subprocess
import sys

from run import command, trace_usage
from study import BASE, ROOT, digest, read, require, validate, verify_freeze, write


def verify_evaluation_freeze(packet):
    frozen = read(ROOT / 'evaluation-freeze.json')
    require(digest((packet / 'freeze.json').read_bytes()) == frozen['packet_freeze_sha256'],
            'Evaluation uses a different request packet')
    repository = ROOT.parents[2]
    for name, expected in frozen['files'].items():
        require(not Path(name).is_absolute() and '..' not in Path(name).parts,
                'Evaluation evidence must use repository-relative paths')
        require(digest((repository / name).read_bytes()) == expected, 'Evaluation source changed: ' + name)
    for name in ['study.py', 'run.py', 'prompt.md', 'protocol.json']:
        require((ROOT / name).read_bytes() == (packet / name).read_bytes(), 'Execution contract changed')


def reference():
    # Reuse the preceding audit's decoder instead of changing event definitions.
    raw = subprocess.check_output([sys.executable, '-B', '-c',
                                   'import json,verify; print(json.dumps(verify.labels()))'], cwd=BASE)
    events = json.loads(raw)
    require(len(events) == 123, 'Frozen event denominator changed')
    baseline = read(BASE / 'review.json')['profiles']['technical']['events']
    require(set(events) == set(baseline), 'Baseline event set differs')
    return events, baseline


def collect(packet, output):
    verify_freeze(packet)
    manifest = read(packet / 'manifest.json')
    ledger = read(output / 'run.json')
    protocol = read(packet / 'protocol.json')
    require(ledger['requested_model'] == protocol['model'] and ledger['effort'] == protocol['effort'],
            'Model or effort differs from the protocol')
    require(ledger['freeze_sha256'] == digest((packet / 'freeze.json').read_bytes()), 'Wrong run inputs')
    rows = {p['page']: p for p in manifest['pages']}
    calls = ledger['calls']
    require(len({c['page'] for c in calls}) == len(calls) and
            {c['page'] for c in calls} <= set(rows), 'Repeated or invented requests')
    predictions = {}; valid_pages = set()
    for call in calls:
        if call['status'] != 'valid':
            continue
        page = call['page']
        require(call['request_sha256'] == rows[page]['request_sha256'], 'Request identity changed')
        args = call['command']
        require(args == command(packet.resolve(), (output / page).resolve(),
                                Path(args[args.index('--cd')+1]), protocol), 'Execution flags changed')
        request = packet / 'requests' / (page + '.json')
        prompt = (packet / 'prompt.md').read_bytes() + b'\n\nDOCUMENT JSON:\n' + request.read_bytes()
        require(digest(prompt) == call['prompt_sha256'], 'Executed prompt differs')
        trace = (output / page / 'events.jsonl').read_text()
        require(call['exit_code'] == 0 and trace_usage(trace) == call['usage'], 'Invalid execution trace')
        messages = [json.loads(line)['item']['text'] for line in trace.splitlines()
                    if json.loads(line).get('type') == 'item.completed' and
                    json.loads(line)['item']['type'] == 'agent_message']
        require(messages and json.loads(messages[-1]) == read(output / page / 'response.json'),
                'Saved response differs from the final CLI message')
        result = validate(read(packet / 'requests' / (page + '.json')),
                          read(output / page / 'response.json'))
        require(result == read(output / page / 'bound.json'), 'Bound result was changed')
        valid_pages.add(rows[page]['reference'])
        for finding in result['findings']:
            predictions[page + '/' + finding['id']] = {**finding, 'path': rows[page]['reference']}
    return predictions, valid_pages, rows


def score(events, baseline, predictions, valid_pages, review, page_paths):
    require(bool(review.get('reviewer')), 'Reviewer identity missing')
    require(valid_pages <= page_paths and all(p['path'] in valid_pages for p in predictions.values()),
            'Predictions claim an unavailable or unknown page')
    findings = review['findings']; credits = review['events']
    require(set(findings) == set(predictions), 'Every emitted finding needs a disposition')
    require(set(credits) == set(events), 'Every frozen event must remain in the ledger')
    for key, judgment in findings.items():
        require(judgment['status'] in ['accepted', 'rejected', 'uncertain'] and
                bool(judgment['rationale'].strip()), 'Missing diagnostic judgment')
        safety = judgment.get('suggestion_safety', {})
        require(safety.get('status') in ['safe', 'unsafe', 'uncertain', 'not_applicable'] and
                isinstance(safety.get('rationale'), str) and bool(safety['rationale'].strip()),
                'Missing independent suggestion safety judgment')
    full = set()
    for key, event in events.items():
        judgment = credits[key]
        require(judgment['status'] in ['full', 'partial', 'missed'] and
                bool(judgment['rationale'].strip()), 'Missing event judgment')
        support = judgment['findings']
        require(isinstance(support, list) and len(support) == len(set(support)), 'Invalid event support')
        require(all(f in predictions and predictions[f]['path'] == event['path'] and
                    findings[f]['status'] == 'accepted' for f in support), 'Unaccepted or unrelated support')
        require(judgment['status'] == 'missed' or support, 'Positive credit has no evidence')
        require(judgment['status'] != 'missed' or not support, 'Missed event has claimed support')
        if judgment['status'] == 'full':
            for target in event['targets']:
                require(any(max(t['span']['start'], target['start']) < min(t['span']['end'], target['end'])
                            for f in support for t in predictions[f]['targets']),
                        'Full credit misses an annotated target')
            full.add(key)
    summary = {}; passed = True
    for cohort in ['whole', 'context', 'confirmation']:
        event_ids = {k for k in events if k.startswith(cohort + '/')}
        ids = {k for k, f in predictions.items() if f['path'].startswith(cohort + '/')}
        accepted = sum(findings[k]['status'] == 'accepted' for k in ids)
        precision = accepted / len(ids) if ids else 0.0
        recall = len(event_ids & full) / len(event_ids)
        prior = {k for k in event_ids if baseline[k]['after']['status'] == 'full'}
        pages = {p for p in page_paths if p.startswith(cohort + '/')}
        complete = pages <= valid_pages
        passed = passed and complete and recall >= .8 and precision >= .85
        safety_counts = {status: sum(findings[k]['suggestion_safety']['status'] == status for k in ids)
                         for status in ['safe', 'unsafe', 'uncertain', 'not_applicable']}
        summary[cohort] = {'pages': len(pages), 'valid_pages': len(pages & valid_pages),
                           'events': len(event_ids), 'full': len(event_ids & full),
                           'recall': recall, 'findings': len(ids), 'accepted': accepted,
                           'accepted_fraction': precision,
                           'suggestion_safety': safety_counts,
                           'accepted_diagnostics_with_unsafe_suggestions': sum(
                               findings[k]['status'] == 'accepted' and
                               findings[k]['suggestion_safety']['status'] == 'unsafe' for k in ids),
                           'baseline_full': len(prior), 'baseline_union_full': len(prior | (event_ids & full))}
    return {'cohorts': summary, 'development_screen_passed': passed,
            'product_qualified': False,
            'scope': 'Exposed same-assistant development; fresh whole-page confirmation still required.'}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('packet', type=Path); parser.add_argument('output', type=Path)
    parser.add_argument('--review', type=Path)
    parser.add_argument('--new-review', type=Path)
    args = parser.parse_args()
    verify_evaluation_freeze(args.packet)
    predictions, valid, pages = collect(args.packet, args.output)
    events, baseline = reference()
    if args.new_review:
        require(not args.new_review.exists(), 'Refuse to overwrite a review')
        write(args.new_review, {'reviewer': '',
                               'findings': {k: {'status': 'unreviewed', 'rationale': '',
                                                'suggestion_safety': {'status': 'unreviewed', 'rationale': ''}}
                                            for k in predictions},
                               'events': {k: {'status': 'unreviewed', 'findings': [], 'rationale': ''} for k in events}})
    else:
        require(args.review is not None, 'Supply an explicit source-bound review')
        print(json.dumps(score(events, baseline, predictions, valid, read(args.review),
                               {p['reference'] for p in pages.values()}), indent=2))


if __name__ == '__main__':
    main()
