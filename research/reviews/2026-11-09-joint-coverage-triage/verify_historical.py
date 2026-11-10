"""Verify the frozen pilot against its historical Makefile, not today's CI recipe."""

import hashlib
import json
import subprocess

import run


# The squash merge that published this pilot retains its frozen Makefile blob.
# The pre-inference packet and its original verifier remain unchanged.
PUBLISHED_COMMIT = "e1aba8105d2062dbb09397e28292367f8edc95ca"


def compare_frozen(frozen, current, historical_makefile_digest):
    if set(frozen) != set(current):
        raise ValueError("frozen file set changed")
    for path, expected in frozen.items():
        actual = historical_makefile_digest if path == "Makefile" else current[path]
        if actual != expected:
            raise ValueError(f"frozen file hash mismatch: {path}")


def historical_makefile():
    result = subprocess.run(
        ["git", "show", f"{PUBLISHED_COMMIT}:Makefile"], cwd=run.REPO,
        check=True, capture_output=True,
    )
    return hashlib.sha256(result.stdout).hexdigest()


def verify():
    run.validate_inputs()
    frozen = run.read(run.PACKET / "freeze.json")
    compare_frozen(frozen, run.frozen_files(), historical_makefile())
    if run.read(run.PACKET / "schema.json") != run.SCHEMA:
        raise ValueError("schema changed")
    if run.read(run.PACKET / "protocol.json") != run.packet_policy():
        raise ValueError("protocol changed")
    actual = {path.name for path in (run.PACKET / "requests").iterdir()}
    wanted = {f"{run.page_name(number)}.txt" for number in run.PAGES}
    if actual != wanted:
        raise ValueError("request set changed")
    for number in run.PAGES:
        if (run.PACKET / "requests" / f"{run.page_name(number)}.txt").read_bytes() != run.request(number):
            raise ValueError("request bytes changed")
    return {"status": "verified", "published_commit": PUBLISHED_COMMIT,
            "frozen_makefile_sha256": frozen["Makefile"],
            "current_makefile_sha256": run.frozen_files()["Makefile"]}


if __name__ == "__main__":
    print(json.dumps(verify(), sort_keys=True))
