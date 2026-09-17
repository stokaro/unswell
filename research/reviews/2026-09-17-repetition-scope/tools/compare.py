#!/usr/bin/env python3
"""Validate source-bound before/after claims, then derive event recall and deltas."""
import argparse
from collections import Counter
from pathlib import Path
import sys

sys.dont_write_bytecode = True
ROOT = Path(__file__).resolve().parent.parent
DEVELOPMENT = ROOT.parent/'2026-09-17-full-page-recall'
sys.path.insert(0, str(DEVELOPMENT/'tools'))
import evaluate as audit

SEMANTICS = dict(audit.SEMANTICS, **{'filler.unscoped-assurance':
                                    {'empty_framing', 'vague_claims', 'unjustified_intensifiers'},
                                    'repetition.duplicate-list-item': {'needless_repetition'}})


def key(finding):
    return audit.encode([finding['rule_id'], [dict(path=l['path'], span=l['span'],
                        segments=l.get('segments', [])) for l in audit.locations(finding)]])


def delta(before, after):
    b = {key(f):i for i,f in enumerate(before['findings'])}
    a = {key(f):i for i,f in enumerate(after['findings'])}
    audit.require(len(b) == len(before['findings']) and len(a) == len(after['findings']), 'Duplicate finding key')
    for k in b.keys() & a.keys():
        # Matcher versions and display text may change; no policy, suppression,
        # evidence, source location or activation change can hide as a retention.
        def comparable(finding):
            return {name:value for name,value in finding.items()
                    if name not in ('id','fingerprint','rule_version','message')}
        audit.require(comparable(before['findings'][b[k]]) == comparable(after['findings'][a[k]]),
                      'Retained diagnostic changed beyond identity/display metadata')
    return dict(added=sorted(a[k] for k in a.keys()-b.keys()),
                removed=sorted(b[k] for k in b.keys()-a.keys()))


def confirmation(root):
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
                      r['phase'] == 'before_confirmation_diagnostic_output_review', 'Review provenance')
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


def run(root, split, phase, profile, pages, files):
    folder = root/'reports'/split/phase
    record = audit.read(folder/'runs.json')[profile]
    path = folder/(profile+'.json.gz')
    audit.require(audit.sha(path.read_bytes()) == record['sha256'], 'Report drift')
    report = audit.read(path)
    audit.require(report['status'] == 'complete' and report['manifest']['complete'] and not report['errors'], 'Incomplete run')
    audit.require(not report.get('abstentions'), 'Published zero-abstention claim does not match the report')
    audit.require(report['manifest']['tool_commit'] == record['tool_commit'] and not report['manifest']['skipped_rules'], 'Engine or capability drift')
    audit.require(not report['manifest']['no_gate'] and report['manifest']['include_source'], 'Changed output policy')
    audit.require(report['manifest']['scoring_profile'] == profile+'-v1' and
                  record['config_sha256'] == audit.sha((DEVELOPMENT/(profile+'.yaml')).read_bytes()), 'Profile drift')
    audit.require(report['manifest']['config_sources'] == [dict(path=profile+'.yaml', kind='config',
                  sha256=record['config_identity_sha256'])], 'Config provenance drift')
    audit.require(record['exit_code'] == (0 if report['gate']['passed'] else 1), 'Operational exit/gate mismatch')
    docs = audit.unique(report['documents'],'name','document')
    audit.require(set(docs) == {p['path'] for p in pages.values()}, 'Document set differs')
    for p in pages.values():
        d = docs[p['path']]
        audit.require(d['source'].encode() == files[p['path']] and d['source_hash'] == p['sha256'] and
                      d['format'] == p['format'], 'Reported source drift')
    for f in report['findings']:
        for loc in audit.locations(f):
            audit.span(files[loc['path']],loc['span'])
            for segment in loc.get('segments',[]):
                audit.span(files[loc['path']],segment)
    return report,record


def validate_claim(finding, row, pages, events):
    audit.require(row['finding_sha256'] == audit.sha(audit.encode(finding)) and row['rationale'], 'Finding review drift')
    audit.require(row['status'] in ('actionable','nonactionable','uncertain'), 'Invalid judgment')
    audit.require(bool(row['events']) == (row['status']=='actionable'), 'Invalid event credit')
    for eid in row['events']:
        event = events[eid]
        audit.require(event['kind'] not in ('uncertain','controls'), 'Uncertain/control credit')
        audit.require(event['category'] in SEMANTICS.get(finding['rule_id'],set()) and
                      audit.matches(finding,event,pages), 'Unsupported credit')


def development_claims(profile, before):
    framing = ROOT.parent/'2026-09-17-framing-recall'
    original = audit.read(DEVELOPMENT/(profile+'.json.gz'))
    rows = audit.read(DEVELOPMENT/'dispositions.json')[profile]
    claims = {key(f):r for f,r in zip(original['findings'],rows,strict=True)}
    prior = audit.read(framing/'reports/development/after'/(profile+'.json.gz'))
    changes = audit.read(framing/'dispositions.json')['development'][profile]
    audit.require(not changes['removed'], 'Update inheritance for prior removals')
    for row in changes['added']:
        claims[key(prior['findings'][row['index']])] = row
    audit.require(before['findings'] == prior['findings'], 'Merged baseline diagnostic drift')
    return [dict(claims[key(f)], index=i, finding_sha256=audit.sha(audit.encode(f)))
            for i,f in enumerate(before['findings'])]


def evaluate(root=ROOT):
    for freeze_name in ('input-freeze.json','annotation-freeze.json'):
        for name,digest in audit.read(root/freeze_name)['files'].items():
            audit.require(audit.sha((root/name).read_bytes()) == digest, 'Frozen artifact drift: '+name)
    _, dp, df, _, de = audit.load(DEVELOPMENT)
    cp, cf, ce = confirmation(root)
    audit.require(not ({p['reference'] for p in dp.values()} & {p['reference'] for p in cp.values()}), 'Confirmation overlaps development')
    prior = audit.read(ROOT.parent/'2026-09-17-framing-recall/confirmation/manifest.json')
    audit.require(not ({p['reference'] for p in prior['pages']} & {p['reference'] for p in cp.values()}), 'Confirmation reused')
    review = audit.read(root/'dispositions.json')
    result = {}
    for split,pages,files,events in [('development',dp,df,de),('confirmation',cp,cf,ce)]:
        result[split] = {}
        for profile in ('technical','strict'):
            before,bcost = run(root,split,'before',profile,pages,files)
            after,acost = run(root,split,'after',profile,pages,files)
            changes = delta(before,after)
            judgments = review[split][profile]
            for kind,report in [('added',after),('removed',before)]:
                rows = audit.unique(judgments[kind],'index','delta disposition')
                audit.require(set(rows) == set(changes[kind]), 'Incomplete delta review')
                for i,row in rows.items():validate_claim(report['findings'][i],row,pages,events)
            if split == 'development':
                bclaims = development_claims(profile, before)
                by_key = {key(f):r for f,r in zip(before['findings'],bclaims,strict=True)}
                added = {r['index']:r for r in judgments['added']}
                aclaims = []
                for i,f in enumerate(after['findings']):
                    row = added[i] if i in added else dict(by_key[key(f)],index=i,finding_sha256=audit.sha(audit.encode(f)))
                    validate_claim(f,row,pages,events)
                    aclaims.append(row)
            else:
                bclaims,aclaims = judgments['before_credits'],judgments['after_credits']
                for report,claims in [(before,bclaims),(after,aclaims)]:
                    ids = set()
                    for row in claims:
                        audit.require(row['index'] not in ids,'Duplicate credit');ids.add(row['index'])
                        validate_claim(report['findings'][row['index']],row,pages,events)
                    candidates = {i for i,f in enumerate(report['findings']) if any(
                        e['kind']=='defects' and e['category'] in SEMANTICS.get(f['rule_id'],set()) and audit.matches(f,e,pages)
                        for e in events.values())}
                    audit.require(ids == candidates, 'Unreviewed possible target credit')
            found = [{e for r in claims for e in r['events'] if events[e]['kind']=='defects'} for claims in (bclaims,aclaims)]
            page_rows = []
            for pid,p in pages.items():
                ids = {e['id'] for e in events.values() if e['page']==pid and e['kind']=='defects'}
                page_rows.append(dict(page=pid,cohort=p['cohort'],selection=p.get('selection','confirmation'),
                    defects=len(ids),before=len(ids & found[0]),after=len(ids & found[1]),
                    missed=sorted(ids-found[1])))
            repetition_ids = {e['id'] for e in events.values()
                              if e['kind'] == 'defects' and e['category'] == 'needless_repetition'}
            repetition = dict(defects=len(repetition_ids), before=len(repetition_ids & found[0]),
                              after=len(repetition_ids & found[1]),
                              missed=sorted(repetition_ids-found[1]))
            result[split][profile] = dict(pages=page_rows,delta=changes,repetition=repetition,
                added_dispositions=dict(Counter(r['status'] for r in judgments['added'])),
                findings=dict(before=len(before['findings']),after=len(after['findings'])),
                gate=dict(before=before['gate'],after=after['gate']),
                costs=dict(before=bcost,after=acost))
    return result


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--output',type=Path)
    args = parser.parse_args()
    result = evaluate()
    if args.output:audit.write(args.output,result)
    else:print(audit.encode(result).decode())
