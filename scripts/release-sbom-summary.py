"""Report the project license each per-platform SBOM declares for its main component.

Original and corrected SBOMs are counted separately, because a release can carry
both when a license correction was published as a supplement.
"""

import json
import sys
from pathlib import Path


def licenses_of(component: dict) -> list[str]:
    names = []
    for entry in component.get("licenses", []):
        license_entry = entry.get("license", {})
        name = license_entry.get("id") or license_entry.get("name")
        if name:
            names.append(name)
    return names


def main() -> None:
    found: dict[str, dict[str, int]] = {"original": {}, "corrected": {}}
    for path in sorted(Path(sys.argv[1]).glob("*.cdx.json")):
        kind = "corrected" if path.name.endswith(".corrected.cdx.json") else "original"
        document = json.loads(path.read_text(encoding="utf-8"))
        component = document.get("metadata", {}).get("component", {})
        for name in licenses_of(component) or ["undeclared"]:
            found[kind][name] = found[kind].get(name, 0) + 1
    print(json.dumps(found, sort_keys=True))


if __name__ == "__main__":
    main()
