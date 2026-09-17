#!/usr/bin/env python3
"""Replay a complete development or confirmation set with an explicit CLI."""
import argparse
import gzip
import json
import platform
import resource
import subprocess
import sys
import time
from pathlib import Path

sys.dont_write_bytecode = True
ROOT = Path(__file__).resolve().parent.parent
DEVELOPMENT = ROOT.parent / '2026-09-17-full-page-recall'
sys.path.insert(0, str(DEVELOPMENT / 'tools'))
from evaluate import archive_sources, read, require, sha, write


def worker(binary, cwd, profile, output):
    command = [binary, 'check', '--config', profile+'.yaml', '--project-root', '.',
               '--include-source', '--report', 'json:'+str(Path(output).resolve()),
               '--timeout', '5m', 'sources']
    start = time.monotonic()
    run = subprocess.run(command, cwd=cwd, capture_output=True, timeout=360)
    wall = time.monotonic()-start
    usage = resource.getrusage(resource.RUSAGE_CHILDREN)
    Path(output+'.log').write_bytes(run.stdout+run.stderr)
    require(run.returncode in (0, 1), 'Operational failure; inspect log')
    memory_scale = 1 if sys.platform == 'darwin' else 1024
    print(json.dumps(dict(exit_code=run.returncode, wall_seconds=wall,
                         cpu_seconds=usage.ru_utime+usage.ru_stime,
                         max_rss_bytes=usage.ru_maxrss*memory_scale,
                         command=['unswell', *command[1:]])))


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--binary', type=Path, required=True)
    parser.add_argument('--output', type=Path, required=True)
    parser.add_argument('--set', choices=['development', 'confirmation', 'exposed_repetition', 'exposed_framing', 'exposed_local', 'exposed_context', 'exposed_instruction', 'exposed_construction', 'exposed_purpose', 'exposed_verb', 'exposed_relations', 'exposed_projection'], required=True)
    args = parser.parse_args()
    binary = args.binary.resolve(strict=True)
    output = args.output.resolve()
    require(not output.exists(), 'Output must be new')
    for name, digest in read(ROOT/'annotation-freeze.json')['sha256'].items():
        require(sha((ROOT/name).read_bytes()) == digest, 'Annotation/input drift')
    sets = {'exposed_projection': ROOT.parent/'2026-09-17-instruction-projection/confirmation', 'development': DEVELOPMENT, 'confirmation': ROOT/'confirmation',
            'exposed_repetition': ROOT.parent/'2026-09-17-repetition-scope/confirmation',
            'exposed_framing': ROOT.parent/'2026-09-17-framing-recall/confirmation',
            'exposed_local': ROOT.parent/'2026-09-17-local-repetition/confirmation',
            'exposed_relations': ROOT.parent/'2026-09-17-rhetoric-relations/confirmation',
            'exposed_verb': ROOT.parent/'2026-09-17-verb-scaffolding/confirmation',
            'exposed_purpose': ROOT.parent/'2026-09-17-purpose-recall/confirmation',
            'exposed_construction': ROOT.parent/'2026-09-17-construction-recall/confirmation',
            'exposed_instruction': ROOT.parent/'2026-09-17-instruction-recall/confirmation',
            'exposed_context': ROOT.parent/'2026-09-17-context-recall/confirmation'}
    archive = sets[args.set] / 'inputs.tar.gz'
    output.mkdir(parents=True)
    inputs = output/'inputs'
    for name, data in archive_sources(archive).items():
        path = inputs/name
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_bytes(data)
    runs = {}
    for profile in ('technical', 'strict'):
        config = (DEVELOPMENT/(profile+'.yaml')).read_bytes()
        (inputs/(profile+'.yaml')).write_bytes(config)
        target = output/(profile+'.json')
        # A fresh worker has exactly one child, so its rusage is per CLI run.
        result = subprocess.run([sys.executable, str(Path(__file__).resolve()), '--worker',
                                 str(binary), str(inputs), profile, str(target)],
                                check=True, capture_output=True, text=True)
        cost = json.loads(result.stdout)
        report = read(target)
        require(report['status'] == 'complete' and report['manifest']['complete'] and not report['errors'], 'Incomplete run')
        name = profile+'.json.gz'
        compressed = gzip.compress(target.read_bytes(), mtime=0)
        (output/name).write_bytes(compressed)
        runs[profile] = dict(path=name, sha256=sha(compressed), binary_sha256=sha(binary.read_bytes()),
                             config_sha256=sha(config), tool_commit=report['manifest']['tool_commit'],
                             config_identity_sha256=report['manifest']['config_sources'][0]['sha256'],
                             host=platform.platform(), **cost)
        print(f'{profile}: {len(report["documents"])} documents, {len(report["findings"])} findings, exit {cost["exit_code"]}', flush=True)
    write(output/'runs.json', runs)


if __name__ == '__main__':
    if len(sys.argv) > 1 and sys.argv[1] == '--worker':
        worker(*sys.argv[2:])
    else:
        main()
