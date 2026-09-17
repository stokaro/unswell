#!/usr/bin/env python3
"""Validate contextual wording evidence and derive complete-page and target metrics."""
import argparse
import importlib.util
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
LOCAL = ROOT.parent/'2026-09-17-local-repetition'
spec = importlib.util.spec_from_file_location('local_repetition_evidence', LOCAL/'tools/summarize.py')
legacy = importlib.util.module_from_spec(spec)
spec.loader.exec_module(legacy)
COMMITS = dict(before='0973ff633aee53bc9fc3592f80803806bf2d62e6', after='c1032187fcc1708617ac9aff53ce12c84c42f70c')


def data_sets():
    _, pages, files, _, events = audit.load(prior.DEVELOPMENT)
    return dict(development=(pages, files, events),
                exposed_repetition=prior.confirmation(SCOPE),
                exposed_framing=prior.confirmation(FRAMING),
                exposed_local=prior.confirmation(LOCAL),
                confirmation=confirmation())


def confirmation():
    root = ROOT
    manifest = audit.read(root/'confirmation/manifest.json')
    pages = audit.unique(manifest['pages'], 'id', 'confirmation page')
    files = audit.archive_sources(root/'confirmation/inputs.tar.gz')
    reviews = audit.unique(audit.read(root/'confirmation/annotations.json'), 'page', 'review')
    audit.require(set(reviews) == set(pages), 'Missing confirmation review')
    events = {}
    for pid,p in pages.items():
        src = files[p['path']]
        audit.require(audit.sha(src) == p['sha256'] and len(src) == p['bytes'], 'Source drift')
        for notice in p['retained_notices']:
            audit.require(audit.sha(files[notice['path']]) == notice['sha256'], 'Notice drift')
        r = reviews[pid]
        audit.require(r['source_sha256'] == p['sha256'] and r['reviewer'] and
                      r['phase'] == 'before_confirmation_diagnostic_output_review_and_before_runtime_changes', 'Review provenance')
        end = 0
        for section in r['coverage']:
            audit.span(src, section)
            audit.require(section['start'] == end and section['rationale'], 'Coverage gap')
            end = section['end']
        audit.require(end == len(src), 'Incomplete coverage')
        for kind in ('defects', 'uncertain', 'controls'):
            for event in r[kind]:
                audit.require(event['id'] not in events and event['rationale'] and event['proposed_edit'], 'Invalid event')
                audit.require(event['category'] in audit.CATEGORIES and event['targets'], 'Invalid target')
                for target in event['targets']:
                    audit.span(src,target)
                events[event['id']] = dict(event,page=pid,kind=kind)
    return pages,files,events


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
    original = audit.read(LOCAL/'reports/development/before'/(profile+'.json.gz'))
    pinned = audit.read(LOCAL/'reports/development/after'/(profile+'.json.gz'))
    rows = legacy.development_claims(profile, original)
    added = audit.read(LOCAL/'dispositions.json')['development'][profile]['added']
    audit.require(before == pinned, 'Pinned baseline drift')
    return legacy.transfer(original, pinned, rows, added)


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



def validate_sources(sets):
    seen = set()
    for pages, _, _ in sets.values():
        refs = {p['reference'] for p in pages.values()}
        audit.require(len(refs) == len(pages) and not refs & seen, 'Reused confirmation source')
        seen |= refs


def validate_probe(root):
    summary = audit.read(root/'summary.json')
    for record in summary['sets']:
        folder = root/record['split']
        for name, digest in record['files'].items():
            audit.require(audit.sha((folder/name).read_bytes()) == digest, 'Probe artifact drift')
        report = audit.read(folder/'report.json.gz')
        audit.require(report['status'] == 'complete' and not report['errors'] and
                      report['manifest']['complete'] and not report['manifest']['skipped_rules'], 'Incomplete probe')
        audit.require(report['manifest']['tool_commit'] == COMMITS['before'] == record['tool_commit'], 'Probe engine drift')
        audit.require(len(report['findings']) == record['findings'] and
                      len(report['documents']) == record['documents'] and
                      Counter(f['rule_id'] for f in report['findings']) == record['rules'] and
                      report.get('abstentions', []) == record['abstentions'], 'Probe summary drift')


def evaluate():
    for filename in ('input-freeze.json', 'annotation-freeze.json'):
        for name, digest in audit.read(ROOT/filename)['files'].items():
            audit.require(audit.sha((ROOT/name).read_bytes()) == digest, 'Frozen artifact drift: '+name)
    sets = data_sets()
    validate_sources(sets)
    validate_probe(ROOT/'all-rules-probe')
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
            complete = split in ('development', 'exposed_local', 'confirmation')
            if split == 'development':
                bclaims = development_claims(profile, before)
                aclaims = transfer(before, after, bclaims, judgments['added'])
            else:
                bclaims, aclaims = judgments['before'], judgments['after']
            for report, rows in [(before, bclaims), (after, aclaims)]:
                check_rows(report, rows, pages, events, complete)
            summary = profile_metrics(before, after, dict(before=bclaims, after=aclaims), pages, events, complete)
            result[split][profile] = dict(summary, delta=changes, gate=dict(before=before['gate'], after=after['gate']),
                costs=dict(before=bc, after=ac))
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
