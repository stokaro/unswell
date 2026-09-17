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
    text=f'''# Instruction projection: limited gains and false matches

This candidate adds five fully detected frozen defects on the 100 exposed pages.
On six separately frozen complete pages, full detection rises from **2/34 to 3/34**
in both profiles. Ptah remains **1/16**; the historical pages rise from **1/18 to
2/18**. Two additional Ptah defects remain only partially detected. The working
80% recall and 85% soft-precision objectives are not met. Keep this candidate in
review; this result does not justify claiming useful broad coverage or immediate
promotion of the generic-reader recognizer.

## Implementation

`filler.instruction-scaffolding` version 4 follows bounded support predicates to
an action and operand. It recognizes nominal ability, generic reader enablement,
modal used-to actions, nested intended-to-enable passive actions and nominal
actions carried out with a method. The original source tokens retain actors,
operands, optionality and conditions. An affirmative ability statement must not
become an obligation or an assertion that an action actually happened.

A capability announcement and its immediate anaphoric method can form one event
across adjacent paragraphs. Both source ranges are retained. Headings, lists,
excluded code, unrelated sentences and compound announcements break the relation.
Independent comments and strings remain separate. Other events retain their
block scope. Table cells remain outside the rule's declared contexts; the table
entry c06-d04 therefore remains a miss, despite the constructed clause regression.
The broader prefix/relative-information hypothesis c06-d05 also remains unmet.

The new projection uses POS and infinitive roles, not dependency parsing or
semantic equivalence. Its bounds are 48 candidate tokens and 96 sentence tokens.
There are still 53 rules and class manifest r8. Weights, thresholds and gates are
unchanged. The old generic users control is now an explicit positive construction;
named actor permission and concrete capability limits remain negative controls.
Tests exercise source mapping, opaque operands, optionality, conditions,
exemptions, occurrence policy, cancellation and structural boundaries.

## Frozen review and source identity

The selector excluded all 100 exposed references and hashes and the original
historical study. Source bytes, selection, notices and protocol were frozen before
reading. The assistant read every extracted prose block and table cell across the
seven editorial categories, then froze labels at `{freeze['frozen_at']}` before
runtime edits or diagnostics: 34 defects, eight uncertainties and 24 controls.

The implementing Codex assistant is the maintainer-accepted reviewer under
[ADR 0041](../../../docs/adr/0041-assistant-review-acceptance.md). These are assistant
development judgments, not human labels or independent qualification. One page
per cohort/length cell supports counts rather than population estimates. No
semantic tuning followed inspection of confirmation output. These 106 identities
are now exposed and cannot serve as fresh confirmation for the next iteration.

{table(['Set','Pages','Frozen defects','Before full','After full'],counts)}

The five exposed gains are a referrer instruction, cookie reuse, transfer-speed
setup, optional field metadata and a two-paragraph Gantt configuration instruction.
The last previously had two partial diagnostics; relating both ranges earns full
event credit without counting the instruction twice. No other frozen event gains
or loses full credit. All five gains are on historical pages.

## Review burden and false matches

Across the exposed pages, the 14 added/replaced strict findings comprise five
frozen actionable findings, four additional actionable wording edits absent from
the frozen labels, three uncertain judgments and two nonactionable findings.
Two old partial Gantt diagnostics are replaced by one complete diagnostic.
Additional post-diagnostic edits receive no frozen recall credit. The older
repetition/framing sets retain their original restricted annotation scopes;
changed findings outside those scopes are still reviewed separately.

The candidate has two false matches. It takes issue-link residue as the subject
of an imperative Allow user changelog entry. It also treats a relative clause
defining an HTML upload form as a reader instruction. Generic feature introductions in the ImGui and alert-rule guides
remain uncertain, as does a previously frozen highlight.js example. These cases
show that the actor boundary and genre/context decision need further work. The
source-bound judgments and four additional repairs are in [CHANGES.md](CHANGES.md).
Do not silently remove these findings or revise frozen labels to improve results.

## Separate confirmation

{table(['Page','Original path','Defects','Detected'],pages)}

Technical emits 17 findings: five actionable (three complete and two partial)
and 12 nonactionable. Strict emits 18 with the same five actionable and 13
nonactionable. The only added finding identifies the nominal ability wrapper in
the Prometheus template reference. All six pages pass both gates.

The [miss ledger](CONFIRMATION-MISSES.md) preserves 31 events without full coverage,
including two partial detections. A length warning overlapping a wording defect
does not earn credit. The adjacent committed tokens have different grammatical
roles; deleting one would break the sentence. That existing repeated-word warning
is recorded as nonactionable, not as detection of the frozen circular explanation.

## Applicability and reproduction

The existing #312 repeated-claim budget abstention remains on exposed instruction
page c05. The old CMake text-mode limit remains explicit. No new abstention or
operational error appears. Technical passes every set; strict fails only the
preexisting forbidden worth-noting phrase in the exposed instruction set.

Before runtime: `{engines['before']}`. After runtime: `{engines['after']}`.
Exposed before reports retain the previous run. New-page before reports and all
after reports use explicit binaries. Reports retain source bytes and ranges,
engine and policy identities, commands, host, wall time and peak RSS.

{table(['Set','Profile','Seconds before → after','Peak MiB before → after'],costs)}

These measurements overlapped ordinary checks and are not a controlled performance
comparison or 2-vCPU qualification. Run `python3 tools/render.py` and
`python3 tools/test_evidence.py` here to validate and regenerate the report and
negative evidence probes. `tools/measure.py` replays an explicit binary and set
into a new output directory. Runtime analysis needs no model call or download.
See [VALIDATION.md](VALIDATION.md) for implementation checks and their limits.
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
                          'Status: '+row['status']+'. Full: '+(', '.join(row['events']) or 'none')+'. Partial: '+(', '.join(row.get('partial_events',[])) or 'none')+'.','']
                if row.get('proposed_edit'): lines += ['Additional repair: '+row['proposed_edit'],'']
                for loc in m.audit.locations(f):
                    span=loc['span'];lines += [f"`{loc['path']}` bytes {span['start']}–{span['end']}",'','```text',files[loc['path']][span['start']:span['end']].decode(),'```','']
    (m.ROOT/'CHANGES.md').write_text('\n'.join(lines))
    _,_,events=sets['confirmation'];lines=['# Remaining confirmation defects','','All 31 lack full coverage in both profiles; two have partial findings.','']
    for eid in confirm['phases']['after']['missed']:
        e=events[eid];lines+=['## '+eid+': '+e['category'],'',e['rationale'],'','Proposed edit: '+e['proposed_edit'],'']
        for span in e['targets']:lines += [f"Bytes {span['start']}–{span['end']}",'','```text',span['quote'],'```','']
    (m.ROOT/'CONFIRMATION-MISSES.md').write_text('\n'.join(lines))


if __name__=='__main__':
    main()
