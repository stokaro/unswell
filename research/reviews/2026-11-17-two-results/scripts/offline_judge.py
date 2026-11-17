#!/usr/bin/env python3
"""Answer a prompt of `unswell compare` from the kept answers of a judge, and never ask a model.

usage: offline_judge.py <kind> <cache dir>     (the prompt on standard input)

A judge that is a local program keeps each answer under a hash of its kind and the prompt. This program gives the
kept answer for the same prompt. For a prompt without a kept answer it fails, so that a replay cannot reach a model.
"""
import hashlib, sys
from pathlib import Path

kind, cache = sys.argv[1], Path(sys.argv[2])
prompt = sys.stdin.buffer.read()
kept = cache / (hashlib.sha256((kind + "\n" + prompt.decode()).encode()).hexdigest() + ".txt")
if not kept.exists():
    sys.exit("offline_judge: no kept answer for this prompt")
sys.stdout.buffer.write(kept.read_bytes())
