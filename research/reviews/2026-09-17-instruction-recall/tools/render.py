#!/usr/bin/env python3
"""Generate tables and source-bound misses from validated instruction evidence."""
import sys
sys.dont_write_bytecode = True
import summarize as m


def pct(value):
    return f'{value*100:.1f}%'


def cell(value):
    return str(value).replace('|', '&#124;').replace('\n', ' ')


def table(headers, rows):
    return '\n'.join(['| '+' | '.join(headers)+' |', '| '+' | '.join(['---']*len(headers))+' |'] +
                     ['| '+' | '.join(cell(v) for v in row)+' |' for row in rows])


def main():
    result = m.evaluate()
    m.audit.write(m.ROOT/'summary.json', result)
    sets = m.data_sets()
    names = {'development': 'Original development', 'exposed_repetition': 'Exposed repetition targets',
             'exposed_framing': 'Exposed framing targets', 'exposed_local': 'Exposed local pages',
             'exposed_context': 'Exposed context pages', 'confirmation': 'New complete-page confirmation'}
    overall = []
    for split, profiles in result.items():
        before, after = (profiles['technical']['phases'][p] for p in ('before', 'after'))
        overall.append([names[split], len(sets[split][0]), before['defects'],
                        f"{before['detected']} ({pct(before['recall'])})", f"{after['detected']} ({pct(after['recall'])})"])
    confirm = result['confirmation']['technical']
    burden = []
    for split in ('exposed_context', 'confirmation'):
        for profile, data in result[split].items():
            before, after = (data['phases'][p] for p in ('before', 'after'))
            counts = after['dispositions']
            burden.append([names[split], profile, f"{before['findings']} → {after['findings']}",
                           counts.get('actionable', 0), counts.get('uncertain', 0), counts.get('nonactionable', 0)])
    bypage = [[p['page']+' `'+p['path']+'`', p['cohort'], p['defects'], p['detected']]
              for p in confirm['pages']['after']]
    groups = []
    for field in ('cohort', 'length_stratum'):
        for group in sorted({p[field] for p in confirm['pages']['after']}):
            phases = [[p for p in confirm['pages'][phase] if p[field] == group] for phase in ('before', 'after')]
            groups.append([field, group, sum(p['defects'] for p in phases[0]),
                           sum(p['detected'] for p in phases[0]), sum(p['detected'] for p in phases[1])])
    categories = [[category, counts['defects'], counts['detected']]
                  for category, counts in confirm['phases']['after']['categories'].items()]
    costs = []
    for split, profiles in result.items():
        for profile, data in profiles.items():
            b, a = (data['costs'][phase] for phase in ('before', 'after'))
            costs.append([names[split], profile, f"{b['wall_seconds']:.3f} → {a['wall_seconds']:.3f}",
                          f"{b['max_rss_bytes']/2**20:.1f} → {a['max_rss_bytes']/2**20:.1f}"])
    sensitivity = confirm['sensitivity']
    engines = m.audit.read(m.ROOT/'engines.json')
    text = f'''# Instruction wording: limited transfer, low overall recall

This iteration continues #309 and #299 with two experimental rules:
`filler.instruction-scaffolding` and `repetition.redundant-predicate`.
On twelve new complete pages, full event detections increase from 4/99 to
12/99 in both profiles. All eight gains occur on one Mermaid page. Ptah
stays at 2/35. The broad-recall objective remains unmet.

The implementing Codex assistant reviewed the source prose under
[ADR 0041](../../../docs/adr/0041-assistant-review-acceptance.md).
These are maintainer-accepted editorial judgments, not human annotations,
independent agreement, authorship labels or population accuracy. Another
reviewer is not a prerequisite for continuing implementation.

## What changed

The first rule identifies bounded indirect capability instructions, reader-goal
prefaces followed by instructions, and adjacent method announcements. It keeps
source ranges for both clauses. The second identifies a path/location subject
with a redundant location predicate, or a reason subject with a redundant
because predicate. The suggested edits retain capability and optionality.
They do not convert a possibility into an obligation.

Negation, permission, failure, conditional applicability, quoted claims and
protected construction tokens prevent matches. Protected operands remain
opaque. The implementation uses existing NLP, clause windows, source maps,
budgets, cancellation and configuration. Both rules are experimental warnings
with `gate: none`; profile thresholds and existing rule weights are unchanged.
The catalog has 53 rules. Class manifest r8 retains r7 and adds two general-style
rules; neither has LLM-specific empirical qualification.

## Frozen reviews and separate denominators

Input identities were frozen at 14:25 UTC on September 17, 2026. Full-page
annotations were frozen at 14:38 UTC, before runtime edits or inspection of
confirmation diagnostics. The review recorded 99 defects, 17 uncertain events
and 43 acceptable controls across all seven categories.

The historical selection frame expands to the pinned long-prose selection
metadata; it excludes all 38 previously selected historical sources. The
Ptah frame remains at its existing snapshot. The deterministic selection
excludes all 64 previously reviewed pages and exact source hashes. Historical
selection admits one page per repository. The manifest preserves candidates,
selection order, source identities and licenses. This does not establish
repository-independent transfer: the new Mermaid page shares a documentation
family with the exposed sequence-diagram page.

{table(['Set', 'Pages', 'Frozen defects', 'Before', 'After'], overall)}

Keep these denominators separate. The repetition and framing reviews cover
specific targets; the other sets have complete-page annotations. Every new
and removed diagnostic is reviewed. Retained new-confirmation diagnostics
are reviewed too. Earlier full reviews are inherited by source identity.

On exposed context pages, six more events receive full credit and three
receive partial credit. One additional warning concerns an unlabelled note
instruction and remains uncertain. On new confirmation, the old wordy-phrase
warning covers only part of c09-d09; the new construction covers it fully.
Partial matches never increase full-event recall. [CHANGES.md](CHANGES.md)
records all additions and their source ranges. Nothing is removed.

## Review burden and applicability

{table(['Set', 'Profile', 'Findings before → after', 'Actionable after', 'Uncertain', 'Nonactionable'], burden)}

Actionable counts include partial findings; event counts deduplicate findings
for the same defect. Unlabelled plausible edits remain uncertain and receive
no primary recall credit. Nonactionable judgments distinguish necessary
technical conditions, reference consistency and incidental length/contrast
warnings from the actual frozen defect. This table is a review-burden measure
for these pages, not a universal false-positive rate.

The new rule adds eight actionable findings on one confirmation page and none
on the other eleven. That concentrated 8/8 result cannot qualify its default
precision or support a broad transfer claim. The redundant-predicate rule has
an exposed positive but no new-confirmation positive.

One existing rule, `repetition.repeated-claim`, exhausts its candidate budget
on c05, the 11,882-word Ptah migration reference. The same abstention occurs
in both engines and both profiles. It remains in every report and in the
summary; the page and its missed defects remain in the recall denominator.
There are no other abstentions or operational errors. Reports are complete
under the current optional-rule budget contract; completeness does not mean
every rule evaluated every page. A future budget repair needs separate evidence.

Technical passes on every set. Strict already fails on confirmation in the
baseline because curl's worth-noting phrase has `gate: forbid`; that result
is unchanged. Other sets pass. The new warnings do not create a gate failure.

{table(['Confirmation page', 'Cohort', 'Defects', 'Detected after'], bypage)}

{table(['Grouping', 'Group', 'Defects', 'Before', 'After'], groups)}

{table(['Category', 'Confirmation defects', 'Detected after'], categories)}

[CONFIRMATION-MISSES.md](CONFIRMATION-MISSES.md) retains all 87 missed events,
source quotes, proposed edits and context references. Broad wordiness,
indirect method statements, repeated explanations and vague claims remain
open work. The 80% recall / 85% soft-diagnostic precision working target is
not met. Lowering a gate would not identify those constructions.

Paired page bootstrap uses 2,000 draws within cohort/length cells and seed
917304. After-recall sensitivity is {pct(sensitivity['after'][0])}–{pct(sensitivity['after'][1])};
paired-gain sensitivity is {pct(sensitivity['delta'][0])}–{pct(sensitivity['delta'][1])}.
These describe this small selected sample, not population confidence intervals.
Shared repositories and templates, genre selection and one assistant reviewer
limit the interpretation. No inference about who wrote a text is made.

## Evaluation corrections

The inherited semantic-credit table lacked `filler.announced-importance`.
This iteration explicitly maps it to `empty_framing` for both engines: the
exact frozen c08-d14 phrase and its diagnostic match. No source annotation or
runtime changes. A regression test rejects unrelated repetition credit.

The inherited replay validator also assumed zero abstentions. The new validator
retains all source, hash, policy, exit and completeness checks and requires
the exact observed c05 budget abstention in both phases. Negative probes reject
its deletion, duplication or appearance in an older set. This records a product
limitation instead of treating missing analysis as a clean result.

## Replay and resource records

Baseline runtime: `{engines['before']}`.
Measured implementation: `{engines['after']}`.
Earlier before reports are byte-identical copies of the preceding after reports.
The confirmation baseline and every after report were measured in this iteration.
Subsequent fixture, evidence and documentation changes do not change runtime.

{table(['Set', 'Profile', 'Wall seconds before → after', 'Peak MiB before → after'], costs)}

These are single macOS arm64 runs with `CGO_ENABLED=0`, not a speedup claim or
the separate two-vCPU Linux performance target. Records include CPU time, host,
binary hashes, configurations and commands. Authorized research reports include
source text; ordinary saved reports still omit it by default.

From the repository root:

```sh
python3 research/reviews/2026-09-17-instruction-recall/tools/render.py
python3 research/reviews/2026-09-17-instruction-recall/tools/test_evidence.py
```

To replay a scan, build the named revision and run `tools/measure.py` with
`--binary PATH --set SET --output NEW_DIR`. It requires a new directory and
uses the retained input archive without downloads. The renderer validates
freeze hashes, source quotes, all diagnostic dispositions and event credits,
then regenerates this page, the change ledger, misses and summary.

The protocol and both freeze manifests remain unchanged. These twelve pages
are now exposed development material; further tuning requires a separate
confirmation sample. #309 remains open for unsupported indirect method and
narrative-tail constructions. #299 remains open for broad useful recall.
'''
    (m.ROOT/'README.md').write_text(text)
    reviews = m.audit.read(m.ROOT/'dispositions.json')
    changes = ['# Source-bound additions', '', 'Technical and strict add the same source-bound constructions. The strict ledger below',
               'uses zero-based report indices; all profile-specific hashes remain in dispositions.json.', '']
    for split in ('exposed_context', 'confirmation'):
        pages, files, _ = sets[split]
        report = m.audit.read(m.ROOT/'reports'/split/'after/strict.json.gz')
        changes += ['## '+names[split], '']
        for row in reviews[split]['strict']['added']:
            finding = report['findings'][row['index']]
            changes += [f"### {row['index']}: `{finding['rule_id']}`", '', row['rationale'], '',
                        'Full credit: '+(', '.join(row['events']) or 'none')+'. Partial credit: '+
                        (', '.join(row.get('partial_events', [])) or 'none')+'.', '']
            for location in m.audit.locations(finding):
                span = location['span']; quote = files[location['path']][span['start']:span['end']].decode()
                changes += [f"`{location['path']}` bytes {span['start']}–{span['end']}:", '', '```text', quote, '```', '']
    (m.ROOT/'CHANGES.md').write_text('\n'.join(changes))
    _, _, events = sets['confirmation']
    misses = ['# Frozen confirmation defects still missed', '',
              'Both profiles miss the same 87 events. IDs and byte ranges refer to the frozen',
              'confirmation archive; full context and proposed edits remain in annotations.json.', '']
    for eid in confirm['phases']['after']['missed']:
        event = events[eid]
        # The frozen rationale names a source quotation. Mark that quotation
        # explicitly in the rendered prose without changing the stored label.
        rationale = event['rationale'].replace('of course very interested', '`of course very interested`')
        misses += [f"## {eid}: {event['category']}", '', rationale, '',
                   'Proposed edit: '+event['proposed_edit'], '']
        for target in event['targets']:
            misses += [f"Bytes {target['start']}–{target['end']}:", '', '```text', target['quote'], '```', '']
    (m.ROOT/'CONFIRMATION-MISSES.md').write_text('\n'.join(misses))


if __name__ == '__main__':
    main()
