#!/usr/bin/env python3
"""Validate clause-grammar inputs, judgments and full-page measurements."""
import argparse
import importlib.util
from pathlib import Path
import sys

sys.dont_write_bytecode = True
ROOT = Path(__file__).resolve().parent.parent
PARENT = ROOT.parent/'2026-09-17-clause-boundaries'
spec = importlib.util.spec_from_file_location('clause_boundary_evidence', PARENT/'tools/summarize.py')
previous = importlib.util.module_from_spec(spec)
spec.loader.exec_module(previous)
audit, prior = previous.audit, previous.prior
metrics = previous.metrics

def frozen_inputs():
    for filename, field in [('input-freeze.json', 'files'), ('annotation-freeze.json', 'sha256')]:
        for name, digest in audit.read(ROOT/filename)[field].items():
            audit.require(audit.sha((ROOT/name).read_bytes()) == digest, 'Frozen artifact drift: '+name)


def confirmation():
    manifest = audit.read(ROOT/'confirmation/manifest.json')
    pages = audit.unique(manifest['pages'], 'id', 'confirmation page')
    files = audit.archive_sources(ROOT/'confirmation/inputs.tar.gz')
    labels = audit.read(ROOT/'confirmation/annotations.json')
    reviews = audit.unique(labels['pages'], 'page', 'review')
    audit.require(set(reviews) == set(pages), 'Missing confirmation review')
    events = {}
    for pid, page in pages.items():
        src = files[page['path']]
        audit.require(audit.sha(src) == page['sha256'] and len(src) == page['bytes'], 'Source drift')
        for notice in page['retained_notices']:
            audit.require(audit.sha(files[notice['path']]) == notice['sha256'], 'Notice drift')
        review = reviews[pid]
        audit.require(review['source_sha256'] == page['sha256'] and review['reviewer'] and
                      review['phase'] == 'before_confirmation_diagnostic_output_review_and_before_runtime_changes', 'Review provenance')
        audit.require(set(review['categories_reviewed']) == audit.CATEGORIES, 'Incomplete rubric coverage')
        end = 0
        for section in review['coverage']:
            audit.span(src, section)
            audit.require(section['start'] == end and section['rationale'], 'Coverage gap')
            end = section['end']
        audit.require(end == len(src), 'Incomplete coverage')
        for kind in ('defects', 'uncertain', 'controls'):
            for event in review[kind]:
                audit.require(event['id'] not in events and event['rationale'] and event['proposed_edit'], 'Invalid event')
                audit.require(event['category'] in audit.CATEGORIES and event['targets'], 'Invalid target')
                for target in event['targets'] + event['context']:
                    audit.span(src, target)
                events[event['id']] = dict(event, page=pid, kind=kind)
    return pages, files, events


def data_sets():
    sets = previous.data_sets()
    sets['exposed_boundaries'] = sets.pop('confirmation')
    sets['confirmation'] = confirmation()
    metrics.validate_sources(sets)
    seen = set()
    for pages, _, _ in sets.values():
        hashes = {p['sha256'] for p in pages.values()}
        audit.require(not seen & hashes, 'Repeated source bytes')
        seen |= hashes
    validate_exposure(sets, audit.read(ROOT/'confirmation/source-overlap.json'))
    return sets


def inherited_claims(split, profile):
    old_split = 'confirmation' if split == 'exposed_boundaries' else split
    return (audit.read(PARENT/'reports'/old_split/'after'/(profile+'.json.gz')),
            audit.read(PARENT/'dispositions.json')[old_split][profile]['after'])


def validate_abstentions(report, split):
    audit.require(not report.get('abstentions'), 'Unexpected abstention; preserve and review it')


def run(split, phase, profile, pages, files):
    # Both phases are remeasured from the merged baseline and semantic freeze.
    folder = ROOT/'reports'/split/phase
    record = audit.read(folder/'runs.json')[profile]
    path = folder/(profile+'.json.gz')
    audit.require(audit.sha(path.read_bytes()) == record['sha256'], 'Report drift')
    report = audit.read(path)
    audit.require(report['status'] == 'complete' and report['manifest']['complete'] and not report['errors'], 'Incomplete run')
    validate_abstentions(report, split)
    audit.require(report['manifest']['tool_commit'] == record['tool_commit'] and not report['manifest']['skipped_rules'], 'Engine or capability drift')
    audit.require(not report['manifest']['no_gate'] and report['manifest']['include_source'], 'Changed output policy')
    audit.require(report['manifest']['scoring_profile'] == profile+'-v1' and
                  record['config_sha256'] == audit.sha((prior.DEVELOPMENT/(profile+'.yaml')).read_bytes()), 'Profile drift')
    audit.require(report['manifest']['config_sources'] == [dict(path=profile+'.yaml', kind='config',
                  sha256=record['config_identity_sha256'])], 'Config provenance drift')
    audit.require(record['exit_code'] == (0 if report['gate']['passed'] else 1), 'Operational exit/gate mismatch')
    docs = audit.unique(report['documents'], 'name', 'document')
    audit.require(set(docs) == {p['path'] for p in pages.values()}, 'Document set differs')
    for page in pages.values():
        doc = docs[page['path']]
        audit.require(doc['source'].encode() == files[page['path']] and doc['source_hash'] == page['sha256'] and
                      doc['format'] == page['format'], 'Reported source drift')
    for finding in report['findings']:
        for loc in audit.locations(finding):
            audit.span(files[loc['path']], loc['span'])
            for segment in loc.get('segments', []):
                audit.span(files[loc['path']], segment)
    return report, record


def finding_delta(before, after):
    b = {prior.key(f): i for i, f in enumerate(before['findings'])}
    a = {prior.key(f): i for i, f in enumerate(after['findings'])}
    audit.require(len(b) == len(before['findings']) and len(a) == len(after['findings']), 'Duplicate finding key')
    def comparable(f):
        return {k:v for k,v in f.items() if k not in ('id','fingerprint','rule_version','message')}
    changed = {key for key in a.keys() & b.keys()
               if comparable(before['findings'][b[key]]) != comparable(after['findings'][a[key]])}
    # A changed suggestion is evidence, so it receives a fresh disposition even
    # when the source location is stable. Other evidence changes do too.
    return dict(added=sorted(a[k] for k in a.keys()-b.keys() | changed),
                removed=sorted(b[k] for k in b.keys()-a.keys() | changed),
                changed_at_same_location=len(changed))


def validate_exposure(sets, record):
    pages, files, _ = sets['confirmation']
    sources = {p['reference']: raw[p['path']] for name, (ps, raw, _) in sets.items()
               if name != 'confirmation' for p in ps.values()}
    audit.require(record['prior_references_scanned'] == len(sources), 'Exposure source count drift')
    matched = set()
    for match in record['matches']:
        audit.require(match['page'] in pages and match['matches'], 'Invalid exposure match')
        source = files[pages[match['page']]['path']]
        audit.span(source, match)
        quote = ' '.join(match['quote'].split())
        audit.require(len(quote.split()) >= 12, 'Exposure match below declared minimum')
        for earlier in match['matches']:
            reference = earlier['reference']
            audit.require(reference in sources and quote in ' '.join(sources[reference].decode().split()),
                          'Exposure quote absent from prior source')
        matched.add(match['page'])
    audit.require(matched == set(record['known_partially_exposed_pages']), 'Hidden partial exposure')


def exposure_metrics(summary):
    exposure = audit.read(ROOT/'confirmation/source-overlap.json')
    known = set(exposure['known_partially_exposed_pages'])
    result = {}
    for name, included in [('without_known_overlap', lambda pid: pid not in known),
                           ('known_partial_exposure', lambda pid: pid in known)]:
        result[name] = {}
        for phase, rows in summary['pages'].items():
            selected = [r for r in rows if included(r['page'])]
            result[name][phase] = dict(pages=len(selected), defects=sum(r['defects'] for r in selected),
                                      detected=sum(r['detected'] for r in selected))
    return result


def evaluate():
    frozen_inputs()
    sets = data_sets()
    review = audit.read(ROOT/'dispositions.json')
    commits = audit.read(ROOT/'engines.json')
    result = {}
    for split, (pages, files, events) in sets.items():
        result[split] = {}
        for profile in ('technical', 'strict'):
            before, bc = run(split, 'before', profile, pages, files)
            after, ac = run(split, 'after', profile, pages, files)
            audit.require(bc['tool_commit'] == commits['before'] and ac['tool_commit'] == commits['after'], 'Engine revision drift')
            delta = finding_delta(before, after)
            judgments = review[split][profile]
            for kind, report in [('added', after), ('removed', before)]:
                rows = audit.unique(judgments[kind], 'index', 'delta review')
                audit.require(set(rows) == set(delta[kind]), 'Incomplete delta review')
                for index, row in rows.items():
                    metrics.validate_claim(report['findings'][index], row, pages, events)
            complete = split not in ('exposed_repetition', 'exposed_framing')
            bclaims, aclaims = judgments['before'], judgments['after']
            for report, rows in [(before, bclaims), (after, aclaims)]:
                metrics.check_rows(report, rows, pages, events, complete)
            summary = metrics.profile_metrics(before, after, dict(before=bclaims, after=aclaims), pages, events, complete)
            if split == 'confirmation':
                summary['sensitivity'] = dict(method='No interval: one page per cohort/length cell; report observed counts only',
                                              reason='Within-cell bootstrap would give a degenerate interval')
            if split == 'confirmation':
                summary['exposure_subsets'] = exposure_metrics(summary)
            result[split][profile] = dict(summary, delta=delta,
                gate=dict(before=before['gate'], after=after['gate']), costs=dict(before=bc, after=ac),
                abstentions=dict(before=before.get('abstentions', []), after=after.get('abstentions', [])))
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
