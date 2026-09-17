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
    text=f'''# Rhetorical relations: three exposed gains, no confirmation gain

The two updated rules add three findings on the 94 exposed pages, all identifying
frozen Ptah defects. On six new complete pages, full detection remains **3/45**
in both profiles. Another defect is only partly identified. These results do not
establish good complete-page recall.

## Implementation and rejected shortcut

`filler.evaluative-closure` version 6 recognizes cognitive-worth modifiers on
information nouns, gerund-anaphor purpose clefts, relative feature-purpose
predicates and discourse subjects declaring importance or editorial integrity.
`filler.unscoped-assurance` version 4 recognizes result tails that predict a
generic reader's preferences or understanding. The implementation uses bounded
source-token relations and the existing POS contract. It does not infer semantic
equality, dependencies or the author's identity.

The new tests also exposed an existing negation bug: a negative nominal-worth
complement could match. It now remains a control. Conditions, real component
purposes, named requirements, measured results, quoted claims and protected
operands have close counterexamples. Source mapping, exemptions and occurrence
policy use the existing engine. Bounds remain 48 candidate tokens and 96 sentence
tokens; defaults retain their weights, thresholds and gates. There are still 53
rules and class manifest r8.

The [scope ablation](scope-ablation.json) disabled two guards only in a temporary
development binary. It added two findings across 94 exposed pages, with no other
changes. That experiment does not justify removing the guards. Their original
behavior is retained. The protocol's broader measured-versus-asserted and
coherence hypotheses are not implemented or demonstrated by this change.

## Frozen review

The selector excluded all 94 exposed source references and hashes and the
original historical study. The assistant froze sources, selection and notices before
reading, then reviewed all six complete extracted pages across the seven rubric
categories. The assistant froze labels at `{freeze['frozen_at']}` before runtime changes
or diagnostic output: 45 defects, five uncertain events and 23 controls.

The implementing Codex assistant is the maintainer-accepted reviewer under
[ADR 0041](../../../docs/adr/0041-assistant-review-acceptance.md). This is assistant
development evidence. Human annotation and independent qualification remain
unperformed. One page per cohort/length cell supports counts; it cannot establish
population recall or a reliable within-cell confidence interval. The selector
retained the short connection guide, which has no definite wording defect. The
review records the unexpanded historical include and excludes its unseen content.

{table(['Set','Pages','Frozen defects','Before full','After full'],counts)}

Both profiles have the same full-event counts. Three new exposed findings cover
an import-purpose restatement, a decision-worth-stating announcement and an
entry-honesty judgment. [Every changed finding](CHANGES.md) has a source-bound
review. No old finding is removed or changed beyond its rule identity.

## Separate confirmation

{table(['Page','Original path','Defects','Detected'],pages)}

Ptah remains at 0/12 and the historical pages at 3/33. Technical emits 35 findings:
29 nonactionable, five actionable and one additional actionable phrase outside
the frozen labels. Strict emits 37: 31 nonactionable, the same five actionable
findings and the same additional phrase. Two of the five actionable findings each
cover only part of one two-block margins instruction; neither gets full credit.
The other three identify existing instruction scaffolding on the Gantt page.
An In-order-to phrase receives no frozen recall credit.

The [miss ledger](CONFIRMATION-MISSES.md) retains all 42 defects without full
coverage, including the partially detected one. Length and punctuation warnings
do not get credit for overlapping a different wording defect. All six pages pass
both gates; that does not establish that they need no editing.

The 80% recall / 85% soft-precision working objectives remain unmet. Successive
bounded-construction additions have produced exposed examples without new-page
improvement. This result argues against treating more such additions as a plan
for broad coverage. General paraphrastic repetition and indirect explanations
remain unresolved in #299 and #305. These 100 pages are now exposed; another
semantic iteration needs a new confirmation set and a different justified
hypothesis. Do not revise these labels or lower gates to obtain a passing result.

## Applicability and reproduction

The existing #312 repeated-claim budget abstention remains on exposed instruction
page c05; its separate fix is outside this branch. The old CMake text-mode limit
also remains visible. No new abstention or operational error appears. Technical
passes every set; strict fails only the preexisting forbidden worth-noting phrase
in the exposed instruction set.

Before runtime: `{engines['before']}`. After runtime: `{engines['after']}`.
Exposed before reports retain the previous run and hashes. The new-page before
run and all after runs use explicit binaries. Reports retain source bytes, source
ranges, runtime/policy identities, commands, host, duration and peak RSS.

{table(['Set','Profile','Seconds before → after','Peak MiB before → after'],costs)}

These scans overlapped ordinary tests and are not a controlled performance
comparison or 2-vCPU qualification. Run `python3 tools/render.py` and
`python3 tools/test_evidence.py` here to validate and regenerate the evidence.
`tools/measure.py` replays a binary and set into a new directory. No model call or
resource download is required. The input and annotation freezes remain unchanged.
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
    _,_,events=sets['confirmation'];lines=['# Remaining confirmation defects','','All 42 lack full coverage in both profiles; one has partial findings.','']
    for eid in confirm['phases']['after']['missed']:
        e=events[eid];lines+=['## '+eid+': '+e['category'],'',e['rationale'],'','Proposed edit: '+e['proposed_edit'],'']
        for span in e['targets']:lines += [f"Bytes {span['start']}–{span['end']}",'','```text',span['quote'],'```','']
    (m.ROOT/'CONFIRMATION-MISSES.md').write_text('\n'.join(lines))


if __name__=='__main__':
    main()
