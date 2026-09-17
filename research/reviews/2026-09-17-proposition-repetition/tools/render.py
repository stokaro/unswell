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
    text=f'''# Bounded proposition repetition: three exposed gains, no new-page gain

This candidate detects three more frozen defects on previously exposed pages.
On six new complete pages, full detection stays at **0/46** in both profiles:
**0/14** in Ptah and **0/32** in historical documentation. The working 80% recall
and 85% soft-precision objectives remain unmet. The new result is evidence that
these bounded comparisons do not address the broader missed constructions.

## Runtime changes

`repetition.repeated-claim` version 2 compares an explicit adjacent reformulation:
an actor performs an action only on objects with a property, followed by the
same actor never performing that action on objects lacking that property.
The second sentence must start with `That is,` or `In other words,`. Actor,
action, tense, property and object must match. A separate bounded dependency-
licensing form preserves the owner and direct-and-transitive scope. It does
not equate different actors, modalities, quantities, conditions or protected
identifiers. This is a surface construction, not general semantic entailment.

`repetition.definition-echo` version 3 recognizes identical gerund actions and
an explicitly bounded both-objects shorthand across a copula.
`repetition.explanatory-restart` version 2 can include an abstract because clause
that repeats the same initial adjective instead of supplying a concrete cause.
A concrete cause is retained outside the warning. Both assertions, and all
three circular judgments when present, have original source spans.

There are still 53 rules and class manifest r8. Thresholds, weights and gates
are unchanged. Blackbox API and compiled CLI tests cover matching restrictions,
distinct actors, scope, tense, properties, conditions, quotes, hidden link
destinations, protected code, Markdown emphasis, Unicode, BOM, CRLF, Go comments,
Python strings, configuration, cancellation and the shared work budget.

## Frozen review and measured changes

The selector excluded all 112 exposed source identities and hashes and the
original historical study. Protocol, sources and notices were frozen before
prose review. The implementing assistant read 310 extracted prose blocks and
table cells, then froze 46 defects, seven uncertainties and 29 controls at
`{freeze['frozen_at']}` before runtime changes or diagnostic output.

The reviewer is the implementing Codex assistant accepted under
[ADR 0041](../../../docs/adr/0041-assistant-review-acceptance.md). These are
assistant development judgments, not human labels or independent qualification.
One page per cohort/length cell supports observed counts, not population recall.
All 118 source identities are now exposed and must be excluded from the next
fresh confirmation. No semantic matcher or frozen label changed after output.

{table(['Set','Pages','Frozen defects','Before full','After full'],counts)}

The three gains are the circular complex/reason/complex explanation, the
permissive-licenses/never-nonpermissive restatement, and typing both flags/still
typing both. The first replaces a partial diagnostic. No full detection is lost.
The Ptah glossary benefit paraphrase in #305 remains unsupported; that issue
must stay open. [CHANGES.md](CHANGES.md) retains all added and removed findings.
Older restricted repetition/framing review scopes remain explicit.

## New complete pages

{table(['Page','Original path','Defects','Detected'],pages)}

Technical emits 14 findings: one additional actionable wording edit absent from
frozen labels, two unresolved length judgments and 11 nonactionable findings.
Strict adds one nonactionable dash-density finding. The additional edit receives
no frozen recall credit. The instruction-scaffolding finding in the Prometheus
proxy explanation is a false positive against a frozen technical control.
All six pages pass both gates; that does not establish satisfactory detection.

The [miss ledger](CONFIRMATION-MISSES.md) records all 46 missed events: 16 empty
framing, 19 wordiness, eight unjustified intensifier, two vague claim and one
needless repetition events. There are no frozen formulaic-transition or
needless-complexity defects in this sample. Indirect action descriptions and
narrated tutorial steps recur among the misses. Another small equivalence
template alone is unlikely to deliver the required broad recall.

## Resource correction after the first replay

The first runtime commit `b963a42da231240e718b3e3f965ed9f51e546f5b`
charged both full sentences before checking whether a reformulation marker was
present. That added budget abstentions on `exposed_framing` page c06 and
`exposed_scope` page c03. A regression test reproduced the problem before the
fix. The corrected loop charges the bounded marker probe for every candidate
and the full comparison only for marked pairs. It keeps the existing budget
and does not change semantic matching or resource limits.

`prototype-reports` preserves the first replay. The evidence validator compares
all findings, source documents and gates byte-for-byte against the corrected
replay, in every set and both profiles. They are identical; only the two new
abstentions disappear. The existing #312 repeated-claim abstention on exposed
instruction page c05 remains explicit. No other abstention or error is accepted.

## Reproduction and limits

Before runtime: `{engines['before']}`. After runtime: `{engines['after']}`.
The after binary was built from a clean detached checkout. Exposed before
reports retain the preceding iteration; all after sets were replayed. New-page
before reports were generated with the pinned before binary. JSON records
source bytes, source mappings, engine and policy identities and resource costs.

Technical passes all sets. Strict fails only the preexisting forbidden
worth-noting phrase in the exposed instruction set. The historical CMake
text-mode limitation remains inherited. These changes are not published to the
playground and remote CI is not treated as verified by these local reports.

{table(['Set','Profile','Seconds before → after','Peak MiB before → after'],costs)}

Measurements overlapped ordinary checks; they are not a controlled performance
comparison or 2-vCPU qualification. Run `python3 tools/render.py` and
`python3 tools/test_evidence.py` here to validate and regenerate the artifacts.
`tools/measure.py` replays an explicit binary into a new output directory.
See [VALIDATION.md](VALIDATION.md) for product checks and actual status.
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
    _,_,events=sets['confirmation'];lines=['# Remaining confirmation defects','',f"All {len(confirm['phases']['after']['missed'])} lack full coverage in both profiles. No partial event receives full credit.",'']
    for eid in confirm['phases']['after']['missed']:
        e=events[eid];lines+=['## '+eid+': '+e['category'],'',e['rationale'],'','Proposed edit: '+e['proposed_edit'],'']
        for span in e['targets']:lines += [f"Bytes {span['start']}–{span['end']}",'','```text',span['quote'],'```','']
    (m.ROOT/'CONFIRMATION-MISSES.md').write_text('\n'.join(lines))


if __name__=='__main__':
    main()
