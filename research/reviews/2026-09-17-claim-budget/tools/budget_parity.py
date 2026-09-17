#!/usr/bin/env python3
"""Validate complete-report parity and the one intended applicability change."""
import copy
import importlib.util
from pathlib import Path
import sys

sys.dont_write_bytecode=True
ROOT=Path(__file__).resolve().parent.parent
PARENT=ROOT.parent/'2026-09-17-construction-recall'
spec=importlib.util.spec_from_file_location('construction_evidence',PARENT/'tools/summarize.py')
previous=importlib.util.module_from_spec(spec);spec.loader.exec_module(previous)
audit=previous.audit
ABSTENTION=dict(path='sources/c05.md',rule_id='repetition.repeated-claim',rule_version='1',
                reason='budget_exhausted',detail='repetition work exceeds max_candidates')


def compare(before,after,split,activations):
    expected=[] if activations or split!='exposed_instruction' else [ABSTENTION]
    audit.require(before.get('abstentions',[])==expected,'Unexpected before abstention')
    audit.require(after.get('abstentions',[])==[],'After abstention')
    audit.require(before['manifest'].get('abstained_rules',[])==([ABSTENTION['rule_id']] if expected else []),'Before applicability manifest differs')
    audit.require(after['manifest'].get('abstained_rules',[])==[],'After applicability manifest differs')
    a,b=copy.deepcopy(before),copy.deepcopy(after)
    for report in [a,b]:
        report['manifest'].pop('tool_commit',None)
        report['manifest'].pop('abstained_rules',None)
        report.pop('abstentions',None)
    audit.require(a==b,'Report semantics or observations differ')
    return len(expected)


def read_run(folder,profile,pages,files,commit,activations):
    record=audit.read(folder/'runs.json')[profile]
    path=folder/(profile+'.json.gz');report=audit.read(path)
    audit.require(audit.sha(path.read_bytes())==record['sha256'],'Report hash drift')
    audit.require(report['manifest']['tool_commit']==record['tool_commit']==commit,'Runtime identity drift')
    audit.require(report['status']=='complete' and report['manifest']['complete'] and not report['errors'],'Incomplete report')
    audit.require(record['exit_code']==(0 if report['gate']['passed'] else 1),'Exit/gate mismatch')
    config=(previous.prior.DEVELOPMENT/(profile+'.yaml')).read_bytes()
    if activations:config+=b'analysis: {max_candidates: 1000000}\n'
    audit.require(audit.sha(config)==record['config_sha256'],'Changed configuration')
    audit.require(report['manifest']['config_sources']==[dict(path=profile+'.yaml',kind='config',sha256=record['config_identity_sha256'])],'Configuration identity differs')
    docs=audit.unique(report['documents'],'name','document')
    audit.require(set(docs)=={p['path'] for p in pages.values()},'Page set drift')
    for page in pages.values():
        doc=docs[page['path']]
        audit.require(doc['source_hash']==page['sha256'] and doc['source'].encode()==files[page['path']] and doc['format']==page['format'],'Source drift')
    for finding in report['findings']:
        for loc in audit.locations(finding):
            audit.span(files[loc['path']],loc['span'])
            for span in loc.get('segments',[]):audit.span(files[loc['path']],span)
    if activations:
        audit.require(bool(report.get('features',{}).get('sources')),'Missing activation evidence')
    return report,record


def evaluate():
    previous.frozen_inputs();engines=audit.read(ROOT/'engines.json');rows=[]
    for name,(pages,files,_) in previous.data_sets().items():
        for profile in ['technical','strict']:
            for kind in ['reports','activations']:
                active=kind=='activations'
                b,bc=read_run(ROOT/kind/name/'before',profile,pages,files,engines['before'],active)
                a,ac=read_run(ROOT/kind/name/'after',profile,pages,files,engines['after'],active)
                if not active:
                    audit.require(b==audit.read(PARENT/'reports'/name/'after'/(profile+'.json.gz')),'Retained baseline differs')
                removed=compare(b,a,name,active)
                rows.append(dict(set=name,profile=profile,kind=kind,pages=len(pages),findings=len(a['findings']),
                                 gate_passed=a['gate']['passed'],removed_abstentions=removed,
                                 complete_report_parity_except_runtime_and_declared_abstention=True,
                                 costs=dict(before=bc,after=ac)))
    return dict(engines=engines,rows=rows)


if __name__=='__main__':
    audit.write(ROOT/'summary.json',evaluate())
    print('All 28 report pairs preserve source, findings, assessments, gates and applicable observations.')
