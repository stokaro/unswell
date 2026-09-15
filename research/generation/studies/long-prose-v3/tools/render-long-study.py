"""Render study tables from saved measurements, without model calls."""
import argparse
import json
import pathlib


def read(path):
    return json.loads(path.read_text())


def pct(value):
    return 'unavailable' if value is None else f'{100 * value:.1f}%'


def number(value):
    return 'unavailable' if value is None else f'{value:.3f}'


def table(lines, headers, rows):
    lines.extend(['', '| ' + ' | '.join(headers) + ' |', '| ' + ' | '.join(['---'] * len(headers)) + ' |'])
    lines.extend('| ' + ' | '.join(map(str, row)) + ' |' for row in rows)
    lines.append('')


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--study', type=pathlib.Path, required=True)
    parser.add_argument('--output', type=pathlib.Path, required=True)
    args = parser.parse_args()
    study = args.study.resolve()
    descriptive = read(study / 'analysis/descriptive.json')
    confirmation = read(study / 'analysis/confirmation.json')
    lengths = read(study / 'analysis/length-sensitivity.json')
    summaries = [s for s in descriptive['strata'] if s['role'] == s['length_band'] == 'all']
    historical = next(s for s in summaries if s['family'] == 'historical')
    lines = ['# Complete-document results, September 15, 2026', '',
             'This record executes protocol version 3 on 38 dated technical documents from',
             '20 repositories and three model families. It measures constructions and review',
             'load. It supplies no human quality labels, authorship probabilities, or default',
             'blocking-rule qualification.', '',
             'The earlier version 2 confirmation remains an underpowered result with 16 global',
             'components. Version 3 uses fresh repositories and a separate frozen comparison.', '',
             '## Coverage', '',
             'Every request has one saved outcome. Truncations and other incomplete responses',
             'remain in coverage and do not become zero-finding documents. Off-length outputs',
             'remain in the primary analysis. Reported tokens include visible reasoning usage',
             'where the harness supplies it; hidden provider settings remain unavailable.']
    coverage = []
    total_tokens = 0
    for family in ['luna', 'haiku', 'qwen']:
        records = read(study / 'runs' / family / 'records.json')
        raw = [read(p) for p in (study / 'runs' / family / 'raw').glob('*/result.json')]
        c = records['coverage']
        total_tokens += sum(r.get('tokens', 0) for r in raw)
        coverage.append([family, c['requests'], c['complete'], c['truncated'], c['refused'], c['error'], c['missing'], c['off_length'], f"{sum(r.get('tokens', 0) for r in raw):,}"])
    table(lines, ['Family', 'Requests', 'Complete', 'Truncated', 'Refused', 'Error', 'Missing', 'Off length', 'Reported tokens'], coverage)
    completion = study / 'generation-completion.json'
    if completion.exists():
        amendment = read(completion)
        lines.extend(['The delivery record includes a declared resource amendment. After 148 Haiku',
                      'responses and 1,962,742 reported tokens across the three models, the runner’s',
                      '65,536-token per-call reservation prevented the final four requests. The',
                      'dispatch ceiling increased from 2,000,000 to 2,150,000 tokens for those fixed',
                      'requests. This happened after Qwen and Luna construction results were inspected;',
                      'the joint primary estimates had not been computed. The original freeze remains',
                      'unchanged in the input archive, and `generation-completion.json` records the',
                      'time, request IDs, wrapper hash, and prior exposure. Sample, prompts, models,',
                      'retry eligibility, deadline, hypotheses, and decision thresholds did not change.',
                      f'The completed run reports {total_tokens:,} tokens in total. This is an explicit',
                      'execution amendment, not an unamended execution of the original resource plan.', ''])
        assert total_tokens <= amendment['amended_token_cap']
    lines.extend(['## Frozen primary comparisons', '',
                  'Each row compares generate/neutral with its own originals. Differences and',
                  '95% percentile intervals use the Go joint component bootstrap (10,000 draws,',
                  'PCG seed 17). Tail probabilities receive Holm correction across all 12 tests.',
                  'A family result also needs an increase of at least three percentage points,',
                  '20 measured groups, and five groups with a response finding.'])
    rows = []
    for r in confirmation['tests']:
        d = r['paired_change']
        interval = ('unavailable' if any(d[k] is None for k in ['value', 'lower', 'upper']) else
                    f"{100*d['value']:+.1f} [{100*d['lower']:+.1f}, {100*d['upper']:+.1f}]")
        rows.append([r['family'], '`' + r['rule_id'] + '`', f"{r['original_with_finding']}/{r['pairs']}",
                     f"{r['response_with_finding']}/{r['pairs']}", interval, 'unavailable' if r['holm_adjusted_p'] is None else f"{r['holm_adjusted_p']:.4g}",
                     f"{r['components']} / {r['response_support_components']}", 'met' if r['association_criteria_met'] else 'not met'])
    table(lines, ['Family', 'Rule', 'Originals', 'Responses', 'Difference and interval (pp)', 'Holm p', 'Groups / support', 'Criteria'], rows)
    lines.extend(['A cross-family claim needs two qualifying families and a positive point estimate',
                  'in the third. A missing or weak result is not proof of equivalence. The frozen',
                  'advisory-load budget is at most one finding per 1,000 historical words and',
                  'findings on at most 25% of historical documents. This is review workload, not',
                  'a false-positive estimate. A load restriction limits an AI-associated advisory',
                  'claim; it does not disable an existing general readability rule.'])
    table(lines, ['Rule', 'Qualifying families', 'Cross-family criteria', 'Historical /1k', 'Historical prevalence', 'Disposition'],
          [[ '`' + r['rule_id'] + '`', r['supporting_families'], str(r['cross_family_criteria_met']).lower(),
             number(r['historical_per_1000_words']), pct(r['historical_prevalence']), r['disposition']] for r in confirmation['decisions']])
    lines.extend(['## Whole-profile load and length', '',
                  'All 40 rules use one explicit measurement policy. These totals are not the',
                  'technical or strict product profiles. Document prevalence grows with exposure;',
                  'per-word load, repository-macro prevalence, and realized-length slices are',
                  'reported alongside it. Historical provenance does not certify clean prose.'])
    table(lines, ['Family', 'Operation', 'Prompt', 'Docs', 'Words', 'Findings', '/1k words', 'Doc prevalence', 'Repo-macro prevalence'],
          [[r['family'], r['operation'], r['prompt'], r['documents'], f"{r['words']:,}", r['findings'], number(r['per_1000_words']),
            pct(r['prevalence']), pct(r['repository_macro_prevalence'])] for r in summaries])
    table(lines, ['Family', 'Operation', 'Prompt', 'Median extracted-length ratio', 'Same band', 'Within 30%'],
          [[r['family'], r['operation'], r['prompt'], number(r['median_extracted_word_ratio']),
            f"{r['same_length_band']}/{r['pairs']}", f"{r['within_30_percent_extracted_words']}/{r['pairs']}"] for r in lengths['arms']])
    lines.extend(['The generation ledger measures whitespace fields; this sensitivity uses',
                  'engine-extracted prose words on both sides. Same-band and within-30% comparisons',
                  'are descriptive restrictions after observing length. They do not replace the',
                  'primary sample or provide a second independent confirmation.', '',
                  '## Context opportunities', '',
                  'Whitespace-adjacent paragraph runs describe available context. Possible sentence',
                  'pairs within eight positions are structural opportunities, not the matcher’s',
                  'rule-specific event clusters. Headings, code, and source gaps affect continuity.',
                  'Cross-block findings directly count diagnostics with evidence in multiple blocks.'])
    context_rows = []
    for r in summaries:
        o = r['opportunities']
        context_rows.append([r['family'], r['operation'], r['prompt'], o['paragraphs'], o['headings'],
                             o['adjacent_heading_paragraph_pairs'], o['run_sentences'], o['runs_with_at_least_eight_sentences'],
                             o['potential_sentence_pairs_within_eight'], sum(q['cross_block_findings'] for q in r['rules'])])
    table(lines, ['Family', 'Operation', 'Prompt', 'Paragraphs', 'Headings', 'Heading/paragraph pairs', 'Run sentences', 'Runs >=8', 'Possible pairs', 'Cross-block findings'], context_rows)
    opportunity_rows = []
    for r in summaries:
        if r['family'] != 'historical' and (r['operation'] != 'generate' or r['prompt'] != 'neutral'):
            continue
        for q in r['rules']:
            if q['rule_id'] not in ['syntax.paired-contrast-density', 'syntax.passive-candidate-density', 'syntax.noun-stack', 'readability.grade-metric']:
                continue
            a = q['applicability']
            eligible = a['applicable_paragraphs']
            opportunity_rows.append([r['family'], '`' + q['rule_id'] + '`', eligible, a['positive_paragraphs'],
                                     pct(a['positive_paragraphs'] / eligible if eligible else None),
                                     a['sentence_opportunities_in_applicable_paragraphs'], q['cross_block_findings']])
    table(lines, ['Family', 'Rule', 'Applicable paragraphs', 'Positive activations', 'Activation share', 'Sentences in applicable paragraphs', 'Cross-block diagnostics'], opportunity_rows)
    lines.extend(['`descriptive.json` also retains engine-owned applicability, reasons for missing',
                  'block activations, and sentence/paragraph bindings for every rule, role, arm,',
                  'and realized-length band. A positive block activation means contributing evidence;',
                  'it is not an additional diagnostic. Sentence opportunities inside applicable',
                  'paragraphs are not independently established per-sentence applicability.', '',
                  'The pre-repair comparison on exposed Ptah data is recorded separately in',
                  '`docs/research/contextual-prose.md`. Its counts are development evidence and are',
                  'not pooled with this new sample.', '',
                  '## Descriptive ablations', '',
                  'These tables remove rule families from saved counts; they do not change the',
                  'engine or repeat model generation. The complete rule-to-family mapping is in',
                  '`descriptive.json` and was assigned for descriptive reporting after the freeze.'])
    table(lines, ['Family', 'All', 'Without phrase', 'Without structural', 'Without contextual'],
          [[r['family']] + [a['findings'] for a in r['ablations']] for r in summaries if r['family']=='historical' or r['operation']=='generate' and r['prompt']=='neutral'])
    lines.extend(['## Every rule', '',
                  'Counts below are descriptive generate/neutral counts, with historical load',
                  'alongside them. A zero is a result within this finite sample, not universal',
                  'absence. General editing rules need not separate model cohorts to remain useful.',
                  'Unlisted hypotheses cannot become confirmatory discoveries after inspection.',
                  '`analysis/evidence-cards.json` retains each rule definition, capability requirements,',
                  'classification, paired estimates for every arm, cases, and disposition.'])
    neutral = {r['family']: {q['rule_id']: q for q in r['rules']} for r in summaries if r['operation']=='generate' and r['prompt']=='neutral'}
    decisions = {r['rule_id']: r['disposition'] for r in confirmation['decisions']}
    rule_rows = []
    cards = []
    paired = {f: read(study / 'analysis' / ('paired-' + f + '.json')) for f in ['luna', 'haiku', 'qwen']}
    measured = read(next(iter(sorted((study / 'analysis/findings').glob('*.json')))))
    catalog = {r['id']: r for r in measured['policy']['rules']}
    classes = {r['rule_id']: r for r in read(study / 'rule-classes.json')['rules']}
    cases = read(study / 'analysis/cases.json')['cases']
    for r in historical['rules']:
        rule = r['rule_id']
        counts = [neutral[f][rule]['findings'] for f in ['luna','haiku','qwen']]
        if rule in decisions:
            disposition = decisions[rule]
        elif r['rule_class'] == 'explicit_prohibition':
            disposition = 'configured policy; no phrase list tested'
        elif not any(q['findings'] for arm in summaries for q in arm['rules'] if q['rule_id'] == rule):
            disposition = 'no observed support in this study; retain narrow scope'
        elif r['rule_class'] == 'general_style':
            disposition = 'retain as general editing measurement'
        else:
            disposition = 'observed candidate; no primary confirmation'
        rule_rows.append(['`'+rule+'`', r['findings'], *counts, disposition])
        cards.append(dict(rule_id=rule, descriptor=catalog[rule], classification=classes[rule],
                          primary=rule in decisions, disposition=disposition,
                          historical=r, comparisons={f: next(q for q in paired[f]['rules'] if q['rule_id'] == rule) for f in paired},
                          cases=[dict(category=c['category'], source_id=c['source_id'], source_sha256=c['source_sha256'],
                                      span=c['diagnostic']['primary']['span']) for c in cases if c['rule_id'] == rule],
                          costs=dict(per_rule_time=None, reason='No isolated per-rule timing in this study; full replay resources are recorded separately.'),
                          quality_qualification='not_performed', default_change=False))
    (study / 'analysis/evidence-cards.json').write_text(json.dumps(dict(version='unswell-long-prose-evidence-cards-v1',
        protocol='unswell-llm-patterns-v3', engine_commit=read(study / 'freeze.json')['engine_commit'],
        policy=measured['policy'], cards=cards), ensure_ascii=False, indent=2, sort_keys=True) + '\n')
    table(lines, ['Rule', 'Historical', 'Luna', 'Haiku', 'Qwen', 'Disposition'], rule_rows)
    lines.extend(['## Inspectable cases and limits', '',
                  'The source-bound Qwen preface examples and their agent-written edits are in',
                  '`e2e/testdata/long_prose_research`: the original prefaces are detected and the',
                  'edited wording retains benchmark numbers, unaffected behavior, file names,',
                  'volunteer delivery, and the absence of release-date promises. The Moment',
                  'counterexample retains two necessary behavioral contrasts. A repeated form',
                  'does not make either distinction expendable.', '',
                  'Ptah’s three contiguous contrast examples and contract-preserving alternatives',
                  'remain in `e2e/testdata/contextual_ptah`. These fixtures run offline through the',
                  'built CLI, JSON, and SARIF. They establish construction behavior and coordinates,',
                  'not human precision or reader preference. The Haiku nominalization fixture',
                  'replaces only `conducts an investigation of` with `investigates`, retaining',
                  'acknowledgment, rejection reasons, policy conditions, and fix-work notification.', '',
                  'Historical passive candidates include ordinary installation and distribution',
                  'descriptions. Repeated installation instructions and shim-command obligations',
                  'are counterexamples to treating parallel openings as automatically defective.',
                  'Grade estimates and noun sequences can reflect required terminology and layout.',
                  'Use the case record with full source spans when assessing a particular finding.', '',
                  'This study covers three fixed low-effort model/harness choices and a small',
                  'English technical source frame. It does not establish behavior for every model,',
                  'genre, language background, or future prompt. Source curation was by an agent.',
                  'Human quality qualification remains deferred to #22/#26, and none of these',
                  'findings is an authorship probability or a reason to block CI by provenance.', '',
                  'See the data card, frozen protocol, archive checksums, and reproduction commands',
                  'beside this report. The earlier study’s placebo and earlier-date sensitivities',
                  'remain under their original policies; their counts are not silently pooled here.'])
    args.output.write_text('\n'.join(lines) + '\n')


if __name__ == '__main__':
    main()
