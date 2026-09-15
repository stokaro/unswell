"""Reproduce saved measurements and tables without sending generation requests."""
import argparse
import hashlib
import json
import pathlib
import resource
import subprocess
import sys
import time


def read(path):
    return json.loads(path.read_text())


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--archives', type=pathlib.Path, required=True)
    parser.add_argument('--study', type=pathlib.Path, required=True, help='A new, nonexistent output directory')
    parser.add_argument('--corpus', type=pathlib.Path, required=True)
    parser.add_argument('--binary', type=pathlib.Path, help='Recompute activation features with this Unswell binary')
    args = parser.parse_args()
    scripts = pathlib.Path(__file__).resolve().parent
    study = args.study.resolve()
    archives = args.archives.resolve()
    began = time.monotonic()

    def run(name, *arguments):
        subprocess.run([sys.executable, str(scripts / name), *map(str, arguments)], check=True)

    run('verify-long-archives.py', '--archives', archives, '--extract-to', study)
    manifests = [read(archives / (name + '.json')) for name in ['outputs', 'measurements']]
    expected = {path: digest for manifest in manifests for path, digest in manifest['files'].items()
                if (path.startswith('runs/') and pathlib.Path(path).name in ['records.json', 'responses.json', 'coverage.json'])
                or path.startswith(('analysis/plans/', 'analysis/candidates/', 'analysis/verifications/', 'analysis/findings/'))
                or path in ['analysis/dataset.json', 'analysis/dataset-plan.json', 'analysis/pinned.json',
                            'analysis/dataset-verification.json', 'analysis/paired-luna.json', 'analysis/paired-haiku.json',
                            'analysis/paired-qwen.json', 'analysis/descriptive.json', 'analysis/length-sensitivity.json',
                            'analysis/confirmation.json', 'analysis/cases.json', 'analysis/evidence-cards.json']}
    shared = ['--study', study, '--corpus', args.corpus.resolve(), '--rebuilt-corpus']
    run('process-long-study.py', *shared, '--phase', 'import')
    run('process-long-study.py', *shared, '--phase', 'dataset')
    run('process-long-study.py', *shared, '--phase', 'measure', '--families', 'historical', 'luna', 'haiku', 'qwen', '--remeasure')
    run('process-long-study.py', *shared, '--phase', 'paired')
    if args.binary:
        # This directory was exclusively created and verified above. Remove only
        # derived activation reports, retaining the published archives unchanged.
        for path in (study / 'analysis/features').glob('*.json'):
            path.unlink()
        for path in (study / 'analysis').glob('features-*.json'):
            path.unlink()
        run('collect-long-features.py', '--study', study, '--binary', args.binary.resolve(), '--cohort', 'historical')
        for family in ['luna', 'haiku', 'qwen']:
            run('collect-long-features.py', '--study', study, '--binary', args.binary.resolve(), '--cohort', 'controlled', '--family', family)
    for script in ['summarize-long-study.py', 'compare-long-lengths.py', 'evaluate-long-study.py', 'collect-long-cases.py']:
        run(script, '--study', study)
    run('render-long-study.py', '--study', study, '--output', study / 'results.md')
    for relative, digest in expected.items():
        with (study / relative).open('rb') as source:
            actual = hashlib.file_digest(source, 'sha256').hexdigest()
        assert actual == digest, 'Reproduction differs: ' + relative
    assert (study / 'results.md').read_bytes() == (archives / 'results.md').read_bytes()
    peak = resource.getrusage(resource.RUSAGE_CHILDREN).ru_maxrss
    record = dict(version='unswell-long-prose-reproduction-v1',
                  verified_files=len(expected), report_identical=True,
                  features_recomputed=args.binary is not None, generation_requests=0,
                  elapsed_seconds=round(time.monotonic() - began, 2),
                  maximum_child_peak_rss_bytes=peak if sys.platform == 'darwin' else peak * 1024,
                  platform=sys.platform,
                  corpus_sha256=hashlib.sha256(args.corpus.read_bytes()).hexdigest())
    (study / 'reproduction.json').write_text(json.dumps(record, indent=2) + '\n')
    print(json.dumps(record, indent=2))


if __name__ == '__main__':
    main()
