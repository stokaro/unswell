#!/usr/bin/env python3
"""Render the local-repetition review from validated reports and dispositions."""
from pathlib import Path
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
    lines = ['# Local repetition: measured improvement and remaining misses', '',
        'This follow-up implements #304 and a limited construction related to #305.',
        'It adds adjacent-word checks, short repeated assertions and explanatory',
        'restarts. The full-page objective remains unmet. The new confirmation pages',
        'gain no findings or event detections. Do not describe these changes as good',
        'general recall or as qualified defaults.', '',
        'The reviewer is the same Codex assistant that implemented the rules, accepted',
        'under [ADR 0041](../../../docs/adr/0041-assistant-review-acceptance.md).',
        'These are subjective development judgments with source-bound proposed edits.',
        'They are not independent human ratings, model calibration or population',
        'accuracy. Neither native-speaker review nor another rater blocks this work.', '',
        '## Complete-page and targeted results', '',
        'Both profiles have the same event counts. Keep the four denominators',
        'separate: the original and new reviews cover all seven rubric categories;',
        'the two intervening confirmation sets have narrower annotation scopes.', '',
        '| Set | Pages | Frozen defects | Before | After |',
        '| --- | ---: | ---: | ---: | ---: |']
    names = dict(development='Original development', exposed_repetition='Exposed repetition targets',
                 exposed_framing='Exposed framing targets', confirmation='New complete-page confirmation')
    for split, profiles in result.items():
        r = profiles['technical']
        b, a = r['phases']['before'], r['phases']['after']
        lines.append(f"| {names[split]} | {len(r['pages']['after'])} | {a['defects']} | {b['detected']} ({pct(b['recall'])}) | {a['detected']} ({pct(a['recall'])}) |")
    lines += ['', 'The original exit-1 explanation is recovered. Three accidental word',
        'duplications are detected in the exposed NATS and ripgrep pages. The',
        'ripgrep complexity explanation receives a useful restart warning, but its',
        'circular final rationale remains unresolved. That partial match does not',
        'receive full event credit. The glossary paraphrase and permissive-license',
        'restatement also remain missed; #305 stays open.', '',
        'The new sample has 12 complete pages: six Ptah and six historical sources,',
        'two per cohort/length cell. All source prose was reviewed before diagnostic',
        'output: 78 defects, 10 uncertain judgments and 20 acceptable controls.',
        'Selection excludes the preceding 41 pages. Source hashes and retained',
        'notices are in [confirmation/manifest.json](confirmation/manifest.json).',
        'The frozen labels are in [confirmation/annotations.json](confirmation/annotations.json).', '',
        'The single detected confirmation event is a repeated document-scope',
        'announcement in BuildKit. The existing metadiscourse warning locates both',
        'announcements. A single scope announcement does not get repetition credit.',
        'A length warning overlapping a different wording defect gets no credit.', '',
        '## Findings and review burden', '',
        'Every confirmation finding was reviewed, including unchanged findings.',
        'An optional rewrite is uncertain, not automatically a true positive.', '',
        '| Set | Profile | Findings before → after | Actionable after | Uncertain | Nonactionable |',
        '| --- | --- | ---: | ---: | ---: | ---: |']
    for split in ('development', 'confirmation'):
        for profile, r in result[split].items():
            b, a = r['phases']['before'], r['phases']['after']; d = a['dispositions']
            lines.append(f"| {names[split]} | {profile} | {b['findings']} → {a['findings']} | {d.get('actionable',0)} | {d.get('uncertain',0)} | {d.get('nonactionable',0)} |")
    lines += ['', 'The new rules add five diagnostics across the exposed sets and remove none.',
        'Four identify complete frozen defects; one identifies part of the circular',
        'explanation. The fresh sample supplies no new-rule positive cases, so it',
        'cannot establish their recall or positive predictive value. Its unchanged',
        'warnings still impose substantial review burden.', '',
        'All scans pass the existing gate. Thresholds and weights of earlier rules',
        'are unchanged. Gate PASS is compatible with missed wording defects.', '',
        'The [complete diagnostic ledger](dispositions.json) preserves judgments and',
        'report hashes. [CHANGES.md](CHANGES.md) lists every added warning.',
        '[CONFIRMATION-MISSES.md](CONFIRMATION-MISSES.md) retains every missed event.', '',
        '## Uncertainty and limits', '',
        'The proposed 80% recall / 85% soft-diagnostic precision target is not met.',
        'A paired page bootstrap resamples within cohort/length cells with seed',
        '917304 and 2,000 draws. It is sensitivity to this small selected page set,',
        'not an interval for all technical writing. Related Ptah pages share one',
        'repository; the single reviewer, fixed source frame, and subjective labels',
        'limit transfer. Changelog entries contribute many vague-claim labels, so',
        'the page and category breakdowns must accompany the aggregate.', '']
    s = result['confirmation']['technical']['sensitivity']
    lines += [f"Confirmation recall sensitivity: {pct(s['after'][0])}–{pct(s['after'][1])}; paired gain: {pct(s['delta'][0])}–{pct(s['delta'][1])}.", '',
        '| Confirmation page | Cohort | Defects | Detected |', '| --- | --- | ---: | ---: |']
    for r in result['confirmation']['technical']['pages']['after']:
        lines.append(f"| {r['page']} {code(r['path'])} | {r['cohort']} | {r['defects']} | {r['detected']} |")
    lines += ['', '| Category | Confirmation defects | Detected |', '| --- | ---: | ---: |']
    for category, r in result['confirmation']['technical']['phases']['after']['categories'].items():
        lines.append(f"| {category} | {r['defects']} | {r['detected']} |")
    lines += ['', '## Runtime and replay identity', '',
        f"Before: `{m.COMMITS['before']}`. Final measured code:",
        f"`{m.COMMITS['after']}`. Builds use `CGO_ENABLED=0`.",
        'The reports include source for authorized offline verification; ordinary',
        'saved reports still omit it by default. The runtime makes no model calls.', '',
        'The [resource amendment](resource-amendment.md) records a correction made',
        'after opening confirmation output. Initial reports remain under',
        '`reports/*/initial`. All findings, assessments, source documents and gates',
        'are identical to the corrected replay. The one initial budget abstention',
        'on the exposed schema-commands page is gone. No final scan abstains.', '',
        '| Set | Profile | Wall seconds before → after | Peak MiB before → after |',
        '| --- | --- | ---: | ---: |']
    for split, profiles in result.items():
        for profile, r in profiles.items():
            b, a = r['costs']['before'], r['costs']['after']
            lines.append(f"| {names[split]} | {profile} | {b['wall_seconds']:.3f} → {a['wall_seconds']:.3f} | {b['max_rss_bytes']/2**20:.1f} → {a['max_rss_bytes']/2**20:.1f} |")
    lines += ['', 'These are single macOS arm64 runs, with CPU time and host recorded in',
        '`runs.json`. They are operational measurements, not evidence of a speedup',
        'or the separate two-vCPU Linux performance target.', '',
        '## Reproduce', '', 'From the repository root:', '', '```sh',
        'python3 research/reviews/2026-09-17-local-repetition/tools/render.py',
        'python3 research/reviews/2026-09-17-local-repetition/tools/test_evidence.py',
        '```', '',
        'The renderer validates source hashes, complete reviews, configuration and',
        'code identity, all diagnostic deltas, source-bound credits and the resource',
        'correction before producing this page. To repeat a scan, build the named',
        'revision and use `tools/measure.py --binary PATH --set SET --output NEW_DIR`.',
        'The original [protocol](protocol.md) and both freeze manifests are unchanged.', '',
        'Next work must address the remaining semantic restatements and contextual',
        'framing/assurance gaps. These confirmation pages are now exposed development',
        'material; further tuning requires another unexposed confirmation sample.', '']
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
