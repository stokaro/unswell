#!/usr/bin/env python3
"""Render the validated complete-page results and source-bound miss ledger."""
import sys
sys.dont_write_bytecode = True
import summarize as m


def table(headers, rows):
    def cell(value):
        return str(value).replace('|', '&#124;').replace('\n', ' ')
    return '\n'.join(['| '+' | '.join(headers)+' |', '| '+' | '.join(['---']*len(headers))+' |']+
                     ['| '+' | '.join(cell(v) for v in row)+' |' for row in rows])


def main():
    results=m.evaluate();m.audit.write(m.ROOT/'summary.json',results);sets=m.data_sets()
    counts=[];burden=[];costs=[]
    for name, profiles in results.items():
        b,a=(profiles['technical']['phases'][p] for p in ('before','after'))
        counts.append([name,len(sets[name][0]),a['defects'],b['detected'],a['detected']])
        for profile, data in profiles.items():
            a=data['phases']['after'];dispositions=a['dispositions']
            if dispositions is not None:
                burden.append([name,profile,a['findings'],dispositions.get('actionable',0),
                               dispositions.get('uncertain',0),dispositions.get('nonactionable',0)])
            b,a=(data['costs'][p] for p in ('before','after'))
            costs.append([name,profile,f"{b['wall_seconds']:.3f} → {a['wall_seconds']:.3f}",
                          f"{b['max_rss_bytes']/2**20:.1f} → {a['max_rss_bytes']/2**20:.1f}"])
    confirm=results['confirmation']['technical'];engines=m.audit.read(m.ROOT/'engines.json')
    pages=[[p['page'],p['path'],p['defects'],p['detected']] for p in confirm['pages']['after']]
    freeze=m.audit.read(m.ROOT/'annotation-freeze.json')
    text=f'''# Purpose recall: ten exposed gains, no new-page gain

Three existing rhetorical rules identify ten additional frozen defects on the
82 exposed pages. On six separately selected complete pages, detection remains
1/28 before and after. This change does not establish broad useful recall.

## Runtime change

The rules match clauses that announce a purpose or declare that a fact is worth
knowing. They also match unsupported claims about what most users want, claims
that output proves a tool understood its input, and document announcements. Local evaluation and reader candidates may follow a longer technical
premise: the candidate stays within 48 tokens and the sentence within 96.
Numbers or reasons in a separate premise no longer erase a local evaluation.
A colon followed by advice does not prove a majority or nearly-always claim.
Concrete component purposes, measurements, quotations, reported claims and
technical conditions remain controls. Some proposed purpose/action constructions
remain unsupported; the protocol lists hypotheses, not guaranteed coverage.

Versions are 5 for `filler.evaluative-closure` and
`filler.document-justification`, and 3 for `filler.unscoped-assurance`.
The catalog retains 53 rules and class manifest r8. These are experimental
warnings; weights, thresholds and gate policy are unchanged.

## Source-only evidence

The implementing assistant froze the [protocol](protocol.md), selection code
and input archive before reading the new pages. The review covered all six
complete extracted sources across the seven rubric categories. The assistant
froze labels at `{freeze['frozen_at']}` before
runtime edits or inspection of diagnostic output: 28 defects, 6 uncertain events
and 15 acceptable controls. Selection excludes all 82 previously exposed source
references and hashes and retains source licenses. The two Ptah MDX pages use
native MDX extraction. Protected code and frontmatter are context only.

The reviewer is the implementing Codex assistant, accepted by the maintainer
under [ADR 0041](../../../docs/adr/0041-assistant-review-acceptance.md).
These are assistant editorial judgments. They are not human labels, independent
agreement, authorship evidence or population estimates. No extra reviewer is
required for this diagnostic work. One page per cohort/length cell does not
support an informative within-cell confidence interval.

{table(['Set','Pages','Frozen defects','Before full','After full'],counts)}

Both profiles detect the same frozen defects. One document-dedication warning
partly covers an exposed framing event; its future-description sentence remains
undetected, so it receives no full credit. Existing partial credits remain
partial. Two changed locations include a conjunction without changing their
editorial meaning; one warning window now covers two separately located events.
Every added and removed finding has a source-bound disposition.

## New-page result and remaining work

{table(['Page','Original path','Defects','Detected'],pages)}

The one detection is the existing worth-stating announcement on Ptah's Protobuf
page. Ptah remains at 1/13; the historical pages remain at 0/15. No confirmation
finding changes. The [miss ledger](CONFIRMATION-MISSES.md) retains all 27 misses.
They include purpose restatements, document self-reference, indirect instructions,
vague rankings and repeated explanations. All six pages pass both profile gates;
passing is not evidence that their wording is clean.

The new rules did not catch any more defects on the six new pages.
Another matching fixture or a lower gate cannot establish the working 80%
recall / 85% soft-diagnostic precision objectives. Those objectives remain open.
The six confirmation pages are now exposed. Further semantic tuning requires
fresh source-only confirmation; do not relabel these misses or overwrite this
negative result. #299, #305 and unsupported #309 constructions remain open.

## Review burden and applicability

{table(['Set','Profile','Findings','Actionable','Uncertain','Nonactionable'],burden)}

Only one warning identifies a frozen defect on the new pages. There are 11
other findings in technical and 14 in strict. Length and punctuation warnings
do not identify the frozen wording edits. The old CMake text-mode limitation and
the single repeated-claim budget abstention on exposed instruction page c05
remain visible; this change does not include the separate #312 optimization.
There are no other abstentions or operational errors. All technical gates pass;
strict still fails only on the exposed instruction set's existing forbidden
worth-noting phrase. No new gate failure is manufactured.

## Reproduction

Before runtime: `{engines['before']}`. After runtime: `{engines['after']}`.
The old before reports retain the prior measurement and original hashes. The new
confirmation before run uses the same preceding runtime binary. All after runs
use the new binary. Reports contain sources, exact byte ranges, model and policy
identities, commands, CPU/RSS, host and artifact hashes.

{table(['Set','Profile','Wall seconds before → after','Peak MiB before → after'],costs)}

These local timings are not a controlled performance comparison or 2-vCPU
qualification; runs occurred at different times and the new scans overlapped
ordinary test work. They do not justify a speed claim.

Run `python3 tools/render.py` and `python3 tools/test_evidence.py` here to validate
and regenerate the evidence without model calls or network access. CLI replay
uses `tools/measure.py` with an explicit binary, set and new output directory.
The validator rejects frozen-input drift, missing review rows, unsupported
cross-page credit, incidental overlap and hidden applicability changes.
'''
    (m.ROOT/'README.md').write_text(text)
    reviews=m.audit.read(m.ROOT/'dispositions.json');lines=['# Reviewed changes','','Indices refer to the retained strict reports.','']
    for name in sets:
        _,files,_=sets[name]
        for kind,phase in [('added','after'),('removed','before')]:
            rows=reviews[name]['strict'][kind]
            if not rows:continue
            lines+=['## '+name+' / '+kind,''];r=m.audit.read(m.ROOT/'reports'/name/phase/'strict.json.gz')
            for row in rows:
                f=r['findings'][row['index']]
                lines += [f"### {row['index']}: {f['rule_id']}",'',row['rationale'],'',
                          'Full: '+(', '.join(row['events']) or 'none')+'. Partial: '+(', '.join(row.get('partial_events',[])) or 'none')+'.','']
                for loc in m.audit.locations(f):
                    span=loc['span'];lines += [f"`{loc['path']}` bytes {span['start']}–{span['end']}",'','```text',files[loc['path']][span['start']:span['end']].decode(),'```','']
    (m.ROOT/'CHANGES.md').write_text('\n'.join(lines))
    _,_,events=sets['confirmation'];lines=['# Remaining confirmation defects','','All 27 remain missed in both profiles.','']
    for eid in confirm['phases']['after']['missed']:
        e=events[eid];lines+=['## '+eid+': '+e['category'],'',e['rationale'],'','Proposed edit: '+e['proposed_edit'],'']
        for span in e['targets']:lines += [f"Bytes {span['start']}–{span['end']}",'','```text',span['quote'],'```','']
    (m.ROOT/'CONFIRMATION-MISSES.md').write_text('\n'.join(lines))


if __name__=='__main__':
    main()
