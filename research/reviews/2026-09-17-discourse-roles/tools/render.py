#!/usr/bin/env python3
"""Render validated full-page counts, changed findings and retained misses."""
import sys
sys.dont_write_bytecode = True
import summarize as m


def table(headers, rows):
    def cell(v):return str(v).replace('|','&#124;').replace('\n',' ')
    return '\n'.join(['| '+' | '.join(headers)+' |','| '+' | '.join(['---']*len(headers))+' |']+
                     ['| '+' | '.join(cell(v) for v in r)+' |' for r in rows])


def main():
    results=m.evaluate();m.audit.write(m.ROOT/'summary.json',results);sets=m.data_sets()
    counts=[];costs=[]
    for name,profiles in results.items():
        b,a=(profiles['technical']['phases'][p] for p in ('before','after'))
        counts.append([name,len(sets[name][0]),a['defects'],b['detected'],a['detected']])
        for profile,data in profiles.items():
            b,a=(data['costs'][p] for p in ('before','after'))
            costs.append([name,profile,f"{b['wall_seconds']:.3f} → {a['wall_seconds']:.3f}",
                          f"{b['max_rss_bytes']/2**20:.1f} → {a['max_rss_bytes']/2**20:.1f}"])
    confirm=results['confirmation']['technical'];engines=m.audit.read(m.ROOT/'engines.json')
    pages=[[p['page'],p['path'],p['defects'],p['detected']] for p in confirm['pages']['after']]
    text=f'''# Discourse roles: development gains, confirmation regression

The new six-page result falls from **2/44 to 1/44** full frozen-event detections
in both profiles. The only new warning is a false positive on the timeout advice
for databases slow to accept connections. The stricter actor check removes a
valid configure-options diagnostic. This candidate has **not** met the 80% recall
or 85% soft-precision objectives and is not ready for acceptance.

On previously exposed pages, 13 new findings identify 11 full frozen events,
one additional wording repair and one false positive. Five removed findings
comprise two known false positives, one unresolved judgment, one full detection
and one partial detection. The result fixes the literal row-counts and checksum
errors, but its losses must remain visible. Development gains do not establish
quality on new prose.

## Runtime changes

The existing evaluative-closure rule, version 8, recognizes information subjects
with selecting importance or attention predicates, author-to-reader notices,
document-outcome endorsements and gerund subjects with abstract functional
clefts. It keeps a following independent clause outside a local judgment while
retaining attached conditions. Bare noun/counts compounds are excluded.

Unscoped-assurance version 5 admits ordinary copular quality and difficulty
predicates, with local condition, measurement and mechanism exclusions.
Instruction-scaffolding version 8 restricts relative support to operation or
actor antecedents. These are bounded surface roles over existing POS tokens,
not dependency edges or semantic entailment. No rule count, threshold, weight,
gate, model or class changed: the catalog remains 53 rules, class manifest r8.

The new negative evidence identifies specific limits:

- The quality matcher accepts a relative restriction inside an instruction as
  if it were the whole sentence's quality claim: the timeout recommendation is
  a concrete symptom/action relationship.
- A clock invocation comparison loses the surrounding coarse-resolution
  tradeoff and precise-timer opt-out. A missing numeric benchmark alone does
  not make that technical comparison an editorial defect.
- The relative actor filter loses a named utility function followed by a
  protected identifier, and a configure-options support description. A supplied
  operand and a named operation require different role checks.
- Most frozen wordiness, reader framing, qualitative claims and discourse-level
  repetition remain undiagnosed. More matches on familiar clauses do not meet
  the complete-page objective.

No runtime adjustment was made after opening these confirmation results. Keep
this candidate in draft until later work addresses those failures with separate
evidence. The archived negative result must not be rewritten as a success.

## Frozen review

Before source reading, the selector froze one short, medium and long page per
cohort, excluding 130 prior references and hashes. The implementing Codex
assistant reviewed all 792 extracted blocks and table cells, then froze 44
defects, seven uncertainties and 28 controls across all seven categories.
This follows [ADR 0041](../../../docs/adr/0041-assistant-review-acceptance.md):
maintainer-accepted assistant review, not human labels or independent review.

The source-only overlap audit found no complete extracted block of at least
12 whitespace tokens in earlier sources after whitespace normalization. It
compared all 130 previous sources. Shorter, partial and paraphrased reuse can
escape that audit; no known overlap is not proof of independent provenance.
All 136 selected source identities are now exposed. These six-page counts do
not estimate population recall, and one page per cohort/length cell cannot
support a meaningful within-cell bootstrap interval.

{table(['Set','Pages','Frozen defects','Before full','After full'],counts)}

The older repetition and framing studies retain their narrower original review
scope. Their counts do not support precision claims over all warnings. Every
changed finding is reviewed separately in [CHANGES.md](CHANGES.md), including
new findings outside those earlier scopes.

## New complete pages

{table(['Page','Original path','Defects','Detected'],pages)}

Ptah remains **1/19**: only the duplicated article is fully detected. Historical
pages fall from **1/25 to 0/25**. The agent diagnostic-code reference has no
required edits and no findings. Passing it is appropriate.

Technical has 27 findings and strict has 29, both before and after. After review,
one finding fully detects a frozen event and another partially detects a
purpose/mandatory wrapper. Five additional repairs were identified only after
diagnostics; they cannot increase frozen recall. Nine findings are unresolved.
The remaining 11 technical and 13 strict findings are nonactionable.

Thus the fraction providing a full frozen-event diagnosis is 1/27 technical and
1/29 strict. Including the partial and additional repairs yields 7/27 and 7/29,
still below 85%. These denominators count findings, not independent documents.
[CONFIRMATION-MISSES.md](CONFIRMATION-MISSES.md) retains all 43 events without
full coverage. All six pages pass both gates; no gate was lowered to force a
failure, and passing does not establish editorial quality.

## Reproduction and checks

Before runtime: `{engines['before']}`. After runtime: `{engines['after']}`.
The latter binary was built from the clean semantic-freeze commit before
opening confirmation diagnostics. Exposed before reports are retained from the
preceding iteration; the fresh before run uses its pinned binary. The new runtime
replayed all 16 sets in both profiles with unchanged input and policy identities.
The known #312 repeated-claim budget abstention on exposed instruction c05 is
unchanged. No additional abstention, incomplete scan or operational error is
accepted by the evidence validator.

{table(['Set','Profile','Seconds before → after','Peak MiB before → after'],costs)}

These host observations are not a controlled performance comparison or a
2-vCPU qualification. Earlier before measurements overlapped other checks.
Run `python3 tools/render.py` and `python3 tools/test_evidence.py` to verify
and regenerate the report. `tools/measure.py` replays an explicit binary into
a new directory. [VALIDATION.md](VALIDATION.md) records actual product checks.
This experiment has not been merged or deployed to the playground.
'''
    (m.ROOT/'README.md').write_text(text)
    reviews=m.audit.read(m.ROOT/'dispositions.json');lines=['# Reviewed changes','','Indices refer to retained strict reports. Exact before and after judgments remain in dispositions.json.','']
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
                if row.get('proposed_edit'):lines+=['Additional repair: '+row['proposed_edit'],'']
                for loc in m.audit.locations(f):
                    z=loc['span'];lines += [f"`{loc['path']}` bytes {z['start']}–{z['end']}",'','```text','\n'.join(line.rstrip() for line in files[loc['path']][z['start']:z['end']].decode().splitlines()),'```','']
    (m.ROOT/'CHANGES.md').write_text('\n'.join(lines))
    _,_,events=sets['confirmation'];lines=['# Remaining confirmation defects','',f"All {len(confirm['phases']['after']['missed'])} events lack full coverage in both profiles. Partial coverage does not count as full.",'']
    for eid in confirm['phases']['after']['missed']:
        e=events[eid];lines+=['## '+eid+': '+e['category'],'',e['rationale'],'','Proposed edit: '+e['proposed_edit'],'']
        for z in e['targets']:lines += [f"Bytes {z['start']}–{z['end']}",'','```text','\n'.join(line.rstrip() for line in z['quote'].splitlines()),'```','']
    (m.ROOT/'CONFIRMATION-MISSES.md').write_text('\n'.join(lines))


if __name__=='__main__':
    main()
