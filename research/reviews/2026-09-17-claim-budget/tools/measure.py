#!/usr/bin/env python3
"""Replay retained pages, optionally collecting generously budgeted activations."""
import argparse
import gzip
import importlib.util
import json
from pathlib import Path
import platform
import resource
import subprocess
import sys
import time

sys.dont_write_bytecode = True
ROOT = Path(__file__).resolve().parent.parent
PREVIOUS = ROOT.parent/'2026-09-17-construction-recall'
spec = importlib.util.spec_from_file_location('construction_evidence', PREVIOUS/'tools/summarize.py')
previous = importlib.util.module_from_spec(spec)
spec.loader.exec_module(previous)
audit = previous.audit


def worker(binary, folder, profile, output, activations):
    command = [binary, 'check', '--project-root', '.', '--config', profile+'.yaml',
               '--include-source', '--report', 'json:'+output, '--timeout', '5m']
    if activations:
        command += ['--feature', 'activation/repetition.repeated-claim']
    command += ['sources']
    start = time.monotonic()
    run = subprocess.run(command, cwd=folder, capture_output=True, timeout=360)
    seconds = time.monotonic()-start
    usage = resource.getrusage(resource.RUSAGE_CHILDREN)
    Path(output+'.log').write_bytes(run.stdout+run.stderr)
    audit.require(run.returncode in (0, 1), 'Operational failure')
    print(json.dumps(dict(exit_code=run.returncode, wall_seconds=seconds,
                         cpu_seconds=usage.ru_utime+usage.ru_stime,
                         max_rss_bytes=usage.ru_maxrss*(1 if sys.platform=='darwin' else 1024),
                         command=['unswell', *command[1:]])))


def main():
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--binary', type=Path, required=True)
    parser.add_argument('--set', choices=list(previous.data_sets()), required=True)
    parser.add_argument('--output', type=Path, required=True)
    parser.add_argument('--activations', action='store_true')
    args=parser.parse_args()
    previous.frozen_inputs()
    binary=args.binary.resolve(strict=True);output=args.output.resolve()
    audit.require(not output.exists(), 'Output must be new')
    _,files,_=previous.data_sets()[args.set]
    inputs=output/'inputs'
    for name,data in files.items():
        path=inputs/name;path.parent.mkdir(parents=True,exist_ok=True);path.write_bytes(data)
    runs={}
    for profile in ('technical','strict'):
        config=(previous.prior.DEVELOPMENT/(profile+'.yaml')).read_bytes()
        if args.activations:
            config+=b'analysis: {max_candidates: 1000000}\n'
        (inputs/(profile+'.yaml')).write_bytes(config)
        target=output/(profile+'.json')
        process=subprocess.run([sys.executable,str(Path(__file__).resolve()),'--worker',str(binary),str(inputs),
                                profile,str(target),'1' if args.activations else '0'],check=True,capture_output=True,text=True)
        costs=json.loads(process.stdout);report=audit.read(target)
        audit.require(report['status']=='complete' and report['manifest']['complete'] and not report['errors'],'Incomplete report')
        compressed=gzip.compress(target.read_bytes(),mtime=0);name=profile+'.json.gz';(output/name).write_bytes(compressed)
        runs[profile]=dict(path=name,sha256=audit.sha(compressed),binary_sha256=audit.sha(binary.read_bytes()),
                           config_sha256=audit.sha(config),config_identity_sha256=report['manifest']['config_sources'][0]['sha256'],
                           tool_commit=report['manifest']['tool_commit'],host=platform.platform(),**costs)
        print(args.set,profile,len(report['findings']),len(report.get('abstentions',[])),flush=True)
    audit.write(output/'runs.json',runs)


if __name__=='__main__':
    if len(sys.argv)>1 and sys.argv[1]=='--worker':
        worker(*sys.argv[2:6],sys.argv[6]=='1')
    else:
        main()
