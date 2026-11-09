#!/usr/bin/env python3
"""Pin fixed public README sources before the offline #349 transfer screen."""

import base64
import hashlib
import json
import subprocess
from pathlib import Path

HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[2]
LOCAL = ROOT / "artifacts/research/2026-11-09-shape-context-transfer/sources"
REPOSITORIES = (
    "go-chi/chi", "go-task/task", "spf13/cobra", "jackc/pgx",
    "rs/zerolog", "nats-io/nats.go", "helm/helm", "golang-migrate/migrate",
    "jmoiron/sqlx", "grpc/grpc-go", "kubernetes-sigs/kustomize",
    "go-playground/validator",
)


def api(path):
    run = subprocess.run(["gh", "api", path], capture_output=True, timeout=30,
                         check=False)
    if run.returncode:
        raise RuntimeError(f"GitHub API request failed: {path}: {run.stderr.decode('utf-8', 'replace')[:200]}")
    return json.loads(run.stdout)


def git_blob_sha1(raw):
    return hashlib.sha1(f"blob {len(raw)}\0".encode() + raw).hexdigest()


def main():
    if LOCAL.exists() or (HERE / "sources.json").exists():
        raise RuntimeError("source output already exists; no overwrite or retry")
    LOCAL.mkdir(parents=True)
    records = []
    for ordinal, repository in enumerate(REPOSITORIES, 1):
        metadata = api(f"repos/{repository}")
        branch = metadata["default_branch"]
        commit = api(f"repos/{repository}/commits/{branch}")["sha"]
        item = api(f"repos/{repository}/contents/README.md?ref={commit}")
        if item["type"] != "file" or item["encoding"] != "base64":
            raise RuntimeError(f"README is not an inline base64 file: {repository}")
        raw = base64.b64decode(item["content"], validate=False)
        raw.decode("utf-8", "strict")
        if len(raw) != item["size"] or git_blob_sha1(raw) != item["sha"]:
            raise RuntimeError(f"GitHub source identity mismatch: {repository}")
        filename = f"page-{ordinal:03d}.source"
        (LOCAL / filename).write_bytes(raw)
        records.append({
            "number": ordinal,
            "partition": "design" if ordinal <= 6 else "transfer",
            "repository": repository,
            "path": item["path"],
            "default_branch": branch,
            "commit_sha": commit,
            "git_blob_sha1": item["sha"],
            "source_sha256": hashlib.sha256(raw).hexdigest(),
            "source_bytes": len(raw),
            "source_file": filename,
        })
    output = {"format": "unswell-349-shape-context-sources-v1", "sources": records}
    (HERE / "sources.json").write_text(json.dumps(output, indent=2, sort_keys=True) + "\n")


if __name__ == "__main__":
    main()
