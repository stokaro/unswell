#!/usr/bin/env python3
"""Render validated counts, reviewed changes and every remaining frozen defect."""
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
    counts=[];costs=[]
    for name, profiles in results.items():
        b,a=(profiles['technical']['phases'][p] for p in ('before','after'))
        counts.append([name,len(sets[name][0]),a['defects'],b['detected'],a['detected']])
        for profile, data in profiles.items():
            b,a=(data['costs'][p] for p in ('before','after'))
            costs.append([name,profile,f"{b['wall_seconds']:.3f} → {a['wall_seconds']:.3f}",
                          f"{b['max_rss_bytes']/2**20:.1f} → {a['max_rss_bytes']/2**20:.1f}"])
    confirm=results['confirmation']['technical'];engines=m.audit.read(m.ROOT/'engines.json')
    pages=[[p['page'],p['path'],p['defects'],p['detected']] for p in confirm['pages']['after']]
    freeze=m.audit.read(m.ROOT/'annotation-freeze.json')
    text=f'''# Verb scaffolding: three exposed gains, no new-page gain

The updated rule adds four warnings on the 88 exposed pages. Three identify
frozen defects; the fourth identifies an additional wording issue absent from
that review. On six new complete pages, full detection stays at **0/43**.
This result does not establish broad useful recall.

## Runtime

`filler.instruction-scaffolding` version 3 recognizes two bounded constructions:
a nominal action or gerund nested between enables/allows and a passive infinitive,
and a nominal action performed through a gerund method. It distinguishes an
action from a concrete object: configuration files being uploaded is a control.
The guidance retains actors, method, conditions and optionality. Ordinary
permissions, simple passives, quotes, questions, protected vocabulary, negation
and conditional actions remain excluded. Candidates stay within 48 tokens.

The [protocol](protocol.md) includes hypotheses that remain unimplemented,
including broader capability chains and role definitions. The catalog still has
53 rules and class manifest r8. Default weights, thresholds and gates are unchanged.
The three gained defects occur in old SQLite and Prometheus documentation; this
step supplies no measured increase on Ptah.

## Frozen source review

The selector excluded all 88 exposed references and hashes and the original
historical study. The assistant froze sources, license notices and selection before
reading prose, then reviewed all six complete extracted pages across all seven
rubric categories. The assistant froze labels at `{freeze['frozen_at']}` before
runtime edits or diagnostic output: 43 defects, 8 uncertain events, 18 controls.
The reviewer is the implementing Codex assistant, accepted under
[ADR 0041](../../../docs/adr/0041-assistant-review-acceptance.md). These are
assistant development judgments, not human labels, independent qualification,
authorship evidence or population recall. One page per cohort/length cell does
not support an informative within-cell confidence interval.

{table(['Set','Pages','Frozen defects','Before full','After full'],counts)}

Profiles have identical full-event counts. Existing partial credits remain
partial. The CSS styling method warning and an existing In-order-to warning on
the new Raft page receive `additional_actionable` dispositions, with explicit
post-diagnostic edits. Neither increases frozen recall or its actionable fraction.
The validator rejects attaching frozen credit to these rows. Updating the advice
to preserve actors changes 24 existing findings without changing their locations
or coverage; each disposition is retained and explicitly reviewed.

## New pages

{table(['Page','Original path','Defects','Detected'],pages)}

Ptah remains at 0/25 and the historical pages at 0/18. Technical emits 41 findings:
39 nonactionable, one uncertain and one additional actionable phrase. Strict emits
48: 46 nonactionable, one uncertain and the same phrase. Length or punctuation
that overlaps a wording defect is not credit for detecting it. The
[miss ledger](CONFIRMATION-MISSES.md) preserves all 43 defects. All six pages pass
both gates. That result does not establish that the pages need no editing.

The new-page result remains negative. More isolated grammatical constructions
have not generalized to the main Ptah problems: repeated justifications, evaluative
tails and indirect explanations. The working 80% recall / 85% soft-precision goals
remain unfulfilled. #299 and the broader repetition work in #305 remain open.
These pages are now exposed; another semantic iteration needs fresh confirmation,
not rewritten labels or lowered gates.

## Applicability and reproduction

The existing #312 repeated-claim budget abstention remains on exposed instruction
page c05; its separate PR is outside this branch. The old CMake text-mode limit
also remains visible. No other abstention or operational error appears. Technical
passes every set; strict fails only the preexisting forbidden worth-noting phrase
in the exposed instruction set. No new gate failure is introduced.

Before runtime: `{engines['before']}`. After runtime: `{engines['after']}`.
Before reports for exposed pages retain the prior run and hashes. New-page before
and all after scans use explicit binaries. Reports retain source bytes, source
ranges, policy and runtime identities, commands, host, time and peak RSS.

{table(['Set','Profile','Seconds before → after','Peak MiB before → after'],costs)}

These local scans overlapped ordinary test work and are not a controlled speed
comparison or 2-vCPU performance qualification. Run `python3 tools/render.py` and
`python3 tools/test_evidence.py` here to validate and regenerate the evidence.
`tools/measure.py` replays an explicit binary and set into a new output directory.
No model calls or downloads are needed. Inputs, annotations and unsupported
hypotheses stay frozen.
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
    _,_,events=sets['confirmation'];lines=['# Remaining confirmation defects','','All 43 remain missed in both profiles.','']
    for eid in confirm['phases']['after']['missed']:
        e=events[eid];lines+=['## '+eid+': '+e['category'],'',e['rationale'],'','Proposed edit: '+e['proposed_edit'],'']
        for span in e['targets']:lines += [f"Bytes {span['start']}–{span['end']}",'','```text',span['quote'],'```','']
    (m.ROOT/'CONFIRMATION-MISSES.md').write_text('\n'.join(lines))


if __name__=='__main__':
    main()
