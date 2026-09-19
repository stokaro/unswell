#!/usr/bin/env python3
"""Optional blackbox checks of the isolated Go executable with local model files."""
import argparse
import copy
import hashlib
import json
import os
from pathlib import Path
import subprocess
import tempfile

STUDY = Path(__file__).resolve().parents[1]


def load(path):
    return json.loads(path.read_text())


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--binary',type=Path,required=True)
    parser.add_argument('--model',type=Path,required=True)
    args = parser.parse_args()
    fixture = STUDY/'fixture'
    original,reference = load(fixture/'data.json'),load(fixture/'vectors.json')
    results = []
    with tempfile.TemporaryDirectory(prefix='unswell-supervised-blackbox-') as directory:
        root = Path(directory)

        def run(name,data,vectors,kind='H'):
            path = root/name
            path.mkdir()
            for filename,value in [('data.json',data),('vectors.json',vectors)]:
                (path/filename).write_text(json.dumps(value)+'\n')
            command = [str(args.binary.resolve()),str(path/'data.json'),str(fixture/'warm.json'),
                       str(path/'vectors.json'),str(args.model.resolve()),str(path/'out'),'0',kind]
            completed = subprocess.run(command,capture_output=True,text=True,
                                       env=dict(os.environ,GOMAXPROCS='2'),timeout=30)
            return completed,path/'out'

        for case in ('label','binding','group','duplicate'):
            data = copy.deepcopy(original)
            vectors = copy.deepcopy(reference)
            if case == 'label':
                data['rows'][0]['label'] = 2
            elif case == 'binding':
                data['rows'][0]['text'] += ' altered'
            elif case == 'group':
                data['rows'][0]['group'] = next(r['group'] for r in data['rows'] if r['fold'] != data['rows'][0]['fold'])
            else:
                data['rows'][1] = copy.deepcopy(data['rows'][0])
                vectors['rows'][1] = copy.deepcopy(vectors['rows'][0])
            result,path = run(case,data,vectors)
            assert result.returncode == 2 and not (path/'report.json').exists(),case
            results.append(dict(case=case,rejected=True))

        data,vectors = copy.deepcopy(original),copy.deepcopy(reference)
        row = copy.deepcopy(next(r for r in data['rows'] if r['fold'] == 0))
        row.update(unit=100,text='red '*300,words=300)
        row['text_sha256'] = hashlib.sha256(row['text'].encode()).hexdigest()
        data['rows'].append(row)
        vectors['rows'].append(dict(page=row['page'],unit=100,text_sha256=row['text_sha256'],values=None))
        result,path = run('sequence-limit',data,vectors)
        assert result.returncode == 0,result.stderr
        report = load(path/'report.json')
        withheld = next(r for r in report['predictions'] if r['unit'] == 100)
        assert withheld['raw_score'] is None and not withheld['selected'] and withheld['reason'] == 'sequence_limit'
        results.append(dict(case='sequence-limit',preserved_null=True))

        result,path = run('long-batches',load(fixture/'long/data.json'),
                          load(fixture/'long/vectors.json'),'FT')
        assert result.returncode == 0,result.stderr
        assert load(path/'report.json')['reload_max_score_error'] <= 1e-5
        results.append(dict(case='long-batches',completed=True))

        paths = []
        for name in ('repeat-a','repeat-b'):
            result,path = run(name,original,reference,'FT')
            assert result.returncode == 0,result.stderr
            paths.append(path)
        for name in ('selected.onnx','head.json'):
            assert (paths[0]/name).read_bytes() == (paths[1]/name).read_bytes(),name
        reports = [load(p/'report.json') for p in paths]
        for report in reports:
            report.pop('elapsed_seconds')
        assert reports[0] == reports[1]
        results.append(dict(case='repeat',numerical_and_artifact_bytes_equal=True))
    print(json.dumps(results,indent=2))


if __name__ == '__main__':
    main()
