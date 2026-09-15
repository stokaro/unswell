"""Import and measure frozen responses with the existing Go corpus commands."""
import argparse
import collections
import hashlib
import json
import pathlib
import subprocess
import time

FAMILIES = {
    'luna': ('openai', 'gpt-5.6-luna', 'codex exec', 'reasoning effort low; decoding parameters unavailable'),
    'haiku': ('claude', 'claude-haiku-4-5-20251001', 'claude print', 'reasoning effort low; decoding parameters unavailable'),
    'qwen': ('qwen', 'HivenetQuant/Qwen3.6-35B-A3B', 'OpenAI-compatible router',
             'temperature 0.7; top_p 0.9; max_tokens 16384; enable_thinking false'),
}


def read(path):
    return json.loads(path.read_text())


def write(path, value):
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(value, ensure_ascii=False, indent=2) + '\n')


def invoke(tool, study, args, output, input_path=None):
    output.parent.mkdir(parents=True, exist_ok=True)
    temporary = output.with_suffix('.pending')
    with temporary.open('wb') as out:
        result = subprocess.run([str(tool)] + [str(a) for a in args], cwd=study,
                                input=input_path.read_bytes() if input_path else None,
                                stdout=out, stderr=subprocess.PIPE)
    if result.returncode:
        raise RuntimeError(' '.join(map(str, args)) + ': ' + result.stderr.decode())
    temporary.replace(output)
    return read(output)


def verify_raw(study, tool, family):
    del tool
    run = study / 'runs' / family
    requests = read(run / 'requests.json')
    assert requests['tasks_sha256'] == hashlib.sha256((study / 'tasks.json').read_bytes()).hexdigest()
    seen = set()
    for request in requests['requests']:
        assert request['id'] not in seen
        seen.add(request['id'])
        directory = run / 'raw' / request['id']
        result = read(directory / 'result.json')
        visible = read(directory / 'input.json')
        assert visible['request'] == request
        assert visible['message'] == requests['harness_instruction'] + '\n\n' + request['text']
        assert visible['system'] == (study / 'system.txt').read_text().strip()
        attempts = sorted(directory.glob('[12]/outcome.json'))
        assert 1 <= len(attempts) <= 2
        outcome = read(attempts[-1])
        assert result['request_id'] == request['id']
        assert {k: v for k, v in result.items() if k != 'request_id'} == outcome
        if family == 'haiku' and result['status'] == 'complete':
            raw = read(attempts[-1].parent / 'stdout')
            assert raw['stop_reason'] == 'end_turn' and not raw['is_error']
            assert raw['result'] == result['text']
        if result['status'] in ['complete', 'truncated', 'refused'] and family != 'luna':
            assert result['served_models'] == [FAMILIES[family][1]]
    print(family, len(seen), 'visible inputs and saved outcomes verified', flush=True)


def collect(study, tool, family):
    verify_raw(study, tool, family)
    run = study / 'runs' / family
    requests = read(run / 'requests.json')
    results = []
    for request in requests['requests']:
        path = run / 'raw' / request['id'] / 'result.json'
        if not path.exists():
            raise RuntimeError(f'{family}: missing outcome for {request["id"]}; do not import a live run')
        result = read(path)
        assert result['request_id'] == request['id']
        results.append({k: result[k] for k in ['request_id', 'text', 'status', 'note', 'duration_ms', 'tokens', 'remote_request_id'] if k in result})
    name, model, agent, parameters = FAMILIES[family]
    responses = dict(version='unswell-responses-v1', run=requests['run'], family=name, model=model,
                     model_basis=('CLI selection; independently served identity unavailable' if family == 'luna' else
                                  'Claude modelUsage in retained stdout' if family == 'haiku' else 'router response model field'),
                     harness='Frozen run-models.py; exact visible system and user inputs in raw/<request_id>/input.json; '
                             'CLI built-in context may remain and is not represented as a complete system transcript',
                     agent_type=agent, generated_on='2026-09-15', parameters=parameters, responses=results)
    write(run / 'responses.json', responses)
    coverage = invoke(tool, study, ['generations', '--tasks', 'tasks.json', '--requests', f'runs/{family}/requests.json',
                     '--responses', f'runs/{family}/responses.json', '--records', f'runs/{family}/records.json',
                     '--work', 'analysis/work', '--output', 'analysis/acquisition/controlled/shards',
                     '--historical', 'analysis/acquisition/historical/records', '--shard-suffix', '-long-v3-' + family],
                      run / 'coverage.json')
    print(family, coverage, flush=True)


def dataset(study, tool):
    analysis = study / 'analysis'
    spec = read(study / 'dataset.json')
    spec['id'] = 'long-prose-v3-measured'
    spec['shards'] = []
    for cohort in ['historical', 'controlled']:
        for path in sorted((analysis / 'acquisition' / cohort / 'shards').glob('*.json')):
            data = path.read_bytes()
            spec['shards'].append(dict(path=str(path.relative_to(analysis)), sha256=hashlib.sha256(data).hexdigest(), bytes=len(data)))
    write(analysis / 'dataset.json', spec)
    plan = invoke(tool, study, ['dataset', 'plan', '--root', 'analysis'], analysis / 'dataset-plan.json', analysis / 'dataset.json')
    assert len(plan['groups']) == 20 and all(g['partition'] == 'final_test' for g in plan['groups'])
    pinned = analysis / 'pinned'
    pinned.mkdir(exist_ok=True)
    # The CLI refuses to overwrite pinned manifests. Reuse only byte-identical output.
    if not (analysis / 'pinned.json').exists():
        invoke(tool, study, ['dataset', 'pin', '--root', 'analysis', '--output', 'analysis/pinned'],
               analysis / 'pinned.json', analysis / 'dataset-plan.json')
    invoke(tool, study, ['dataset', 'verify', '--root', 'analysis', '--pinned', 'analysis/pinned'],
           analysis / 'dataset-verification.json', analysis / 'dataset-plan.json')
    print('dataset', len(spec['shards']), 'shards', len(plan['groups']), 'groups', flush=True)


def measure(study, tool, family, force=False):
    analysis = study / 'analysis'
    cohort = 'historical' if family == 'historical' else 'controlled'
    pattern = '*.json' if family == 'historical' else '*-long-v3-' + family + '.json'
    for path in sorted((analysis / 'acquisition' / cohort / 'shards').glob(pattern)):
        stem = path.stem
        manifest = read(path)
        root = analysis / 'work' / cohort / manifest['sources'][0]['repository'].replace('/', '__')
        plan_path = analysis / 'plans' / (stem + '.json')
        candidates = analysis / 'candidates' / (stem + '.json')
        findings = analysis / 'findings' / (stem + '.json')
        if findings.exists() and not force:
            saved = read(findings)
            assert saved['failed_documents'] == 0
            continue
        invoke(tool, study, ['plan'], plan_path, path)
        invoke(tool, study, ['extract', '--root', root], candidates, plan_path)
        invoke(tool, study, ['verify', '--root', root], analysis / 'verifications' / (stem + '.json'), candidates)
        start = time.monotonic()
        result = invoke(tool, study, ['measure', '--root', root, '--policy', study / 'policy.yaml'], findings, candidates)
        assert result['failed_documents'] == 0
        print(stem, len(result['documents']), 'documents', round(time.monotonic() - start, 2), 'seconds', flush=True)


def paired(study, tool, family):
    analysis = study / 'analysis'
    args = ['paired', '--plan', 'analysis/dataset-plan.json', '--tasks', 'tasks.json',
            '--records', f'runs/{family}/records.json', '--classes', 'rule-classes.json']
    for path in sorted((analysis / 'findings').glob('*.json')):
        args += ['--findings', path]
    result = invoke(tool, study, args, analysis / ('paired-' + family + '.json'))
    assert result['version'] == 'unswell-paired-tables-v3'
    print(family, 'paired', result['arms'], flush=True)


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--study', type=pathlib.Path, required=True)
    parser.add_argument('--corpus', type=pathlib.Path, required=True)
    parser.add_argument('--phase', choices=['import', 'measure', 'dataset', 'paired', 'verify-raw'], required=True)
    parser.add_argument('--families', nargs='+', choices=[*FAMILIES, 'historical'], default=list(FAMILIES))
    parser.add_argument('--remeasure', action='store_true', help='Replace derived measurement outputs with a fresh engine scan')
    parser.add_argument('--rebuilt-corpus', action='store_true', help='Explicitly allow a rebuilt binary; verify reproduced outputs against the published results')
    args = parser.parse_args()
    if 'historical' in args.families and args.phase != 'measure':
        parser.error('historical is supported only by the measure phase')
    study, tool = args.study.resolve(), args.corpus.resolve()
    freeze = read(study / 'freeze.json')
    binary_sha = hashlib.sha256(tool.read_bytes()).hexdigest()
    if binary_sha != freeze['corpus_binary_sha256'] and not args.rebuilt_corpus:
        raise RuntimeError('Corpus binary differs from freeze; an independent rebuild requires --rebuilt-corpus and output verification')
    if args.phase == 'dataset':
        dataset(study, tool)
    else:
        for family in args.families:
            if args.phase == 'measure':
                measure(study, tool, family, args.remeasure)
            else:
                {'import': collect, 'paired': paired, 'verify-raw': verify_raw}[args.phase](study, tool, family)


if __name__ == '__main__':
    main()
