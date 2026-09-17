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
    text=f'''# Clause-boundary repairs and new confirmation losses

The six new pages fall from **4/56 to 2/56** full frozen-event detections in
both profiles. The stricter noun-phrase check loses two useful simplicity
judgments. This candidate is **not ready for acceptance** and does not meet
the 80% recall or 85% soft-precision targets.

The previously exposed sample gains three full detections: a named parsing
function, configure options, and a process-is-simple judgment after a goal.
It loses no exposed full detection and removes two established false positives:
timeout advice and the clock comparison fragment beginning with `and`.
These repairs do not compensate for the fresh confirmation regression.

## Changes and limitations

Instruction-scaffolding version 9 resolves a protected identifier after an
explicit function/method head and recognizes configuration-option antecedents.
The checksum, token, filename and label operand exclusions remain covered.
Protected text stays opaque and never supplies a role keyword.

Unscoped-assurance version 6 requires a main subject and rejects relative
restrictions inside instructions. A goal prefix can precede a separate main
quality assertion without making that assertion scoped. Source spans isolate
the assertion and preserve the goal. The catalog remains 53 rules, manifest r8;
no gate, weight, score threshold, model, dependency or public API changed.

The clock false positive was a simpler structural error than the initial
paragraph-scope hypothesis: `and` was accepted as a subject. Excluding that
fragment fixes this instance without introducing broad paragraph suppression.
This change does not claim general cross-sentence mechanism resolution.

The new losses expose an overbroad determiner-sequence restriction:
`The way the --type flag functions is simple` has a nested nominal construction,
and `Setting up a configuration file is simple` has a particle before its
object. Both are valid subjects for a quality judgment. No runtime tuning was
made after these confirmation outputs were opened. Preserve this negative result
and add both losses to the next development regression set.

## Frozen review

The source selector excluded all 136 previous references and hashes. One short,
medium and long page per cohort were selected before prose review. The same
implementing Codex assistant read all 477 extracted blocks/table cells and
froze 56 defects, six uncertainties and 26 technical controls across all seven
categories before runtime edits. This is maintainer-accepted assistant review
under [ADR 0041](../../../docs/adr/0041-assistant-review-acceptance.md), not
human or independent annotation.

A source-only audit found no complete extracted block of at least 12 whitespace
tokens in previous sources after whitespace normalization. Shorter, partial and
paraphrased reuse can escape that audit. All 142 selected source identities are
now exposed. One page per cohort/length cell does not support a within-cell
bootstrap or a population recall estimate.

{table(['Set','Pages','Frozen defects','Before full','After full'],counts)}

The older repetition and framing reviews retain their narrower scope; they do
not establish precision over every warning. [CHANGES.md](CHANGES.md) contains
every added and removed diagnostic, including lost detections.

## New complete pages

{table(['Page','Original path','Defects','Detected'],pages)}

Ptah stays at **0/20**. Historical pages fall from **4/36 to 2/36**. The short
Ptah Assist page has no required edits and no findings, so passing it is
appropriate. The other missed events remain visible in
[CONFIRMATION-MISSES.md](CONFIRMATION-MISSES.md): 54 lack full coverage.

Technical findings fall from 39 to 37; strict findings fall from 42 to 40.
After review, two findings fully diagnose frozen events, two offer additional
post-diagnostic phrase repairs, and 14 remain uncertain. The remaining 19
technical and 22 strict findings are nonactionable on this review. Full-event
diagnostic fractions are 2/37 and 2/40; including additional repairs gives 4/37
and 4/40. Those repairs cannot increase frozen recall.

Technical passes all six pages. Strict fails the ripgrep page on the existing
announced-importance phrase; all three Ptah pages still pass. No threshold was
changed to force a failure. A gate result is not a completeness measurement.

## Reproduction

Before runtime: `{engines['before']}`. After runtime: `{engines['after']}`.
The after binary was built from the clean semantic-freeze commit. Exposed
before reports are retained from the previous iteration, and the fresh before
run uses its pinned binary. All 17 sets were replayed in both profiles with
unchanged sources and policy. The existing #312 budget abstention on exposed
instruction c05 remains explicit; the validator rejects any other abstention,
error or skipped rule.

{table(['Set','Profile','Seconds before → after','Peak MiB before → after'],costs)}

These are host observations, not a controlled comparison or 2-vCPU performance
qualification. Ordinary tests overlapped part of the measurement. Run
`python3 tools/render.py` and `python3 tools/test_evidence.py` to validate and
regenerate the report. `tools/measure.py` takes an explicit CLI binary and a
new output directory. [VALIDATION.md](VALIDATION.md) records software checks.
This candidate has not been merged or deployed to the playground.
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
