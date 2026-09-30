#!/usr/bin/env python3
"""Replay the pinned public TETRA inventory through Go and check every pair."""

import argparse
import base64
import hashlib
import json
from pathlib import Path
import re
import subprocess
import xml.etree.ElementTree as ET

ROOT = Path(__file__).resolve().parent
REVISION = "054ba8b308b7ed5a3a2aca9619427bc8f8fbd405"
MANIFEST_SHA256 = "7893eeed77d4e9390581c3bf6b6d6e63954a232f93414903181bb6a3b6dd4465"
VERSION = "unswell-tetra-import-v1"
RENDERING = ("Join decoded XML text/edit leaves with one ASCII space; "
             "spans refer only to reconstructed UTF-8 sections.")
TARGET = "Upstream editor preference; neither revision necessity nor technical safety is inferred."
SECTIONS = ("title", "abstract", "introduction")
SOURCE_PATH = re.compile(r"original/([PW][0-9]{2}-[0-9]{4})-([ABC])\.xml")


def require(condition, reason):
    if not condition:
        raise ValueError(reason)


def sha256(raw):
    return hashlib.sha256(raw).hexdigest()


def git_blob(raw):
    return hashlib.sha1(b"blob " + str(len(raw)).encode() + b"\0" + raw,
                        usedforsecurity=False).hexdigest()


def json_bytes(value):
    return (json.dumps(value, ensure_ascii=False, indent=2) + "\n").encode()


def read_inventory(source_root, manifest_path):
    raw_manifest = manifest_path.read_bytes()
    require(sha256(raw_manifest) == MANIFEST_SHA256, "Changed acquisition manifest")
    manifest = json.loads(raw_manifest)
    require(manifest["repository"] == "chemicaltree/tetra" and
            manifest["revision"] == REVISION, "Changed repository identity")
    paths = [f["path"] for f in manifest["files"]]
    require(len(paths) == len(set(paths)) == 194, "Changed source inventory")
    require({p for p in paths if not SOURCE_PATH.fullmatch(p)} ==
            {"README.md", "aspect_edit-type_map.txt"}, "Unexpected inventory path")
    files = []
    for entry in manifest["files"]:
        path = source_root / entry["path"]
        require(path.resolve().is_relative_to(source_root.resolve()), "Source escaped root")
        raw = path.read_bytes()
        require(len(raw) == entry["bytes"] and sha256(raw) == entry["sha256"] and
                git_blob(raw) == entry["git_blob"], "Changed source: " + entry["path"])
        if SOURCE_PATH.fullmatch(entry["path"]):
            files.append({k: entry[k] for k in ("path", "sha256", "git_blob")} |
                         {"xml_base64": base64.b64encode(raw).decode()})
    require(len(files) == 192, "Changed XML denominator")
    return files


def decoded_leaf(element):
    value = element.text or ""
    for child in element:
        require(child.tag is ET.Comment, "Nested markup or processing instruction")
        value += child.tail or ""
    return value


def reconstruct(section, paper, editor):
    original, revised, edits = [], [], []
    before_offset = after_offset = 0
    leaves = [child for child in section if child.tag is not ET.Comment]
    for index, element in enumerate(leaves):
        before = decoded_leaf(element)
        after = element.attrib["crr"] if element.tag == "edit" else before
        if index:
            before_offset += 1
            after_offset += 1
        if element.tag == "edit":
            raw_type = element.attrib.get("type", "")
            edits.append({
                "id": f"{paper}-{editor}/{section.tag}/{index:04d}",
                "element_index": index,
                "types": sorted({t.strip().lower() for t in raw_type.split(";") if t.strip()}),
                "type_available": "type" in element.attrib,
                "original_type": raw_type,
                "comment": element.attrib.get("comments"),
                "before": before,
                "after": after,
                "original_span": {"start": before_offset, "end": before_offset + len(before.encode())},
                "revised_span": {"start": after_offset, "end": after_offset + len(after.encode())},
                "unchanged": before == after,
            })
        before_offset += len(before.encode())
        after_offset += len(after.encode())
        original.append(before)
        revised.append(after)
    before, after = " ".join(original), " ".join(revised)
    return {"name": section.tag, "original": before, "revised": after,
            "original_sha256": sha256(before.encode()), "revised_sha256": sha256(after.encode()),
            "edits": edits}


def no_unowned_text(element):
    require(not (element.text or "").strip(" \t\r\n"), "Text outside text/edit leaves")
    for child in element:
        require(not (child.tail or "").strip(" \t\r\n"), "Text after text/edit leaves")


def expected_file(source):
    paper, editor = SOURCE_PATH.fullmatch(source["path"]).groups()
    raw = base64.b64decode(source["xml_base64"], validate=True)
    file = {k: source[k] for k in ("path", "sha256", "git_blob", "xml_base64")}
    file.update(paper=paper, editor=editor, bytes=len(raw), declared_metadata={},
                sections=[], status="valid", source_eligible=True, issue_code=None)
    try:
        parser = ET.XMLParser(target=ET.TreeBuilder(insert_comments=True, insert_pis=True))
        root = ET.fromstring(raw, parser=parser)
    except ET.ParseError:
        file.update(status="invalid_xml", source_eligible=False, issue_code="xml_syntax_error")
        return file
    try:
        require(root.tag == "doc" and set(root.attrib) <=
                {"id", "editor", "format", "position", "region"}, "Unsupported root")
        require("id" in root.attrib and "editor" in root.attrib, "Missing root identity")
        no_unowned_text(root)
        sections = [s for s in root if s.tag is not ET.Comment]
        require(tuple(s.tag for s in sections) == SECTIONS, "Unexpected sections")
        for section in sections:
            require(not section.attrib, "Section attributes")
            no_unowned_text(section)
            for leaf in section:
                if leaf.tag is ET.Comment:
                    continue
                require(leaf.tag in ("text", "edit"), "Unsupported leaf")
                allowed = {"type", "crr", "comments"} if leaf.tag == "edit" else set()
                require(set(leaf.attrib) <= allowed, "Unsupported leaf attributes")
                require(leaf.tag != "edit" or "crr" in leaf.attrib, "Missing correction")
                decoded_leaf(leaf)
    except ValueError:
        file.update(status="invalid_xml", source_eligible=False, issue_code="unsupported_xml")
        return file
    file["declared_metadata"] = root.attrib
    file["sections"] = [reconstruct(s, paper, editor) for s in sections]
    if root.attrib["id"] != paper or root.attrib["editor"] != editor:
        file.update(status="identity_conflict", source_eligible=False, issue_code="identity_conflict")
    return file


def expected_report(files):
    expected = sorted((expected_file(f) for f in files), key=lambda f: f["path"])
    bad_papers = {f["paper"] for f in expected if not f["source_eligible"]}
    for file in expected:
        if file["paper"] in bad_papers:
            file["source_eligible"] = False
            if file["status"] == "valid":
                file.update(status="quarantined", issue_code="paper_group_quarantined")
    papers = sorted({f["paper"] for f in expected})
    groups = [{"paper": p, "files": [f["path"] for f in expected if f["paper"] == p],
               "source_eligible": p not in bad_papers} for p in papers]
    counts = {
        "files": len(expected),
        "parsed_files": sum(f["status"] != "invalid_xml" for f in expected),
        "invalid_files": sum(f["status"] == "invalid_xml" for f in expected),
        "identity_conflicts": sum(f["status"] == "identity_conflict" for f in expected),
        "quarantined_files": sum(not f["source_eligible"] for f in expected),
        "groups": len(groups),
        "eligible_groups": sum(g["source_eligible"] for g in groups),
        "edits": sum(len(s["edits"]) for f in expected for s in f["sections"]),
        "eligible_edits": sum(len(s["edits"]) for f in expected if f["source_eligible"]
                              for s in f["sections"]),
    }
    return expected, groups, counts


def verify(report, files):
    expected, groups, counts = expected_report(files)
    require(report["version"] == VERSION and report["revision"] == REVISION and
            report["repository"] == "https://github.com/chemicaltree/tetra" and
            report["license"] == "CC-BY-4.0", "Changed import identity")
    require(report["task_labels_assigned"] is False and report["product_qualified"] is False,
            "Source preference was promoted into a task verdict")
    require(report["rendering"] == RENDERING and report["target"] == TARGET,
            "Changed section or target semantics")
    require(report["complete"] is (counts["quarantined_files"] == 0), "Incorrect completion")
    require(report["groups"] == groups and report["counts"] == counts, "Group or count mismatch")
    require(len(report["files"]) == len(expected), "Missing source record")
    for got, want in zip(report["files"], expected):
        require(set(got) == (set(want) - {"issue_code"} | {"issues"}), "Unexpected file fields")
        for key, value in want.items():
            if key != "issue_code":
                require(got[key] == value, f"Source mismatch: {want['path']} / {key}")
        issues = got["issues"]
        require(len(issues) == (want["issue_code"] is not None), "Missing or extra issue")
        if issues:
            require(issues[0]["code"] == want["issue_code"] and
                    isinstance(issues[0]["reason"], str) and issues[0]["reason"],
                    "Missing source failure reason")
        for section in got["sections"]:
            for edit in section["edits"]:
                for side, text in (("original", "before"), ("revised", "after")):
                    span = edit[side + "_span"]
                    require(section[side].encode()[span["start"]:span["end"]].decode() == edit[text],
                            "Broken UTF-8 edit span")
    return counts


def replay(args):
    files = read_inventory(args.source_root, args.manifest)
    payload = {"version": VERSION, "repository": "https://github.com/chemicaltree/tetra",
               "revision": REVISION, "license": "CC-BY-4.0", "include_xml": True, "files": files}
    args.output_dir.mkdir(mode=0o700)
    request = json_bytes(payload)
    (args.output_dir / "input.json").write_bytes(request)
    binary = args.binary.resolve()
    completed = subprocess.run([str(binary)], input=request, capture_output=True, timeout=60,
                               check=False)
    (args.output_dir / "output.json").write_bytes(completed.stdout)
    (args.output_dir / "stderr.txt").write_bytes(completed.stderr)
    require(completed.returncode in (0, 2), "Go importer failed operationally")
    counts = verify(json.loads(completed.stdout), files)
    require(completed.returncode == (2 if counts["quarantined_files"] else 0),
            "CLI exit does not match quarantine")
    summary = {"version": "unswell-tetra-import-verification-v1", "verified": True,
               "revision": REVISION, "manifest_sha256": MANIFEST_SHA256,
               "binary_sha256": sha256(binary.read_bytes()), "input_sha256": sha256(request),
               "output_sha256": sha256(completed.stdout), "cli_exit": completed.returncode,
               "counts": counts, "task_labels_assigned": False, "product_qualified": False}
    repository = ROOT.parents[2]
    implementation = ["internal/tetra/" + name for name in
                      ("types.go", "import.go", "xml.go", "command.go", "input.schema.json")]
    implementation.append("cmd/importtetra/main.go")
    summary["implementation_sha256"] = {
        "research/annotation/" + name: sha256((repository / "research/annotation" / name).read_bytes())
        for name in implementation}
    summary["verifier_sha256"] = sha256(Path(__file__).read_bytes())
    (args.output_dir / "verification.json").write_bytes(json_bytes(summary))
    print(json.dumps(summary, sort_keys=True))


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--source-root", type=Path, required=True)
    parser.add_argument("--binary", type=Path, required=True)
    parser.add_argument("--output-dir", type=Path, required=True)
    parser.add_argument("--manifest", type=Path, default=ROOT / "acquisition-manifest.json")
    replay(parser.parse_args())


if __name__ == "__main__":
    main()
