#!/usr/bin/env python3
"""Join frozen candidate witnesses to v30 and separated-judgment outcomes."""

import argparse
from collections import Counter, defaultdict
import hashlib
import json
from pathlib import Path
import re

from verify import DENOMINATORS, cohort, digest, read, require


def check_expanded(original, grades, source_manifest, source_root):
    require(grades['complete'] and len(original['original']) == 36 and
            len(original['paths']) == 36 and len(original['groups']) == 13 and
            len(original['reference']['events']) == len(grades['events']) == 123,
            'Expanded scope changed')
    sources = {row['source_sha256']: row for row in source_manifest['sources']}
    require(len(sources) == 36, 'Source manifest has missing or repeated hashes')
    candidates = {}
    pages = {}
    raw_by_path = {}
    group_by_page = {}
    for group, numbers in original['groups'].items():
        for number in numbers:
            page = f'page-{number:03}'
            require(page not in group_by_page, 'Page belongs to two frozen source groups')
            group_by_page[page] = group
    require(set(group_by_page) == set(original['original']),
            'Frozen source group coverage changed')
    for page, data in original['original'].items():
        path = original['paths'][page]
        document = data['document']
        raw = document['source'].encode('utf-8')
        source_hash = hashlib.sha256(raw).hexdigest()
        source = sources.get(source_hash)
        require(source is not None and source['path'].endswith('/inputs/' + path)
                and (source_root / 'inputs' / path).read_bytes() == raw,
                'Expanded source identity changed')
        units = {unit['id']: unit for unit in data['raw_units']}
        require(len(units) == len(data['raw_units']), 'Repeated source unit')
        for unit in units.values():
            span = unit['span']
            require(raw[span['start']:span['end']].decode('utf-8') == unit['raw'],
                    'Raw source-unit slice changed')
        pages[page] = {'path': path, 'source_sha256': source_hash,
                       'source_group': group_by_page[page],
                       'repository': source['repository']}
        raw_by_path[path] = raw
        for candidate in data['candidates']:
            identifier = candidate['id']
            require(identifier not in candidates and candidate['targets'],
                    'Repeated or unlocated candidate')
            targets = []
            for target in candidate['targets']:
                span = target['span']
                require(raw[span['start']:span['end']].decode('utf-8') ==
                        target['quote'], 'Candidate quote or byte span changed')
                unit_ids = target.get('units', [target.get('unit')])
                require(unit_ids and all(item in units for item in unit_ids),
                        'Candidate has unknown unit')
                targets.append({'start': span['start'], 'end': span['end'],
                                'unit_ids': unit_ids})
            candidates[identifier] = {'page': page, 'targets': targets,
                                      'category': candidate['category'],
                                      'origin': candidate['origin'],
                                      'claim_sha256': hashlib.sha256(
                                          (candidate['diagnostic'] + '\n' +
                                           candidate['rationale']).encode('utf-8')).hexdigest()}
    require(len(candidates) == len(grades['findings']) == 1154 and
            set(candidates) == set(grades['findings']) == set(original['labels']),
            'Candidate and grade inventory differs')
    require(Counter(row['status'] for row in grades['findings'].values()) ==
            {'accepted': 795, 'uncertain': 320, 'rejected': 39},
            'Frozen grade counts changed')
    for event_id, event in original['reference']['events'].items():
        require(event_id in grades['events'] and event['path'] in
                {page['path'] for page in pages.values()},
                'Reference event or source page missing')
        for target in event['targets']:
            span = target
            require(raw_by_path[event['path']][span['start']:span['end']].decode('utf-8')
                    == target['quote'], 'Reference event quote or byte span changed')
        for group in grades['events'][event_id]['full_support_groups']:
            require(group and all(item in candidates for item in group),
                    'Event has unknown complete-support candidate')
    return pages, candidates


def check_separated(private, public, public_manifest, original, pages, candidates):
    require(private['original_scope'] == public['original_scope'] ==
            {'pages': 36, 'units': 4549, 'judgments': 804, 'events': 123},
            'Separated scope changed')
    by_path = {row['path']: row for row in public_manifest['pages']}
    require(len(by_path) == 36 and len(pages) == 36,
            'Separated page manifest incomplete')
    for page in pages.values():
        row = by_path.get(page['path'])
        require(row is not None and row['source_sha256'] == page['source_sha256'],
                'Separated source identity differs')
    selected = set(private['all_selected_actual_aliases'])
    require(len(selected) == len(private['all_selected_actual_aliases']) ==
            sum(row['delivered'] for row in private['cohorts'].values()) == 2847,
            'Separated stream has hidden or repeated delivery')
    for alias in selected:
        if alias.startswith('original/'):
            require(alias.removeprefix('original/') in candidates,
                    'Selected original alias has no candidate identity')
    require(len(private['complete_witnesses']) == 123,
            'Separated event inventory changed')
    for event_id, witness in private['complete_witnesses'].items():
        require(event_id in original['reference']['events'],
                'Separated event is outside original reference')
        event_path = original['reference']['events'][event_id]['path']
        require(witness['status'] in {'full', 'partial', 'missed'},
                'Unknown separated event status')
        groups = witness['full_groups']
        require((witness['status'] == 'full') == bool(groups),
                'Unsupported separated FULL credit')
        for group in groups:
            require(group and set(group) <= selected,
                    'Separated witness was not delivered')
            for alias in group:
                if alias.startswith('original/'):
                    candidate_id = alias.removeprefix('original/')
                    require(candidate_id in candidates and
                            pages[candidates[candidate_id]['page']]['path'] == event_path,
                            'Original witness joined to another source')
                else:
                    match = re.search(r'page-\d{3}', alias)
                    require(match is not None and match.group() in pages and
                            pages[match.group()]['path'] == event_path,
                            'Separated witness joined to another source')
    for part, row in private['cohorts'].items():
        published_row = public['cohorts'][part]
        require(row['events'] == published_row['events'] == DENOMINATORS[part]
                and row['full'] == published_row['full'] ==
                sum(event_id.startswith(part + '/') and witness['status'] == 'full'
                    for event_id, witness in private['complete_witnesses'].items())
                and row['delivered'] == published_row['delivered']
                and row['accepted'] == published_row['accepted']
                and row['states'].get('uncertain', 0) == published_row['uncertain']
                and row['states'].get('rejected', 0) == published_row['rejected']
                and sum(row['states'].values()) == row['delivered'],
                'Separated public and sealed totals differ')
    return selected


def assemble(original, grades, pages, candidates, baseline, separated, selected):
    accepted = {identifier for identifier, grade in grades['findings'].items()
                if grade['status'] == 'accepted'}
    safe = {identifier for identifier in accepted
            if grades['findings'][identifier].get('suggestion_safety', {}).get('status')
            in {'safe', 'not_applicable'}}
    require(set(baseline) == set(original['reference']['events']) ==
            set(separated['complete_witnesses']), 'Event inventory join differs')
    by_path = {value['path']: value for value in pages.values()}
    rows = []
    used_candidates = set()
    counts = defaultdict(Counter)
    deltas = defaultdict(lambda: {'gained': [], 'lost': []})
    for event_id, event in original['reference']['events'].items():
        require(baseline[event_id]['original_event'] == event,
                'Baseline and expanded event differ')
        part = cohort(event['path'])
        all_groups = grades['events'][event_id]['full_support_groups']
        accepted_groups = [group for group in all_groups if set(group) <= accepted]
        safe_groups = [group for group in all_groups if set(group) <= safe]
        used_candidates.update(item for group in all_groups for item in group)
        base_status = baseline[event_id]['after']['status']
        separate = separated['complete_witnesses'][event_id]
        separate_status = separate['status']
        original_selected = [group for group in accepted_groups if
                             {'original/' + item for item in group} <= selected]
        states = {'baseline': base_status, 'candidate_oracle': bool(all_groups),
                  'accepted_oracle': bool(accepted_groups),
                  'assessed_safe_oracle': bool(safe_groups),
                  'separated_delivered': separate_status,
                  'original_candidate_support_selected': bool(original_selected)}
        counts[part]['baseline_full'] += base_status == 'full'
        counts[part]['candidate_full_ceiling'] += bool(all_groups)
        counts[part]['accepted_full_ceiling'] += bool(accepted_groups)
        counts[part]['safe_full_ceiling'] += bool(safe_groups)
        counts[part]['separated_full'] += separate_status == 'full'
        if base_status != 'full' and separate_status == 'full':
            deltas[part]['gained'].append(event_id)
        if base_status == 'full' and separate_status != 'full':
            deltas[part]['lost'].append(event_id)
        if not all_groups:
            earliest = 'no_complete_candidate_in_saved_pool'
        elif not accepted_groups:
            earliest = 'no_accepted_complete_candidate'
        elif not safe_groups:
            earliest = 'complete_candidate_advice_unassessed_or_unsafe'
        elif separate_status != 'full':
            earliest = 'selected_stream_lacks_complete_delivered_witness'
        else:
            earliest = 'complete_witness_in_selected_stream'
        rows.append({'event_id': event_id, 'cohort': part,
                     'source_group': by_path[event['path']]['source_group'],
                     'repository': by_path[event['path']]['repository'],
                     'source_sha256': by_path[event['path']]['source_sha256'],
                     'category': event['category'],
                     'target_spans': [{'start': item['start'], 'end': item['end']}
                                      for item in event['targets']],
                     'states': states, 'earliest_supported_stage': earliest,
                     'baseline_witness_ids': baseline[event_id]['after'].get('finding_ids', []),
                     'accepted_candidate_groups': accepted_groups,
                     'assessed_safe_candidate_groups': safe_groups,
                     'selected_original_groups': original_selected,
                     'separated_complete_groups': separate['full_groups']})
    require({part: len([row for row in rows if row['cohort'] == part])
             for part in DENOMINATORS} == DENOMINATORS,
            'Changed event denominators')
    witness_candidates = {}
    for identifier in sorted(used_candidates):
        candidate = candidates[identifier]
        page = pages[candidate['page']]
        grade = grades['findings'][identifier]
        witness_candidates[identifier] = {
            'source_sha256': page['source_sha256'],
            'source_group': page['source_group'],
            'repository': page['repository'],
            'byte_targets': candidate['targets'],
            'category': candidate['category'], 'origin': candidate['origin'],
            'claim_sha256': candidate['claim_sha256'],
            'frozen_grade': grade['status'],
            'assessed_advice_safety': grade.get('suggestion_safety', {}).get('status', 'unassessed')}
    return rows, witness_candidates, counts, deltas


def assert_research_only(output):
    forbidden = {'selector', 'predictions', 'inference_features',
                 'model_weights', 'threshold', 'runtime_policy'}
    require(output['kind'] == 'retrospective_event_loss_ledger' and
            not output['inference_selector_trained_or_evaluated'] and
            output['retrospective_grades_are_not_inference_inputs'] and
            not forbidden.intersection(output),
            'Frozen grades or event labels leaked into an inference policy')


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--expanded-root', type=Path, required=True)
    parser.add_argument('--separated-evaluation', type=Path, required=True)
    parser.add_argument('--separated-public', type=Path, required=True)
    parser.add_argument('--separated-manifest', type=Path, required=True)
    parser.add_argument('--baseline-events', type=Path, required=True)
    parser.add_argument('--source-manifest', type=Path, required=True)
    parser.add_argument('--source-root', type=Path, required=True)
    parser.add_argument('--output', type=Path, required=True)
    args = parser.parse_args()
    files = {'expanded_originals': args.expanded_root / 'expanded-originals.json',
             'expanded_grades': args.expanded_root / 'expanded-grades.json',
             'separated_evaluation': args.separated_evaluation,
             'separated_public': args.separated_public,
             'separated_manifest': args.separated_manifest,
             'baseline_events': args.baseline_events,
             'source_manifest': args.source_manifest}
    original, grades, private, public, public_manifest, baseline, sources = (
        read(files[key]) for key in files)
    require(digest(args.separated_evaluation) == public['private_evaluation_sha256'],
            'Sealed evaluation differs from published identity')
    pages, candidates = check_expanded(original, grades, sources,
                                        args.source_root)
    selected = check_separated(private, public, public_manifest, original, pages,
                               candidates)
    rows, witnesses, counts, deltas = assemble(original, grades, pages,
                                                candidates, baseline, private,
                                                selected)
    output = {'issue': 421, 'kind': 'retrospective_event_loss_ledger',
              'scope': {'pages': 36, 'units': 4549, 'judgments': 804,
                        'events': 123, 'source_groups': 13,
                        'distinct_repository_labels': len({page['repository']
                                                           for page in pages.values()}),
                        'expanded_candidates': 1154,
                        'separated_deliveries': len(selected)},
              'input_sha256': {key: digest(path) for key, path in files.items()},
              'cohorts': {part: dict(counts[part]) for part in DENOMINATORS},
              'baseline_to_separated': {part: deltas[part] for part in DENOMINATORS},
              'event_rows': rows, 'candidate_witnesses': witnesses,
              'inference_selector_trained_or_evaluated': False,
              'retrospective_grades_are_not_inference_inputs': True,
              'prospective_pages_used': False,
              'product_qualified': False}
    assert_research_only(output)
    args.output.write_text(json.dumps(output, indent=2, sort_keys=True) + '\n')
    print(json.dumps({'cohorts': output['cohorts'],
                      'baseline_to_separated_counts': {part: {
                          name: len(ids) for name, ids in deltas[part].items()}
                          for part in DENOMINATORS},
                      'witness_candidates': len(witnesses)}, sort_keys=True))


if __name__ == '__main__':
    main()
