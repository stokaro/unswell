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
    text = text.replace('Both SSL reader-desire wrappers are identified as related occurrences.', 'The diagnostic relates both SSL reader-desire wrappers.')
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
    text=f'''# Action grammar: bounded gains, broad recall still low

Full frozen-event detection on the six selected pages rises from **2/57 to 6/57**
in both profiles. All four gains are on Libevent 2.0, a page with a known shared
paragraph in previously reviewed Libevent 2.1 notes. On the other five pages,
recall remains **2/24**. Neither result meets the 80% recall objective.

The 16 added findings on previously exposed pages provide six full detections,
one partial detection, seven additional wording repairs, one unresolved judgment
and one false positive. The new engine retains every prior finding and full detection.
The prior Curl referrer and speed-condition losses are recovered.
These development gains do not establish performance on unseen prose.

## Runtime change

`filler.instruction-scaffolding` version 7 uses explicit infinitive and imperative
roles for reader goals, including coordinated actions with a shared object.
The narrower verb vocabulary for impersonal possibility statements is unchanged.
An adjacent nominal antecedent can carry a relative support chain; bare reader
capabilities still need another support layer or a related method.

Copular capable-of gerunds, imperative be-sure-to instructions and purpose-to-steps
announcements also qualify. A later used-operand layer can supply the missing
support relation. An immediate by-using method needs an explicit reader action
and a repeated two-noun object to link it to the preceding capability.

The matcher uses bounded surface roles. It does not establish dependency edges
or semantic entailment.
Keep prerequisites, actors, optionality and operational conditions in revisions.
The checksum false positive below demonstrates the limit: a nominal antecedent
can be an operand for an operation, not its actor. Permission, negation,
attribution, protected vocabulary and structural-boundary controls remain tested.
No threshold, weight, gate, model, class or catalog size changed: 53 rules, r8.

## Frozen review and overlap

Protocol, selector, six sources and notices were frozen before source reading.
The implementing assistant reviewed 554 extracted blocks and table cells, then
froze 57 defects, nine uncertainties and 42 controls at `{freeze['frozen_at']}`.
All seven rubric categories were considered. The review follows
[ADR 0041](../../../docs/adr/0041-assistant-review-acceptance.md); it is assistant
development evidence, not human annotation or independent qualification.

The selector excluded 124 prior source identities and hashes. Source-only review
then found that c04 reused the incompleteness notice from an earlier Libevent
file. The assistant froze the [amendment](source-exposure-amendment.md) before changing
the engine or reading diagnostics. It retains all six pages and their full denominator,
while separating c04 from the other five. No page was replaced.

The exact-block audit compared extracted blocks of at least 12 whitespace tokens
with all prior sources after whitespace normalization. Its one match is retained
in [source-overlap.json](confirmation/source-overlap.json). Shorter fragments,
partial overlap and paraphrases can escape that check. The other five pages have
no known overlap; that is not proof of independent provenance. All 130 selected
source identities are now exposed. One page per cohort/length cell supports
observed counts, not population recall or a meaningful within-cell bootstrap.
No matcher or frozen label changed after confirmation output was opened.

{table(['Set','Pages','Frozen defects','Before full','After full'],counts)}

Two older studies limit review to their original categories. Their counts do
not support precision claims about all warnings. [CHANGES.md](CHANGES.md) reviews
every changed finding, including new warnings outside those older scopes.
The exposed annotation-information defect receives only partial credit: detecting
the support chain does not also identify its repeated information wording.

## Complete selected pages

{table(['Page','Original path','Defects','Detected'],pages)}

Ptah remains 2/22 and the historical cohort rises from 0/35 to 4/35. With c04
excluded, the remaining historical pages stay 0/2, and the five-page total stays
2/24. Two pages have no definite defects: the MCP tool reference and raft example.
Their repeated API and procedure structure is useful, not automatically defective.

Technical findings rise from 31 to 34; strict rises from 33 to 36. Five findings
cover the six full events because one diagnostic relates both SSL alternatives.
There is one additional post-diagnostic repair and nine unresolved judgments.
The other 19 technical and 21 strict findings are nonactionable in this review.
The frozen-label actionable fraction is **5/34** for technical and **5/36** for
strict; including the additional repair gives **6/34** and **6/36**. The 85%
soft-precision objective is not met. All six pages pass both gates.

The [miss ledger](CONFIRMATION-MISSES.md) retains all 51 remaining defects.
They include unsupported benefit claims, indirect paraphrases, reader judgments,
page narration and repetition that a local action grammar does not resolve.
The shared Libevent notice itself remains missed. Detections elsewhere in that
file do not make it an unexposed page.

Two false positives need fixes. The existing evaluative-closure rule reads
`row counts` as a claim of value, although it names a measured field. The new
relative-support path mistakes a checksum's role in caching for an indirect
instruction. Review found both after the semantic freeze. They remain in the
measured result, with no later repair counted as a confirmation gain.

## Reproduction and limits

Before runtime: `{engines['before']}`. After runtime: `{engines['after']}`.
The build used the clean semantic-freeze checkout. The before reports for exposed
pages come from the prior iteration. The new engine replayed all 15 sets. The six-page before run uses the pinned preceding binary. Source
bytes, mappings, policy identities, resource costs and actual engine commits
remain in the archived JSON reports.

The known #312 repeated-claim budget abstention on exposed instruction page c05
is unchanged. No other abstention or operational error is accepted. Technical
passes all sets. Strict retains the earlier forbidden worth-noting failure in
the exposed instruction set. Text-mode Libevent extraction includes C examples;
they were not treated as prose defects. The inherited historical CMake text-mode
limitation also remains explicit.

{table(['Set','Profile','Seconds before → after','Peak MiB before → after'],costs)}

Measurements overlapped ordinary checks and are not a controlled performance
comparison or 2-vCPU qualification. Run `python3 tools/render.py` and
`python3 tools/test_evidence.py` to validate and regenerate the report.
`tools/measure.py` replays an explicit binary into a new directory. See
[VALIDATION.md](VALIDATION.md) for actual product-check results. This iteration
has not been merged or deployed to the playground.
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
