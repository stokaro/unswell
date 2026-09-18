#!/usr/bin/env python3
"""Render reviewed whole-page outcomes without converting partial repairs to recall."""
import sys
sys.dont_write_bytecode = True
import summarize as m


def table(headers, rows):
    return '\n'.join(['| '+' | '.join(headers)+' |', '| '+' | '.join(['---']*len(headers))+' |']+
                     ['| '+' | '.join(str(v).replace('|','\\|').replace('\n',' ') for v in row)+' |' for row in rows])


def main():
    result=m.evaluate();m.audit.write(m.ROOT/'summary.json',result)
    sets=m.data_sets();review=m.audit.read(m.ROOT/'dispositions.json');engines=m.audit.read(m.ROOT/'engines.json')
    rows=[];costs=[]
    for name,profiles in result.items():
        b,a=(profiles['technical']['phases'][p] for p in ('before','after'))
        rows.append([name,len(sets[name][0]),a['defects'],b['detected'],a['detected']])
        for profile,data in profiles.items():
            b,a=(data['costs'][p] for p in ('before','after'))
            costs.append([name,profile,f"{b['wall_seconds']:.2f} → {a['wall_seconds']:.2f}",f"{b['max_rss_bytes']/2**20:.1f} → {a['max_rss_bytes']/2**20:.1f}"])
    confirmation=result['confirmation'];a=confirmation['technical']['phases']['after'];missed=a['missed']
    pages=confirmation['technical']['pages']['after'];ptah=[p for p in pages if p['cohort']=='ptah']
    decision=dict(status='incremental_candidate_broad_recall_unmet',baseline_commit=engines['before'],candidate_commit=engines['after'],
        confirmation_defects=a['defects'],confirmation_detected_before=confirmation['technical']['phases']['before']['detected'],
        confirmation_detected_after=a['detected'],ptah_defects=sum(p['defects'] for p in ptah),ptah_detected=sum(p['detected'] for p in ptah),
        reason='Two new confirmation warnings are useful, one fully and one partially. Existing detections are preserved, but broad recall remains poor.',
        rejected_development_hypothesis='Isolated subject clefts: legitimate technical focus dominates the disputed findings.',
        qualification='Single-assistant editorial review accepted under ADR0041; no population, human-model, authorship, or calibrated-probability claim.')
    m.audit.write(m.ROOT/'decision.json',decision)
    dispositions=[]
    for profile,p in confirmation.items():
        for phase,d in p['phases'].items():
            c=d['dispositions'];dispositions.append([profile,phase,d['findings'],c.get('actionable',0),c.get('additional_actionable',0),c.get('uncertain',0),c.get('nonactionable',0)])
    body=f'''# Structural wordiness: useful local repairs, low whole-page recall

The candidate improves full detection from **146/947 to 150/947** on 160 exposed
pages and from **1/63 to 2/63** on twelve new pages. On the six new Ptah pages,
**0/18** defects are fully diagnosed in either phase. These results do not meet
the 80% recall or 85% soft-diagnostic precision goals. A passing gate does not
establish that a page is clean.

The two new rules recognize nominal support around an action or modifier, and
repeated markers of the same grammatical relationship. The instruction rule can
also relate an adjacent nominal instrument to its action announcement. On the
new pages, one warning fully identifies a duplicate additive relationship and
another identifies only the nominal support inside a larger narrated instruction.
Neither rule establishes machine authorship.

An isolated subject-cleft rule was rejected during development: many matches
expressed defensible technical focus. Its source snapshot and every development
judgment are retained in [development-decisions.md](development-decisions.md) and
[development-screening.json](development-screening.json). It is absent from the
frozen confirmation candidate.

## Scope and review

The same Codex assistant selected sources, reviewed them, implemented the
candidate and reviewed diagnostics. This is maintainer-accepted assistant review
under ADR0041. It is not independent human annotation, model qualification or
population accuracy. Frozen labels cover all seven rubric categories and contain
63 defects, 16 uncertain events and 53 controls. Two Ptah pages have no defects
and were retained in the denominator of page-level outcomes.

The prior 160 source identities are exposed development data. The new sample has
four short and two medium Ptah pages, plus two short, two medium and two long
historical pages from six repositories. The pinned Ptah frame has no unused long
pages. Historical expansion and metadata ranking were fixed before source review;
new historical results cannot establish recall on fresh long Ptah pages. Exact
whole-block overlap checks found no prior matches of at least twelve whitespace
tokens. Shorter or paraphrased overlap remains possible. Dates do not prove human
origin.

All twenty sets were replayed in both profiles with the two recorded binaries.
Runs are complete, with no operational errors, skipped rules or abstentions.
Two older targeted sets retain their narrower review scope; every changed warning
still has a disposition. On exposed pages the candidate adds nineteen reviewed
warnings: four full frozen detections, four partial repairs and eleven additional
edits. One shorter instruction finding is replaced by the finding that also
includes its method. No existing fully detected event is lost.

{table(['Set','Pages','Frozen defects','Baseline full','Candidate full'],rows)}

## Confirmation outcomes

{table(['Page','Stored source','Defects','Fully detected'],[[p['page'],p['path'],p['defects'],p['detected']] for p in pages])}

Every finding in both phases and profiles was reviewed. Partial repairs do not
increase full-event recall. Additional edits discovered after diagnostics have
no frozen recall credit. Controls and uncertain labels are unchanged.

{table(['Profile','Phase','Findings','Frozen actionable','Additional repairs','Uncertain','Nonactionable'],dispositions)}

The seven actionable strict findings include only two complete events; the others
partially address four events. Three existing duplicate-word warnings conflate
the Boolean operator name `AND` with the English conjunction `and`. These are
recorded as false positives, not counted as repeated prose. The frozen runtime
is unchanged after confirmation; [#333](https://github.com/stokaro/unswell/issues/333)
tracks this separate correctness repair.

[CHANGES.md](CHANGES.md) lists each diagnostic change. [MISSES.md](MISSES.md) keeps
all {len(missed)} remaining confirmation defects and their source-bound repairs.
The larger misses concern multi-part narration, indirect claims and repeated
explanation. More local cue additions alone have not shown sufficient transfer.
Future experiments must compare broader representations with this rules baseline
on separately grouped pages, preserving contextual controls and the origin/quality
distinction.

## Reproduce and validate

Baseline runtime: `{engines['before']}`.
Candidate runtime: `{engines['after']}`.
The candidate binary was built from the clean commit before the first
confirmation diagnostic output. `code-freeze.json` records all 491 source-input
hashes and the binary hash. The archive container was subsequently normalized to
remove directory entries; every frozen file digest stayed identical. Research
catalog revision 9 and its integration tests were added later and do not change
runtime source inputs.

Build either revision with `CGO_ENABLED=0`, `-trimpath`, and
`-ldflags '-X github.com/stokaro/unswell.BuildCommit=<full revision>'`, then run:

```sh
python3 tools/measure.py --binary /path/to/unswell --set confirmation --output /new/output
python3 tools/test_evidence.py
python3 tools/render.py
```

The runner restores sources outside any enclosing Git checkout, preventing its
index from silently excluding untracked replay inputs. Output directories must
be new. Fourteen evidence tests reject missing judgments, invented credit, source
reuse, changed runtime files and unresolved analysis failures. Public API tests
cover protected operands, quotations, legal restrictions, separate actions,
Unicode, mappings and paragraph boundaries. The annotated CLI fixture includes
revisions and controls with BOM and CRLF. Existing golden changes are catalog
identity hashes only. Race, active fuzzing and coverage remain deferred to #123.
Repository validation outcomes are retained in [validation.json](validation.json).

[COSTS.md](COSTS.md) reports complete local scan costs. They do not qualify the
separate 100,000-word production performance target.
'''
    (m.ROOT/'README.md').write_text(body)
    (m.ROOT/'COSTS.md').write_text('# Complete scan costs\n\nRecorded per fresh CLI process; peak RSS includes extraction, NLP and reporting.\n\n'+table(['Set','Profile','Seconds baseline → candidate','Peak MiB baseline → candidate'],costs)+'\n')
    lines=['# Reviewed diagnostic changes','','Indices refer to retained strict reports. Both profiles have saved dispositions.','']
    for name in result:
        for kind,phase in [('added','after'),('removed','before')]:
            for row in review[name]['strict'][kind]:
                report=m.audit.read(m.ROOT/'reports'/name/phase/'strict.json.gz');f=report['findings'][row['index']]
                lines += [f"## {name}: {kind} {row['index']} ({f['rule_id']})",'',f"Status: {row['status']}. Full frozen credit: {', '.join(row['events']) or 'none'}.",'',row['rationale'],'']
                for loc in m.audit.locations(f):
                    raw=sets[name][1][loc['path']];span=loc['span']
                    lines += [f"{loc['path']} bytes {span['start']}–{span['end']}:",'','> '+' '.join(raw[span['start']:span['end']].decode().split()),'']
    (m.ROOT/'CHANGES.md').write_text('\n'.join(lines).rstrip()+'\n')
    lines=['# Remaining new-page defects','',f'These {len(missed)} frozen events are not fully diagnosed. Four have partial coverage. Origin and an overlapping warning cannot substitute for the stated repair.','']
    for eid in missed:
        e=sets['confirmation'][2][eid];lines += [f"## {eid}: {e['category']}",'',e['rationale'],'',e['proposed_edit'],'']
        for t in e['targets']:lines += ['> '+' '.join(t['quote'].split()),'']
    (m.ROOT/'MISSES.md').write_text('\n'.join(lines).rstrip()+'\n')


if __name__=='__main__':
    main()
