#!/usr/bin/env python3
"""Verify window ledgers and reuse the frozen complete-page scoring contract."""
import argparse
import json
from pathlib import Path

import accounting as a
import execute

prior_evaluate = execute.cli_compat.evaluate


def collect(packet, output):
    packet = packet.resolve(); output = output.resolve()
    manifest, protocol = execute.verify(packet); ledger = a.read(output / 'run.json')
    a.require(ledger['model'] == protocol['model'] and ledger['effort'] == protocol['effort'] and
              ledger['packet_freeze_sha256'] == a.digest((packet / 'freeze.json').read_bytes()),
              'Run identity differs')
    rows = manifest['requests']; calls = ledger['calls']
    a.require(len(calls) <= len(rows) and [c['request'] for c in calls] ==
              [r['request'] for r in rows[:len(calls)]], 'Repeated, reordered, or extra requests')
    complete = len(calls) == len(rows) and all(c['status'] == 'valid' for c in calls)
    a.require(ledger['complete'] is complete and ledger['unattempted'] ==
              [r['request'] for r in rows[len(calls):]], 'False completion or missing tail')
    invalid = [i for i, c in enumerate(calls) if c['status'] != 'valid']
    a.require(not invalid or invalid == [len(calls) - 1], 'Request made after terminal failure')
    a.require(all(c['status'] in ['valid', 'invalid'] for c in calls), 'Run is not terminal')
    if not complete:
        a.require(ledger.get('stop_reason') == ('first_failed_request_no_retry' if invalid
                                              else 'total_time_budget'), 'Missing failure policy')
    valid_responses = {}; request_data = {}; page_rows = {}
    for row in rows:
        page_rows.setdefault(row['page'], []).append(row)
        request_data[row['request']] = a.read(packet / 'requests' / (row['request'] + '.json'))
    for row, call in zip(rows, calls):
        name = row['request']; args = call['command']; target = output / name
        a.require(call['page'] == row['page'] and call['request_sha256'] == row['request_sha256'],
                  'Call identity changed')
        recorded_packet = Path(args[args.index('--output-schema') + 1]).parent
        recorded_target = Path(args[args.index('--output-last-message') + 1]).parent
        a.require(recorded_target.name == name, 'Call output identity changed')
        a.require(args == execute.prior_run.command(recorded_packet, recorded_target,
                  Path(args[args.index('--cd') + 1]), protocol), 'Execution flags changed')
        request = packet / 'requests' / (name + '.json')
        a.require(a.digest(execute.prompt(packet, request)) == call['prompt_sha256'], 'Prompt changed')
        if call['status'] != 'valid':
            # Source-invalid completed responses remain auditable; never replace
            # one with a fixed quote or grant its findings research credit.
            if call.get('exit_code') == 0 and not call.get('timed_out'):
                try:
                    execute.check_response(packet, output, row, dict(call))
                except (ValueError, KeyError, TypeError, OSError):
                    pass
                else:
                    raise ValueError('Recorded invalid response now validates')
            continue
        a.require(call['exit_code'] == 0 and call.get('timed_out') is False, 'Valid call failed')
        checked_call = dict(call)
        result = execute.check_response(packet, output, row, checked_call)
        a.require(result == a.read(target / 'bound.json') and
                  checked_call['usage'] == call['usage'], 'Bound output or usage changed')
        valid_responses[name] = a.read(target / 'response.json')
    predictions = {}; partial = {}; complete_pages = set(); page_paths = set()
    for page, members in page_rows.items():
        path = members[0]['reference']; page_paths.add(path)
        a.require(all(row['reference'] == path for row in members), 'Page reference changed')
        names = [r['request'] for r in members]
        if all(name in valid_responses for name in names):
            result = a.assemble([request_data[n] for n in names], [valid_responses[n] for n in names])
            complete_pages.add(path)
            for finding in result['findings']:
                predictions[finding['id']] = {**finding, 'path': path}
        else:
            for name in names:
                if name not in valid_responses: continue
                result = a.validate(request_data[name], valid_responses[name])
                for finding in result['findings']:
                    key = name + '/' + finding['id']
                    partial[key] = {**finding, 'id': key, 'path': path}
    return predictions, partial, complete_pages, page_paths


def score(predictions, partial, complete_pages, page_paths, review):
    events, baseline = prior_evaluate.reference()
    all_predictions = {**predictions, **partial}
    a.require(set(review['findings']) == set(all_predictions), 'Every emission needs review')
    for row in review['findings'].values():
        a.require(row['status'] in ['accepted', 'rejected', 'uncertain'] and row['rationale'].strip(),
                  'Missing diagnostic disposition')
        safety = row['suggestion_safety']
        a.require(safety['status'] in ['safe', 'unsafe', 'uncertain', 'not_applicable'] and
                  safety['rationale'].strip(), 'Missing suggestion safety judgment')
    full_review = {**review, 'findings': {key: review['findings'][key] for key in predictions}}
    result = prior_evaluate.score(events, baseline, predictions, complete_pages, full_review, page_paths)
    result['all_valid_request_emissions'] = {}
    for cohort in ['whole', 'context', 'confirmation']:
        keys = [k for k, p in all_predictions.items() if p['path'].startswith(cohort + '/')]
        accepted = sum(review['findings'][k]['status'] == 'accepted' for k in keys)
        fraction = accepted / len(keys) if keys else 0.0
        result['all_valid_request_emissions'][cohort] = {
            'findings': len(keys), 'accepted': accepted, 'accepted_fraction': fraction,
            'partial_page_findings': sum(k in partial for k in keys)}
        result['development_screen_passed'] &= fraction >= .85
    return result


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('packet', type=Path); parser.add_argument('output', type=Path)
    parser.add_argument('--review', type=Path); parser.add_argument('--new-review', type=Path)
    args = parser.parse_args()
    predictions, partial, pages, paths = collect(args.packet, args.output)
    if args.new_review:
        a.require(not args.new_review.exists(), 'Refuse to overwrite judgments')
        events, _ = prior_evaluate.reference()
        a.write(args.new_review, {'reviewer': '', 'findings': {
            key: {'status': 'unreviewed', 'rationale': '', 'suggestion_safety': {
                'status': 'unreviewed', 'rationale': ''}} for key in {**predictions, **partial}},
            'events': {key: {'status': 'unreviewed', 'findings': [], 'rationale': ''} for key in events}})
    else:
        a.require(args.review is not None, 'Explicit review required')
        print(json.dumps(score(predictions, partial, pages, paths, a.read(args.review)), indent=2))
