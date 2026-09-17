#!/usr/bin/env python3
"""Render the retained comparison and missed-event records."""
import sys
sys.dont_write_bytecode = True
import summarize as m


def table(headers, rows):
    return '\n'.join(['| '+' | '.join(headers)+' |', '| '+' | '.join(['---']*len(headers))+' |']+
                     ['| '+' | '.join(str(v).replace('|','\\|').replace('\n',' ') for v in row)+' |' for row in rows])


def main():
    result = m.evaluate()
    m.audit.write(m.ROOT/'summary.json', result)
    sets = m.data_sets()
    review = m.audit.read(m.ROOT/'dispositions.json')
    engines = m.audit.read(m.ROOT/'engines.json')
    rows, costs = [], []
    for name, profiles in result.items():
        b, a = (profiles['technical']['phases'][p] for p in ('before','after'))
        rows.append([name,len(sets[name][0]),a['defects'],b['detected'],a['detected']])
        for profile, data in profiles.items():
            b, a = (data['costs'][p] for p in ('before','after'))
            costs.append([name,profile,f"{b['wall_seconds']:.3f} → {a['wall_seconds']:.3f}",
                          f"{b['max_rss_bytes']/2**20:.1f} → {a['max_rss_bytes']/2**20:.1f}"])
    pages = result['confirmation']['technical']['pages']['after']
    body=f'''# Clause grammar and attention-frame repairs

The previously exposed boundary sample improves from **2/56 to 7/56** fully
diagnosed defects in both profiles. The changes restore two lost simplicity
judgments and diagnose three other frozen defects, including two Ptah attention
frames. No existing finding or full detection disappears across the 18 sets.
One additional capability rewrite was identified after diagnostics in the older
repetition sample; it earns no frozen-event credit.

The six new pages remain at **2/34**, or 5.9%, in both profiles. Ptah contributes
1/17 and the historical sources 1/17. One more Ptah framing event is partially
diagnosed. These results do not meet the 80% recall or 85% soft-precision targets.
This is a bounded repair, with no measured improvement on the new pages.

## What changed

Unscoped-assurance version 7 accepts a gerund with a particle and object and a
way clause with its own finite predicate. The incomplete-goal check applies at
the boundary before a main subject, rather than rejecting every later article.
Relative restrictions inside instructions and subjectless conjunctions remain
controls. The source span excludes the introductory goal.

Instruction-scaffolding version 10 resolves supports/provides/offers/gives plus
an ability/capability and an optional generic reader, including bounded nested
support chains. Specific actors, permission restrictions, negative clauses and
protected operands remain controls. A capability still does not imply a duty.

Evaluative-closure version 9 recognizes impersonal quality/cognitive notices,
reader-attention selections and whole-point judgments about reading an
explanation. Operational verification, ordering requirements and concrete
component purposes remain controls. The existing announced-importance phrase
rule keeps ownership of its exact sentence openings.

The catalog remains 53 rules, manifest r8. Gate thresholds, weights, models,
public APIs and runtime dependencies are unchanged. Focused public-engine tests
and CLI checks preserve source mapping, revisions and technical controls.

## Frozen evidence and limits

Before runtime changes or confirmation diagnostics, the implementing Codex
assistant reviewed all seven editorial categories and froze 34 defects,
12 uncertain cases and 28 technical controls. This is maintainer-accepted
assistant review under [ADR 0041](../../../docs/adr/0041-assistant-review-acceptance.md),
not independent human annotation or a population estimate.

The selector excluded 142 prior source identities. A source-only audit found no
whole normalized block of at least 12 whitespace tokens in those sources;
shorter or paraphrased overlap is not ruled out. All 148 identities are now
exposed. The historical long selection is a CMake build file classified as text
by source metadata. Its size includes code; this set has no ordinary historical
long-prose page. It remains in the denominator rather than being replaced after
inspection. [Source notes](source-review-notes.md) retain this limitation and
the change from a draft-branch baseline to merged main.

Both phases were remeasured for all 18 sets. The merged baseline produces the
same findings as the prior retained results, but its long-reference candidate
budget abstention is gone after the separate budget repair. Every current run
is complete, with no errors, skipped rules or abstentions. There are no within-cell
bootstrap intervals: each new cohort/length cell contains only one source, and
one of those sources is mixed code and prose.

{table(['Set','Pages','Frozen defects','Before full','After full'],rows)}

The old repetition and framing sets retain narrower targeted review scope;
they do not establish precision for all their diagnostics.
[CHANGES.md](CHANGES.md) records every changed finding, and
[MISSES.md](MISSES.md) records all 32 remaining new-page defects.

## New pages and review burden

{table(['Page','Stored source','Defects','Detected'],[[p['page'],p['path'],p['defects'],p['detected']] for p in pages])}

Technical emits 39 findings: three actionable diagnoses (including one partial
event), one additional post-diagnostic wording repair, seven uncertain findings
and 28 nonactionable findings. Strict emits 47: the same three actionable, one
additional, seven uncertain and 36 nonactionable findings. The eight extra
punctuation warnings do not diagnose additional frozen defects. The seven
code-spanning CMake length warnings are retained as false alarms, relevant to
[#307](https://github.com/stokaro/unswell/issues/307).

The preexisting instruction warning on the local-editor workflow is a false
positive: its condition selects a real alternative to browser editing. The new
rules do not repair that case. Most remaining defects require fuller rhetorical
or discourse relations; a matching word, length warning or contrast marker
receives no credit for a different diagnosis. No semantic tuning followed
confirmation output review.

## Reproduction

Before runtime: `{engines['before']}`.
After runtime: `{engines['after']}`.
The after binary was built from the clean semantic-freeze commit.
`code-freeze.json` pins runtime and focused test files; `runtime-inputs.tar.gz`
retains those bytes from the recorded revision so later repository changes do
not invalidate historical report verification. Saved reports include
source bytes, identities, finding spans and per-run resource records.

Run `python3 tools/test_evidence.py` and `python3 tools/render.py` from this
study directory. The tests reject missing findings, source drift, invented
credit, unexpected abstention, runtime changes and credit for unreviewed repairs.
To replay a set, build the recorded revision and run:

```sh
python3 tools/measure.py --binary /path/to/unswell --set confirmation --output /new/output
```

The following figures cover the complete CLI scan of each fixed set, with a
fresh process for each profile. They are local measurements, not evidence for
the separate 100,000-word production target. Host identities and binary hashes
are recorded in `reports/*/*/runs.json`.

{table(['Set','Profile','Seconds before → after','Peak MiB before → after'],costs)}
'''
    (m.ROOT/'README.md').write_text(body)
    lines=['# Reviewed diagnostic changes','','Indices refer to the retained strict reports. Findings and dispositions for both profiles are saved.','']
    for name,data in result.items():
        for kind,phase in [('added','after'),('removed','before')]:
            for row in review[name]['strict'][kind]:
                report=m.audit.read(m.ROOT/'reports'/name/phase/'strict.json.gz');f=report['findings'][row['index']]
                lines += [f"## {name}: {kind} {row['index']} ({f['rule_id']})",'',row['rationale'],'']
                for loc in m.audit.locations(f):
                    raw=sets[name][1][loc['path']];span=loc['span']
                    lines += [f"{loc['path']} bytes {span['start']}–{span['end']}:",'','> '+' '.join(raw[span['start']:span['end']].decode().split()),'']
    (m.ROOT/'CHANGES.md').write_text('\n'.join(lines).rstrip()+'\n')
    lines=['# Remaining new-page defects','','These 32 frozen events are not fully diagnosed. One has partial coverage. No count or origin label substitutes for the stated wording repair.','']
    events=sets['confirmation'][2]
    for eid in result['confirmation']['technical']['phases']['after']['missed']:
        e=events[eid];lines += [f"## {eid}: {e['category']}",'',e['rationale'],'',e['proposed_edit'],'']
        for target in e['targets']:lines += ['> '+' '.join(target['quote'].split()),'']
    (m.ROOT/'MISSES.md').write_text('\n'.join(lines).rstrip()+'\n')


if __name__=='__main__':
    main()
