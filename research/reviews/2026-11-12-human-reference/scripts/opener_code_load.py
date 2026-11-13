#!/usr/bin/env python3
"""Count the findings of scaffold.chat-preamble on source code of open-source packages.

usage: opener_code_load.py <unswell binary> <out json> <root> [<root> ...]

A finding of the listed openers needs the text "sure," or "sure!" in the file,
in any case. The script lists the source files under the roots, keeps those
that contain that text, and runs the rule on each of them from standard input.
It writes the number of source files, of candidates, of candidates that failed
to run with the reasons, and the findings with their lines.
"""
import collections, concurrent.futures, json, re, subprocess, sys, tempfile
from pathlib import Path

EXTENSIONS = {".go", ".py", ".js", ".ts", ".tsx", ".rs", ".java", ".c", ".cc", ".cpp", ".h", ".cs", ".yaml", ".yml", ".sh"}
TEXT = re.compile(rb"sure[,!]", re.I)
POLICY = "version: 1\nextends: [builtin:custom]\nrules:\n  scaffold.chat-preamble:\n    enabled: true\n"
binary, out, roots = sys.argv[1], Path(sys.argv[2]), sys.argv[3:]

files = [path for root in roots for path in Path(root).rglob("*") if path.suffix in EXTENSIONS and path.is_file()]


def candidate(path):
    try:
        return bool(TEXT.search(path.read_bytes()))
    except OSError:
        return False


def run(path):
    data = path.read_bytes()
    with tempfile.TemporaryDirectory() as scratch:
        Path(scratch, "policy.yaml").write_text(POLICY)
        done = subprocess.run([binary, "check", "--stdin", "--filename", path.name, "--config", "policy.yaml",
                               "--no-gate", "--allow-empty", "--report", "json:-"], input=data, cwd=scratch, capture_output=True)
    if done.returncode != 0:
        message = done.stderr.decode("utf-8", "replace").strip().splitlines()[-1]
        return path, None, "exceeds max_file_bytes" if "max_file_bytes" in message else "does not parse"
    found = []
    for finding in json.loads(done.stdout).get("findings") or []:
        span = finding["primary"]["span"]
        found.append({"line": data.count(b"\n", 0, span["start"]) + 1, "phrase": data[span["start"]:span["end"]].decode("utf-8", "replace")})
    return path, found, None


with concurrent.futures.ThreadPoolExecutor(8) as pool:
    candidates = [path for path, keep in zip(files, pool.map(candidate, files)) if keep]
    results = list(pool.map(run, candidates))
failures = collections.Counter(reason for _, _, reason in results if reason)
findings = [{"file": str(path), **item} for path, found, _ in results for item in found or []]
summary = {"files": len(files), "by_extension": dict(collections.Counter(path.suffix for path in files).most_common()),
           "candidates": len(candidates), "failed": sum(failures.values()), "failures": dict(failures), "findings": findings}
out.write_text(json.dumps(summary, indent=1, sort_keys=True) + "\n")
print({key: value for key, value in summary.items() if key != "findings"}, "findings", len(findings))
