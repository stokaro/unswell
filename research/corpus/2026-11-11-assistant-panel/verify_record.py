#!/usr/bin/env python3
"""Verify the assistant-panel corpus record offline.

usage: verify_record.py [--write-digests] [record directory]

The check covers file digests, the record schema, the binding of labels and
reviews to items, repository separation between development and test sets, and
the published summary. The schema admits coordinates and decisions only, so a
record that carries page text fails.
"""
import collections
import gzip
import hashlib
import json
import sys
from pathlib import Path

HERE = Path(__file__).resolve().parent
DIGESTS = "files.sha256"
DATA_DIRECTORIES = ("audit", "configs", "contextual", "items", "labels", "manifests", "phrases", "reviews", "rounds")
DOCUMENT_KEYS = {"doc", "repository", "path", "sha256", "reference", "origin", "family", "operation", "items"}
ITEM_KEYS = {"id", "kind", "segments"}
LABEL_KEYS = {"doc", "annotator", "model", "effort", "rubric", "prompt_sha256", "flags", "unbound_flags"}
FLAG_KEYS = {"item", "label", "categories"}
REVIEW_KEYS = {"id", "set", "doc", "rule", "group", "phrase", "reviews", "severity", "span"}
CONTEXT_KEYS = {"id", "set", "doc", "caller", "categories", "span", "reviews"}
LABELS = {"needs_revision", "uncertain"}
VERDICTS = {"yes", "no", "uncertain", None}
SEVERITIES = {"note", "warning", "error"}
MAX_PHRASE_WORDS = 6


def fail(message):
    raise ValueError(message)


def data_files(root):
    """Return the relative names of every data file of the record."""
    names = ["summary.json"] if (root / "summary.json").exists() else []
    for directory in DATA_DIRECTORIES:
        names += [p.relative_to(root).as_posix() for p in (root / directory).rglob("*") if p.is_file()]
    return sorted(names)


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def write_digests(root):
    lines = [f"{digest(root / name)}  {name}\n" for name in data_files(root)]
    (root / DIGESTS).write_text("".join(lines))


def check_digests(root):
    listed = {}
    for line in (root / DIGESTS).read_text().splitlines():
        value, name = line.split("  ", 1)
        listed[name] = value
    present = data_files(root)
    if sorted(listed) != present:
        changed = sorted(set(listed) ^ set(present))
        fail(f"data files differ from {DIGESTS}: {', '.join(changed[:5])}")
    for name in present:
        if digest(root / name) != listed[name]:
            fail(f"digest mismatch: {name}")


def rows(path):
    with gzip.open(path, "rt", encoding="utf-8") as handle:
        for line in handle:
            yield json.loads(line)


def set_name(path):
    return path.name.removesuffix(".jsonl.gz")


def check_span(span, where):
    ok = (isinstance(span, list) and len(span) == 2 and all(isinstance(v, int) for v in span)
          and 0 <= span[0] < span[1])
    if not ok:
        fail(f"invalid byte span in {where}")


def load_items(root):
    """Return {set: {doc: document row}} after checking the item schema."""
    sets = {}
    for path in sorted((root / "items").glob("*.jsonl.gz")):
        documents = sets.setdefault(set_name(path), {})
        for row in rows(path):
            if set(row) != DOCUMENT_KEYS:
                fail(f"unexpected document fields in {path.name}: {sorted(set(row) ^ DOCUMENT_KEYS)}")
            if row["doc"] in documents:
                fail(f"duplicate document {row['doc']}")
            seen = set()
            for item in row["items"]:
                if set(item) != ITEM_KEYS:
                    fail(f"unexpected item fields in {row['doc']}: {sorted(set(item) ^ ITEM_KEYS)}")
                if item["id"] in seen:
                    fail(f"duplicate item {item['id']} in {row['doc']}")
                seen.add(item["id"])
                for segment in item["segments"]:
                    check_span(segment, f"{row['doc']} {item['id']}")
            documents[row["doc"]] = row
    return sets


def load_labels(root, sets, panel):
    """Return {set: {doc: {annotator: set of items marked for revision}}}."""
    marks = {}
    for path in sorted((root / "labels").glob("*.jsonl.gz")):
        name = set_name(path)
        for row in rows(path):
            if set(row) != LABEL_KEYS:
                fail(f"unexpected label fields in {path.name}: {sorted(set(row) ^ LABEL_KEYS)}")
            member = panel.get(row["annotator"])
            if not member or "annotator" not in member["roles"] or member["model"] != row["model"]:
                fail(f"annotator {row['annotator']} is not an annotator of the panel")
            document = sets.get(name, {}).get(row["doc"])
            if document is None:
                fail(f"label for unknown document {row['doc']} in {name}")
            known = {item["id"] for item in document["items"]}
            marked = set()
            for flag in row["flags"]:
                if set(flag) != FLAG_KEYS or flag["label"] not in LABELS:
                    fail(f"invalid flag in {row['doc']}")
                if flag["item"] not in known:
                    fail(f"flag for unknown item {flag['item']} in {row['doc']}")
                if flag["label"] == "needs_revision":
                    marked.add(flag["item"])
            by_annotator = marks.setdefault(name, {}).setdefault(row["doc"], {})
            if row["annotator"] in by_annotator:
                fail(f"duplicate labels from {row['annotator']} for {row['doc']}")
            by_annotator[row["annotator"]] = marked
    return marks


def load_reviews(root, sets, panel):
    """Return {round: [review rows]} after checking the review schema."""
    rounds = {}
    reviewers = {name for name, member in panel.items() if "reviewer" in member["roles"]}
    for path in sorted((root / "reviews").glob("*.jsonl.gz")):
        seen = set()
        listed = rounds.setdefault(set_name(path), [])
        for row in rows(path):
            if set(row) != REVIEW_KEYS:
                fail(f"unexpected review fields in {path.name}: {sorted(set(row) ^ REVIEW_KEYS)}")
            if row["id"] in seen:
                fail(f"duplicate finding {row['id']} in {path.name}")
            seen.add(row["id"])
            if row["doc"] not in sets.get(row["set"], {}):
                fail(f"review for unknown document {row['doc']} in {row['set']}")
            if not set(row["reviews"]) <= reviewers:
                fail(f"review by a model outside the panel: {sorted(set(row['reviews']) - reviewers)}")
            if any(verdict not in VERDICTS for verdict in row["reviews"].values()):
                fail(f"invalid verdict for finding {row['id']}")
            if row["severity"] not in SEVERITIES:
                fail(f"invalid severity for finding {row['id']}")
            if len(row["phrase"].split()) > MAX_PHRASE_WORDS:
                fail(f"finding {row['id']} carries text instead of a matched phrase")
            check_span(row["span"], f"finding {row['id']}")
            listed.append(row)
    return rounds


def check_contextual(root, sets, panel):
    """Check the flags of calling assistants: coordinates, categories, and verdicts only."""
    reviewers = {name for name, member in panel.items() if "reviewer" in member["roles"]}
    for path in sorted((root / "contextual").glob("*.jsonl.gz")):
        seen = set()
        for row in rows(path):
            if set(row) != CONTEXT_KEYS:
                fail(f"unexpected flag fields in {path.name}: {sorted(set(row) ^ CONTEXT_KEYS)}")
            if row["id"] in seen:
                fail(f"duplicate flag {row['id']} in {path.name}")
            seen.add(row["id"])
            if "caller" not in panel.get(row["caller"], {}).get("roles", []):
                fail(f"flag by an assistant outside the panel: {row['caller']}")
            if row["doc"] not in sets.get(row["set"], {}):
                fail(f"flag for unknown document {row['doc']} in {row['set']}")
            if not set(row["reviews"]) <= reviewers - {row["caller"]}:
                fail(f"flag {row['id']} is judged by its own caller or outside the panel")
            if any(verdict not in VERDICTS for verdict in row["reviews"].values()):
                fail(f"invalid verdict for flag {row['id']}")
            if not row["categories"] or any(" " in category for category in row["categories"]):
                fail(f"flag {row['id']} carries text instead of categories")
            check_span(row["span"], f"flag {row['id']}")


def check_panel(panel):
    models = [member["model"] for member in panel.values()]
    if len(set(models)) != len(models):
        fail("panel members share a model")
    for name, member in panel.items():
        if set(member) != {"model", "effort", "roles"}:
            fail(f"unexpected panel fields for {name}")
        if "human" in member["roles"]:
            fail(f"panel member {name} is an assistant and cannot hold a human role")


def check_separation(sets, roles):
    """Reject a repository that appears under two different set roles."""
    owner = {}
    for name, documents in sorted(sets.items()):
        for row in documents.values():
            repository = row["repository"].lower()
            previous = owner.setdefault(repository, name)
            if roles[previous] != roles[name]:
                fail(f"repository {row['repository']} appears in {previous} and {name}")


def summarize(sets, marks, rounds, roles, panel):
    """Recompute the summary from the data files."""
    summary = {"panel": panel, "sets": [], "rounds": []}
    for name, documents in sets.items():
        annotated = collections.Counter()
        for doc, by_annotator in marks.get(name, {}).items():
            if len(by_annotator) != 2:
                continue
            first, second = by_annotator["astra"], by_annotator["claude"]
            annotated["documents"] += 1
            annotated["items"] += len(documents[doc]["items"])
            annotated["astra"] += len(first)
            annotated["claude"] += len(second)
            annotated["both"] += len(first & second)
        summary["sets"].append({
            "set": name, "role": roles[name], "documents": len(documents),
            "repositories": len({row["repository"] for row in documents.values()}),
            "items": sum(len(row["items"]) for row in documents.values()), "annotated": dict(annotated)})
    for name, listed in rounds.items():
        table = collections.defaultdict(collections.Counter)
        for row in listed:
            first = row["reviews"].get("claude")
            second = row["reviews"].get("astra") or row["reviews"].get("qwen")
            if not first or not second:
                continue
            counts = table[row["rule"]]
            counts["reviewed"] += 1
            counts["both_accept"] += first == "yes" and second == "yes"
            counts["any_accept"] += first == "yes" or second == "yes"
            counts["both_reject"] += first == "no" and second == "no"
        summary["rounds"].append({"round": name, "findings": len(listed),
                                  "rules": {rule: dict(counts) for rule, counts in sorted(table.items())}})
    return summary


def normalized(summary):
    summary = json.loads(json.dumps(summary))
    summary["sets"].sort(key=lambda entry: entry["set"])
    summary["rounds"].sort(key=lambda entry: entry["round"])
    return summary


def verify(root=HERE):
    """Raise ValueError when the record at root is inconsistent."""
    root = Path(root)
    check_digests(root)
    published = json.loads((root / "summary.json").read_text())
    panel = published["panel"]
    check_panel(panel)
    roles = {entry["set"]: entry["role"] for entry in published["sets"]}
    sets = load_items(root)
    if set(sets) != set(roles):
        fail("summary lists other sets than the item files")
    marks = load_labels(root, sets, panel)
    rounds = load_reviews(root, sets, panel)
    check_contextual(root, sets, panel)
    check_separation(sets, roles)
    if normalized(summarize(sets, marks, rounds, roles, panel)) != normalized(published):
        fail("summary differs from the data files")
    return published


def main(arguments):
    write = "--write-digests" in arguments
    paths = [a for a in arguments if a != "--write-digests"]
    root = Path(paths[0]) if paths else HERE
    if write:
        write_digests(root)
    summary = verify(root)
    documents = sum(entry["documents"] for entry in summary["sets"])
    findings = sum(entry["findings"] for entry in summary["rounds"])
    print(f"record verified: {documents} documents, {findings} reviewed findings")


if __name__ == "__main__":
    main(sys.argv[1:])
