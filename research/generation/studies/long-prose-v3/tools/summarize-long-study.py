"""Aggregate engine outputs; statistical resampling remains in corpus paired."""
import argparse
import collections
import hashlib
import json
import pathlib
import statistics

PRIMARY = ['syntax.paired-contrast-density', 'syntax.passive-candidate-density',
           'syntax.noun-stack', 'readability.grade-metric']
BANDS = ['under_400', '400_799', '800_1599', '1600_3199', '3200_plus']


def read(path):
    return json.loads(path.read_text())


def band(words):
    return BANDS[next((i for i, end in enumerate([400, 800, 1600, 3200]) if words < end), 4)]


def fraction(n, d):
    return n / d if d else None


def scope(rule):
    if rule.startswith('format.'):
        return 'structural'
    if rule.startswith(('repetition.', 'density.')) or rule in ['filler.section-announcement', 'filler.summary-echo'] or rule.endswith('-density'):
        return 'contextual'
    if rule.startswith(('syntax.', 'readability.', 'structure.', 'format.')):
        return 'structural'
    return 'phrase'


def structural_opportunities(units, data):
    """Count observable adjacency, not the matcher's rule-specific event windows."""
    runs, run, previous = [], 0, None
    for unit in units:
        vals = {v['id']: v.get('number') for v in unit['values']}
        sentences = vals.get('prose-sentences')
        eligible = unit['kind'] == 'paragraph' and not unit.get('excluded', False) and sentences is not None
        adjacent = (previous is not None and previous['span']['end'] <= unit['span']['start'] and
                    not data[previous['span']['end']:unit['span']['start']].strip())
        if not eligible or not adjacent:
            if run:
                runs.append(run)
            run = 0
        if eligible:
            run += int(sentences)
            previous = unit
        else:
            previous = None
    if run:
        runs.append(run)
    # All sentence pairs separated by fewer than eight positions within a run.
    heading_pairs = sum(a['kind'] == 'heading' and b['kind'] == 'paragraph' and
                        a['span']['end'] <= b['span']['start'] and
                        not data[a['span']['end']:b['span']['start']].strip(b' \t\r\n#=-')
                        for a, b in zip(units, units[1:]))
    return dict(paragraphs=sum(u['kind'] == 'paragraph' for u in units),
                headings=sum(u['kind'] == 'heading' for u in units),
                adjacent_heading_paragraph_pairs=heading_pairs,
                whitespace_adjacent_paragraph_runs=len(runs), run_sentences=sum(runs),
                runs_with_at_least_two_sentences=sum(n >= 2 for n in runs),
                runs_with_at_least_eight_sentences=sum(n >= 8 for n in runs),
                potential_sentence_pairs_within_eight=sum(sum(min(7, n-i-1) for i in range(n)) for n in runs))


def feature_summary(feature, diagnostics, data, rules):
    activations = {r: dict(applicable_blocks=0, positive_blocks=0, applicable_paragraphs=0,
                          positive_paragraphs=0, sentence_opportunities_in_applicable_paragraphs=0,
                          absent_reasons=collections.Counter()) for r in rules}
    for unit in feature['units']:
        sentences = next((v.get('number') for v in unit['values'] if v['id'] == 'prose-sentences'), None)
        for value in unit['values']:
            if not value['id'].startswith('activation/'):
                continue
            cell = activations[value['id'].removeprefix('activation/')]
            if value.get('number') is None:
                cell['absent_reasons'][value.get('reason', 'unspecified')] += 1
            else:
                cell['applicable_blocks'] += 1
                cell['positive_blocks'] += value['number'] > 0
                if unit['kind'] == 'paragraph':
                    cell['applicable_paragraphs'] += 1
                    cell['positive_paragraphs'] += value['number'] > 0
                    cell['sentence_opportunities_in_applicable_paragraphs'] += int(sentences or 0)
    cross = collections.Counter()
    for finding in diagnostics:
        blocks = set(o['block_id'] for o in finding['evidence']['occurrences'])
        if len(blocks) > 1:
            cross[finding['rule_id']] += 1
    return dict(source_hash=feature['source_hash'], opportunities=structural_opportunities(feature['units'], data),
                activations=activations, cross_block_findings=dict(cross),
                diagnostic_counts=collections.Counter(f['rule_id'] for f in diagnostics))


def assemble(study):
    analysis = study / 'analysis'
    classes = read(study / 'rule-classes.json')['rules']
    rules = [r['rule_id'] for r in classes]
    sources = {}
    for path in (analysis / 'acquisition').glob('*/shards/*.json'):
        for source in read(path)['sources']:
            sources[source['id']] = source
    tasks = {t['id']: t for t in read(study / 'tasks.json')['tasks']}
    response_meta = {}
    for family in ['luna', 'haiku', 'qwen']:
        path = study / 'runs' / family / 'records.json'
        if path.exists():
            generation = read(path)
            for record in generation['records']:
                if record['status'] != 'complete':
                    continue
                response_path = 'generated/' + generation['run'] + '/' + record['response_id'] + '.md'
                response_meta[response_path] = dict(family=family, operation=record['operation'], prompt=record['prompt'],
                                                    task_id=record['task_id'], off_length=record['off_length'],
                                                    requested_words=record['requested_words'])
    global_plan = analysis / 'dataset-plan.json'
    groups = {s['id']: s['group'] for s in read(global_plan)['sources']} if global_plan.exists() else {}
    findings, units = {}, collections.defaultdict(list)
    policy = None
    catalog = None
    for path in (analysis / 'findings').glob('*.json'):
        artifact = read(path)
        identity = (artifact['policy']['config_hash'], artifact['policy']['ruleset_hash'])
        assert policy is None or identity == policy
        policy = identity
        catalog = artifact['policy']['rules']
        for doc in artifact['documents']:
            assert doc['status'] == 'measured', (path, doc['source_id'], doc['status'])
            assert doc['source_id'] not in findings
            findings[doc['source_id']] = doc
        for unit in artifact['units']:
            units[unit['source_id']].append(unit)
    features = {}
    feature_files = []
    for cohort in ['historical', 'controlled']:
        split = sorted((analysis / 'features').glob(cohort + '-*.json'))
        original = analysis / ('features-' + cohort + '.json')
        feature_files.extend(split if split else [original] if original.exists() else [])
    for path in feature_files:
        result = read(path)
        # CLI identity includes the config bundle; corpus uses the raw config.
        # Require the frozen policy bytes and catalog, then every diagnostic count.
        assert result['manifest']['ruleset_hash'] == policy[1]
        assert result['manifest']['rules'] == catalog
        assert read(study / 'freeze.json')['files']['policy.yaml'] == hashlib.sha256((study / 'policy.yaml').read_bytes()).hexdigest()
        assert result['manifest']['config_sources'][0]['path'] == '../../policy.yaml'
        diagnostics = collections.defaultdict(list)
        for finding in result['findings']:
            diagnostics[finding['primary']['path']].append(finding)
        for source in result['features']['sources']:
            source_path = source['path']
            assert source_path not in features
            data = (analysis / 'work' / source_path).read_bytes()
            assert hashlib.sha256(data).hexdigest() == source['source_hash']
            features[source_path] = feature_summary(source, diagnostics[source_path], data, rules)
        del result
    rows = []
    for source_id, doc in sorted(findings.items()):
        source = sources[source_id]
        path = doc['cohort'] + '/' + source['repository'].replace('/', '__') + '/' + source['path']
        meta = response_meta[source['path']] if doc['cohort'] == 'controlled' else dict(family='historical', operation='original', prompt='original')
        row = dict(source_id=source_id, source_hash=source['sha256'], path=path, repository=source['repository'],
                   group=groups.get(source_id, source['repository']), role=source['role'], **meta,
                   words=doc['prose_words'], length_band=band(doc['prose_words']), findings=doc['findings'],
                   by_rule=doc['by_rule'], abstentions=doc.get('abstained', {}), units={})
        for kind in ['paragraph', 'sentence', 'fragment']:
            selected = [u for u in units[source_id] if u['kind'] == kind]
            counts = collections.Counter(f['rule_id'] for u in selected for f in u['findings'])
            with_finding = collections.Counter(r for u in selected for r in set(f['rule_id'] for f in u['findings']))
            row['units'][kind] = dict(count=len(selected), words=sum(u['words'] for u in selected),
                                      findings=dict(counts), with_finding=dict(with_finding))
        if path in features:
            feature = features[path]
            assert feature['source_hash'] == source['sha256']
            assert feature['diagnostic_counts'] == collections.Counter(doc['by_rule'])
            row.update({k: feature[k] for k in ['opportunities', 'activations', 'cross_block_findings']})
        rows.append(row)
    return rows, classes, policy


def group_summary(rows, classes):
    n = len(rows)
    words = sum(r['words'] for r in rows)
    repositories = sorted(set(r['repository'] for r in rows))
    summary = dict(documents=n, repositories=len(repositories), words=words,
                   with_finding=sum(r['findings'] > 0 for r in rows), findings=sum(r['findings'] for r in rows))
    summary['prevalence'] = fraction(summary['with_finding'], n)
    summary['per_1000_words'] = fraction(1000 * summary['findings'], words)
    summary['repository_macro_prevalence'] = statistics.mean(
        fraction(sum(r['findings'] > 0 for r in rows if r['repository'] == repo), sum(r['repository'] == repo for r in rows)) for repo in repositories)
    summary['opportunities'] = dict(sum((collections.Counter(r.get('opportunities', {})) for r in rows), collections.Counter()))
    summary['rules'] = []
    for cls in classes:
        rule = cls['rule_id']
        admitted = [r for r in rows if r['role'] in cls['roles'] and rule not in r['abstentions']]
        count = sum(r['by_rule'].get(rule, 0) for r in admitted)
        active = sum(r['by_rule'].get(rule, 0) > 0 for r in admitted)
        rule_words = sum(r['words'] for r in admitted)
        row = dict(rule_id=rule, rule_class=cls['class'], documents=len(admitted), findings=count,
                   documents_with_finding=active, prevalence=fraction(active, len(admitted)),
                   per_1000_words=fraction(1000 * count, rule_words),
                   cross_block_findings=sum(r.get('cross_block_findings', {}).get(rule, 0) for r in admitted))
        row['applicability'] = {}
        cells = [r['activations'][rule] for r in admitted if 'activations' in r]
        for key in ['applicable_blocks', 'positive_blocks', 'applicable_paragraphs', 'positive_paragraphs', 'sentence_opportunities_in_applicable_paragraphs']:
            row['applicability'][key] = sum(c[key] for c in cells)
        row['applicability']['absent_reasons'] = dict(sum((c['absent_reasons'] for c in cells), collections.Counter()))
        row['units'] = {}
        for kind in ['paragraph', 'sentence', 'fragment']:
            row['units'][kind] = dict(count=sum(r['units'][kind]['count'] for r in admitted),
                words=sum(r['units'][kind]['words'] for r in admitted),
                findings=sum(r['units'][kind]['findings'].get(rule, 0) for r in admitted),
                with_finding=sum(r['units'][kind]['with_finding'].get(rule, 0) for r in admitted))
        summary['rules'].append(row)
    summary['ablations'] = []
    for name in ['all', 'without_phrase', 'without_structural', 'without_contextual']:
        counts = [sum(v for k, v in r['by_rule'].items() if name == 'all' or scope(k) != name.removeprefix('without_')) for r in rows]
        summary['ablations'].append(dict(name=name, findings=sum(counts), with_finding=sum(c > 0 for c in counts),
                                          per_1000_words=fraction(1000 * sum(counts), words)))
    return summary


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--study', type=pathlib.Path, required=True)
    args = parser.parse_args()
    study = args.study.resolve()
    rows, classes, policy = assemble(study)
    groups = collections.defaultdict(list)
    for row in rows:
        base = (row['family'], row['operation'], row['prompt'])
        for role, length in [('all', 'all'), (row['role'], 'all'), ('all', row['length_band']), (row['role'], row['length_band'])]:
            groups[base + (role, length)].append(row)
    result = dict(version='unswell-long-prose-descriptive-v1', policy=dict(config_hash=policy[0], ruleset_hash=policy[1]),
                  semantics=dict(unit_findings='Primary diagnostic bindings in a complete-document scan; nested units are not independent trials.',
                                 activation='Engine-owned block applicability; positive means contributing evidence, not a separate diagnostic.',
                                 sentence_opportunity='Sentences inside applicable paragraphs, not independent per-sentence rule applicability.',
                                 structural_window='Whitespace-adjacent paragraph runs and possible sentence pairs; not evaluated rule-specific event clusters.',
                                 ablations='Post-freeze descriptive grouping by rule IDs; no tuning or changed rule execution.'),
                  ablation_groups={c['rule_id']: scope(c['rule_id']) for c in classes}, documents=rows, strata=[])
    for key, members in sorted(groups.items()):
        result['strata'].append(dict(zip(['family', 'operation', 'prompt', 'role', 'length_band'], key), **group_summary(members, classes)))
    (study / 'analysis/descriptive.json').write_text(json.dumps(result, ensure_ascii=False, indent=2, sort_keys=True) + '\n')
    print(len(rows), 'documents,', len(groups), 'strata')


if __name__ == '__main__':
    main()
