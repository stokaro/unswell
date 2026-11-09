"""Verify the published freeze chain and reported transfer arithmetic offline."""

from __future__ import annotations

import gzip
import hashlib
import json
import tempfile
from collections import Counter
from pathlib import Path

from select_transfer_v2 import select


ROOT = Path(__file__).resolve().parent


def require(condition: bool, message: str) -> None:
    if not condition:
        raise ValueError(message)


def read_json(name: str) -> tuple[dict, str]:
    data = (ROOT / name).read_bytes()
    return json.loads(data), hashlib.sha256(data).hexdigest()


def verify_transfer_result(inventory: dict, labels: dict, result: dict,
                           inventory_sha: str, labels_sha: str,
                           selector_sha: str) -> None:
    require(labels["inventory_sha256"] == inventory_sha, "labels bind wrong inventory")
    targets = inventory["targets"]
    reviews = labels["labels"]
    rows = result["rows"]
    require(len(targets) == len(reviews) == len(rows), "transfer row count changed")
    require(result["inventory_sha256"] == inventory_sha, "result binds wrong inventory")
    require(result["labels_sha256"] == labels_sha, "result binds wrong labels")
    require(result["selector_sha256"] == selector_sha, "result binds wrong selector")
    require(result["target_count"] == len(targets), "reported target count changed")
    require(result["source_count"] == len(inventory["sources"]), "source count changed")
    require(not result["product_qualified"], "negative research claims qualification")
    matrix = Counter()
    by_query = {}
    by_repo = {}
    for index, (target, review, row) in enumerate(zip(targets, reviews, rows, strict=True)):
        require(review["index"] == row["index"] == index, "index order changed")
        require(review["role"] in {"admit", "withhold", "abstain"}, "invalid frozen role")
        require(row["gold"] == review["role"], "result changes frozen role")
        require(row["decision"] in {"admit", "withhold", "abstain"}, "invalid decision")
        require(row["repo"] == target["repo"], "result changes source repo")
        require(row["query_index"] == target["query_index"], "result changes query")
        require(target["end"] - target["start"] == len(target["quote"].encode()),
                "target quote length differs from original byte span")
        key = f'{row["gold"]}->{row["decision"]}'
        matrix[key] += 1
        by_query.setdefault(str(row["query_index"]), Counter())[key] += 1
        by_repo.setdefault(row["repo"], Counter())[key] += 1
    require(dict(matrix) == result["matrix"], "reported matrix differs from rows")
    require({k: dict(v) for k, v in by_query.items()} == result["by_query"],
            "reported query matrix differs from rows")
    require({k: dict(v) for k, v in by_repo.items()} == result["by_repo"],
            "reported repository matrix differs from rows")
    positives = sum(row["gold"] == "admit" for row in rows)
    positive_repos = {row["repo"] for row in rows if row["gold"] == "admit"}
    true_admits = matrix["admit->admit"]
    false_admits = sum(count for key, count in matrix.items()
                       if key.endswith("->admit") and key != "admit->admit")
    require(result["positive_count"] == positives, "positive count changed")
    require(result["positive_repo_count"] == len(positive_repos),
            "positive repository count changed")
    require(result["true_admissions"] == true_admits, "true admission count changed")
    require(result["false_admissions"] == false_admits, "false admission count changed")
    if false_admits:
        expected_status = "fail"
    elif positives < 5 or len(positive_repos) < 3:
        expected_status = "underpowered"
    else:
        expected_status = "pass" if true_admits / positives >= 0.8 else "fail"
    require(result["status"] == expected_status,
            "reported status violates the frozen decision rule")


def verify() -> None:
    selection, selection_sha = read_json("transfer_v2_selection.json")
    sources, sources_sha = read_json("transfer_v2_sources.json")
    inventory, inventory_sha = read_json("transfer_v2_inventory.json")
    labels, labels_sha = read_json("transfer_v2_labels.json")
    result, _ = read_json("transfer_v2_result.json")
    selector_sha = hashlib.sha256((ROOT / "selector.py").read_bytes()).hexdigest()
    with tempfile.TemporaryDirectory() as folder:
        search_dir = Path(folder)
        for index, item in enumerate(selection["search_pages"], 1):
            raw = gzip.decompress((ROOT / "search_v2" / f"{index}.json.gz").read_bytes())
            require(hashlib.sha256(raw).hexdigest() == item["sha256"],
                    "search response hash changed")
            (search_dir / f"{index}.json").write_bytes(raw)
        require(select(search_dir) == selection, "source selection changed")
    require(len(sources["sources"]) == len(selection["selected"]),
            "source revision count changed")
    for selected, source in zip(selection["selected"], sources["sources"], strict=True):
        require(all(source[key] == selected[key] for key in selected),
                "source revisions change selected file")
        require(len(source["commit"]) == 40 and source["url"].endswith(
            f'/{source["commit"]}/{source["path"]}'), "invalid source revision URL")
    require(inventory["selection_sha256"] == sources_sha,
            "inventory binds wrong source revisions")
    require(len(inventory["sources"]) == len(sources["sources"]),
            "inventory source count changed")
    for selected, source in zip(sources["sources"], inventory["sources"], strict=True):
        require(all(source[key] == selected[key] for key in
                    ("repo", "path", "commit", "query_index", "url")),
                "inventory changes source identity")
        require(len(source["sha256"]) == 64 and len(source["git_blob_sha1"]) == 40,
                "invalid source digest")
        require(source["git_blob_sha1"] == selected["search_blob_sha1"],
                "source blob differs from search result")
    verify_transfer_result(inventory, labels, result, inventory_sha, labels_sha,
                           selector_sha)
    require(result["false_admissions"] == 5 and result["positive_count"] == 1,
            "frozen negative outcome changed")
    print("cue transfer verified: 20 sources, 12 targets, 5 false admissions; selector retired")


if __name__ == "__main__":
    verify()
