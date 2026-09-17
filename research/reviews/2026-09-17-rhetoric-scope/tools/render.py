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
    text=f'''# Clause scope: four exposed gains, one new-page gain

This candidate repairs two confirmed instruction false positives and detects
four more frozen defects on the 106 exposed pages. On six separately frozen
complete pages, full detection rises from **5/72 to 6/72** in both profiles.
Ptah rises from **1/29 to 2/29**; the historical pages stay at **4/43**. The working
80% recall and 85% soft-precision objectives are not met. This is a bounded
correctness improvement, not evidence that the broad detection problem is solved.

## What changed

`filler.evaluative-closure` version 7 scopes attribution through the candidate
clause. An operational report after a colon no longer suppresses a preceding
value announcement. An attribution before that colon still protects reported
speech. Quoted judgments remain excluded; a quoted noun phrase can be the
subject of an unquoted reading-value judgment.

Information judgments now recognize discourse subjects declaring importance,
cognitive-worth predicates with bounded manner/frequency qualifiers and abstract
whole-value announcements. Two adjacent evaluative predicates can share one
source-bound diagnostic. Concrete purposes, costs, measured qualifiers and
conditional judgments remain controls. These are bounded surface constructions;
they do not infer semantic equivalence or dependency edges.

`filler.instruction-scaffolding` version 5 requires a nominal actor and excludes
restrictive relative subjects. This removes the imperative changelog entry
preceded by an issue link and the HTML upload-form definition reported in PR319.
The tagger's adjective label for a noun such as library is accepted only with a
nominal determiner. Existing protected identifier actors remain opaque.

The rule count remains 53 and the class manifest remains r8. Weights, gates and
thresholds are unchanged. Blackbox API and CLI tests cover reported speech,
quoted information, technical controls, original byte ranges, Unicode, CRLF,
Markdown emphasis, Go comments, Python strings, term exemptions and occurrence
policy. Existing cancellation and resource tests remain in the required checks.

## Frozen review

The selector excluded all 106 exposed references and hashes and the original
historical study. Protocol, source identities, notices and bytes were frozen
before prose review. The implementing assistant read all 1,162 extracted prose
blocks and table cells, then froze 72 defects, seven uncertainties and 24 controls
at `{freeze['frozen_at']}` before runtime edits or diagnostic output. The Ptah
home page uses native MDX extraction and has no definite frozen defect.

The reviewer is the implementing Codex assistant accepted under
[ADR 0041](../../../docs/adr/0041-assistant-review-acceptance.md). These are
assistant development judgments, not human labels or independent qualification.
One page per cohort/length cell supports counts, not population estimates.
No semantic tuning followed inspection of confirmation output. All 112 source
identities are now exposed and must be excluded from future fresh confirmation.

{table(['Set','Pages','Frozen defects','Before full','After full'],counts)}

The exposed gains are the whole-value announcement in the provider guide, the
combined whole-guarantee and worth-stating announcement, quoted reading advice,
and the importance declaration in the component reference. All four are Ptah
framing events. One replaces a previously partial diagnostic. No frozen full
detection is lost. Both removed instruction findings were recorded false positives.
The generic feature introductions that remained uncertain in PR319 are unchanged.

## New complete pages

{table(['Page','Original path','Defects','Detected'],pages)}

The only new-page gain is the whole-criterion announcement in the Atlas command
reference. Technical emits 68 findings: six frozen actionable findings, one
additional actionable wording edit absent from frozen labels, four unresolved
judgments and 57 nonactionable findings. Strict emits 75 with the same six,
one and four, plus 64 nonactionable findings. All six pages pass both gates.

The one additional wording edit receives no frozen recall credit. The four
unresolved length judgments are retained rather than counted as successful
wording detection. Most other length and punctuation findings identify concrete
technical conditions or overlap wording defects without identifying their repair.
The unchanged test-suite sentence-opener warning repeats a necessary named subject.

The [miss ledger](CONFIRMATION-MISSES.md) records all 66 missed events. Of the
72 frozen defects, only 2/21 empty-framing and 4/27 wordiness events are detected;
none of the eight needless repetitions, eleven unjustified intensifiers or five
vague claims is detected. There are no frozen formulaic-transition or needless-
complexity defects in this sample. Fixing scope alone has little effect on broad
recall. More narrow templates must not be presented as completion of that goal.

[CHANGES.md](CHANGES.md) records every added, replaced and removed diagnostic.
`dispositions.json` also reviews all retained confirmation findings. Original
restricted scopes for the older repetition/framing studies remain explicit.

## Reproduction and limits

Before runtime: `{engines['before']}`. After runtime: `{engines['after']}`.
The after binary was built from a clean detached checkout of the runtime commit.
An earlier dirty-checkout build was retained only in temporary scratch space and
was replaced by a full replay of the clean binary before acceptance of reports.
Exposed before reports retain the previous run; every after set was replayed.
JSON preserves source bytes, spans, engine and policy identities and resource
records. No source or frozen annotation was revised after output review.

The existing #312 repeated-claim budget abstention remains on exposed instruction
page c05. No new abstention or operational error appears. Technical passes all
sets; strict fails only the preexisting forbidden worth-noting phrase in the
exposed instruction set. The old CMake text-mode limitation remains inherited.

{table(['Set','Profile','Seconds before → after','Peak MiB before → after'],costs)}

Measurements overlapped ordinary checks; they are not a controlled performance
comparison or 2-vCPU qualification. Run `python3 tools/render.py` and
`python3 tools/test_evidence.py` here to validate and regenerate these results.
`tools/measure.py` replays an explicit binary into a new output directory.
See [VALIDATION.md](VALIDATION.md) for product checks and their actual status.
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
