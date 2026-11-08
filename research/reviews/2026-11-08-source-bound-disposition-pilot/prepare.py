"""Build an exact-byte source inventory request from pinned public README files."""

import hashlib
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parent


def prepare(root=ROOT):
    root = Path(root)
    manifest = json.loads((root / "sources.json").read_text(encoding="utf-8"))
    if manifest["format"] != "unswell-349-disposition-source-manifest-v1":
        raise ValueError("unexpected source manifest")
    pages = []
    ids = set()
    for source in manifest["sources"]:
        raw = (root / source["local_path"]).read_bytes()
        blob = hashlib.sha1(b"blob " + str(len(raw)).encode() + b"\0" + raw).hexdigest()
        digest = hashlib.sha256(raw).hexdigest()
        if (len(raw) != source["bytes"] or blob != source["git_blob_sha1"] or
                digest != source["source_sha256"] or source["id"] in ids or
                source["id"] != source["repository"] + "/" + source["path"]):
            raise ValueError("source identity mismatch: " + source["id"])
        ids.add(source["id"])
        pages.append({"id": source["id"], "format": "markdown",
                      "sha256": digest, "text": raw.decode("utf-8")})
    if len(pages) != 3:
        raise ValueError("expected three distinct source pages")
    data = {"version": 1, "basis": "source-function-inventory", "pages": pages}
    return (json.dumps(data, ensure_ascii=False, separators=(",", ":")) + "\n").encode("utf-8")


if __name__ == "__main__":
    (ROOT / "input.json").write_bytes(prepare())
