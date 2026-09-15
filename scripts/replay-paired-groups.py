#!/usr/bin/env python3
"""Recompute paired tables under both group contracts from saved responses."""

import argparse
import hashlib
import json
import os
from pathlib import Path
import subprocess


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def finding_index(root):
    index = []
    for path in sorted(root.glob("*.json")):
        data = json.loads(path.read_text())
        cohorts = {row.get("cohort") for row in data["documents"]}
        runs = {row["path"].split("/")[1] for row in data["documents"]
                if row.get("cohort") == "controlled" and row["path"].startswith("generated/")}
        index.append((path, cohorts, runs))
    return index


def measure(binary, run, inputs, args, variant):
    command = [str(binary), "paired", "--records", str(run / "records.json"),
               "--tasks", str(run / "tasks.json"), "--classes", str(args.classes)]
    if variant == "after":
        command += ["--plan", str(args.plan)]
    for path in inputs:
        command += ["--findings", str(path)]
    result = subprocess.run(command, check=True, stdout=subprocess.PIPE,
                            env={**os.environ, "GOMAXPROCS": "2"})
    target = args.output / (run.name + "-" + variant + ".json")
    target.write_bytes(result.stdout)
    return json.loads(result.stdout), digest(target)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    for flag in ("before", "after", "runs", "findings", "classes", "plan", "output"):
        parser.add_argument("--" + flag, type=Path, required=True)
    args = parser.parse_args()
    args.output.mkdir(parents=True, exist_ok=False)
    index = finding_index(args.findings)
    summary = {"version": "unswell-paired-group-repair-v1", "status": "in_progress",
               "plan_sha256": digest(args.plan), "before_binary_sha256": digest(args.before),
               "after_binary_sha256": digest(args.after), "runs": []}
    for task_file in sorted(args.runs.glob("*/tasks.json")):
        run = task_file.parent
        if not (run / "records.json").is_file():
            continue
        tasks = json.loads(task_file.read_text())
        inputs = [p for p, cohorts, runs in index if tasks["cohort"] in cohorts or run.name in runs]
        before, before_hash = measure(args.before, run, inputs, args, "before")
        after, after_hash = measure(args.after, run, inputs, args, "after")
        summary["runs"].append({
            "run": run.name, "tasks_sha256": digest(task_file),
            "records_sha256": digest(run / "records.json"),
            "before_report_sha256": before_hash, "after_report_sha256": after_hash,
            "before_arms": before["arms"], "after_arms": after["arms"],
            "before_rules": before["rules"], "after_rules": after["rules"],
        })
        (args.output / "summary.json").write_text(json.dumps(summary, separators=(",", ":")) + "\n")
        print(run.name, [(a["operation"], a["prompt"], a["measured"], a["components"])
                         for a in after["arms"]], flush=True)
    summary["status"] = "complete"
    (args.output / "summary.json").write_text(json.dumps(summary, separators=(",", ":")) + "\n")


if __name__ == "__main__":
    main()
