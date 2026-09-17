#!/usr/bin/env python3
"""Validate local-repetition evidence and derive complete-page and target metrics."""
import argparse
from collections import Counter
from pathlib import Path
import random
import sys

sys.dont_write_bytecode = True
ROOT = Path(__file__).resolve().parent.parent
SCOPE = ROOT.parent/'2026-09-17-repetition-scope'
FRAMING = ROOT.parent/'2026-09-17-framing-recall'
sys.path.insert(0, str(SCOPE/'tools'))
import compare as prior
audit = prior.audit
SEMANTICS = dict(prior.SEMANTICS, **{
    'repetition.adjacent-word': {'needless_repetition'},
    'repetition.repeated-claim': {'needless_repetition'},
    'repetition.explanatory-restart': {'needless_repetition'},
})
COMMITS = dict(before='b3ffaae00dbc62dfea1aad828e2cc8ea42f20614',
               initial='299cf61f02216993f761db6a9d70f68f7b84b9d3',
               after='0973ff633aee53bc9fc3592f80803806bf2d62e6')


def data_sets():
    _, pages, files, _, events = audit.load(prior.DEVELOPMENT)
    return dict(development=(pages, files, events),
                exposed_repetition=prior.confirmation(SCOPE),
                exposed_framing=prior.confirmation(FRAMING),
                confirmation=prior.confirmation(ROOT))


def credits_match(finding, event, pages):
    semantics = SEMANTICS.get(finding['rule_id'], set())
    # A grouped self-description diagnostic can identify two announcements of
    # the same scope. A single scope announcement cannot receive repeat credit.
    if finding['rule_id'] == 'filler.document-metadiscourse' and len(audit.locations(finding)) > 1:
        semantics = semantics | {'needless_repetition'}
    return event['category'] in semantics and audit.matches(finding, event, pages)


def validate_claim(finding, row, pages, events):
    audit.require(row['finding_sha256'] == audit.sha(audit.encode(finding)) and row['rationale'], 'Finding review drift')
    audit.require(row['status'] in ('actionable', 'nonactionable', 'uncertain'), 'Invalid judgment')
    full, partial = row['events'], row.get('partial_events', [])
    audit.require(bool(full or partial) == (row['status'] == 'actionable'), 'Invalid event credit')
    audit.require(not set(full) & set(partial) and len(set(full + partial)) == len(full + partial), 'Duplicate credit')
    for eid in full + partial:
        event = events[eid]
        audit.require(event['kind'] not in ('uncertain', 'controls') and credits_match(finding, event, pages), 'Unsupported credit')


def development_claims(profile, before):
    old_before = audit.read(SCOPE/'reports/development/before'/(profile+'.json.gz'))
    old_after = audit.read(SCOPE/'reports/development/after'/(profile+'.json.gz'))
    claims = {prior.key(f): row for f, row in zip(old_before['findings'], prior.development_claims(profile, old_before), strict=True)}
    changes = audit.read(SCOPE/'dispositions.json')['development'][profile]
    for row in changes['removed']:
        del claims[prior.key(old_before['findings'][row['index']])]
    for row in changes['added']:
        claims[prior.key(old_after['findings'][row['index']])] = row
    audit.require(before['findings'] == old_after['findings'], 'Merged baseline diagnostic drift')
    return [dict(claims[prior.key(f)], index=i, finding_sha256=audit.sha(audit.encode(f)))
            for i, f in enumerate(before['findings'])]


def transfer(before, after, rows, added):
    by_key = {prior.key(f): r for f, r in zip(before['findings'], rows, strict=True)}
    additions = audit.unique(added, 'index', 'added disposition')
    return [additions[i] if i in additions else dict(by_key[prior.key(f)], index=i,
            finding_sha256=audit.sha(audit.encode(f))) for i, f in enumerate(after['findings'])]


def check_rows(report, rows, pages, events, complete):
    index = audit.unique(rows, 'index', 'finding review')
    expected = set(range(len(report['findings']))) if complete else {
        i for i, f in enumerate(report['findings']) if any(
            e['kind'] == 'defects' and credits_match(f, e, pages) for e in events.values())}
    audit.require(set(index) == expected, 'Missing or unexpected finding review')
    for i, row in index.items():
        validate_claim(report['findings'][i], row, pages, events)


def page_metrics(report, rows, pages, events, full):
    found = {eid for row in rows for eid in row['events'] if events[eid]['kind'] == 'defects'}
    index = {row['index']: row for row in rows}
    result = []
    for pid, page in pages.items():
        ids = {eid for eid, e in events.items() if e['page'] == pid and e['kind'] == 'defects'}
        findings = [i for i, f in enumerate(report['findings']) if f['primary']['path'] == page['path']]
        counts = Counter(index[i]['status'] for i in findings if i in index)
        result.append(dict(page=pid, path=page['original_path'], cohort=page['cohort'],
                           length_stratum=page['length_stratum'], format=page['format'],
                           defects=len(ids), detected=len(ids & found), missed=sorted(ids - found),
                           findings=len(findings), dispositions=dict(counts) if full else None))
    return result, found


def profile_metrics(before, after, claims, pages, events, complete):
    phases, page_rows = {}, {}
    for phase, report in [('before', before), ('after', after)]:
        rows, found = page_metrics(report, claims[phase], pages, events, complete)
        page_rows[phase] = rows
        ids = {eid for eid, e in events.items() if e['kind'] == 'defects'}
        counts = Counter(row['status'] for row in claims[phase])
        phases[phase] = dict(defects=len(ids), detected=len(found), recall=audit.ratio(len(found), len(ids)),
            missed=sorted(ids - found), findings=len(report['findings']),
            dispositions=dict(counts) if complete else None,
            actionable_fraction=audit.ratio(counts['actionable'], len(report['findings'])) if complete else None,
            partial_events=sorted({eid for row in claims[phase] for eid in row.get('partial_events', [])}),
            categories={c: dict(defects=sum(e['kind'] == 'defects' and e['category'] == c for e in events.values()),
                               detected=sum(events[eid]['category'] == c for eid in found)) for c in sorted(audit.CATEGORIES)})
    return dict(phases=phases, pages=page_rows, review_scope='all diagnostics' if complete else 'targeted credits only',
                sensitivity=paired_sensitivity(page_rows) if complete else None)


def paired_sensitivity(rows):
    cells = {}
    for i, row in enumerate(rows['before']):
        cells.setdefault((row['cohort'], row['length_stratum']), []).append(i)
    rng, values = random.Random(917304), []
    for _ in range(2000):
        sample = [rng.choice(cell) for cell in cells.values() for _ in cell]
        count = sum(rows['before'][i]['defects'] for i in sample)
        if count:
            values.append(tuple(sum(rows[p][i]['detected'] for i in sample)/count for p in ('before', 'after')))
    def interval(values):
        values = sorted(values)
        return [values[int((len(values)-1)*q)] for q in (.025, .975)] if values else None
    return dict(method='paired page bootstrap within cohort/length cells; descriptive sensitivity only',
                seed=917304, replicates=len(values), before=interval([v[0] for v in values]),
                after=interval([v[1] for v in values]), delta=interval([v[1]-v[0] for v in values]))


def resource_check(split, profile, after):
    folder = ROOT/'reports'/split/'initial'
    record = audit.read(folder/'runs.json')[profile]
    path = folder/(profile+'.json.gz')
    audit.require(audit.sha(path.read_bytes()) == record['sha256'], 'Initial replay drift')
    initial = audit.read(path)
    audit.require(initial['manifest']['tool_commit'] == record['tool_commit'] == COMMITS['initial'], 'Initial revision drift')
    for field in ('findings', 'assessments', 'documents', 'gate'):
        audit.require(initial[field] == after[field], 'Resource correction changed '+field)
    abstentions = initial.get('abstentions', [])
    expected = [dict(path='sources/c06.mdx', rule_id='repetition.repeated-claim', rule_version='1',
                     reason='budget_exhausted', detail='repetition work exceeds max_candidates')] if split == 'exposed_framing' else []
    audit.require(abstentions == expected, 'Unexpected initial abstention')
    return dict(findings_assessments_documents_gate_identical=True, initial_abstentions=abstentions, costs=record)


def evaluate():
    for filename in ('input-freeze.json', 'annotation-freeze.json'):
        for name, digest in audit.read(ROOT/filename)['files'].items():
            audit.require(audit.sha((ROOT/name).read_bytes()) == digest, 'Frozen artifact drift: '+name)
    sets = data_sets()
    seen = set()
    for pages, _, _ in sets.values():
        refs = {p['reference'] for p in pages.values()}
        audit.require(not refs & seen, 'Reused confirmation source')
        seen |= refs
    review = audit.read(ROOT/'dispositions.json')
    result = {}
    for split, (pages, files, events) in sets.items():
        result[split] = {}
        for profile in ('technical', 'strict'):
            before, bc = prior.run(ROOT, split, 'before', profile, pages, files)
            after, ac = prior.run(ROOT, split, 'after', profile, pages, files)
            audit.require(bc['tool_commit'] == COMMITS['before'] and ac['tool_commit'] == COMMITS['after'], 'Engine revision drift')
            changes = prior.delta(before, after)
            judgments = review[split][profile]
            for kind, report in [('added', after), ('removed', before)]:
                rows = audit.unique(judgments[kind], 'index', 'delta review')
                audit.require(set(rows) == set(changes[kind]), 'Incomplete delta review')
                for i, row in rows.items():
                    validate_claim(report['findings'][i], row, pages, events)
            complete = split in ('development', 'confirmation')
            if split == 'development':
                bclaims = development_claims(profile, before)
                aclaims = transfer(before, after, bclaims, judgments['added'])
            else:
                bclaims, aclaims = judgments['before'], judgments['after']
            for report, rows in [(before, bclaims), (after, aclaims)]:
                check_rows(report, rows, pages, events, complete)
            summary = profile_metrics(before, after, dict(before=bclaims, after=aclaims), pages, events, complete)
            result[split][profile] = dict(summary, delta=changes, gate=dict(before=before['gate'], after=after['gate']),
                costs=dict(before=bc, after=ac), resource_correction=resource_check(split, profile, after))
    return result


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--output', type=Path)
    args = parser.parse_args()
    result = evaluate()
    if args.output:
        audit.write(args.output, result)
    else:
        print(audit.encode(result).decode())
