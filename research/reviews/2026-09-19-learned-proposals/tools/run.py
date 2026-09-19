#!/usr/bin/env python3
"""Run the frozen Go comparison and retain the actual execution identities."""
import argparse
import gzip
import hashlib
import io
import json
from pathlib import Path
import platform
import subprocess
import tarfile
import time

STUDY = Path(__file__).resolve().parents[1]
ROOT = STUDY.parents[2]


def sha(data):
    return hashlib.sha256(data).hexdigest()


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--binary', type=Path, required=True)
    parser.add_argument('--input', type=Path, required=True)
    args = parser.parse_args()
    record_path = STUDY/'measurement-record.json'
    if record_path.exists():
        raise ValueError('A measurement already exists; do not overwrite it')
    raw = gzip.decompress(args.input.read_bytes())
    plan = json.loads((STUDY/'split-plan.json').read_text())
    if sha(raw) != plan['input_json_sha256'] or sha((STUDY/'protocol.md').read_bytes()) != plan['protocol_sha256']:
        raise ValueError('Frozen input or protocol changed')
    files = sorted(list((ROOT/'research/annotation/internal/reviewbaseline').glob('*.go')) +
                   list((ROOT/'research/annotation/cmd/reviewbaseline').glob('*.go')))
    archive = io.BytesIO()
    identities = {}
    with tarfile.open(fileobj=archive, mode='w') as output:
        for path in files:
            data = path.read_bytes()
            name = str(path.relative_to(ROOT))
            identities[name] = sha(data)
            item = tarfile.TarInfo(name)
            item.size = len(data)
            item.mode = 0o644
            output.addfile(item, io.BytesIO(data))
    snapshot = gzip.compress(archive.getvalue(), mtime=0)
    (STUDY/'trainer-source.tar.gz').write_bytes(snapshot)
    record = dict(version=1, runtime_base=subprocess.check_output(['git','rev-parse','HEAD'],cwd=ROOT,text=True).strip(),
                  tool_files=identities, source_snapshot_sha256=sha(snapshot), binary_sha256=sha(args.binary.read_bytes()),
                  input_json_sha256=sha(raw), protocol_sha256=plan['protocol_sha256'],
                  split_plan_sha256=sha((STUDY/'split-plan.json').read_bytes()),
                  go_version=subprocess.check_output(['go','version'],text=True).strip(), platform=platform.platform(),
                  scope='exposed assistant development; not product qualification', status='started')
    record_path.write_text(json.dumps(record,indent=2)+'\n')
    started = time.monotonic()
    completed = subprocess.run([str(args.binary.resolve())],input=raw,stdout=subprocess.PIPE,stderr=subprocess.PIPE)
    record.update(exit_code=completed.returncode, elapsed_seconds=time.monotonic()-started,
                  stderr=completed.stderr.decode(errors='replace'),
                  status='complete' if completed.returncode == 0 else 'failed')
    if completed.returncode == 0:
        result = json.loads(completed.stdout)
        if result['input_sha256'] != sha(raw):
            raise ValueError('Run consumed a different input')
        packed = gzip.compress(completed.stdout,mtime=0)
        (STUDY/'predictions.json.gz').write_bytes(packed)
        record['predictions_sha256'] = sha(packed)
    record_path.write_text(json.dumps(record,indent=2)+'\n')
    print(json.dumps({k:v for k,v in record.items() if k != 'tool_files'},indent=2))
    raise SystemExit(completed.returncode)


if __name__ == '__main__':
    main()
