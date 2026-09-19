#!/usr/bin/env python3
"""Retain prior full-event rule credit separately from learned retrieval."""
import hashlib
import json
from pathlib import Path

STUDY = Path(__file__).resolve().parents[1]
REFERENCE = STUDY.parent/'2026-09-18-structural-wordiness'
SPLITS = {
    'development':'2026-09-17-full-page-recall',
    'exposed_local':'2026-09-17-local-repetition',
    'exposed_context':'2026-09-17-context-recall',
    'exposed_instruction':'2026-09-17-instruction-recall',
    'exposed_construction':'2026-09-17-construction-recall',
    'exposed_purpose':'2026-09-17-purpose-recall',
    'exposed_verb':'2026-09-17-verb-scaffolding',
    'exposed_relations':'2026-09-17-rhetoric-relations',
    'exposed_projection':'2026-09-17-instruction-projection',
    'exposed_scope':'2026-09-17-rhetoric-scope',
    'exposed_proposition':'2026-09-17-proposition-repetition',
    'exposed_clauses':'2026-09-17-instruction-clauses',
    'exposed_action':'2026-09-17-action-grammar',
    'exposed_roles':'2026-09-17-discourse-roles',
    'exposed_boundaries':'2026-09-17-clause-boundaries',
    'exposed_grammar':'2026-09-18-clause-grammar',
    'exposed_discourse':'2026-09-18-discourse-patterns',
    'confirmation':'2026-09-18-structural-wordiness',
}


def reference(audit, summary):
    prior = json.loads((REFERENCE/'summary.json').read_text())
    events = [e for e in audit['events'] if e['kind']=='defects']
    known,missed,rows = set(),set(),[]
    for split,prefix in SPLITS.items():
        reviewed = prior[split]['technical']['phases']['after']
        members = {(e['page'],e['id']) for e in events if e['page'].startswith(prefix+'/')}
        missing = {key for key in members if key[1] in reviewed['missed']}
        if len(members) != reviewed['defects'] or len(missing) != len(reviewed['missed']):
            raise ValueError('Rule review denominator does not bind to the audit: '+split)
        if len(members)-len(missing) != reviewed['detected']:
            raise ValueError('Rule full-event credit changed: '+split)
        known.update(members)
        missed.update(missing)
        rows.append(dict(split=split,defects=len(members),diagnosed=len(members)-len(missing)))
    if len(known) != len(events):
        raise ValueError('Rule comparison omitted or duplicated a source event')
    overlaps = {}
    for kind,model in summary['models'].items():
        retrieved = {(e['page'],e['id']) for e in model['retrieved_events']}
        overlaps[kind] = dict(retrieved=len(retrieved),previously_diagnosed=len(retrieved & (known-missed)),
                              previously_missed=len(retrieved & missed),new_diagnostic_credit=0)
    return dict(basis='retained source-bound diagnostic review; learned output earns retrieval credit only',
                engine=json.loads((REFERENCE/'engines.json').read_text())['after'],profile='technical',
                source_files={str(p.relative_to(STUDY.parents[2])):hashlib.sha256(p.read_bytes()).hexdigest()
                              for p in [REFERENCE/'summary.json',REFERENCE/'dispositions.json',REFERENCE/'engines.json']},
                defects=len(known),diagnosed=len(known-missed),splits=rows,learned_retrieval_overlap=overlaps)


if __name__ == '__main__':
    import gzip
    audit = json.loads(gzip.decompress((STUDY.parent/'2026-09-19-representation-audit/measurements.json.gz').read_bytes()))
    result = reference(audit,json.loads((STUDY/'summary.json').read_text()))
    path = STUDY/'rule-reference.json'
    if path.exists():
        raise ValueError('Reference already exists')
    path.write_text(json.dumps(result,indent=2)+'\n')
    print(json.dumps({key:result[key] for key in ('defects','diagnosed','learned_retrieval_overlap')},indent=2))
