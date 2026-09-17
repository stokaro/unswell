#!/usr/bin/env python3
"""Validate instruction-recall inputs, judgments and full-page measurements."""
import argparse
import importlib.util
from pathlib import Path
import sys

sys.dont_write_bytecode = True
ROOT = Path(__file__).resolve().parent.parent
PARENT = ROOT.parent/'2026-09-17-context-recall'
spec = importlib.util.spec_from_file_location('context_recall_evidence', PARENT/'tools/summarize.py')
previous = importlib.util.module_from_spec(spec)
spec.loader.exec_module(previous)
audit, prior = previous.audit, previous.prior
previous.SEMANTICS.update({'filler.instruction-scaffolding': {'wordiness', 'empty_framing'},
                          'repetition.redundant-predicate': {'wordiness', 'needless_repetition'}})


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
    sets['exposed_context'] = sets.pop('confirmation')
    sets['confirmation'] = confirmation()
    previous.validate_sources(sets)
    seen = set()
    for pages, _, _ in sets.values():
        hashes = {p['sha256'] for p in pages.values()}
        audit.require(not seen & hashes, 'Repeated source bytes')
        seen |= hashes
    return sets


def inherited_claims(split, profile):
    old_split = 'confirmation' if split == 'exposed_context' else split
    judgments = audit.read(PARENT/'dispositions.json')[old_split][profile]
    before = audit.read(PARENT/'reports'/old_split/'before'/(profile+'.json.gz'))
    after = audit.read(PARENT/'reports'/old_split/'after'/(profile+'.json.gz'))
    if split == 'development':
        rows = previous.development_claims(profile, before)
        return after, previous.transfer(before, after, rows, judgments['added'])
    return after, judgments['after']


def evaluate():
    frozen_inputs()
    sets = data_sets()
    review = audit.read(ROOT/'dispositions.json')
    commits = audit.read(ROOT/'engines.json')
    result = {}
    for split, (pages, files, events) in sets.items():
        result[split] = {}
        for profile in ('technical', 'strict'):
            before, bc = prior.run(ROOT, split, 'before', profile, pages, files)
            after, ac = prior.run(ROOT, split, 'after', profile, pages, files)
            audit.require(bc['tool_commit'] == commits['before'] and ac['tool_commit'] == commits['after'], 'Engine revision drift')
            delta = prior.delta(before, after)
            judgments = review[split][profile]
            for kind, report in [('added', after), ('removed', before)]:
                rows = audit.unique(judgments[kind], 'index', 'delta review')
                audit.require(set(rows) == set(delta[kind]), 'Incomplete delta review')
                for index, row in rows.items():
                    previous.validate_claim(report['findings'][index], row, pages, events)
            complete = split not in ('exposed_repetition', 'exposed_framing')
            if split == 'confirmation':
                bclaims, aclaims = judgments['before'], judgments['after']
            else:
                pinned, bclaims = inherited_claims(split, profile)
                audit.require(before == pinned, 'Inherited baseline drift')
                if complete:
                    aclaims = previous.transfer(before, after, bclaims, judgments['added'])
                else:
                    aclaims = judgments['after']
            for report, rows in [(before, bclaims), (after, aclaims)]:
                previous.check_rows(report, rows, pages, events, complete)
            summary = previous.profile_metrics(before, after, dict(before=bclaims, after=aclaims), pages, events, complete)
            result[split][profile] = dict(summary, delta=delta,
                gate=dict(before=before['gate'], after=after['gate']), costs=dict(before=bc, after=ac))
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
