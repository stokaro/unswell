#!/usr/bin/env python3
"""Collect the disclosed CLI-compatibility run using the frozen scoring contract."""
import argparse
import json
from pathlib import Path

import evaluate
from cli_compat import trace_usage
from study import ROOT, digest, read, require


def collect(packet, output):
    evaluate.verify_evaluation_freeze(packet)
    amendment = read(output / 'cli-compat-amendment.json')
    original = read(output / 'run-original.json'); current = read(output / 'run.json')
    require(digest((output / 'run-original.json').read_bytes()) == amendment['original_run_sha256'],
            'Original failure ledger changed')
    require(digest((ROOT / 'cli_compat.py').read_bytes()) == amendment['adapter_sha256'] and
            (ROOT / 'cli_compat.py').read_bytes() == (output / 'cli_compat.py').read_bytes(),
            'Compatibility adapter changed')
    require(digest((output / 'cli-compat-amendment.json').read_bytes()) ==
            current['compatibility_amendment_sha256'], 'Compatibility amendment changed')
    require(amendment['prompt_model_effort_schema_labels_thresholds_changed'] is False and
            amendment['new_model_requests_at_most'] == 35, 'Compatibility scope changed')
    require(len(original['calls']) == 1 and original['calls'][0]['status'] == 'invalid',
            'Different original execution')
    old, first = original['calls'][0], current['calls'][0]
    for name in ['page', 'command', 'request_sha256', 'prompt_sha256', 'exit_code', 'seconds']:
        require(old[name] == first[name], 'Original request was replaced')
    require(first['original_error'] == old['error'] and first['original_status'] == old['status'],
            'Original rejection was lost')
    rows = read(packet / 'manifest.json')['pages']
    require(len(current['calls']) <= 36 and
            [c['page'] for c in current['calls']] == [p['page'] for p in rows[:len(current['calls'])]],
            'Changed order, repeated request, or expanded scope')
    # Only trace classification is replaced. Frozen source, model, prompt, and scoring checks remain.
    # Retain the original commands verbatim, but allow offline evidence relocation.
    # Only the two filesystem roots differ; all flags still use the frozen builder.
    args = first['command']
    recorded_packet = Path(args[args.index('--output-schema') + 1]).parent
    first_output = Path(args[args.index('--output-last-message') + 1])
    require(first_output.name == 'response.json' and first_output.parent.name == rows[0]['page'],
            'Original response path changed')
    recorded_output = first_output.parent.parent
    previous_command = evaluate.command
    protocol = read(packet / 'protocol.json')
    for call in current['calls']:
        page = call['page']; command_args = call['command']
        empty = Path(command_args[command_args.index('--cd') + 1])
        require(command_args == previous_command(recorded_packet, recorded_output / page, empty, protocol),
                'Execution flags changed')
        request = packet / 'requests' / (page + '.json')
        require(digest(request.read_bytes()) == call['request_sha256'], 'Request identity changed')
        prompt = (packet / 'prompt.md').read_bytes() + b'\n\nDOCUMENT JSON:\n' + request.read_bytes()
        require(digest(prompt) == call['prompt_sha256'], 'Executed prompt differs')
        if call['status'] == 'invalid' and 'usage' in call:
            trace = (output / page / 'events.jsonl').read_text()
            require(trace_usage(trace) == call['usage'], 'Invalid response trace changed')
            messages = [e['item']['text'] for e in map(json.loads, trace.splitlines())
                        if e.get('type') == 'item.completed' and e['item']['type'] == 'agent_message']
            require(messages and json.loads(messages[-1]) == read(output / page / 'response.json'),
                    'Invalid response differs from its original final message')

    def recorded_command(packet_path, output_path, empty, protocol):
        require(packet_path == packet.resolve() and output_path.parent == output.resolve(),
                'Unexpected offline evidence path')
        return previous_command(recorded_packet, recorded_output / output_path.name, empty, protocol)

    previous = evaluate.trace_usage
    try:
        evaluate.trace_usage = trace_usage
        evaluate.command = recorded_command
        return evaluate.collect(packet, output)
    finally:
        evaluate.trace_usage = previous
        evaluate.command = previous_command


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('packet', type=Path); parser.add_argument('output', type=Path)
    parser.add_argument('--review', type=Path, required=True)
    args = parser.parse_args()
    predictions, valid, rows = collect(args.packet, args.output)
    events, baseline = evaluate.reference()
    result = evaluate.score(events, baseline, predictions, valid, read(args.review),
                            {row['reference'] for row in rows.values()})
    result['execution_amendment'] = 'Exact unstable-feature startup notice classified separately; original failure retained.'
    print(json.dumps(result, indent=2))
