#!/usr/bin/env python3
"""Reconstruct the frozen supervised inputs from prior verified Go exports."""
import argparse
import gzip
import hashlib
import json
from pathlib import Path

STUDY = Path(__file__).resolve().parents[1]


def load(path):
    raw = path.read_bytes()
    return json.loads(gzip.decompress(raw) if path.suffix == '.gz' else raw)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--input', type=Path, required=True)
    parser.add_argument('--texts', type=Path, required=True)
    parser.add_argument('--out', type=Path, required=True)
    args = parser.parse_args()
    plan = load(STUDY/'plan.json')
    raw = args.input.read_bytes()
    if args.input.suffix == '.gz':
        raw = gzip.decompress(raw)
    assert hashlib.sha256(raw).hexdigest() == plan['parent_input_sha256']
    source, texts = json.loads(raw), load(args.texts)
    assert texts['input_sha256'] == plan['parent_input_sha256']
    prior = load(STUDY.parent/'2026-09-19-context-encoders/predictions.json.gz')
    pages = {p['id']:p for p in source['pages']}
    predictions = {(r['page'],r['unit']):r for m in prior['models']
                   if m['kind'] == 'E' for r in m['predictions']}
    rows = []
    for row in texts['rows']:
        page = pages[row['page']]
        unit = page['units'][row['unit']]
        prior_row = predictions[(row['page'],row['unit'])]
        assert unit['binding']['text_sha256'] == row['text_sha256']
        assert hashlib.sha256(row['text'].encode()).hexdigest() == row['text_sha256']
        assert unit['label'] == prior_row['label']
        rows.append(dict(row, fold=page['fold'], group=page['group'],
                         cohort=page['cohort'], label=unit['label'], words=prior_row['words']))
    data = dict(version=1,input_sha256=plan['parent_input_sha256'],rows=rows)
    warm = dict(version=1,input_sha256=plan['parent_input_sha256'],
                encoder_sha256=prior['encoder_sha256'],models=[
                    dict(fold=m['evaluation_fold'],parameters=m['parameters'],
                         training_hash=m['training_sha256'],training_groups=m['training_groups'])
                    for m in prior['models'] if m['kind'] == 'E'])
    args.out.mkdir(exist_ok=False)
    for name, value in [('data.json',data),('warm.json',warm)]:
        content = (json.dumps(value,ensure_ascii=True,separators=(',',':'))+'\n').encode()
        assert hashlib.sha256(content).hexdigest() == plan['files'][name], name
        (args.out/name).write_bytes(content)
    print(json.dumps({'rows':len(rows),'status':'frozen input hashes reproduced'}))


if __name__ == '__main__':
    main()
