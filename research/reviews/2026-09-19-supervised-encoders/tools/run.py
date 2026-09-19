#!/usr/bin/env python3
"""Orchestrate the isolated Go fits; Python does not train the model."""
import argparse
import datetime
import hashlib
import json
import os
from pathlib import Path
import subprocess
import time

STUDY = Path(__file__).resolve().parents[1]


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    for name in ('binary','sources','inputs','vectors','model','out'):
        parser.add_argument('--'+name,type=Path,required=True)
    args = parser.parse_args()
    plan = json.loads((STUDY/'plan.json').read_text())
    execution = json.loads((STUDY/'execution.json').read_text())
    for name in ('data.json','warm.json'):
        assert digest(args.inputs/name) == plan['files'][name], name
    assert digest(args.vectors) == execution['vectors_sha256']
    for name,expected in execution['source_files'].items():
        assert digest(args.sources/name) == expected,name
    manifest = json.loads((STUDY.parent/'2026-09-19-context-encoders/feasibility/tiny-model-manifest.json').read_text())
    for name,meta in manifest['files'].items():
        path = args.model/('model.onnx' if name == 'onnx/model.onnx' else name)
        assert digest(path) == meta['sha256'],name
    args.out.mkdir(exist_ok=False)
    record = dict(started_at=datetime.datetime.now(datetime.timezone.utc).isoformat(),
                  binary_sha256=digest(args.binary),gomaxprocs=2,runs=[])
    record_path = args.out/'run-record.json'
    record_path.write_text(json.dumps(record,indent=2)+'\n')
    for fold in range(5):
        for kind in ('H','FT'):
            name = f'{kind.lower()}-{fold}'
            command = [str(args.binary.resolve()),str(args.inputs/'data.json'),
                       str(args.inputs/'warm.json'),str(args.vectors),str(args.model),
                       str(args.out/name),str(fold),kind]
            start = time.monotonic()
            with (args.out/(name+'.log')).open('wb') as log:
                result = subprocess.run(command,stdout=log,stderr=subprocess.STDOUT,
                                        env=dict(os.environ,GOMAXPROCS='2'),timeout=1860)
            row = dict(name=name,exit_code=result.returncode,seconds=time.monotonic()-start)
            record['runs'].append(row)
            record_path.write_text(json.dumps(record,indent=2)+'\n')
            print(json.dumps(row),flush=True)
            if result.returncode:
                raise SystemExit(result.returncode)
    record['status'] = 'complete'
    record_path.write_text(json.dumps(record,indent=2)+'\n')


if __name__ == '__main__':
    main()
