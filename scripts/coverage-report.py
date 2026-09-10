"""Aggregate Go coverage profiles into package, area and overall statement coverage.

The report gates the areas named in issue #27: engine and scoring, rules,
configuration, and source mapping at 90%, and every measured package together at
85%. Packages without executable statements are reported with a null percentage
and excluded from the totals.
"""

import json
import os
import sys
from pathlib import Path

MODULE = "github.com/stokaro/unswell"

AREAS = {
    "engine-and-scoring": [MODULE],
    "rules": [f"{MODULE}/rule", f"{MODULE}/ruleset", f"{MODULE}/builtin"],
    "configuration": [f"{MODULE}/config", f"{MODULE}/internal/appconfig"],
    "source-mapping": [f"{MODULE}/extract", f"{MODULE}/document", f"{MODULE}/internal/mapping"],
}

TARGETS = {
    "engine-and-scoring": 90.0,
    "rules": 90.0,
    "configuration": 90.0,
    "source-mapping": 90.0,
    "overall": 85.0,
}


def read_profile(path):
    """Return {package: {block: (statements, covered)}} for one coverage profile."""
    packages = {}
    for number, line in enumerate(path.read_text().splitlines()):
        if number == 0:
            if not line.startswith("mode:"):
                raise SystemExit(f"{path}: missing mode header")
            continue
        location, statements, count = line.rsplit(" ", 2)
        package = location.rsplit("/", 1)[0]
        blocks = packages.setdefault(package, {})
        previous = blocks.get(location, (int(statements), False))
        blocks[location] = (int(statements), previous[1] or int(count) > 0)
    return packages


def merge(profiles):
    packages = {}
    for profile in profiles:
        for package, blocks in read_profile(profile).items():
            target = packages.setdefault(package, {})
            for location, (statements, covered) in blocks.items():
                previous = target.get(location, (statements, False))
                target[location] = (statements, previous[1] or covered)
    return packages


def totals(blocks):
    statements = sum(count for count, _ in blocks.values())
    covered = sum(count for count, hit in blocks.values() if hit)
    return statements, covered


def percentage(statements, covered):
    return None if statements == 0 else round(100.0 * covered / statements, 1)


def main():
    if len(sys.argv) < 2:
        raise SystemExit("usage: coverage-report.py PROFILE [PROFILE ...]")
    packages = merge(Path(argument) for argument in sys.argv[1:])
    measured = {name: totals(blocks) for name, blocks in packages.items()}

    report = {
        "format": "unswell-coverage-observation-v1",
        "environment": {
            name: os.environ.get(variable, "")
            for name, variable in (
                ("commit", "UNSWELL_COVERAGE_COMMIT"),
                ("go_version", "UNSWELL_COVERAGE_GO"),
                ("goos", "UNSWELL_COVERAGE_GOOS"),
                ("goarch", "UNSWELL_COVERAGE_GOARCH"),
            )
        },
        "packages": {
            name: {"statements": statements, "covered": covered, "percent": percentage(statements, covered)}
            for name, (statements, covered) in sorted(measured.items())
        },
        "areas": {},
    }
    for area, members in AREAS.items():
        missing = [member for member in members if member not in measured]
        if missing:
            raise SystemExit(f"area {area} has unmeasured packages: {', '.join(missing)}")
        statements = sum(measured[member][0] for member in members)
        covered = sum(measured[member][1] for member in members)
        report["areas"][area] = {
            "packages": members,
            "statements": statements,
            "covered": covered,
            "percent": percentage(statements, covered),
            "target": TARGETS[area],
        }
    statements = sum(count for count, _ in measured.values())
    covered = sum(hit for _, hit in measured.values())
    report["overall"] = {
        "statements": statements,
        "covered": covered,
        "percent": percentage(statements, covered),
        "target": TARGETS["overall"],
    }

    failures = [
        f"{name}: {entry['percent']}% below {entry['target']}%"
        for name, entry in list(report["areas"].items()) + [("overall", report["overall"])]
        if entry["percent"] is None or entry["percent"] < entry["target"]
    ]
    report["failures"] = failures
    print(json.dumps(report, indent=2, sort_keys=True))
    if failures:
        raise SystemExit(1)


main()
