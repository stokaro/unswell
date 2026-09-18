#!/usr/bin/env python3
"""Render a rejected candidate's retained measurements and review decisions."""
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
    pages=confirmation['technical']['pages']['after']
    decision=dict(status='rejected_for_runtime_promotion',baseline_commit=engines['before'],candidate_commit=engines['after'],
                  confirmation_defects=a['defects'],confirmation_detected_before=confirmation['technical']['phases']['before']['detected'],
                  confirmation_detected_after=a['detected'],reason='No new full frozen-defect detections on confirmation; the only new finding remains uncertain.',
                  runtime_policy='Restore the baseline runtime. Retain the candidate commit and source archive for reproduction; do not tune it on these confirmation outputs.')
    m.audit.write(m.ROOT/'decision.json',decision)
    dispositions=[]
    for profile,p in confirmation.items():
        for phase,d in p['phases'].items():
            c=d['dispositions'];dispositions.append([profile,phase,d['findings'],c.get('actionable',0),c.get('additional_actionable',0),c.get('uncertain',0),c.get('nonactionable',0)])
    body=f'''# Discourse constructions: candidate rejected

The candidate adds seven full detections on 841 previously exposed defect records,
but none on the twelve new pages: **2/106 before and 2/106 after** (1.9%).
Both full detections are in the historical Dear ImGui FAQ: a wordy causal phrase
and an accidentally repeated preposition. Ptah remains **0/31** on these six pages.
The only new confirmation warning concerns `Unfortunately` in kafka-go's
motivations section, which was already labeled uncertain before diagnostics.

The candidate is **not promoted into the runtime**. The final change restores
production code, tests and the generated rule catalog to merged main. It retains
this negative experiment and its missed-defect inventory. The goal of reliable
whole-page detection remains open; neither a low index nor a passing gate proves
that these pages are clean.

## Evidence and boundaries

The same Codex assistant selected the sample, reviewed sources, implemented the
candidate and reviewed diagnostics. This is maintainer-accepted assistant
agreement under ADR0041, not independent human annotation or population accuracy.
All seven rubric categories were reviewed before runtime changes or diagnostic
output. The frozen labels contain 106 defects, 26 uncertain events and 62 controls.

The final sample has two short, three medium and one long page per cohort.
Only one unused eligible long page remained per cohort; the metadata-only
selection amendment preceded source reading and is retained in
[source-selection-notes.md](source-selection-notes.md). All 148 earlier source
identities were excluded. The exact-block overlap check found no matches of at
least twelve whitespace tokens; it does not rule out shorter or paraphrased reuse.
Historical dates are source provenance, not proof of human authorship.

All 19 sets were replayed with fresh CLI processes for both profiles and both
revisions. Runs are complete, with no operational errors, skipped rules or
abstentions. The two oldest targeted sets have narrower review scope; they do not
establish precision for all their findings. These are observed counts without a
population interval; there is only one long page per cohort.

{table(['Set','Pages','Frozen defects','Baseline full','Candidate full'],rows)}

## New-page review

{table(['Page','Stored source','Defects','Fully detected'],[[p['page'],p['path'],p['defects'],p['detected']] for p in pages])}

Every new-page finding in both phases and profiles has a retained disposition.
Three findings address frozen defects: two fully and one partially. Four other
small wording repairs were found after diagnostics; they cannot increase recall.
Sentence length, punctuation density and overlapping contrast markers receive no
credit for unrelated frozen defects. Controls and uncertainty were not relabeled
to make the candidate look better.

{table(['Profile','Phase','Findings','Frozen actionable','Additional repairs','Uncertain','Nonactionable'],dispositions)}

The candidate introduces ten warnings on exposed sets: seven address frozen
misses and three identify additional tone edits. No old findings are removed.
On confirmation it adds one uncertain warning. The earlier development gains
do not justify a claim of improved transfer or the 80% recall / 85% soft-precision
targets. [CHANGES.md](CHANGES.md) records every changed warning;
[MISSES.md](MISSES.md) retains all {len(missed)} remaining new-page defects.

## Why the candidate was rejected

The candidate recognizes narrated tutorial actions, comma-delimited reader
inference and evaluative cues, and more copular quality modifiers. It still
requires narrow local surface structures. The new defects mostly concern longer
wording dependencies, unsupported scope, repeated explanation and indirect
claims. A finite verb mislabeled as a noun can suppress a valid candidate, while
a generic regret cue does not distinguish an unnecessary technical endorsement
from a relevant author reaction in a motivations section.

The miss inventory also contains a concrete separate boundary bug: the FAQ's
`when when` is missed in a sentence containing the quoted name `ImGui`, although
the duplicate is outside the quotation. This needs local quote scope, with
positive and negative source-mapping tests in [#329](https://github.com/stokaro/unswell/issues/329).
The post-diagnostic [minimal probe](quote-boundary-probe.json) records the three
cases and their complete CLI outcomes; it does not add frozen recall credit.

The next experiment should test substantially broader relations or resolve
measurable extraction/role failures. Extending another short cue list without
showing transfer would repeat this result. The new confirmation pages are now
exposed and may be used for development, not reused as fresh confirmation.

## Reproduction and validation

Baseline runtime: `{engines['before']}`.
Candidate runtime: `{engines['after']}`.
The candidate binary was built from the clean candidate commit before the
confirmation reports were opened. `code-freeze.json` and `runtime-inputs.tar.gz`
retain its runtime and focused tests. The final checkout intentionally contains
the baseline runtime; rebuilding HEAD does not reproduce the candidate.

Build the specified revision with `CGO_ENABLED=0`, `-trimpath`, and
`-ldflags '-X github.com/stokaro/unswell.BuildCommit=<full revision>'`.
Then run from this study directory:

```sh
python3 tools/measure.py --binary /path/to/unswell --set confirmation --output /new/output
python3 tools/test_evidence.py
python3 tools/render.py
```

The candidate's focused public API tests and strict lint passed. Its broader
root test run found stale rule-version expectations (`10` versus candidate `11`)
in existing tests; those failures are retained in `validation.json`. No candidate
release is claimed. Final repository checks run against the restored runtime and
research artifacts. Race, active fuzzing and coverage remain deferred to #123.
The fifteen evidence tests reject source drift, missing judgments, invented
credit, unexpected abstention and changes to the frozen runtime archive.

The following costs cover complete local CLI scans, including extraction,
analysis and reporting. They are not a qualification of the separate
100,000-word production target. Each run records host and binary identities.

{table(['Set','Profile','Seconds baseline → candidate','Peak MiB baseline → candidate'],costs)}
'''
    (m.ROOT/'README.md').write_text(body)
    lines=['# Reviewed diagnostic changes','','Indices refer to retained strict reports. Both profiles have complete saved dispositions.','']
    for name in result:
        for kind,phase in [('added','after'),('removed','before')]:
            for row in review[name]['strict'][kind]:
                report=m.audit.read(m.ROOT/'reports'/name/phase/'strict.json.gz');f=report['findings'][row['index']]
                lines += [f"## {name}: {kind} {row['index']} ({f['rule_id']})",'',f"Status: {row['status']}. Full frozen credit: {', '.join(row['events']) or 'none'}.",'',row['rationale'],'']
                for loc in m.audit.locations(f):
                    raw=sets[name][1][loc['path']];span=loc['span']
                    lines += [f"{loc['path']} bytes {span['start']}–{span['end']}:",'','> '+' '.join(raw[span['start']:span['end']].decode().split()),'']
    (m.ROOT/'CHANGES.md').write_text('\n'.join(lines).rstrip()+'\n')
    lines=['# Remaining new-page defects','',f'These {len(missed)} frozen events are not fully diagnosed. One has partial coverage. No origin label or overlapping warning substitutes for the stated repair.','']
    for eid in missed:
        e=sets['confirmation'][2][eid];lines += [f"## {eid}: {e['category']}",'',e['rationale'],'',e['proposed_edit'],'']
        for t in e['targets']:lines += ['> '+' '.join(t['quote'].split()),'']
    (m.ROOT/'MISSES.md').write_text('\n'.join(lines).rstrip()+'\n')


if __name__=='__main__':
    main()
