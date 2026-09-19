#!/usr/bin/env python3
"""Repeat the frozen Go fit from retained vectors, without an encoder runtime."""
import argparse
import gzip
import hashlib
import json
from pathlib import Path
import subprocess
import tempfile
import time

STUDY = Path(__file__).resolve().parents[1]


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--binary',type=Path,required=True)
    parser.add_argument('--input',type=Path,required=True)
    parser.add_argument('--record',type=Path,required=True)
    args = parser.parse_args()
    if args.record.exists():
        raise ValueError('Refusing to overwrite repeat evidence')
    raw = args.input.read_bytes()
    if args.input.suffix == '.gz':
        raw = gzip.decompress(raw)
    plan = json.loads((STUDY/'split-plan.json').read_text())
    if hashlib.sha256(raw).hexdigest() != plan['input_json_sha256']:
        raise ValueError('Frozen source input changed')
    with tempfile.TemporaryDirectory(prefix='unswell-encoder-repeat-') as directory:
        vectors = Path(directory)/'embeddings.json'
        vectors.write_bytes(gzip.decompress((STUDY/'embeddings.json.gz').read_bytes()))
        started = time.monotonic()
        result = subprocess.run([str(args.binary.resolve()),str(vectors)],input=raw,
                                stdout=subprocess.PIPE,stderr=subprocess.PIPE,check=True)
    expected = gzip.decompress((STUDY/'predictions.json.gz').read_bytes())
    record = dict(elapsed_seconds=time.monotonic()-started,byte_identical=result.stdout==expected,
                  predictions_sha256=hashlib.sha256(result.stdout).hexdigest(),
                  binary_sha256=hashlib.sha256(args.binary.read_bytes()).hexdigest(),
                  stderr=result.stderr.decode(),models=10,evaluation_rows=59500)
    args.record.write_text(json.dumps(record,indent=2)+'\n')
    if not record['byte_identical']:
        raise ValueError('Repeated Go fit differs from retained output')
    print(json.dumps(record))


if __name__ == '__main__':
    main()
