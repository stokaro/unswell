#!/usr/bin/env python3
"""Reconstruct inputs and verify the retained representation measurements."""
import gzip
import json
from pathlib import Path
import tarfile

from evaluate import evaluate
from prepare import ROOT, digest, prepare
from supervision import derive

STUDY = Path(__file__).resolve().parents[1]


def verify():
    record = json.loads((STUDY/'measurement-record.json').read_text())
    for name, wanted in record['input_files'].items():
        if digest((ROOT/name).read_bytes()) != wanted:
            raise ValueError('Reviewed input changed: '+name)
    dataset = prepare()
    raw = json.dumps(dataset, ensure_ascii=False, sort_keys=True).encode()+b'\n'
    if digest(raw) != record['input_json_sha256']:
        raise ValueError('Reconstructed input changed')
    packed = (STUDY/'representation.json.gz').read_bytes()
    if digest(packed) != record['report_sha256']:
        raise ValueError('Report changed')
    report = json.loads(gzip.decompress(packed))
    if report['manifest']['tool_commit'] != record['tool_commit']:
        raise ValueError('Wrong audit build identity')
    archive = STUDY/'exporter-source.tar.gz'
    if digest(archive.read_bytes()) != record['source_snapshot_sha256']:
        raise ValueError('Exporter snapshot changed')
    with tarfile.open(archive) as source:
        actual = {m.name:digest(source.extractfile(m).read()) for m in source.getmembers() if m.isfile()}
    if actual != record['exporter_files']:
        raise ValueError('Exporter source identity mismatch')
    actual = evaluate(dataset, report)
    expected = json.loads(gzip.decompress((STUDY/'measurements.json.gz').read_bytes()))
    if actual != expected:
        raise ValueError('Stored measurements do not match the retained report')
    labels = json.loads(gzip.decompress((STUDY/'supervision.json.gz').read_bytes()))
    if derive(dataset, report) != labels:
        raise ValueError('Derived supervision changed')
    return actual['summary']


if __name__ == '__main__':
    print(json.dumps(verify(), indent=2))
