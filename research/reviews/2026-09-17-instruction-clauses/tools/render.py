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


def display_rationale(text):
    # The JSON retains exact review wording. Avoid repeating general advice in
    # each displayed occurrence; quote a discussed modifier as a literal.
    for repeated in (
        ' The revised guidance explicitly retains actors; the matched construction and event coverage are unchanged.',
        ' The changed advice preserves prerequisites and distinguishes capability from obligation; source coverage and this judgment remain unchanged.',
    ):
        text = text.replace(repeated, '')
    return text.replace('Very adds emphasis', '`Very` adds emphasis').replace(
        'Very minimal gives', '`Very minimal` gives')


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
    text=f'''# Instruction clauses: precision repair without new-page recall gain

On the six fresh complete pages, full frozen-event detection remains **4/58**
in both profiles: **1/19** in Ptah and **3/39** in historical documentation.
Three false positives about concrete capabilities disappear. Previously exposed
pages gain four full detections and two partial detections, but lose two useful
full detections and one additional wording edit. The net exposed full-event gain
is two. Neither the 80% recall nor the 85% soft-precision objective is met.

## Runtime changes and limits

`filler.instruction-scaffolding` version 6 recognizes first-person tutorial setup
with a stated prerequisite and nested need/make-sure layers. It separates the
narrator from the action while retaining prerequisite, actor and ordering.
An actual condition, another actor's state, permission, uncertainty or a quoted
instruction is not enough to match.

Named method/function subjects establish an operation role for used-to or
used-for complements. Optional support adverbs, gerund complements, queried
whether/if values and nested intended-to-allow actions are handled within the
existing bounded projection. Standalone allows-user capability descriptions
now require another support layer or an explicit adjacent method announcement.
The latter restriction removes false positives, but also suppresses real
redundancy when the indirect layer occurs later in the operand or the following
method uses a different construction. Those losses remain in the measurements.

These are bounded surface relations, not full dependency analysis or semantic
entailment. There are still 53 rules and class manifest r8. Thresholds, weights
and gates did not change. Blackbox API and compiled CLI tests cover prerequisites,
actors, permission, uncertainty, named operations, guarded clauses, same- and
cross-paragraph methods, Unicode, BOM, CRLF, comments, strings, policy exemptions,
occurrence allowances, cancellation and resource abstention.

## Frozen review

The selector excluded all 118 exposed source identities and hashes and the
original historical study. Protocol, input bytes, identities and notices were
frozen before prose review. The implementing assistant read 412 extracted prose
blocks and table cells, then froze 58 defects, eight uncertainties and 34 controls
at `{freeze['frozen_at']}` before runtime changes or diagnostic output.

The implementing Codex assistant is the reviewer accepted under
[ADR 0041](../../../docs/adr/0041-assistant-review-acceptance.md). These are
assistant development judgments, not human labels or independent qualification.
One page per cohort/length cell supports observed counts, not population recall.
All 124 identities are now exposed. No matcher or frozen label changed after
opening confirmation output.

{table(['Set','Pages','Frozen defects','Before full','After full'],counts)}

The exposed full gains are Query's indirect lookup, Load's repeated action,
the pull-image narrator setup and the task-wait narrator setup. The tracing-
purpose and repeated based-off constructions receive only partial credit.
The lost full detections are Curl's referrer wrapper and its transfer-speed
setup; its time-condition wrapper loses an additional post-diagnostic repair.
A Curl interface-lookup wrapper is a new additional repair without frozen credit.
The Prometheus proxy explanation loses a false positive. Older restricted
repetition/framing scopes remain explicit and are not pooled into all-warning
precision. [CHANGES.md](CHANGES.md) retains all changed and removed findings.

## New complete pages

{table(['Page','Original path','Defects','Detected'],pages)}

Technical findings fall from 36 to 33; strict falls from 39 to 36. Both profiles
retain four full events and one partial event. There are three additional useful
repairs found only after diagnostics, eight unresolved judgments, and 17
nonactionable findings in technical (20 in strict). The three removed warnings
are frozen controls: shared TypeVar relationships, overlay/snapshot support,
and CRIU-dependent migration. Their removal improves precision, not recall.

The full-or-partial frozen-label actionable fraction is 5/33 in technical and
5/36 in strict. Even counting the three additional repairs gives only 8/33 and
8/36; neither is evidence for the 85% soft-precision objective. The fragment
in-order-to does not cover perform-the-following-steps, and long-sentence or
contrast-density overlap does not diagnose the frozen wordiness underneath.
All six pages pass both gates. A pass is not proof of satisfactory wording.

The [miss ledger](CONFIRMATION-MISSES.md) retains all 54 events without full
coverage, including one partial event: 14 empty framing, 23 wordiness, six
unjustified intensifier, nine vague claim and two needless repetition events.
No formulaic-transition or needless-complexity defect was frozen in this sample.
The remaining work includes narrated desire and setup, nominal action wrappers,
metaphorical paraphrases, repeated page promises and ungrounded benefit claims.
Another small construction extension cannot by itself establish broad recall.

## Reproduction and validation

Before runtime: `{engines['before']}`. After runtime: `{engines['after']}`.
The after binary was built from the clean task checkout immediately after the
semantic freeze commit. Exposed before reports retain the preceding iteration;
all after sets were replayed. The new before reports use the pinned before
binary. JSON records original bytes, source mappings, policy/engine identities,
resource costs and the unchanged #312 budget abstention on exposed instruction
page c05. No other abstention or operational error is accepted by the validator.

Technical passes all sets. Strict fails only the preexisting forbidden
worth-noting phrase in the exposed instruction set. The historical CMake
text-mode limitation remains inherited. This work is not deployed to the
playground; local evidence does not establish remote CI or merged-commit status.

{table(['Set','Profile','Seconds before → after','Peak MiB before → after'],costs)}

Measurements overlapped ordinary checks and are not a controlled performance
comparison or 2-vCPU qualification. Run `python3 tools/render.py` and
`python3 tools/test_evidence.py` to validate and regenerate artifacts.
`tools/measure.py` replays an explicit binary into a new output directory.
See [VALIDATION.md](VALIDATION.md) for actual product-check results.
'''
    (m.ROOT/'README.md').write_text(text)
    reviews=m.audit.read(m.ROOT/'dispositions.json');lines=['# Reviewed changes','','Indices refer to the retained strict reports. Replaced findings link to their after review; exact before and after judgments remain in dispositions.json.','']
    for name in sets:
        _,files,_=sets[name]
        after_report=m.audit.read(m.ROOT/'reports'/name/'after/strict.json.gz')
        replacements={m.prior.key(after_report['findings'][r['index']]):r['index'] for r in reviews[name]['strict']['added']}
        for kind,phase in [('added','after'),('removed','before')]:
            rows=reviews[name]['strict'][kind]
            if not rows:continue
            lines+=['## '+name+' / '+kind,''];r=m.audit.read(m.ROOT/'reports'/name/phase/'strict.json.gz')
            for row in rows:
                f=r['findings'][row['index']]
                if kind=='removed' and m.prior.key(f) in replacements:
                    lines += [f"### {row['index']}: {f['rule_id']}",'',f"Replaced at the same source location by after finding {replacements[m.prior.key(f)]}, reviewed above. The new advice preserves prerequisites, actors and capability semantics.",'']
                    continue
                lines += [f"### {row['index']}: {f['rule_id']}",'',display_rationale(row['rationale']),'',
                          'Status: '+row['status']+'. Full: '+(', '.join(row['events']) or 'none')+'. Partial: '+(', '.join(row.get('partial_events',[])) or 'none')+'.','']
                if row.get('proposed_edit'): lines += ['Additional repair: '+row['proposed_edit'],'']
                for loc in m.audit.locations(f):
                    span=loc['span'];lines += [f"`{loc['path']}` bytes {span['start']}–{span['end']}",'','```text','\n'.join(line.rstrip() for line in files[loc['path']][span['start']:span['end']].decode().splitlines()),'```','']
    (m.ROOT/'CHANGES.md').write_text('\n'.join(lines))
    _,_,events=sets['confirmation'];lines=['# Remaining confirmation defects','',f"All {len(confirm['phases']['after']['missed'])} lack full coverage in both profiles. No partial event receives full credit.",'']
    for eid in confirm['phases']['after']['missed']:
        e=events[eid];lines+=['## '+eid+': '+e['category'],'',display_rationale(e['rationale']),'','Proposed edit: '+e['proposed_edit'],'']
        for span in e['targets']:lines += [f"Bytes {span['start']}–{span['end']}",'','```text','\n'.join(line.rstrip() for line in span['quote'].splitlines()),'```','']
    (m.ROOT/'CONFIRMATION-MISSES.md').write_text('\n'.join(lines))


if __name__=='__main__':
    main()
