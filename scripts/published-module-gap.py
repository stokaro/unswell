"""List public packages the working tree's consumer imports that a tagged consumer did not.

A newer consumer can use packages a published module does not contain. Naming
them keeps a passing compatibility check from implying that today's API shipped.
"""

import re
import sys
from pathlib import Path

IMPORT = re.compile(r'"(github\.com/stokaro/unswell(?:/[a-z0-9/]+)?)"')


def imports(directory: Path) -> set[str]:
    found: set[str] = set()
    for path in sorted(directory.glob("*.go")):
        found.update(IMPORT.findall(path.read_text(encoding="utf-8")))
    return found


def main() -> None:
    current = imports(Path(sys.argv[1]) / "examples" / "consumer")
    released = imports(Path(sys.argv[2]))
    print(__import__("json").dumps(sorted(current - released)))


if __name__ == "__main__":
    main()
