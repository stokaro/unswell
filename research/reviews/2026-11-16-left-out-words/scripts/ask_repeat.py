#!/usr/bin/env python3
"""Answer one prompt with a panel model through ask_model.py; a repeated prompt within one run is asked again.

usage: ask_repeat.py <ask_model.py> <kind> <cache dir> <run directory>

ask_model.py keeps each answer and gives it again for the same prompt. A command that repeats a question after an
answer it could not read would get that answer again. The run directory remembers which prompts this run has asked:
the second time a prompt comes, the kept answer is dropped first.
"""
import hashlib, os, subprocess, sys
from pathlib import Path

program, kind, cache, run = sys.argv[1], sys.argv[2], Path(sys.argv[3]), Path(sys.argv[4])
prompt = sys.stdin.buffer.read()
key = hashlib.sha256((kind + "\n" + prompt.decode()).encode()).hexdigest()
run.mkdir(parents=True, exist_ok=True)
asked = run / key
if asked.exists():
    kept = cache / (key + ".txt")
    if kept.exists():
        kept.unlink()
asked.touch()
done = subprocess.run([sys.executable, program, kind, str(cache)], input=prompt, capture_output=True)
sys.stdout.buffer.write(done.stdout)
sys.stderr.buffer.write(done.stderr)
sys.exit(done.returncode)
