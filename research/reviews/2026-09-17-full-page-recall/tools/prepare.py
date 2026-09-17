#!/usr/bin/env python3
"""Select complete pages from existing pinned frames; never select by findings."""

import argparse
import hashlib
import json
from pathlib import Path, PurePosixPath
import tarfile


def sha(data):
    return hashlib.sha256(data).hexdigest()


def read(path):
    return json.loads(path.read_text(encoding="utf-8"))


def write(path, value):
    path.write_text(json.dumps(value, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")


def tier(words):
    return "short" if words < 800 else "medium" if words < 2000 else "long"


def rank(row):
    key = "\n".join(("full-page-recall-v1", row["cohort"], row["repository"], row["original_path"]))
    return sha(key.encode())


def member(archive, name):
    info = archive.getmember(name)
    if not info.isfile() or info.size > 4 * 1024 * 1024:
        raise ValueError("Invalid source archive member: " + name)
    return archive.extractfile(info).read()


def save_source(root, relative, data, expected):
    path = PurePosixPath(relative)
    if path.is_absolute() or ".." in path.parts or sha(data) != expected:
        raise ValueError("Invalid source path or digest: " + relative)
    target = root / path
    target.parent.mkdir(parents=True, exist_ok=True)
    target.write_bytes(data)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--ptah-sources", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True, help="New output directory")
    args = parser.parse_args()
    if args.output.exists():
        raise ValueError("Output must be new; never overwrite frozen annotations")
    root = Path(__file__).resolve().parents[4]
    prior = root / "research/reviews/2026-09-16-ptah-rhetoric"
    historical = root / "research/generation/studies/long-prose-v3"
    ptah = read(prior / "inputs.json")
    # Deliberately project only input identity/length, ignoring outcome fields.
    lengths = {p["path"].removeprefix("sources/"): p["words"]
               for p in read(prior / "technical.json")["after"]["pages"]}
    frame = []
    for source in ptah["pages"]:
        frame.append(dict(cohort="ptah", repository="stokaro/ptah",
                          original_path=source["source_path"], input_path=source["path"],
                          sha256=source["sha256"], bytes=source["bytes"],
                          format="mdx" if source["path"].endswith(".mdx") else "markdown",
                          words=lengths[source["path"]], commit=ptah["ptah_commit"],
                          reference=f"https://github.com/stokaro/ptah/blob/{ptah['ptah_commit']}/{source['source_path']}",
                          rights={"license": "MIT", "evidence": "Maintainer-authorized research reuse; retained Ptah notice",
                                  "allowed_uses": ["annotation", "evaluation", "redistribute_source", "redistribute_annotations"]},
                          origin={"label": "unknown", "scope": "repository", "evidence": "Maintainer reports nearly all Ptah prose is AI-generated; no unit-level record"}))
    archive_path = historical / "inputs.tar.gz"
    archive_record = read(historical / "inputs.json")
    if sha(archive_path.read_bytes()) != archive_record["sha256"]:
        raise ValueError("Historical archive drift")
    with tarfile.open(archive_path) as archive:
        selected = json.loads(member(archive, "selection.json"))
        for item in selected:
            slug = item["repository"].replace("/", "__")
            manifest = json.loads(member(archive, "pinned/shards/long-prose-v3-" + slug + ".json"))
            source = next(s for s in manifest["sources"] if s["id"] == item["source_id"])
            frame.append(dict(cohort="historical", repository=item["repository"],
                              original_path=source["path"], input_path="sources/historical/" + slug + "/" + source["path"],
                              sha256=source["sha256"], bytes=source["bytes"], format=source["format"],
                              words=item["words"], reference=source["reference"], rights=source["rights"],
                              snapshot=source["snapshot"], notices=source["notices"], origin=source["origin"]))
        for row in frame:
            row.update(rank=rank(row), length_stratum=tier(row["words"]))
        chosen, cells = [], []
        used_repositories = set()
        for cohort in ("ptah", "historical"):
            formats = ("markdown", "mdx") if cohort == "ptah" else ("any",)
            for fmt in formats:
                for length in ("long", "medium", "short"):
                    candidates = sorted((r for r in frame if r["cohort"] == cohort and r["length_stratum"] == length
                                         and (fmt == "any" or r["format"] == fmt)), key=lambda r: r["rank"])
                    wanted = 2 if cohort == "ptah" else 4
                    accepted = []
                    for row in candidates:
                        if cohort == "historical" and row["repository"] in used_repositories:
                            continue
                        accepted.append(row)
                        used_repositories.add(row["repository"])
                        if len(accepted) == wanted:
                            break
                    chosen.extend(dict(r, selection="sample") for r in accepted)
                    cells.append(dict(cohort=cohort, format=fmt, length=length, eligible=len(candidates),
                                      requested=wanted, selected=len(accepted)))
        for path in ("databases/postgresql.md", "concepts/database-urls-and-dev-databases.md"):
            row = next(r for r in frame if r["cohort"] == "ptah" and r["input_path"] == path)
            if not any(r["reference"] == row["reference"] for r in chosen):
                chosen.append(dict(row, selection="exposed_anchor"))
        args.output.mkdir(parents=True)
        for index, row in enumerate(chosen):
            row["id"] = f"p{index + 1:03d}"
            suffix = PurePosixPath(row["original_path"]).suffix
            row["path"] = "sources/" + row["id"] + suffix
            if row["cohort"] == "ptah":
                data = (args.ptah_sources / row["input_path"]).read_bytes()
                notices = {"notices/ptah/LICENSE": (root / "e2e/rhetoricdata/LICENSE.ptah").read_bytes()}
            else:
                data = member(archive, row["input_path"])
                base = "sources/historical/" + row["repository"].replace("/", "__") + "/"
                notices = {}
                for notice in row["notices"]:
                    notice_data = member(archive, base + notice["path"])
                    if sha(notice_data) != notice["sha256"]:
                        raise ValueError("Notice digest mismatch")
                    notices["notices/" + row["repository"].replace("/", "__") + "/" + notice["path"]] = notice_data
            if len(data) != row["bytes"]:
                raise ValueError("Source size mismatch")
            save_source(args.output, row["path"], data, row["sha256"])
            row["retained_notices"] = []
            for path, data in notices.items():
                save_source(args.output, path, data, sha(data))
                row["retained_notices"].append(dict(path=path, sha256=sha(data)))
        write(args.output / "manifest.json", dict(version="unswell-full-page-inputs-v1",
              protocol="unswell-full-page-recall-v1", engine_commit="cc79256188310f9d63a86f259044df4c3a27af1d",
              historical_archive_sha256=archive_record["sha256"], cells=cells, frame=frame, pages=chosen))
        for row in chosen:
            print(row["id"], row["cohort"], row["format"], row["words"], row["selection"], row["repository"], row["original_path"])


if __name__ == "__main__":
    main()
