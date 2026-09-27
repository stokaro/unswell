#!/usr/bin/env python3
"""Retain and classify one CLI startup notice without retrying model requests."""
import argparse
import json
import os
from pathlib import Path
import signal
import subprocess
import tempfile
import time

import evaluate
import run
from study import digest, read, require, validate, verify_freeze, write

NOTICE = ('Under-development features enabled: skip_host_skill_discovery. '
          'Under-development features are incomplete and may behave unpredictably. '
          'To suppress this warning, set `suppress_unstable_features_warning = true` '
          'in /Users/buster/.codex/config.toml.')


def trace_usage(raw):
    events = [json.loads(line) for line in raw.splitlines()]
    retained = []; removed = 0; started = False
    for event in events:
        if event.get('type') == 'turn.started': started = True
        item = event.get('item', {})
        if (event.get('type') == 'item.completed' and item.get('type') == 'error' and
                item.get('message') == NOTICE):
            require(not started and removed == 0 and len(retained) == 1 and
                    retained[0].get('type') == 'thread.started', 'Notice outside startup position')
            removed += 1
        else:
            retained.append(event)
    return run.trace_usage('\n'.join(json.dumps(event) for event in retained))


def check_answer(packet, output, row, call):
    path = output / row['page']
    require(call['exit_code'] == 0, 'CLI did not complete')
    call['usage'] = trace_usage((path / 'events.jsonl').read_text())
    response = path / 'response.json'
    require(response.stat().st_size <= 8 << 20, 'Response exceeds byte limit')
    result = validate(read(packet / 'requests' / (row['page'] + '.json')), read(response))
    messages = [event['item']['text'] for event in
                map(json.loads, (path / 'events.jsonl').read_text().splitlines())
                if event.get('type') == 'item.completed' and event['item']['type'] == 'agent_message']
    require(messages and json.loads(messages[-1]) == read(response), 'Final response differs from trace')
    write(path / 'bound.json', result)
    call['status'] = 'valid'


def continue_run(packet, output, authorized):
    require(authorized, 'Explicit execution authorization required')
    packet = packet.resolve(); output = output.resolve()
    frozen = verify_freeze(packet); evaluate.verify_evaluation_freeze(packet)
    protocol = read(packet / 'protocol.json'); rows = read(packet / 'manifest.json')['pages']
    ledger = read(output / 'run.json')
    require(len(rows) == protocol['max_calls'] == 36 and len(ledger['calls']) == 1,
            'Continuation is only for the single recorded startup-notice stop')
    require(ledger['calls'][0]['page'] == rows[0]['page'] and ledger['calls'][0]['exit_code'] == 0 and
            ledger['calls'][0]['error'] == 'A tool or unexpected item appeared in the trace' and
            ledger['stop_reason'] == 'first_failed_request_no_retry', 'Different failure requires investigation')
    require(not (output / 'run-original.json').exists(), 'This continuation was already attempted')
    amendment = read(output / 'cli-compat-amendment.json')
    require(amendment['adapter_sha256'] == digest(Path(__file__).read_bytes()), 'Adapter changed')
    require(amendment['original_run_sha256'] == digest((output / 'run.json').read_bytes()), 'Original run changed')
    (output / 'run-original.json').write_bytes((output / 'run.json').read_bytes())
    first = ledger['calls'][0]
    check_answer(packet, output, rows[0], first)
    first['original_status'] = 'invalid'; first['original_error'] = first.pop('error')
    ledger['compatibility_amendment_sha256'] = digest((output / 'cli-compat-amendment.json').read_bytes())
    ledger.pop('stop_reason'); ledger['complete'] = False
    write(output / 'run.json', ledger)
    with tempfile.TemporaryDirectory(prefix='unswell-contextual-empty-') as empty:
        for row in rows[1:]:
            remaining = amendment['original_deadline_unix'] - time.time()
            if remaining <= 0:
                ledger['stop_reason'] = 'original_total_time_budget'; break
            page = row['page']; target = output / page; target.mkdir(exist_ok=False)
            request = packet / 'requests' / (page + '.json')
            require(digest(request.read_bytes()) == row['request_sha256'], 'Request changed')
            args = run.command(packet, target, Path(empty), protocol)
            prompt = (packet / 'prompt.md').read_bytes() + b'\n\nDOCUMENT JSON:\n' + request.read_bytes()
            call = {'page': page, 'request_sha256': row['request_sha256'],
                    'prompt_sha256': digest(prompt), 'command': args, 'status': 'started'}
            ledger['calls'].append(call); write(output / 'run.json', ledger)
            before = time.monotonic()
            proc = subprocess.Popen(args, stdin=subprocess.PIPE, stdout=subprocess.PIPE,
                                    stderr=subprocess.PIPE, start_new_session=True)
            timed_out = False
            try:
                stdout, stderr = proc.communicate(prompt, timeout=min(remaining, protocol['per_call_timeout_seconds']))
            except subprocess.TimeoutExpired:
                timed_out = True; os.killpg(proc.pid, signal.SIGTERM)
                try: stdout, stderr = proc.communicate(timeout=10)
                except subprocess.TimeoutExpired:
                    os.killpg(proc.pid, signal.SIGKILL); stdout, stderr = proc.communicate()
            (target / 'events.jsonl').write_bytes(stdout); (target / 'stderr.txt').write_bytes(stderr)
            call.update(exit_code=proc.returncode, seconds=time.monotonic()-before)
            try:
                require(not timed_out, 'Request timed out')
                check_answer(packet, output, row, call)
            except (ValueError, KeyError, TypeError, OSError) as error:
                call.update(status='invalid', error=str(error))
            write(output / 'run.json', ledger); print(page, call['status'], flush=True)
            if call['status'] != 'valid':
                ledger['stop_reason'] = 'first_failed_request_no_retry'; break
    require(verify_freeze(packet) == frozen, 'Frozen packet changed')
    ledger['complete'] = len(ledger['calls']) == 36 and all(c['status'] == 'valid' for c in ledger['calls'])
    ledger['unattempted'] = [p['page'] for p in rows if p['page'] not in {c['page'] for c in ledger['calls']}]
    write(output / 'run.json', ledger)
    return ledger['complete']


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('packet', type=Path); parser.add_argument('output', type=Path)
    parser.add_argument('--authorize-model-calls', action='store_true')
    args = parser.parse_args()
    if not continue_run(args.packet, args.output, args.authorize_model_calls): raise SystemExit(2)
