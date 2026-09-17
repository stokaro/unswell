#!/usr/bin/env python3
"""Render contextual wording results, diagnostic deltas and frozen misses."""
import sys

sys.dont_write_bytecode = True
import summarize as m
ROOT = m.ROOT


def pct(value):
    return 'n/a' if value is None else f'{100*value:.1f}%'


def code(text):
    return '`' + ' '.join(text.replace('`', "'").replace('|', r'\|').split()) + '`'


def note(notes, text):
    if text not in notes:
        notes[text] = 'n'+str(len(notes)+1)
    ident = notes[text]
    return f'[{ident}](#{ident})'


def append_notes(lines, notes):
    lines += ['## Review notes', '']
    for text, ident in notes.items():
        lines += ['### '+ident, '', text, '']


def render():
    result = m.evaluate()
    m.audit.write(ROOT/'summary.json', result)
    lines = ['# Contextual wording: known-case gains without confirmation gain', '',
        'This iteration continues #299 with bounded extensions to three existing',
        'rules. Five added warnings identify six more frozen defects on previously',
        'reviewed pages. The eleven new complete pages gain no detections: only',
        'one of 42 defects is found. The broad-recall objective remains unmet.', '',
        'The implementing Codex assistant reviewed the prose under the maintainer',
        'acceptance in [ADR 0041](../../../docs/adr/0041-assistant-review-acceptance.md).',
        'The labels are subjective, source-bound editorial judgments. They are not',
        'human annotations, independent agreement or population accuracy. Another',
        'reviewer is not a prerequisite for continuing this work.', '',
        '## Separate evaluation sets', '',
        'Both profiles have the same event counts. Keep these denominators separate:',
        'the framing and repetition sets have narrower annotation scopes than the',
        'three complete-page reviews. An overlap with a length warning does not',
        'count when the warning misses the actual wording construction.', '',
        '| Set | Pages | Frozen defects | Before | After |',
        '| --- | ---: | ---: | ---: | ---: |']
    names = dict(development='Original development', exposed_repetition='Exposed repetition targets',
                 exposed_framing='Exposed framing targets', exposed_local='Exposed complete pages',
                 confirmation='New complete-page confirmation')
    for split, profiles in result.items():
        r = profiles['technical']; b, a = r['phases']['before'], r['phases']['after']
        lines.append(f"| {names[split]} | {len(r['pages']['after'])} | {a['defects']} | {b['detected']} ({pct(b['recall'])}) | {a['detected']} ({pct(a['recall'])}) |")
    lines += ['', 'The added warnings cover a page explaining its own explanation, a',
        'cognitive-worth preface, two unrestricted reader claims, and two exaggerated',
        'quality claims grouped into one diagnostic. [CHANGES.md](CHANGES.md) retains',
        'their original source spans, credit and review rationale. No warning is removed.', '',
        'Matcher versions change to `filler.document-justification` v3,',
        '`filler.evaluative-closure` v3 and `filler.unscoped-assurance` v2.',
        'The catalog remains at 51 rules. Weights, gate thresholds, profiles and',
        'rule classes are unchanged. Clause analysis preserves explicit scope,',
        'measurements, quoted claims, negation and protected text.', '',
        '## New confirmation and review burden', '',
        'The assistant froze the source identities and selection method, then read',
        'all source prose. Both steps preceded implementation and any inspection of',
        'diagnostic output. The review recorded 42 defects, 13 uncertain events and',
        '25 acceptable controls.',
        'Selection excludes the 53 preceding sources. The requested historical-long',
        'cell had one unused page left, so the sample contains eleven pages rather',
        'than twelve. This deficit was recorded before reading the selected prose.', '',
        'The sole detected defect is an accidental adjacent-word duplication in the',
        'Mermaid page. The new constructions produce no confirmation positives.',
        'Every confirmation warning, including unchanged ones, has a disposition.', '',
        '| Set | Profile | Findings before → after | Actionable after | Uncertain | Nonactionable |',
        '| --- | --- | ---: | ---: | ---: | ---: |']
    for split in ('development', 'exposed_local', 'confirmation'):
        for profile, r in result[split].items():
            b, a = r['phases']['before'], r['phases']['after']; d = a['dispositions']
            lines.append(f"| {names[split]} | {profile} | {b['findings']} → {a['findings']} | {d.get('actionable',0)} | {d.get('uncertain',0)} | {d.get('nonactionable',0)} |")
    lines += ['', 'All scans pass the existing gate. Those passes coexist with substantial',
        'missed wording problems and review burden. The fresh sample cannot qualify',
        'the new matchers: it contains no new-rule positive diagnostics. Do not turn',
        'the known-case improvement into a general precision or recall claim.', '',
        'The proposed 80% event recall and 85% soft-diagnostic precision target is',
        'not met. [CONFIRMATION-MISSES.md](CONFIRMATION-MISSES.md) lists the remaining',
        'frozen defects, including introductory scaffolding, vague claims, repetition',
        'and wordiness. The [ledger](dispositions.json) binds judgments to report',
        'hashes; [summary.json](summary.json) contains every page/category breakdown.', '',
        '| Confirmation page | Cohort | Defects | Detected |', '| --- | --- | ---: | ---: |']
    for r in result['confirmation']['technical']['pages']['after']:
        lines.append(f"| {r['page']} {code(r['path'])} | {r['cohort']} | {r['defects']} | {r['detected']} |")
    lines += ['', '| Category | Confirmation defects | Detected |', '| --- | ---: | ---: |']
    for category, r in result['confirmation']['technical']['phases']['after']['categories'].items():
        lines.append(f"| {category} | {r['defects']} | {r['detected']} |")
    s=result['confirmation']['technical']['sensitivity']
    lines += ['', 'Paired page bootstrap within cohort/length cells uses seed 917304 and',
        '2,000 draws. It describes sensitivity to this small sample, not a population',
        'confidence interval. Shared repositories, the single reviewer, genre mix',
        'and the exhausted historical-long cell limit transfer.', '',
        f"Confirmation recall sensitivity: {pct(s['after'][0])}–{pct(s['after'][1])}; paired gain: {pct(s['delta'][0])}–{pct(s['delta'][1])}.", '',
        '## All-rule inspection', '',
        'A separate exploratory scan enabled all 51 existing rules on the exposed',
        'sets. On the twelve recently reviewed pages it increased findings from',
        '66 to 166. Passive-voice candidates and readability grades accounted for',
        '87 of the 100 additions. This is not a completed all-rule recall study:',
        'the extra warnings were inspected by rule and context, not exhaustively',
        'credited against every defect. Reports, policy and hashes remain under',
        '`all-rules-probe/`. The experiment does not change shipped defaults.', '',
        '## Replay and resources', '',
        f"Baseline runtime: `{m.COMMITS['before']}`.",
        f"Measured implementation: `{m.COMMITS['after']}`.",
        'The four exposed baseline report sets come from the preceding iteration.',
        'Their original resource records remain intact. This iteration ran the',
        'confirmation baseline and all after scans. No scan abstained or reported',
        'an operational error.', '',
        '| Set | Profile | Wall seconds before → after | Peak MiB before → after |',
        '| --- | --- | ---: | ---: |']
    for split, profiles in result.items():
        for profile, r in profiles.items():
            b, a = r['costs']['before'], r['costs']['after']
            lines.append(f"| {names[split]} | {profile} | {b['wall_seconds']:.3f} → {a['wall_seconds']:.3f} | {b['max_rss_bytes']/2**20:.1f} → {a['max_rss_bytes']/2**20:.1f} |")
    lines += ['', 'These are single macOS arm64 runs, not a speedup claim or the separate',
        'two-vCPU Linux target. Records retain CPU time, host and binary hashes.',
        'Builds use `CGO_ENABLED=0`. Source is included in these authorized offline',
        'research artifacts; ordinary saved reports still omit it by default.', '',
        'From the repository root:', '', '```sh',
        'python3 research/reviews/2026-09-17-context-recall/tools/render.py',
        'python3 research/reviews/2026-09-17-context-recall/tools/test_evidence.py',
        '```', '',
        'The renderer validates source hashes, freeze manifests, complete diagnostic',
        'reviews, code and policy identity, every delta and source-bound event credit.',
        'To repeat a scan, build the named revision and run',
        '`tools/measure.py --binary PATH --set SET --output NEW_DIR` from this directory.',
        'It requires a new output directory and performs no downloads.', '',
        'The [protocol](protocol.md) and both freeze manifests remain unchanged.',
        'This confirmation set is now exposed. Further tuning needs fresh pages',
        'and a larger historical source frame. The negative result supports moving',
        'beyond expansions of narrow surface constructions: semantic restatements',
        'in #305 and broader contextual wordiness still require implementation.', '']
    (ROOT/'README.md').write_text('\n'.join(lines))
    changes = ['# Added and removed diagnostics', '',
        'All added findings are listed below. There are no removals. Partial credit',
        'does not enter event recall. Indices are zero-based within the saved report.',
        'Identical evidence across profiles shares one entry and one review note.', '']
    review = m.audit.read(ROOT/'dispositions.json')
    groups, notes = {}, {}
    for split, profiles in review.items():
        for profile, r in profiles.items():
            report = m.audit.read(ROOT/'reports'/split/'after'/(profile+'.json.gz'))
            for row in r['added']:
                f = report['findings'][row['index']]
                key = (split, m.prior.key(f))
                group = groups.setdefault(key, dict(finding=f, review=row, profiles=[]))
                m.audit.require(group['review']['rationale'] == row['rationale'] and group['review']['events'] == row['events'] and
                                group['review'].get('partial_events', []) == row.get('partial_events', []), 'Profile review differs')
                group['profiles'].append(f"{profile}: {row['index']}")
    for (split, _), group in groups.items():
        f, row = group['finding'], group['review']
        changes += [f"## {split} / {f['rule_id']} / {row['index']}", '', '; '.join(group['profiles']), '',
                    'Review: '+note(notes, row['rationale']), '']
        for loc in m.audit.locations(f):
            span = loc['span']
            changes += [f"- {loc['path']} [{span['start']}, {span['end']}): {code(loc['snippet'])}"]
        changes += ['', 'Full events: '+str(row['events'])+'. Partial events: '+str(row.get('partial_events', []))+'.', '']
    append_notes(changes, notes)
    (ROOT/'CHANGES.md').write_text('\n'.join(changes))
    misses = ['# Remaining confirmation defects', '',
        'These frozen judgments were made before viewing diagnostic output. Both',
        'profiles miss the same events. The entries are development targets after',
        'this measurement; they must not become a reused confirmation set.', '']
    _, _, events = m.data_sets()['confirmation']
    notes = {}
    for eid in result['confirmation']['technical']['phases']['after']['missed']:
        e = events[eid]
        misses += [f"## {eid}: {e['category']}", '', 'Review: '+note(notes, e['rationale']), '',
                   'Proposed edit: '+note(notes, e['proposed_edit']), '']
        for target in e['targets']:
            misses += [f"- [{target['start']}, {target['end']}): {code(target['quote'])}"]
        misses += ['']
    append_notes(misses, notes)
    (ROOT/'CONFIRMATION-MISSES.md').write_text('\n'.join(misses))


if __name__ == '__main__':
    render()
