#!/usr/bin/env python3
"""Verify the terminal pilot and score all frozen events without repair advice."""
import argparse
import copy
import importlib.util
import json
from pathlib import Path

import diagnosis as a
import execute

SPEC = importlib.util.spec_from_file_location('frozen_contextual_evaluator',
                                               a.OLD.PRIOR / 'evaluate.py')
prior_evaluate = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(prior_evaluate)


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
        artifacts = execute.artifacts(target)
        a.require(call['artifacts'] == artifacts, 'Raw call artifact changed')
        if call.get('exit_code') is not None:
            a.require(artifacts['events.jsonl'] is not None and
                      artifacts['stderr.txt'] is not None,
                      'Completed call lost its raw trace')
        recorded_packet = Path(args[args.index('--output-schema') + 1]).parent
        recorded_target = Path(args[args.index('--output-last-message') + 1]).parent
        a.require(recorded_packet.resolve() == packet, 'Call used another packet')
        a.require(recorded_target.resolve() == target, 'Call used another output directory')
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
        a.require(artifacts['response.json'] is not None, 'Valid call lost its response')
        checked_call = dict(call)
        result = execute.check_response(packet, output, row, checked_call)
        a.require(result == a.read(target / 'bound.json') and
                  checked_call['usage'] == call['usage'], 'Bound output or usage changed')
        valid_responses[name] = a.read(target / 'response.json')
    predictions = {}; partial = {}; complete_pages = set()
    assessment_keys = set()
    decisions = {'no_burden': 0, 'diagnose': 0, 'uncertain': 0}
    partial_pages = 0
    page_paths = set(manifest['source_references'])
    a.require(len(page_paths) == 36, 'Original page denominators changed')

    def count_assessments(path, result):
        for assessment in result['assessments']:
            key = (path, assessment['unit'])
            a.require(key not in assessment_keys, 'Unit assessed more than once')
            assessment_keys.add(key)
            decisions[assessment['burden_decision']] += 1

    for page, members in page_rows.items():
        path = members[0]['reference']
        a.require(path in page_paths, 'Pilot page missing from original source')
        a.require(all(row['reference'] == path for row in members), 'Page reference changed')
        names = [r['request'] for r in members]
        if all(name in valid_responses for name in names):
            result = a.assemble([request_data[n] for n in names], [valid_responses[n] for n in names])
            complete_pages.add(path)
            count_assessments(path, result)
            for finding in result['findings']:
                predictions[finding['id']] = {**finding, 'path': path}
        else:
            partial_pages += any(name in valid_responses for name in names)
            for name in names:
                if name not in valid_responses: continue
                result = a.validate(request_data[name], valid_responses[name])
                count_assessments(path, result)
                for finding in result['findings']:
                    key = name + '/' + finding['id']
                    partial[key] = {**finding, 'id': key, 'path': path}
    focus_units = sum(len(request['focus']) for request in request_data.values())
    valid_units = sum(len(request_data[name]['focus']) for name in valid_responses)
    a.require(len(assessment_keys) == valid_units, 'Valid-request unit coverage changed')
    invalid_pages = {row['page'] for row, call in zip(rows, calls)
                     if call['status'] == 'invalid'}
    attempted_pages = {call['page'] for call in calls}
    coverage = {'requests': len(rows), 'valid_requests': len(valid_responses),
                'failed_requests': len(calls) - len(valid_responses),
                'unattempted_requests': len(rows) - len(calls),
                'pages': len(page_rows), 'complete_pages': len(complete_pages),
                'partial_pages': partial_pages, 'invalid_pages': len(invalid_pages),
                'unattempted_pages': len(page_rows) - len(attempted_pages),
                'incomplete_pages': len(page_rows) - len(complete_pages),
                'focus_units': focus_units, 'assessed_units': valid_units,
                'unassessed_units': focus_units - valid_units,
                'burden_decisions': decisions}
    return predictions, partial, complete_pages, page_paths, coverage


def score(predictions, partial, complete_pages, page_paths, review, coverage):
    events, baseline = prior_evaluate.reference()
    a.require(complete_pages <= set(a.PAGES) and complete_pages <= page_paths and
              all(event['path'] in page_paths for event in events.values()),
              'Available pages or reference paths changed')
    counters = [coverage['valid_requests'], coverage['failed_requests'],
                coverage['unattempted_requests'], coverage['assessed_units'],
                coverage['unassessed_units'], coverage['partial_pages'],
                coverage['invalid_pages'], coverage['unattempted_pages'],
                *coverage['burden_decisions'].values()]
    a.require(all(type(value) is int and value >= 0 for value in counters),
              'Invalid pilot coverage counter')
    a.require(coverage['requests'] == 7 and coverage['pages'] == 6 and
              coverage['focus_units'] == 310 and
              coverage['complete_pages'] == len(complete_pages) and
              coverage['incomplete_pages'] == 6 - len(complete_pages) and
              coverage['valid_requests'] + coverage['failed_requests'] +
              coverage['unattempted_requests'] == 7 and
              coverage['assessed_units'] + coverage['unassessed_units'] == 310 and
              sum(coverage['burden_decisions'].values()) == coverage['assessed_units'] and
              set(coverage['burden_decisions']) == {'no_burden', 'diagnose', 'uncertain'} and
              coverage['partial_pages'] <= coverage['incomplete_pages'] and
              coverage['invalid_pages'] + coverage['unattempted_pages'] <=
              coverage['incomplete_pages'],
              'Pilot coverage accounting changed')
    all_predictions = {**predictions, **partial}
    a.require(set(review['findings']) == set(all_predictions), 'Every emission needs review')
    for row in review['findings'].values():
        a.require(row['status'] in ['accepted', 'rejected', 'uncertain'] and row['rationale'].strip(),
                  'Missing diagnostic disposition')
        safety = row['suggestion_safety']
        a.require(safety['status'] == 'not_applicable' and safety['rationale'].strip(),
                  'A diagnosis-only pilot cannot claim repair safety')
    a.require(set(review['events']) == set(events), 'Every frozen event must remain in the ledger')
    scored_events = {}
    for key, event in events.items():
        judgment = review['events'][key]
        a.require(isinstance(judgment['rationale'], str) and judgment['rationale'].strip(),
                  'Missing event rationale')
        if event['path'] in complete_pages:
            a.require(judgment['status'] in ['full', 'partial', 'missed'],
                      'Available event needs a source judgment')
            scored_events[key] = judgment
        else:
            a.require(judgment['status'] == 'unavailable' and judgment['findings'] == [],
                      'Unavailable event cannot claim a semantic result')
            scored_events[key] = {'status': 'missed', 'findings': [],
                                  'rationale': 'Unavailable page in this pilot; no semantic miss is claimed.'}
    full_review = {**review,
                   'findings': {key: review['findings'][key] for key in predictions},
                   'events': scored_events}
    result = prior_evaluate.score(events, baseline, predictions, complete_pages, full_review, page_paths)
    result['all_valid_request_emissions'] = {}
    for cohort in ['whole', 'context', 'confirmation']:
        keys = [k for k, p in all_predictions.items() if p['path'].startswith(cohort + '/')]
        dispositions = {status: sum(review['findings'][k]['status'] == status for k in keys)
                        for status in ['accepted', 'rejected', 'uncertain']}
        accepted = dispositions['accepted']
        fraction = accepted / len(keys) if keys else None
        result['all_valid_request_emissions'][cohort] = {
            'findings': len(keys), **dispositions, 'accepted_fraction': fraction,
            'partial_page_findings': sum(k in partial for k in keys)}
        cohort_result = result['cohorts'][cohort]
        unavailable = sum(key.startswith(cohort + '/') and
                          event['path'] not in complete_pages for key, event in events.items())
        cohort_result['unavailable_events'] = unavailable
        cohort_result['available_events'] = cohort_result['events'] - unavailable
        cohort_result['available_recall'] = (cohort_result['full'] /
                                             cohort_result['available_events']
                                             if cohort_result['available_events'] else None)
        if cohort_result['valid_pages'] != cohort_result['pages']:
            cohort_result['recall'] = None
        if not cohort_result['findings']:
            cohort_result['accepted_fraction'] = None
    context = result['cohorts']['context']
    all_context = result['all_valid_request_emissions']['context']
    a.require(len(page_paths) == 36 and set(a.PAGES) <= page_paths and
              result['cohorts']['whole']['pages'] == 26 and
              context['pages'] == 6 and context['events'] == 46 and
              result['cohorts']['confirmation']['pages'] == 4,
              'Changed frozen page denominator')
    result['pilot_screen_passed'] = (
        context['valid_pages'] == 6 and context['full'] >= 37 and
        all_context['accepted_fraction'] is not None and
        all_context['accepted_fraction'] >= .85)
    result['development_screen_passed'] = False
    result['product_qualified'] = False
    result['pilot_coverage'] = copy.deepcopy(coverage)
    result['scope'] = ('Exposed six-page context pilot only; 30 pages and their events remain '
                       'unavailable. A full development screen and separate prospective '
                       'confirmation are still required.')
    return result


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('packet', type=Path); parser.add_argument('output', type=Path)
    parser.add_argument('--review', type=Path); parser.add_argument('--new-review', type=Path)
    args = parser.parse_args()
    predictions, partial, pages, paths, coverage = collect(args.packet, args.output)
    if args.new_review:
        a.require(not args.new_review.exists(), 'Refuse to overwrite judgments')
        events, _ = prior_evaluate.reference()
        a.write(args.new_review, {'reviewer': '', 'findings': {
            key: {'status': 'unreviewed', 'rationale': '', 'suggestion_safety': {
                'status': 'not_applicable',
                'rationale': 'No repair or suggestion is emitted by this pilot.'}}
            for key in {**predictions, **partial}},
            'events': {key: {'status': 'unreviewed', 'findings': [], 'rationale': ''}
                       if event['path'] in pages else
                       {'status': 'unavailable', 'findings': [],
                        'rationale': 'No complete valid page is available in this pilot.'}
                       for key, event in events.items()}})
    else:
        a.require(args.review is not None, 'Explicit review required')
        print(json.dumps(score(predictions, partial, pages, paths, a.read(args.review), coverage),
                         indent=2))
