#!/usr/bin/env python3
"""Repeat the fixed FT fold-zero fit and compare numerical outputs and model bytes."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import resource
import subprocess
import sys
import time


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    for name in ('binary','inputs','vectors','model','expected','out','record'):
        parser.add_argument('--'+name,type=Path,required=True)
    args = parser.parse_args()
    if args.out.exists() or args.record.exists():
        raise ValueError('Refusing to overwrite repeat evidence')
    command = [str(args.binary.resolve()),str(args.inputs/'data.json'),str(args.inputs/'warm.json'),
               str(args.vectors),str(args.model),str(args.out),'0','FT']
    start = time.monotonic()
    result = subprocess.run(command,capture_output=True,text=True,
                            env=dict(os.environ,GOMAXPROCS='2'),timeout=1860)
    if result.returncode:
        raise RuntimeError(result.stderr)
    expected = json.loads((args.expected/'report.json').read_text())
    actual = json.loads((args.out/'report.json').read_text())
    expected.pop('elapsed_seconds')
    actual.pop('elapsed_seconds')
    identical = {name:(args.expected/name).read_bytes() == (args.out/name).read_bytes()
                 for name in ('selected.onnx','head.json')}
    peak = resource.getrusage(resource.RUSAGE_CHILDREN).ru_maxrss
    record = dict(scope='fixed FT fold zero only; not an all-fit or cross-platform repeat',
                  numerical_report_identical=expected == actual,files_byte_identical=identical,
                  elapsed_seconds=time.monotonic()-start,platform=sys.platform,
                  peak_rss_bytes=int(peak if sys.platform == 'darwin' else peak*1024),gomaxprocs=2,
                  binary_sha256=digest(args.binary),data_sha256=digest(args.inputs/'data.json'),
                  warm_sha256=digest(args.inputs/'warm.json'),vectors_sha256=digest(args.vectors),
                  snapshot_sha256=digest(args.out/'selected.onnx'),
                  report_sha256=digest(args.out/'report.json'),log=result.stdout+result.stderr)
    args.record.write_text(json.dumps(record,indent=2)+'\n')
    if not record['numerical_report_identical'] or not all(identical.values()):
        raise ValueError('Repeated training differs from the selected original fit')
    print(json.dumps({k:v for k,v in record.items() if k != 'log'},indent=2))


if __name__ == '__main__':
    main()
