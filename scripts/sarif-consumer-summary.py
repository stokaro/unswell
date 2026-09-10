"""Summarize a SARIF validation log by result level.

The consumer writes its own SARIF report; this counts its results so the shell
script can record them without embedding a program in a string.
"""

import json
import sys


def main() -> None:
    document = json.load(open(sys.argv[1], encoding="utf-8"))
    levels: dict[str, int] = {}
    for run in document.get("runs", []):
        for result in run.get("results", []):
            level = result.get("level", "warning")
            levels[level] = levels.get(level, 0) + 1
    print(json.dumps(levels, sort_keys=True))


if __name__ == "__main__":
    main()
