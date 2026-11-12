#!/usr/bin/env python3
"""Answer one prompt from stdin with one panel model and cache the answer.

usage: ask_model.py <claude|astra|qwen> <cache dir>

The review command starts this program once per question. The answer is the
model's plain text; no output schema is set, so the command's own parser is
what reads it. A cached answer is reused for an identical prompt, which lets
several confirming models judge the same first review.
"""
import hashlib, os, subprocess, sys, tempfile
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))
import annotate as base

kind, cache = sys.argv[1], Path(sys.argv[2])
prompt = sys.stdin.read()
cache.mkdir(parents=True, exist_ok=True)
target = cache / (hashlib.sha256((kind + "\n" + prompt).encode()).hexdigest() + ".txt")
if target.exists():
    sys.stdout.write(target.read_text())
    sys.exit(0)
spec = base.ANNOTATORS[kind]
with tempfile.TemporaryDirectory() as scratch:
    if spec["kind"] == "codex":
        answer = Path(scratch) / "answer.txt"
        args = ["codex", "exec", *base.CODEX_OPTIONS, "--model", spec["model"], "--cd", scratch,
                "--output-last-message", str(answer), "-c", f'model_reasoning_effort="{spec["effort"]}"']
        for setting in base.CODEX_SETTINGS:
            args += ["-c", setting]
        base.run(args + ["-"], prompt.encode(), scratch)
        text = answer.read_text()
    elif spec["kind"] == "claude":
        args = ["claude", "-p", "--model", spec["model"], "--tools", "", "--strict-mcp-config",
                "--setting-sources", "", "--no-session-persistence",
                "--system-prompt", "You answer with the requested JSON object and nothing else."]
        text = base.run(args, prompt.encode(), scratch).decode()
    elif spec["kind"] == "router":
        import json, urllib.request
        body = {"model": os.environ["ROUTER_MODEL"], "temperature": 0, "max_tokens": 24000, "stream": False,
                "messages": [{"role": "user", "content": prompt}]}
        request = urllib.request.Request(os.environ["ROUTER_URL"], data=json.dumps(body).encode(),
                                         headers={"Content-Type": "application/json",
                                                  "Authorization": "Bearer " + os.environ["ROUTER_KEY"]})
        with urllib.request.urlopen(request, timeout=base.TIMEOUT) as response:
            text = json.load(response)["choices"][0]["message"].get("content") or ""
    else:
        raise SystemExit("unsupported kind " + kind)
if not text.strip():
    raise SystemExit("empty answer")
temporary = target.with_suffix(f".{os.getpid()}.tmp")
temporary.write_text(text)
temporary.rename(target)
sys.stdout.write(text)
