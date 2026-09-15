"""Apply the frozen twelve-test decision rule to Go paired estimates."""
import argparse
import json
import pathlib

PRIMARY = ['syntax.paired-contrast-density', 'syntax.passive-candidate-density',
           'syntax.noun-stack', 'readability.grade-metric']
FAMILIES = ['luna', 'haiku', 'qwen']


def read(path):
    return json.loads(path.read_text())


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--study', type=pathlib.Path, required=True)
    args = parser.parse_args()
    study = args.study.resolve()
    descriptive = read(study / 'analysis/descriptive.json')
    historical = next(s for s in descriptive['strata'] if s['family'] == 'historical' and s['role'] == s['length_band'] == 'all')
    loads = {r['rule_id']: r for r in historical['rules']}
    tests = []
    for family in FAMILIES:
        paired = read(study / 'analysis' / ('paired-' + family + '.json'))
        assert paired['version'] == 'unswell-paired-tables-v3'
        for rule in PRIMARY:
            row = next(r for r in paired['rules'] if r['rule_id'] == rule)
            arm = next(a for a in row['arms'] if a['operation'] == 'generate' and a['prompt'] == 'neutral')
            tests.append(dict(family=family, rule_id=rule, **arm))
    # Missing probabilities remain part of the twelve-test multiplicity family.
    largest = 0.0
    for rank, row in enumerate(sorted(tests, key=lambda r: r['paired_p_value'] if r['paired_p_value'] is not None else 1.0)):
        p = row['paired_p_value']
        largest = max(largest, min(1.0, (len(tests) - rank) * (p if p is not None else 1.0)))
        row['holm_adjusted_p'] = largest if p is not None else None
    for row in tests:
        effect = row['paired_change']
        failures = []
        if row['components'] < 20:
            failures.append('fewer_than_20_measured_components')
        if row['response_support_components'] < 5:
            failures.append('fewer_than_5_response_support_components')
        if effect['value'] is None or effect['lower'] is None or effect['upper'] is None:
            failures.append('estimate_or_interval_unavailable')
        if effect['value'] is None or effect['value'] < 0.03:
            failures.append('increase_below_three_percentage_points')
        if row['holm_adjusted_p'] is None or row['holm_adjusted_p'] > 0.05:
            failures.append('holm_threshold_not_met')
        row['association_criteria_met'] = not failures
        row['unmet_criteria'] = failures
    decisions = []
    for rule in PRIMARY:
        rows = [r for r in tests if r['rule_id'] == rule]
        positive = all(r['paired_change']['value'] is not None and r['paired_change']['value'] > 0 for r in rows)
        admitted = sum(r['association_criteria_met'] for r in rows)
        load = loads[rule]
        within_load = load['per_1000_words'] <= 1 and load['prevalence'] <= 0.25
        decisions.append(dict(rule_id=rule, supporting_families=admitted, all_family_estimates_positive=positive,
                              cross_family_criteria_met=admitted >= 2 and positive,
                              historical_per_1000_words=load['per_1000_words'], historical_prevalence=load['prevalence'],
                              within_advisory_load_budget=within_load,
                              disposition=('restricted_by_historical_review_load' if not within_load else
                                           'association_supported_advisory_only' if admitted >= 2 and positive else
                                           'association_not_established'),
                              human_quality_qualification='not_performed', default_promotion=False))
    result = dict(version='unswell-long-prose-confirmation-v1', protocol='unswell-llm-patterns-v3',
                  multiplicity={'method': 'Holm', 'tests': 12, 'alpha': 0.05}, tests=tests, decisions=decisions)
    (study / 'analysis/confirmation.json').write_text(json.dumps(result, indent=2, sort_keys=True) + '\n')
    for row in tests:
        print(row['family'], row['rule_id'], row['pairs'], row['components'], row['paired_change']['value'], row['holm_adjusted_p'], row['association_criteria_met'])
    for decision in decisions:
        print(decision)


if __name__ == '__main__':
    main()
