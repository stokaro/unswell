#!/usr/bin/env python3
"""Collect a complete corrected run and derive its tables from stored predictions."""
import argparse
import gzip
import hashlib
import json
from pathlib import Path
import shutil
import check


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('run',type=Path)
    args = parser.parse_args()
    root = check.STUDY
    record = check.load(args.run/'run-record.json')
    expected = [f'{kind}-{fold}' for fold in range(5) for kind in ('h','ft')]
    check.require(record.get('status') == 'complete' and
                  [r['name'] for r in record['runs']] == expected and
                  all(r['exit_code'] == 0 for r in record['runs']),'Incomplete run')
    models = [check.load(args.run/'runs'/name/'report.json') for name in expected]
    prior,metadata = check.source_metadata()
    report = dict(version=1,basis='exposed assistant development; selection is not diagnostic credit',
                  input_sha256=prior['input_sha256'],encoder_sha256=prior['encoder_sha256'],
                  rows=len(metadata),models=models)
    check.validate(report,metadata,prior)
    predictions = root/'predictions.json.gz'
    check.require(not predictions.exists(),'Refusing to overwrite collected predictions')
    predictions.write_bytes(gzip.compress((json.dumps(report,separators=(',',':'))+'\n').encode(),mtime=0))
    summary,rules = check.derive(report)
    for name,value in [('summary.json',summary),('rule-reference.json',rules)]:
        (root/name).write_text(json.dumps(value,indent=2)+'\n')
    numerical = [check.load(args.run/'numerical'/(name+'.json')) for name in expected]
    (root/'numerical-reference.json').write_text(json.dumps(numerical,indent=2)+'\n')
    shutil.copy2(args.run/'run-record.json',root/'run-record.json')
    logs = root/'training-logs'
    logs.mkdir(exist_ok=False)
    snapshots = []
    for name in expected:
        shutil.copy2(args.run/'runs'/(name+'.log'),logs/(name+'.log'))
        path = args.run/'runs'/name
        snapshots.append(dict(fit=name,files={file:dict(bytes=(path/file).stat().st_size,
                              sha256=hashlib.sha256((path/file).read_bytes()).hexdigest())
                              for file in ('selected.onnx','head.json','report.json')}))
    (root/'snapshot-manifest.json').write_text(json.dumps(snapshots,indent=2)+'\n')
    print(json.dumps({kind:{key:row[key] for key in ('true_positive','false_positive',
                     'unlabeled_selected','event_retrieval','selected_epochs','cohorts')}
                     for kind,row in summary['models'].items()},indent=2))


if __name__ == '__main__':
    main()
