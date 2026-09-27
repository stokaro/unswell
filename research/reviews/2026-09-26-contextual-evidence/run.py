#!/usr/bin/env python3
"""One-attempt CLI runner; preparation and planning never call a model."""
import argparse
import json
import os
from pathlib import Path
import signal
import subprocess
import tempfile
import time

from study import digest, read, require, validate, verify_freeze, write


def command(packet, destination, workdir, protocol):
    args = ['codex', 'exec', '--ignore-user-config', '--strict-config', '--ephemeral',
            '--skip-git-repo-check', '--sandbox', 'read-only', '--color', 'never',
            '--json', '--model', protocol['model'], '--cd', str(workdir),
            '--output-schema', str(packet / 'response-schema.json'),
            '--output-last-message', str(destination / 'response.json')]
    settings = {
        'model_reasoning_effort': protocol['effort'], 'project_doc_max_bytes': 0,
        'web_search': 'disabled', 'agents.enabled': False, 'cloud.skills.enabled': False,
        'orchestrator.mcp.enabled': False, 'apps._default.enabled': False,
        'features.shell_tool': False, 'features.unified_exec': False,
        'features.apply_patch_freeform': False, 'features.code_mode': False,
        'features.multi_agent': False, 'features.multi_agent_v2': False,
        'features.plugins': False, 'features.plugin_hooks': False,
        'features.skill_search': False, 'features.skip_host_skill_discovery': True,
    }
    for key, value in settings.items():
        args += ['-c', key + '=' + json.dumps(value)]
    return args + ['-']


def trace_usage(raw):
    completed = []
    for line in raw.splitlines():
        item = json.loads(line)
        kind = item.get('type')
        require(kind in ['thread.started', 'turn.started', 'turn.completed',
                         'item.started', 'item.updated', 'item.completed'],
                'Unexpected CLI event or failure')
        if kind in ['item.started', 'item.updated', 'item.completed']:
            require(item['item']['type'] in ['agent_message', 'reasoning'],
                    'A tool or unexpected item appeared in the trace')
        if kind == 'turn.completed':
            completed.append(item)
    require(len(completed) == 1, 'Expected exactly one completed turn')
    return completed[0].get('usage')


def execute(packet, output, authorized=False):
    require(authorized, 'Live model execution is not authorized')
    packet = packet.resolve(strict=True)
    frozen = verify_freeze(packet)
    protocol = read(packet / 'protocol.json')
    manifest = read(packet / 'manifest.json')
    require(len(manifest['pages']) == protocol['pages'] == protocol['max_calls'] == 36,
            'Request scope changed')
    output = output.resolve()
    output.mkdir(parents=True, exist_ok=False)
    started = time.monotonic()
    ledger = {'requested_model': protocol['model'], 'effort': protocol['effort'],
              'freeze_sha256': digest((packet / 'freeze.json').read_bytes()),
              'cli_version': subprocess.check_output(['codex', '--version'], text=True).strip(),
              'monetary_cost': None, 'calls': [], 'complete': False}
    write(output / 'run.json', ledger)
    with tempfile.TemporaryDirectory(prefix='unswell-contextual-empty-') as empty:
        for entry in manifest['pages']:
            remaining = protocol['total_timeout_seconds'] - (time.monotonic() - started)
            if remaining <= 0:
                ledger['stop_reason'] = 'total_time_budget'; break
            request = packet / 'requests' / (entry['page'] + '.json')
            require(digest(request.read_bytes()) == entry['request_sha256'], 'Request changed')
            target = output / entry['page']; target.mkdir()
            args = command(packet, target, Path(empty), protocol)
            prompt = (packet / 'prompt.md').read_bytes() + b'\n\nDOCUMENT JSON:\n' + request.read_bytes()
            record = {'page': entry['page'], 'request_sha256': entry['request_sha256'],
                      'prompt_sha256': digest(prompt), 'command': args, 'status': 'started'}
            ledger['calls'].append(record)
            write(output / 'run.json', ledger)
            before = time.monotonic()
            proc = subprocess.Popen(args, stdin=subprocess.PIPE, stdout=subprocess.PIPE,
                                    stderr=subprocess.PIPE, start_new_session=True)
            try:
                stdout, stderr = proc.communicate(prompt, timeout=min(remaining, protocol['per_call_timeout_seconds']))
            except subprocess.TimeoutExpired:
                os.killpg(proc.pid, signal.SIGTERM)
                try:
                    stdout, stderr = proc.communicate(timeout=10)
                except subprocess.TimeoutExpired:
                    os.killpg(proc.pid, signal.SIGKILL)
                    stdout, stderr = proc.communicate()
                record['status'] = 'timeout'
            (target / 'events.jsonl').write_bytes(stdout)
            (target / 'stderr.txt').write_bytes(stderr)
            record.update(exit_code=proc.returncode, seconds=time.monotonic()-before)
            try:
                require(record['status'] != 'timeout' and proc.returncode == 0, 'CLI did not complete')
                record['usage'] = trace_usage(stdout.decode())
                response = target / 'response.json'
                require(response.stat().st_size <= 8 << 20, 'Response exceeds byte limit')
                result = validate(read(request), read(response))
                write(target / 'bound.json', result)
                record['status'] = 'valid'
            except (ValueError, KeyError, TypeError, OSError) as error:
                record.update(status='invalid', error=str(error))
            write(output / 'run.json', ledger)
            print(entry['page'], record['status'], flush=True)
            if record['status'] != 'valid':
                ledger['stop_reason'] = 'first_failed_request_no_retry'; break
    require(verify_freeze(packet) == frozen, 'Frozen packet changed during execution')
    ledger['complete'] = len(ledger['calls']) == protocol['pages'] and all(
        x['status'] == 'valid' for x in ledger['calls'])
    ledger['unattempted'] = [p['page'] for p in manifest['pages'] if p['page'] not in
                            {c['page'] for c in ledger['calls']}]
    write(output / 'run.json', ledger)
    return ledger['complete']


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('packet', type=Path)
    parser.add_argument('--output', type=Path)
    parser.add_argument('--authorize-model-calls', action='store_true')
    args = parser.parse_args()
    verify_freeze(args.packet)
    if not args.authorize_model_calls:
        protocol = read(args.packet / 'protocol.json')
        print(json.dumps({'mode': 'offline_plan', 'model_calls': 0, 'proposed': protocol}, indent=2))
        return
    require(args.output is not None, 'A new output directory is required')
    if not execute(args.packet, args.output, authorized=True):
        raise SystemExit(2)


if __name__ == '__main__':
    main()
