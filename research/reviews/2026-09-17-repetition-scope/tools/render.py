#!/usr/bin/env python3
"""Render source-bound changes and recall from validated saved reports."""
import sys
sys.dont_write_bytecode = True
import compare as c

r=c.ROOT

def markdown(name, lines):
    text = '\n'.join(line.rstrip() for line in '\n'.join(lines).splitlines()).rstrip() + '\n'
    (r/name).write_text(text)

results=c.evaluate()
c.audit.write(r/'results.json',results)
lines=['# Before and after','','Counts describe agreement with the same assistant\'s frozen judgments on these','pages. They are not population recall or authorship accuracy.','',
'| Set | Profile | Defects | Before | After | Findings before / after | Added / removed |',
'| --- | --- | ---: | ---: | ---: | ---: | ---: |']
for split,profiles in results.items():
 for profile,data in profiles.items():
  totals={k:sum(p[k] for p in data['pages']) for k in ('defects','before','after')}
  lines.append(f'| {split} | {profile} | {totals["defects"]} | {totals["before"]} | {totals["after"]} | {data["findings"]["before"]} / {data["findings"]["after"]} | {len(data["delta"]["added"])} / {len(data["delta"]["removed"])} |')
lines += ['', '## Repetition category', '',
          'The development denominator above spans all editorial categories. Restricting',
          'it to the frozen needless-repetition labels gives the following result.', '',
          '| Set | Profile | Repetition defects | Before | After |',
          '| --- | --- | ---: | ---: | ---: |']
for split, profiles in results.items():
    for profile, data in profiles.items():
        d = data['repetition']
        lines.append(f'| {split} | {profile} | {d["defects"]} | {d["before"]} | {d["after"]} |')
lines+=['','The gain is p014-d01; the loss is p004-d01. Development recall is unchanged.','The glossary paraphrase p012-d01 remains missed. The six confirmation pages','contain five annotated defects, five uncertain judgments and fourteen controls.','None of their five defects is detected before or after.','',
'All scans complete with passing gates and no abstentions. The index and gate','thresholds are unchanged. Declining finding counts do not imply improved recall.','',
'| Page | Cohort / selection | Defects | Before | After |','| --- | --- | ---: | ---: | ---: |']
for profiles in results.values():
 for p in profiles['technical']['pages']:
  lines.append(f'| {p["page"]} | {p["cohort"]} / {p["selection"]} | {p["defects"]} | {p["before"]} | {p["after"]} |')
lines+=['','Per-page event credit is identical between profiles.','', '## Resource observations','','One fresh process per scan, including model startup and reports. These are local','observations, not an isolated performance comparison. Other development checks','ran on the host. Order: development-before, development-after, confirmation-before,','confirmation-after; technical preceded strict within each pair.','',
'| Set / profile | Wall seconds before / after | CPU seconds before / after | Peak RSS MiB before / after |','| --- | ---: | ---: | ---: |']
for split,profiles in results.items():
 for profile,data in profiles.items():
  b,a=data['costs']['before'],data['costs']['after']
  lines.append(f'| {split} / {profile} | {b["wall_seconds"]:.3f} / {a["wall_seconds"]:.3f} | {b["cpu_seconds"]:.3f} / {a["cpu_seconds"]:.3f} | {b["max_rss_bytes"]/2**20:.1f} / {a["max_rss_bytes"]/2**20:.1f} |')
markdown('RESULTS.md', lines)
review=c.audit.read(r/'dispositions.json')
lines=['# Diagnostic changes','','Both profiles add one actionable diagnostic and remove twelve development','diagnostics: eleven nonactionable findings and one actionable finding.','Confirmation removes one nonactionable finding and adds none.','',
'Every change below has source-bound review in `dispositions.json`. Unchanged','development findings inherit the original audit and framing follow-up judgments.','This is assistant review; no independent agreement statistic is claimed.',
'Rendered snippets trim trailing line whitespace; the JSON reports retain exact text.','']
for split in ('development','confirmation'):
 for kind,phase in [('added','after'),('removed','before')]:
  report=c.audit.read(r/'reports'/split/phase/'technical.json.gz')
  for row in review[split]['technical'][kind]:
   f=report['findings'][row['index']]
   lines += [f'## {split}: {kind} {f["rule_id"]} ({row["status"]})','',row['rationale'],'']
   for loc in c.audit.locations(f):
    s=loc['span'];lines += [f'{loc["path"]}, bytes [{s["start"]}, {s["end"]}):','','```text',loc['snippet'],'```','']
markdown('CHANGES.md', lines)
pages,_,events=c.confirmation(r)
lines=['# Remaining confirmation defects','','All five defects below remain missed. Adjacent-word repetitions, repeated','clauses and paraphrases are distinct from complete list-item duplicates.','The labels were frozen before diagnostic outputs were opened.','']
for event in events.values():
 if event['kind']!='defects':continue
 p=pages[event['page']];lines += [f'## {event["id"]}: {event["subtype"]}','',f'[{p["repository"]}/{p["original_path"]}]({p["reference"]})','']
 for target in event['targets']:lines+=['```text',target['quote'],'```','']
 lines += [event['rationale'],'','Proposed edit: '+event['proposed_edit'],'']
markdown('CONFIRMATION-MISSES.md', lines)
