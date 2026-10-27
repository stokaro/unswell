#!/usr/bin/env python3
"""Run a separately authorized diagnosis-first pilot once, if approved."""
import argparse
import json
import os
from pathlib import Path
import signal
import subprocess
import sys
import tempfile
import time

import diagnosis as a

# Reuse the checked command and startup-notice classifier verbatim. These files
# are bound by the separate evaluation freeze before any new execution.
sys.path.insert(0, str(a.OLD.PRIOR))
import run as prior_run
import cli_compat
sys.path.pop(0)


def verify(packet):
    a.REFERENCE.verify_freeze(packet)
    frozen = a.read(a.ROOT / 'evaluation-freeze.json')
    a.require(a.digest((packet / 'freeze.json').read_bytes()) == frozen['packet_freeze_sha256'],
              'Different prepared packet')
    for name, expected in frozen['files'].items():
        path = Path(name)
        a.require(not path.is_absolute() and '..' not in path.parts, 'Unsafe dependency path')
        a.require(a.digest((a.ROOT.parents[2] / path).read_bytes()) == expected,
                  'Frozen dependency changed: ' + name)
    for name in ['diagnosis.py', 'execute.py', 'evaluate.py', 'freeze.py',
                 'prompt.md', 'protocol.json', 'DESIGN.md', 'README.md']:
        a.require((packet / name).read_bytes() == (a.ROOT / name).read_bytes(), 'Execution file changed')
    manifest = a.read(packet / 'manifest.json'); protocol = a.read(packet / 'protocol.json')
    source_manifest_path = packet / 'source-manifest.json'
    source_manifest = a.read(source_manifest_path)
    source_rows = {row['page']: row for row in source_manifest['pages']}
    source_references = [row['reference'] for row in source_manifest['pages']]
    a.require(a.digest(source_manifest_path.read_bytes()) == manifest['source_manifest_sha256'] and
              len(source_rows) == 36 and source_references == manifest['source_references'] and
              len(set(source_references)) == 36 and
              sum(ref.startswith('whole/') for ref in source_references) == 26 and
              sum(ref.startswith('context/') for ref in source_references) == 6 and
              sum(ref.startswith('confirmation/') for ref in source_references) == 4,
              'Original page denominators changed')
    a.require(a.read(packet / 'response-schema.json') == a.schema(),
              'Response schema differs from validator')
    a.require(protocol['model'] == 'gpt-6-astra' and protocol['effort'] == 'medium' and
              protocol['concurrency'] == 1 and protocol['attempts_per_request'] == 1 and
              protocol['automatic_retries'] is False and protocol['tools_allowed'] is False and
              protocol['suggestions_in_responses'] is False and
              protocol['per_call_timeout_seconds'] == 600 and
              protocol['total_timeout_seconds'] == 5400,
              'Changed execution contract')
    a.require(len(manifest['requests']) == protocol['max_calls'] == 7 and
              manifest['source_pages'] == protocol['source_pages'] == 36 and
              manifest['selected_pages'] == protocol['pilot_pages'] == 6 and
              manifest['model_calls'] == 0 and
              manifest['max_focus_units'] == protocol['max_focus_units'] == a.MAX_UNITS and
              sum(row['focus_units'] for row in manifest['requests']) == 310,
              'Changed request budget')
    a.require(manifest['prior_packet_freeze_sha256'] ==
              a.read(a.OLD.PRIOR / 'evaluation-freeze.json')['packet_freeze_sha256'] and
              manifest['accounting_code_sha256'] ==
              a.digest((a.ACCOUNTING / 'accounting.py').read_bytes()) and
              manifest['source_binder_sha256'] ==
              a.digest((a.OLD.PRIOR / 'study.py').read_bytes()),
              'Source ancestry changed')
    grouped = {}; references = {}; ordered_refs = []
    for row in manifest['requests']:
        request_path = packet / 'requests' / (row['request'] + '.json')
        a.require(a.digest(request_path.read_bytes()) == row['request_sha256'], 'Changed request')
        request = a.read(request_path); a.OLD.verify_request(request)
        a.require(request['request'] == row['request'] and request['page'] == row['page'] and
                  request['source_sha256'] == row['source_sha256'] and
                  source_rows[row['page']]['reference'] == row['reference'] and
                  source_rows[row['page']]['source_sha256'] == row['source_sha256'] and
                  len(request['focus']) == row['focus_units'] and
                  request_path.stat().st_size == row['request_bytes'],
                  'Changed manifest identity')
        if row['page'] not in grouped:
            references[row['page']] = row['reference']
            ordered_refs.append(row['reference'])
        else:
            a.require(references[row['page']] == row['reference'], 'Changed page reference')
        grouped.setdefault(row['page'], []).append(request)
    a.require(len(grouped) == 6 and ordered_refs == a.PAGES,
              'Missing or reordered context pages')
    for rows in grouped.values():
        a.require(rows == a.OLD.requests(a.OLD.verify_request(rows[0])),
                  'Changed page partition')
    return manifest, protocol


def prompt(packet, request):
    return (packet / 'prompt.md').read_bytes() + b'\n\nDOCUMENT JSON:\n' + request.read_bytes()


def artifacts(target):
    """Hash raw CLI output, including invalid and incomplete responses."""
    result = {}
    for name in ['events.jsonl', 'stderr.txt', 'response.json']:
        path = target / name
        result[name] = a.digest(path.read_bytes()) if path.is_file() else None
    return result


def check_response(packet, output, row, call):
    target = output / row['request']; raw = (target / 'events.jsonl').read_text()
    call['usage'] = cli_compat.trace_usage(raw)
    messages = [e['item']['text'] for e in map(json.loads, raw.splitlines())
                if e.get('type') == 'item.completed' and e['item']['type'] == 'agent_message']
    path = target / 'response.json'
    a.require(path.stat().st_size <= 8 << 20, 'Response byte limit exceeded')
    response = a.read(path)
    a.require(messages and json.loads(messages[-1]) == response, 'Saved response differs from trace')
    return a.validate(a.read(packet / 'requests' / (row['request'] + '.json')), response)


def execute(packet, output, authorized=False):
    a.require(authorized, 'A new explicit model-run authorization is required')
    packet = packet.resolve(strict=True); manifest, protocol = verify(packet)
    output = output.resolve(); output.mkdir(parents=True, exist_ok=False)
    ledger = {'model': protocol['model'], 'effort': protocol['effort'],
              'packet_freeze_sha256': a.digest((packet / 'freeze.json').read_bytes()),
              'cli_version': subprocess.check_output(['codex', '--version'], text=True).strip(),
              'monetary_cost': None, 'calls': [], 'complete': False}
    a.write(output / 'run.json', ledger); started = time.monotonic()
    with tempfile.TemporaryDirectory(prefix='unswell-diagnosis-empty-') as empty:
        for row in manifest['requests']:
            remaining = protocol['total_timeout_seconds'] - (time.monotonic() - started)
            if remaining <= 0:
                ledger['stop_reason'] = 'total_time_budget'; break
            target = output / row['request']; target.mkdir()
            request = packet / 'requests' / (row['request'] + '.json')
            message = prompt(packet, request)
            command = prior_run.command(packet, target, Path(empty), protocol)
            call = {'request': row['request'], 'page': row['page'], 'status': 'started',
                    'request_sha256': row['request_sha256'], 'prompt_sha256': a.digest(message),
                    'command': command}
            ledger['calls'].append(call); a.write(output / 'run.json', ledger)
            before = time.monotonic(); timed_out = False
            try:
                proc = subprocess.Popen(command, stdin=subprocess.PIPE, stdout=subprocess.PIPE,
                                        stderr=subprocess.PIPE, start_new_session=True)
                try:
                    stdout, stderr = proc.communicate(message, timeout=min(
                        remaining, protocol['per_call_timeout_seconds']))
                except subprocess.TimeoutExpired:
                    timed_out = True; os.killpg(proc.pid, signal.SIGTERM)
                    try: stdout, stderr = proc.communicate(timeout=10)
                    except subprocess.TimeoutExpired:
                        os.killpg(proc.pid, signal.SIGKILL); stdout, stderr = proc.communicate()
                (target / 'events.jsonl').write_bytes(stdout)
                (target / 'stderr.txt').write_bytes(stderr)
                call.update(exit_code=proc.returncode, seconds=time.monotonic() - before,
                            timed_out=timed_out)
                a.require(not timed_out and proc.returncode == 0, 'CLI did not complete')
                a.write(target / 'bound.json', check_response(packet, output, row, call))
                call['status'] = 'valid'
            except (ValueError, KeyError, TypeError, OSError) as error:
                call.update(status='invalid', error=str(error), seconds=time.monotonic() - before)
            call['artifacts'] = artifacts(target)
            a.write(output / 'run.json', ledger)
            print(row['request'], call['status'], flush=True)
            if call['status'] != 'valid':
                ledger['stop_reason'] = 'first_failed_request_no_retry'; break
    verify(packet)
    ledger['complete'] = len(ledger['calls']) == len(manifest['requests']) and all(
        c['status'] == 'valid' for c in ledger['calls'])
    ledger['unattempted'] = [r['request'] for r in manifest['requests'][len(ledger['calls']):]]
    a.write(output / 'run.json', ledger)
    return ledger['complete']


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('packet', type=Path); parser.add_argument('--output', type=Path)
    parser.add_argument('--authorize-model-calls', action='store_true')
    args = parser.parse_args(); manifest, protocol = verify(args.packet)
    if not args.authorize_model_calls:
        print(json.dumps({'mode': 'offline_plan', 'model_calls': 0, 'proposed': protocol}, indent=2))
    else:
        a.require(args.output is not None, 'Supply a new output directory')
        if not execute(args.packet, args.output, True): raise SystemExit(2)
