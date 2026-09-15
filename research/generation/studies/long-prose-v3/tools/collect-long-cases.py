"""Select source-bound examples in stable order; assign no quality labels."""
import argparse
import collections
import hashlib
import json
import pathlib
import re


def read(path):
    return json.loads(path.read_text())


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--study', type=pathlib.Path, required=True)
    args = parser.parse_args()
    study = args.study.resolve()
    analysis = study / 'analysis'
    sources = {}
    for path in (analysis / 'acquisition').glob('*/shards/*.json'):
        for source in read(path)['sources']:
            cohort = source['snapshot']['cohort']
            name = cohort + '/' + source['repository'].replace('/', '__') + '/' + source['path']
            sources[name] = source
    response_meta = {}
    for family in ['luna', 'haiku', 'qwen']:
        path = study / 'runs' / family / 'records.json'
        if not path.exists():
            continue
        generation = read(path)
        for record in generation['records']:
            name = 'generated/' + generation['run'] + '/' + record['response_id'] + '.md'
            response_meta[name] = dict(family=family, operation=record['operation'], prompt=record['prompt'],
                                       request_id=record['response_id'], task_id=record['task_id'])
    paths = []
    for cohort in ['historical', 'controlled']:
        split = sorted((analysis / 'features').glob(cohort + '-*.json'))
        original = analysis / ('features-' + cohort + '.json')
        paths.extend(split if split else [original] if original.exists() else [])
    chosen = collections.defaultdict(list)
    totals = collections.Counter()
    for path in paths:
        result = read(path)
        for finding in result['findings']:
            name = finding['primary']['path']
            source = sources[name]
            meta = response_meta[source['path']] if source['snapshot']['cohort'] == 'controlled' else dict(family='historical', operation='original', prompt='original')
            category = meta['family'] if meta['operation'] in ['original', 'generate'] and meta['prompt'] in ['original', 'neutral'] else 'other_controlled_arms'
            key = (finding['rule_id'], category)
            totals[key] += 1
            row = dict(source=name, source_id=source['id'], source_sha256=source['sha256'], reference=source['reference'],
                       role=source['role'], **meta, diagnostic=finding)
            chosen[key].append(row)
    cases = []
    for (rule, category), rows in sorted(chosen.items()):
        for row in sorted(rows, key=lambda r: (r['source'], r['diagnostic']['primary']['span']['start']))[:2]:
            source = (analysis / 'work' / row['source']).read_bytes()
            assert hashlib.sha256(source).hexdigest() == row['source_sha256']
            spans = [row['diagnostic']['primary']['span']] + [r['span'] for r in row['diagnostic']['related']]
            first, last = min(s['start'] for s in spans), max(s['end'] for s in spans)
            separators = list(re.finditer(rb'\r?\n[ \t]*\r?\n', source))
            start = max((m.end() for m in separators if m.end() <= first), default=0)
            end = min((m.start() for m in separators if m.start() >= last), default=len(source))
            assert start <= first <= last <= end <= len(source)
            # Preserve the entire affected paragraph range, including related evidence.
            excerpt = source[start:end]
            row.update(rule_id=rule, category=category, context_span=dict(start=start, end=end), context=excerpt.decode(),
                       context_sha256=hashlib.sha256(excerpt).hexdigest(), category_findings=totals[(rule, category)],
                       review_status='construction_example_without_quality_label')
            cases.append(row)
    result = dict(version='unswell-long-prose-cases-v1',
                  selection='First two findings by source path and primary byte offset in each rule/category; categories are historical, each generate/neutral family, and other controlled arms.',
                  interpretation='Automatically selected source-bound inspection material, not a human annotation round. Reviewed rewrite and technical-counterexample judgments are described separately.',
                  cases=cases)
    (analysis / 'cases.json').write_text(json.dumps(result, ensure_ascii=False, indent=2, sort_keys=True) + '\n')
    print(len(cases), 'source-bound cases', len(set(r['rule_id'] for r in cases)), 'rules')


if __name__ == '__main__':
    main()
