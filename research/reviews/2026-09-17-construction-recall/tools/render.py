#!/usr/bin/env python3
"""Render checked counts, source-bound deltas and remaining defects."""
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
    names=dict(development='Original development',exposed_repetition='Exposed repetition targets',
               exposed_framing='Exposed framing targets',exposed_local='Exposed local pages',
               exposed_context='Exposed context pages',exposed_instruction='Exposed instruction pages',
               confirmation='New complete-page confirmation')
    overall=[];burden=[];costs=[]
    for split,profiles in results.items():
        b,a=(profiles['technical']['phases'][p] for p in ['before','after'])
        overall.append([names[split],len(sets[split][0]),a['defects'],b['detected'],a['detected']])
        for profile,data in profiles.items():
            a=data['phases']['after'];counts=a['dispositions']
            if counts is not None:
                burden.append([names[split],profile,a['findings'],counts.get('actionable',0),counts.get('uncertain',0),counts.get('nonactionable',0)])
            b,a=(data['costs'][phase] for phase in ['before','after'])
            costs.append([names[split],profile,f"{b['wall_seconds']:.3f} → {a['wall_seconds']:.3f}",
                          f"{b['max_rss_bytes']/2**20:.1f} → {a['max_rss_bytes']/2**20:.1f}"])
    confirm=results['confirmation']['technical'];engines=m.audit.read(m.ROOT/'engines.json')
    page_rows=[[p['page'],p['path'],p['cohort'],p['length_stratum'],p['defects'],p['detected']] for p in confirm['pages']['after']]
    categories=[[name,value['defects'],value['detected']] for name,value in confirm['phases']['after']['categories'].items()]
    frozen=m.audit.read(m.ROOT/'annotation-freeze.json')['frozen_at']
    text=f'''# Construction recall: useful exposed gains, limited confirmation

Five existing rule families now recognize more grammatical scaffolding and
repeated definitions. On exposed instruction pages, full detections increase
from 12/99 to 27/99. On six separately selected complete pages, they increase
from 0/32 to 1/32. The broad useful-recall objective remains unmet.

## Changes and limits

Cognitive-worth announcements and impersonal modal notices retain attached
reasons and conditions. Indirect instructions cover gerund methods, nominalized
methods, reader-purpose clauses and additional supported actions. Document
maintenance announcements and two bounded relational repetitions are included.
The engine still uses local NLP, source maps, clause limits and existing budgets.
Quoted and protected construction words remain controls; ordinary actors,
qualified relations and operational constraints must remain in the proposed edit.

Rule versions are 4 for evaluative closure and document justification, and 2 for
instruction scaffolding, redundant predicates and definition echo. The catalog
still has 53 rules and class manifest r8 is unchanged. All five remain experimental
warnings. Weights, profile thresholds and gates are unchanged.

The implementing Codex assistant reviewed sources under
[ADR 0041](../../../docs/adr/0041-assistant-review-acceptance.md). These are the
maintainer-accepted reviewer's judgments on the selected pages. They are not
human labels, independent agreement, authorship evidence or population accuracy.
No additional reviewer is required for this diagnostic work.

## Frozen inputs and observed recall

The [protocol](protocol.md) and input identities were frozen before reading the
selected prose. Annotations were frozen at `{frozen}`, before runtime changes or
inspection of diagnostic output. They contain 32 defects, 7 uncertain events
and 17 acceptable controls across all seven categories. Selection excludes all
76 previously reviewed sources and hashes, uses one page per cohort/length cell,
and retains source licenses. Shared repositories do not establish independent
domain transfer. No interval is estimated from a single page in each cell.

{table(['Set','Pages','Frozen defects','Before full','After full'],overall)}

Both profiles give the same full-event credits. The 17 new exposed detections
comprise two local-page events and 15 instruction-page events. One exposed
notice covers only part of a wordiness event; one additional indirect method
has no frozen defect label and remains uncertain. Source-only labels stay intact.

The new confirmation gain is the worth-and-surprise announcement on the Ptah
target-layout page. Ptah moves from 0/18 to 1/18; historical sources remain 0/14.
Five pages gain no full detection. An existing wordy-phrase finding covers only
part of the containerd introduction and never counts as a full detection.
The working 80% recall / 85% soft-diagnostic precision goals remain unfulfilled.

{table(['Page','Original path','Cohort','Length','Defects','Detected'],page_rows)}

{table(['Category','Defects','Detected'],categories)}

## Review burden and applicability

{table(['Set','Profile','Findings','Actionable','Uncertain','Nonactionable'],burden)}

Actionable includes partial matches. Full recall counts each frozen event once.
Unlabeled plausible edits stay uncertain. The one added confirmation warning
is actionable, but a single positive cannot qualify the rule's default precision.
The large nonactionable count is an existing product limitation, not evidence
that every long technical sentence or contrast needs revision.

The historical long cell selected curl's CMakeLists.txt in its inherited text
format. It produces 24 warnings over program structure, including repeated
feature/protocol/backend inventories. This is a source-format limitation:
comments and option descriptions were reviewed, but code and copyright text are
not editorial defects. The page remains in all measurements and denominators.
It reinforces the embedded-code follow-up in #307; no sample was replaced.

The existing repeated-claim budget abstention remains on exposed instruction
page c05 in both phases and profiles (#312). There are no other abstentions or
operational errors. All technical gates pass. Strict fails only on the exposed
instruction set, as it did before this iteration, because of the preexisting
forbidden worth-noting phrase. All new confirmation pages pass both gates.

## Evidence and replay

Before: `{engines['before']}`. After: `{engines['after']}`.
The initial implementation is recorded in refactor-equivalence.json; replay after
splitting helper functions preserved every document, finding, gate, abstention,
error and status. Only the tool commit changed in report manifests.

Every added and removed finding has a disposition. Changed editing guidance is
reviewed as a replacement even with an unchanged location. Cognitive-worth and
notice diagnostics may identify either wordiness or empty framing in the rubric;
the same semantic eligibility applies to both engines, and source-bound full or
partial judgments are still required. This changes no frozen label or baseline
credit. [CHANGES.md](CHANGES.md) lists all reviewed changes. The
[miss ledger](CONFIRMATION-MISSES.md) retains all 31 missed confirmation defects.

{table(['Set','Profile','Wall seconds before → after','Peak MiB before → after'],costs)}

These are local observations, not a controlled speed comparison or the roadmap
2-vCPU qualification. The original reference runs and initial implementation
measurement ran at different times; final after runs used GOMAXPROCS=4 after the
full test run. Reports retain per-process CPU, RSS, commands, host and hashes.

Run `python3 tools/render.py` and `python3 tools/test_evidence.py` in this review
directory to validate and regenerate the retained evidence without network or
model calls. CLI replay uses tools/measure.py with an explicit binary and a new
output directory. Frozen inputs and labels are checked before execution.

These six pages are now exposed. Further semantic tuning needs a separately
selected, source-only confirmation. #299 remains open for broad recall; #305 and
unsupported #309 constructions remain open. More matching fixtures cannot by
themselves establish the requested product quality.
'''
    (m.ROOT/'README.md').write_text(text)
    reviews=m.audit.read(m.ROOT/'dispositions.json');lines=['# Reviewed diagnostic changes','','Zero-based indices refer to the retained strict reports. Each phase keeps its own hashes.','']
    for split in sets:
        _,files,_=sets[split]
        after=m.audit.read(m.ROOT/'reports'/split/'after'/'strict.json.gz')
        def location_key(f):
            return (f['rule_id'], str(m.audit.locations(f)))
        additions={location_key(after['findings'][r['index']]):r['index'] for r in reviews[split]['strict']['added']}
        for kind,phase in [('added','after'),('removed','before')]:
            rows=reviews[split]['strict'][kind]
            if not rows:continue
            report=m.audit.read(m.ROOT/'reports'/split/phase/'strict.json.gz');lines+=['## '+names[split]+' — '+kind,'']
            for row in rows:
                f=report['findings'][row['index']]
                if kind=='removed' and location_key(f) in additions:
                    lines += [f"- Before index {row['index']} has updated guidance at after index {additions[location_key(f)]}; see its reviewed entry above. The before judgment remains in dispositions.json.",'']
                    continue
                lines += [f"### {row['index']}: `{f['rule_id']}`",'',row['rationale'],'',
                   'Full credit: '+(', '.join(row['events']) or 'none')+'. Partial credit: '+(', '.join(row.get('partial_events',[])) or 'none')+'.','']
                for loc in m.audit.locations(f):
                    span=loc['span'];lines += [f"`{loc['path']}` bytes {span['start']}–{span['end']}:",'','```text',files[loc['path']][span['start']:span['end']].decode(),'```','']
    (m.ROOT/'CHANGES.md').write_text('\n'.join(lines))
    _,_,events=sets['confirmation'];lines=['# Remaining frozen confirmation defects','','Both profiles miss these 31 events. Partial detection is identified in summary.json.','']
    for eid in confirm['phases']['after']['missed']:
        e=events[eid];lines += ['## '+eid+': '+e['category'],'',e['rationale'],'','Proposed edit: '+e['proposed_edit'],'']
        for span in e['targets']:lines += [f"Bytes {span['start']}–{span['end']}:",'','```text',span['quote'],'```','']
    (m.ROOT/'CONFIRMATION-MISSES.md').write_text('\n'.join(lines))


if __name__=='__main__':
    main()
