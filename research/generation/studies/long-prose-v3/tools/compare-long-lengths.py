"""Describe realized-length overlap without changing the primary paired sample."""
import argparse
import collections
import json
import pathlib
import statistics

parser = argparse.ArgumentParser()
parser.add_argument('--study', type=pathlib.Path, required=True)
args = parser.parse_args()
study = args.study.resolve()
descriptive = json.loads((study / 'analysis/descriptive.json').read_text())
rows = descriptive['documents']
tasks = {t['id']: t for t in json.loads((study / 'tasks.json').read_text())['tasks']}
originals = {r['source_id']: r for r in rows if r['family'] == 'historical'}
rules = list(descriptive['ablation_groups'])
arms = collections.defaultdict(list)
for row in rows:
    if row['family'] != 'historical':
        arms[(row['family'], row['operation'], row['prompt'])].append(row)
output = []
for arm, responses in sorted(arms.items()):
    pairs = [(originals[tasks[r['task_id']]['source_id']], r) for r in responses]
    item = dict(zip(['family', 'operation', 'prompt'], arm), pairs=len(pairs),
                median_extracted_word_ratio=statistics.median(r['words'] / o['words'] for o, r in pairs),
                requested_length_deviations=sum(r['off_length'] for r in responses),
                same_length_band=sum(o['length_band'] == r['length_band'] for o, r in pairs),
                within_30_percent_extracted_words=sum(0.7 <= r['words'] / o['words'] <= 1.3 for o, r in pairs), rules=[])
    for rule in rules:
        cells = {}
        for condition in ['all_completed_pairs', 'same_realized_band', 'within_30_percent_extracted_words']:
            selected = [(o, r) for o, r in pairs if rule not in o['abstentions'] and rule not in r['abstentions'] and
                        (condition == 'all_completed_pairs' or condition == 'same_realized_band' and o['length_band'] == r['length_band'] or
                         condition == 'within_30_percent_extracted_words' and 0.7 <= r['words'] / o['words'] <= 1.3)]
            n = len(selected)
            a = sum(o['by_rule'].get(rule, 0) > 0 for o, _ in selected)
            b = sum(r['by_rule'].get(rule, 0) > 0 for _, r in selected)
            cells[condition] = dict(pairs=n, groups=len(set(r['group'] for _, r in selected)), original_with_finding=a,
                                    response_with_finding=b, prevalence_difference=(b-a)/n if n else None)
        item['rules'].append(dict(rule_id=rule, comparisons=cells))
    output.append(item)
result = dict(version='unswell-long-prose-length-sensitivity-v1',
              interpretation='Descriptive post-output restrictions; no independent confirmation, no replacement of the primary all-completed-pair estimand.',
              length_basis='Engine-extracted prose words on both sides. The generation record uses whitespace fields and has a separate deviation flag.', arms=output)
(study / 'analysis/length-sensitivity.json').write_text(json.dumps(result, indent=2, sort_keys=True) + '\n')
print(len(output), 'arms')
