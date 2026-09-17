#!/usr/bin/env python3
"""Render tables and source-bound deltas from validated saved evidence."""
from collections import defaultdict
import sys
sys.dont_write_bytecode = True
import compare as c

root = c.ROOT
result = c.evaluate()
c.audit.write(root/'results.json',result)
lines = ['# Before and after', '', 'These counts use the same assistant\'s frozen editorial labels. They are not',
         'population recall or authorship measurements.', '',
         '| Set | Profile | Defects | Before | After | Findings before / after | Added / removed |',
         '| --- | --- | ---: | ---: | ---: | ---: | ---: |']
for split,profiles in result.items():
    for profile,r in profiles.items():
        totals = {k:sum(p[k] for p in r['pages']) for k in ('defects','before','after')}
        lines.append(f'| {split} | {profile} | {totals["defects"]} | {totals["before"]} | {totals["after"]} | '
                     f'{r["findings"]["before"]} / {r["findings"]["after"]} | {len(r["delta"]["added"])} / {len(r["delta"]["removed"])} |')
lines += ['', 'Both profiles retain passing gates for all pages. No rule abstentions or',
          'operational errors occurred. The earlier dispositions of retained development',
          'findings remain in force. Confirmation credit evaluates only the three target',
          'families; other existing diagnostics were not assigned new quality labels.', '',
          '| Page | Cohort / selection | Defects | Before | After |', '| --- | --- | ---: | ---: | ---: |']
for split in result.values():
    for p in split['technical']['pages']:
        lines.append(f'| {p["page"]} | {p["cohort"]} / {p["selection"]} | {p["defects"]} | {p["before"]} | {p["after"]} |')
lines += ['', 'Per-page credit is identical between profiles.', '', '## Resource observations', '',
          'One fresh CLI process per cell, on the host recorded in `runs.json`.',
          'These are observations from this replay, not a stable comparative benchmark.',
          'Order was development-after, confirmation-before, confirmation-after,',
          'development-before. Cold start and unrelated host activity are not controlled.', '',
          '| Set / profile | Wall seconds before / after | CPU seconds before / after | Peak RSS MiB before / after |',
          '| --- | ---: | ---: | ---: |']
for split,profiles in result.items():
    for profile,r in profiles.items():
        b,a = r['costs']['before'],r['costs']['after']
        lines.append(f'| {split} / {profile} | {b["wall_seconds"]:.3f} / {a["wall_seconds"]:.3f} | '
                     f'{b["cpu_seconds"]:.3f} / {a["cpu_seconds"]:.3f} | '
                     f'{b["max_rss_bytes"]/2**20:.1f} / {a["max_rss_bytes"]/2**20:.1f} |')
(root/'RESULTS.md').write_text('\n'.join(lines)+'\n')
_,pages,_,_,events = c.audit.load(c.DEVELOPMENT)
report = c.audit.read(root/'reports/development/after/technical.json.gz')
rows = c.audit.read(root/'dispositions.json')['development']['technical']['added']
lines = ['# Added diagnostic review', '', 'Seven source locations are added in each profile, with no removals. All seven',
         'match previously annotated development events. Strict has the same additions.', '',
         'Version and display-message updates to retained findings are not counted as',
         'additions. The comparison rejects concealed policy or evidence changes.', '']
for row in rows:
    f = report['findings'][row['index']]; e = events[row['events'][0]]; p = pages[e['page']]
    lines += ['## '+e['id']+' — '+f['rule_id'], '', '['+p['original_path']+']('+p['reference']+')', '',
              '```text', f['primary']['snippet'], '```', '', row['rationale'], '', 'Proposed edit: '+e['proposed_edit'], '']
lines += ['The MDX configuration excerpt retains its source admonition marker in the',
          'diagnostic span. The learning-promise excerpt includes its leading conjunction.',
          'Both overlap the intended clause; no automatic edit is offered.', '',
          '## Confirmation', '', 'No added or removed findings. The existing honest-answer diagnostic still',
          'receives the single credit, `c02-d02`. The remaining ten target defects stay',
          'missed. Six uncertain cases are excluded from the positive denominator.', '']
(root/'CHANGES.md').write_text('\n'.join(lines)+'\n')
cp,_,ce = c.confirmation(root)
found = {r['page']:r for r in result['confirmation']['technical']['pages']}
lines = ['# Confirmation misses', '', 'These ten judgments were frozen before diagnostic output review. They did not',
         'become tuning inputs in this iteration. A later iteration may use them as',
         'development material only with a new confirmation selection.', '']
for pid,p in cp.items():
    for eid in found[pid]['missed']:
        e=ce[eid]
        lines += ['## '+eid+' — '+e['category'],'','['+p['original_path']+']('+p['reference']+')','',
                  '```text', '\n'.join(t['quote'] for t in e['targets']), '```','',e['rationale'],'',
                  'Proposed edit: '+e['proposed_edit'],'']
(root/'CONFIRMATION-MISSES.md').write_text('\n'.join(lines)+'\n')
