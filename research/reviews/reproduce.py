#!/usr/bin/env python3
"""Reproduce the September 13 diagnostic replay using existing local clones.

Build preparation may use the Go module cache or module proxy. Every scan runs
locally without a model server. No generation, annotation, or training occurs.
The output directory must be new. Temporary linked worktrees are removed on exit.
"""

import argparse
from contextlib import ExitStack
import hashlib
import json
import os
from pathlib import Path
import subprocess
import sys

sys.dont_write_bytecode = True
from compare import PAIRS, compare, digest, require, write


SUBJECTS = {"ptah": "7b47e7cfb5d4ff32a38375345069f5533bff892f",
            "unswell": "a1c6d7af67a89ab573997aba9cec7b51a09f4dbb"}
ENGINES = {"old-a1c6d7a": SUBJECTS["unswell"],
           "old-d3c9f23": "d3c9f231f372694c2f38307e209e65503bf3b1e1",
           "new": "06b38615d4483970b65a6fbe2187aeb4864830dc"}
ENV = {**os.environ, "CGO_ENABLED": "0", "GOMAXPROCS": "2", "GOFLAGS": "-p=2", "GOWORK": "off"}
HERE = Path(__file__).resolve().parent


def git(repo, *args):
    return subprocess.check_output(["git", "-C", str(repo), *args])


def checkout(stack, repo, commit, destination):
    git(repo, "worktree", "add", "--detach", str(destination), commit)
    stack.callback(git, repo, "worktree", "remove", str(destination))
    return destination


def repository_manifest(root, subject):
    sources = []
    for raw in git(root, "ls-files", "-z").split(b"\0")[:-1]:
        name = raw.decode("utf-8")
        path = root / name
        require(path.resolve().is_relative_to(root), f"Source escapes checkout: {name}")
        sources.append({"path": name, "bytes": path.stat().st_size, "sha256": digest(path)})
    return {"version": "unswell-review-replay-sources-v1", "repository": f"stokaro/{subject}",
            "commit": SUBJECTS[subject], "sources": sources}


def responses(repo, root):
    root.mkdir()
    rows, seen, inputs = [], set(), {}
    commit = SUBJECTS["unswell"]

    def source(path):
        return git(repo, "show", f"{commit}:{path}")

    for run in ("2026-09-11-pilot", "2026-09-11-run2"):
        prefix = f"research/generation/runs/{run}"
        raw = source(prefix + "/records.json")
        inputs[prefix + "/records.json"] = hashlib.sha256(raw).hexdigest()
        records = json.loads(raw)["records"]
        tasks = {t["id"]: t for t in json.loads(source(prefix + "/tasks.json"))["tasks"]}
        declared = {}
        for shard in git(repo, "ls-tree", "-r", "--name-only", commit, prefix + "/shards").decode().splitlines():
            for item in json.loads(source(shard))["sources"]:
                key = item["repository"], item["path"]
                require(key not in declared, "Duplicate shard declaration")
                declared[key] = item
        for record in records:
            require(record["status"] == "complete", "Incomplete original response")
            data = record["text"].encode("utf-8")
            require(hashlib.sha256(data).hexdigest() == record["output_sha256"], "Response text hash mismatch")
            data += b"\n"
            repository = tasks[record["task_id"]]["repository"]
            relative = f"generated/{record['response_id']}.md"
            item = declared[repository, relative]
            sha = hashlib.sha256(data).hexdigest()
            require(item["sha256"] == sha and item["bytes"] == len(data), "Original shard mismatch")
            name = repository.replace("/", "__") + "/" + relative
            require(name not in seen, "Cross-run response path collision")
            seen.add(name)
            destination = root / name
            require(destination.resolve().is_relative_to(root), "Response path escapes root")
            destination.parent.mkdir(parents=True, exist_ok=True)
            destination.write_bytes(data)
            rows.append({"path": name, "sha256": sha, "bytes": len(data), "run": run,
                         **{key: record[key] for key in ("response_id", "task_id", "operation", "prompt")}})
    require(len(rows) == 800, "Expected the original 800 responses")
    return {"version": "unswell-review-replay-sources-v1", "source_commit": commit,
            "reconstruction": "The original importer writes record.text followed by one LF.",
            "records_sha256": inputs, "sources": sorted(rows, key=lambda row: row["path"])}


def scan(binary, source, paths, policy, directory):
    directory.mkdir(parents=True)
    policy_path = source / ".unswell-replay.yaml"
    require(not policy_path.exists(), "Policy scratch file already exists")
    policy_path.write_bytes(policy)
    try:
        args = [str(binary), "check", *paths, "--config", str(policy_path), "--no-gate", "--include-source",
                "--timeout", "20m", "--jobs", "2", "--report", f"json:{directory}/result.json"]
        with (directory / "stdout.txt").open("w") as stdout, (directory / "stderr.txt").open("w") as stderr:
            process = subprocess.run(args, cwd=source, env=ENV, stdout=stdout, stderr=stderr, timeout=1300, check=False)
        write(directory / "invocation.json", {"args": args, "exit_code": process.returncode,
                                             "binary_sha256": digest(binary),
                                             "policy_sha256": hashlib.sha256(policy).hexdigest()})
        require(process.returncode in (0, 2), f"Unexpected scan exit {process.returncode}")
        require((directory / "result.json").is_file(), "Scan produced no report")
    finally:
        policy_path.unlink()


def run(unswell_repo, ptah_repo, output):
    output.mkdir(parents=True, exist_ok=False)
    (output / "bin").mkdir()
    roots = output / "worktrees"
    roots.mkdir()
    write(output / "environment.json", {"go": subprocess.check_output(["go", "version"], env=ENV, text=True).strip(),
                                       "CGO_ENABLED": "0", "GOMAXPROCS": "2", "GOFLAGS": "-p=2", "GOWORK": "off",
                                       "engine_commits": ENGINES, "subject_commits": SUBJECTS})
    with ExitStack() as stack:
        engine_roots = {}
        for name, commit in ENGINES.items():
            root = checkout(stack, unswell_repo, commit, roots / name)
            engine_roots[name] = root
            subprocess.run(["go", "build", "-trimpath", "-buildvcs=false", "-ldflags",
                            f"-X github.com/stokaro/unswell.BuildCommit={commit}",
                            "-o", str(output / "bin" / name), "./cmd/unswell"], cwd=root, env=ENV, check=True)
        subjects = {"unswell": engine_roots["old-a1c6d7a"],
                    "ptah": checkout(stack, ptah_repo, SUBJECTS["ptah"], roots / "ptah")}
        for subject, root in subjects.items():
            write(output / f"{subject}-sources.json", repository_manifest(root, subject))
        subjects["responses"] = output / "response-inputs"
        manifest = responses(unswell_repo, subjects["responses"])
        write(output / "responses-sources.json", manifest)
        for subject, profile, old in PAIRS:
            policy = (b"version: 1\nextends: [builtin:technical]\n" if profile == "default" else
                      (HERE / f"2026-09-11-{subject}" / f"config-{profile}.yaml").read_bytes())
            paths = [item["path"] for item in manifest["sources"]] if subject == "responses" else ["."]
            directory = output / subject if subject == "responses" else output / subject / profile
            for engine in (old, "new"):
                print(f"Scanning {subject}/{profile}/{engine}", flush=True)
                scan(output / "bin" / engine, subjects[subject], paths, policy, directory / engine)
    compare(output, output / "compared")


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--unswell", type=Path, required=True, help="Existing local Unswell clone containing pinned commits")
    parser.add_argument("--ptah", type=Path, required=True, help="Existing local Ptah clone containing the pinned commit")
    parser.add_argument("--output", type=Path, required=True, help="New directory for binaries, inputs, raw reports, and comparisons")
    options = parser.parse_args()
    run(options.unswell.resolve(), options.ptah.resolve(), options.output.resolve())
