#!/usr/bin/env python3
"""Report full-scope model evidence; never infer semantic credit from overlap."""
import argparse
import json
from collections import Counter
from pathlib import Path

from evaluate import reference, score
from evaluate_compat import collect
from study import read, require, validate


def report(packet, output, review_path):
    predictions, valid, pages = collect(packet, output)
    events, baseline = reference(); review = read(review_path)
    result = score(events, baseline, predictions, valid, review,
                   {row['reference'] for row in pages.values()})
    ledger = read(output / 'run.json'); calls = {c['page']: c for c in ledger['calls']}
    require(all(c['status'] in ['valid', 'invalid'] for c in calls.values()), 'Run is not terminal')
    require(ledger['unattempted'] == [page for page in pages if page not in calls],
            'Unattempted page ledger differs')
    require(ledger['complete'] == (len(valid) == len(pages)), 'Completion ledger differs')
    for cohort, metrics in result['cohorts'].items():
        available = [e for key, e in events.items()
                     if key.startswith(cohort + '/') and e['path'] in valid]
        metrics['events_on_valid_pages'] = len(available)
        metrics['events_on_unavailable_pages'] = metrics['events'] - len(available)
        metrics['recall_on_valid_pages_only'] = metrics['full'] / len(available) if available else None
    result['execution'] = {
        'requested_model': ledger['requested_model'], 'effort': ledger['effort'],
        'cli_version': ledger['cli_version'], 'calls': len(calls),
        'complete': ledger['complete'], 'unattempted': ledger['unattempted'],
        'stop_reason': ledger.get('stop_reason'), 'monetary_cost': ledger['monetary_cost'],
        'compatibility_amendment_sha256': ledger['compatibility_amendment_sha256'],
        'usage': {name: sum(c.get('usage', {}).get(name, 0) for c in calls.values())
                  for name in ['input_tokens', 'cached_input_tokens', 'cache_write_input_tokens',
                               'output_tokens', 'reasoning_output_tokens']},
        'request_seconds': sum(c.get('seconds', 0) for c in calls.values()),
    }
    result['invalid_responses'] = []
    for page, call in calls.items():
        if call['status'] != 'invalid':
            continue
        response_path = output / page / 'response.json'
        raw_count = None
        if response_path.exists():
            response = read(response_path)
            raw_count = len(response.get('findings', []))
            try:
                validate(read(packet / 'requests' / (page + '.json')), response)
            except ValueError as error:
                require(str(error) == call['error'], 'Invalid response no longer reproduces its recorded error')
            else:
                require(call['error'] != 'Quotation is not in the selected source unit',
                        'Recorded quotation failure no longer occurs')
        result['invalid_responses'].append({'page': page, 'reason': call['error'],
                                            'raw_findings_excluded': raw_count})
    result['pages'] = []
    for page, row in pages.items():
        path = row['reference']; keys = [k for k, e in events.items() if e['path'] == path]
        findings = [k for k, p in predictions.items() if p['path'] == path]
        statuses = Counter(review['events'][k]['status'] for k in keys)
        dispositions = Counter(review['findings'][k]['status'] for k in findings)
        result['pages'].append({
            'page': page, 'reference': path, 'source_sha256': row['source_sha256'],
            'bytes': row['bytes'], 'units': row['units'],
            'execution_status': calls.get(page, {}).get('status', 'unattempted'),
            'events': len(keys), 'event_statuses': dict(statuses),
            'findings': len(findings), 'finding_dispositions': dict(dispositions),
            'full_event_recall': statuses['full']/len(keys) if keys else None,
            'accepted_findings_fraction': dispositions['accepted']/len(findings) if findings else None,
        })
    result['categories'] = []
    for cohort in ['whole', 'context', 'confirmation']:
        categories = sorted({e['category'] for key, e in events.items() if key.startswith(cohort+'/')})
        for category in categories:
            keys = [key for key, e in events.items()
                    if key.startswith(cohort+'/') and e['category'] == category]
            full = sum(review['events'][key]['status'] == 'full' for key in keys)
            result['categories'].append({'cohort': cohort, 'category': category,
                                         'events': len(keys), 'full': full, 'recall': full/len(keys)})
    require(sum(p['events'] for p in result['pages']) == len(events) == 123,
            'The page report dropped frozen events')
    require(sum(p['findings'] for p in result['pages']) == len(predictions),
            'The page report dropped emitted findings')
    result['limitations'] = [
        'Exposed development pages and judgments by the same Codex assistant; not independent human review or population accuracy.',
        'Exact-source validation is not evidence of diagnostic correctness. Full credit is supplied by explicit reviewed event mappings.',
        'A safe suggestion is not necessarily a complete repair. No generated edits were applied.',
        'The one exact CLI startup notice was classified by a disclosed compatibility amendment; original trace and failed ledger remain.',
        'Model, prompt, and request contract differ from earlier model trials; this is not a causal model-only ablation.',
        'Unavailable events stay in planned recall denominators; they are not observed semantic misses. Valid-page recall is supplementary and cannot pass completeness.',
    ]
    return result


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('packet', type=Path); parser.add_argument('output', type=Path)
    parser.add_argument('--review', type=Path, required=True)
    args = parser.parse_args()
    print(json.dumps(report(args.packet, args.output, args.review), indent=2))
