#!/usr/bin/env python3
"""Run the frozen WikiSQE vague-label transfer screen for issue 349."""

from __future__ import annotations

from collections import Counter, defaultdict
import gzip
import hashlib
import json
from pathlib import Path
import platform
import sys

import pyarrow
import pyarrow.parquet as parquet
import sklearn
from sklearn.feature_extraction.text import TfidfVectorizer
from sklearn.linear_model import LogisticRegression
from sklearn.metrics import accuracy_score, precision_recall_fscore_support


HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[2]
REVIEWS = ROOT / "research/reviews"
sys.path.insert(0, str(REVIEWS / "2026-09-26-outcome-announcements"))
from replay import sources  # noqa: E402
from verify import labels  # noqa: E402

DATASET_REVISION = "e6a3da39b44df3efa2c75706ccba1dc26dd17076"
DATASET_SHA256 = {
    "train": "25b2d9609cdd34d88dadcf8ab687bd05d7a6212766a9b010513ae5741d4a4bbe",
    "val": "dcfdd685a6082df6909f8fb1c03c81ac63c9a83376d262e1dc14584b2d1bc5aa",
    "test": "37d789c7ba762bb7b40b95091a3ddf29e2eb46b9702d757dac1dad44e515c72f",
}
SOURCE_FILES = {
    "sentences": REVIEWS / "2026-11-07-ape-proposition-screen/scores.json.gz",
    "source_hashes": REVIEWS / "2026-11-07-ape-proposition-screen/source-hashes.json",
    "events": REVIEWS / "2026-11-07-current-main-editorial-review/events.json",
    "frontier": REVIEWS / "2026-11-08-dual-gate-feasibility/category-frontier.json",
}


def digest(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def require(condition: bool, message: str) -> None:
    if not condition:
        raise ValueError(message)


def read_split(data_dir: Path, split: str) -> tuple[list[str], list[int]]:
    path = data_dir / f"{split}.parquet"
    require(digest(path) == DATASET_SHA256[split], f"WikiSQE {split} file changed")
    table = parquet.read_table(path, columns=["text", "label"])
    texts = table.column("text").to_pylist()
    labels = table.column("label").to_pylist()
    require(len(texts) == len(labels) and all(isinstance(s, str) and s for s in texts),
            f"Invalid {split} text")
    require(set(labels) == {0, 1}, f"Invalid {split} labels")
    return texts, labels


def overlapping(row: dict, target: dict) -> bool:
    return any(max(span["start"], target["start"]) <
               min(span["end"], target["end"]) for span in row["spans"])


def main() -> None:
    if len(sys.argv) != 2:
        raise SystemExit("usage: run.py DATA_DIR")
    data_dir = Path(sys.argv[1])
    train_text, train_y = read_split(data_dir, "train")
    val_text, val_y = read_split(data_dir, "val")
    test_text, test_y = read_split(data_dir, "test")
    require(len(val_text) == len(val_y) == len(test_text) == len(test_y) == 1000,
            "WikiSQE validation/test size changed")

    vectorizer = TfidfVectorizer(ngram_range=(1, 2), min_df=2,
                                 max_features=150000, sublinear_tf=True,
                                 lowercase=True, strip_accents=None, norm="l2",
                                 token_pattern=r"(?u)\b\w\w+\b")
    train_x = vectorizer.fit_transform(train_text)
    model = LogisticRegression(C=1.0, max_iter=1000, class_weight=None,
                               random_state=0)
    model.fit(train_x, train_y)
    test_pred = model.predict(vectorizer.transform(test_text))
    precision, recall, f1, _ = precision_recall_fscore_support(
        test_y, test_pred, average="binary", zero_division=0)

    raw = json.loads(gzip.decompress(SOURCE_FILES["sentences"].read_bytes()))
    rows = raw["sentences"]
    require(len(rows) == 4139 and all(row["index"] == i and row["text"]
                                      for i, row in enumerate(rows)),
            "Source sentence stream changed")
    source_pages = sources()
    require(len(source_pages) == 36 and
            set(row["page"] for row in rows) == set(source_pages),
            "Source page set changed")
    source_hashes = json.loads(SOURCE_FILES["source_hashes"].read_text())
    require(source_hashes == {name: hashlib.sha256(data).hexdigest()
                              for name, data in source_pages.items()},
            "Source bytes changed")
    for row in rows:
        source = source_pages[row["page"]]
        require(row["spans"] and all(0 <= span["start"] < span["end"] <= len(source)
                                     for span in row["spans"]),
                "Invalid original-byte span")

    probabilities = model.predict_proba(vectorizer.transform(
        [row["text"] for row in rows]))[:, 1].tolist()
    by_page: dict[str, list[int]] = defaultdict(list)
    for i, row in enumerate(rows):
        by_page[row["page"]].append(i)
    selected = []
    for page in sorted(source_pages):
        ranked = sorted(by_page[page], key=lambda i: (
            -probabilities[i], rows[i]["spans"][0]["start"], i))
        selected.extend(ranked[:3])
    require(len(selected) == 108 and len(set(selected)) == 108,
            "Candidate count changed")

    events = json.loads(SOURCE_FILES["events"].read_text())
    frozen = labels()
    frontier = json.loads(SOURCE_FILES["frontier"].read_text())
    additions = set(frontier["reviewed_additions"])
    require(len(events) == len(frozen) == 123 and
            {event["event"] for event in events} == set(frozen),
            "Frozen event ledger changed")
    require(len(additions) == 7 and all(
        event["status"] == "missed" for event in events
        if event["event"] in additions), "Reviewed additions changed")

    possible = []
    counts: dict[str, Counter] = defaultdict(Counter)
    for event in events:
        label = frozen[event["event"]]
        page = label["path"]
        for target in label["targets"]:
            require(source_pages[page][target["start"]:target["end"]].decode() ==
                    target["quote"], "Frozen target quote changed")
        matches = [i for i in selected if rows[i]["page"] == page and
                   any(overlapping(rows[i], target) for target in label["targets"])]
        complete = all(any(overlapping(rows[i], target) for i in matches)
                       for target in label["targets"])
        current_full = event["status"] == "full" or event["event"] in additions
        cohort = event["event"].split("/", 1)[0]
        counts[cohort]["events"] += 1
        counts[cohort]["baseline_full"] += current_full
        counts[cohort]["span_ceiling"] += complete
        counts[cohort]["new_span_ceiling"] += complete and not current_full
        if complete:
            possible.append({"event": event["event"], "category": label["category"],
                             "baseline_full": current_full,
                             "selected_indices": matches})

    new_possible = [item for item in possible if not item["baseline_full"]]
    coef_digest = hashlib.sha256(model.coef_.tobytes() +
                                 model.intercept_.tobytes()).hexdigest()
    result = {
        "format": "unswell-349-wikisqe-vague-screen-v1",
        "status": "negative_early_stop" if len(new_possible) < 8 else "requires_editorial_review",
        "source_tree": "e1aba8105d2062dbb09397e28292367f8edc95ca",
        "dataset": "ando55/WikiSQE_experiment/vague",
        "dataset_revision": DATASET_REVISION,
        "dataset_sha256": DATASET_SHA256,
        "local_input_sha256": {name: digest(path) for name, path in SOURCE_FILES.items()},
        "protocol_sha256": digest(HERE / "PROTOCOL.md"),
        "run_script_sha256": digest(HERE / "run.py"),
        "versions": {"python": platform.python_version(), "sklearn": sklearn.__version__,
                     "pyarrow": pyarrow.__version__},
        "train_examples": len(train_text),
        "train_labels": dict(Counter(train_y)),
        "test_examples": len(test_text),
        "test_labels": dict(Counter(test_y)),
        "test_sanity": {"accuracy": accuracy_score(test_y, test_pred),
                        "precision": precision, "recall": recall, "f1": f1},
        "model_coef_sha256": coef_digest,
        "source_sentences": len(rows),
        "source_pages": len(source_pages),
        "selected_candidates": len(selected),
        "cohorts": {cohort: dict(counter) for cohort, counter in sorted(counts.items())},
        "new_full_event_span_ceiling": len(new_possible),
        "new_possible_categories": dict(Counter(item["category"]
                                                 for item in new_possible)),
        "admission_min_new_full_events": 8,
        "admission_possible_by_span": len(new_possible) >= 8,
        "editorial_review_completed": False,
        "possible_events": possible,
        "credit_note": "Span overlap is only an optimistic bound, not diagnosis credit.",
    }
    scores = {"probabilities": probabilities, "selected_indices": selected}
    with (HERE / "scores.json.gz").open("wb") as file:
        with gzip.GzipFile(fileobj=file, mode="wb", mtime=0) as compressed:
            compressed.write(json.dumps(scores, separators=(",", ":")).encode())
    (HERE / "top-candidates.json").write_text(json.dumps([
        {"index": i, "page": rows[i]["page"], "text": rows[i]["text"],
         "spans": rows[i]["spans"], "score": probabilities[i]}
        for i in selected], indent=2) + "\n", encoding="utf-8")
    result["scores_sha256"] = digest(HERE / "scores.json.gz")
    result["top_candidates_sha256"] = digest(HERE / "top-candidates.json")
    (HERE / "result.json").write_text(json.dumps(result, indent=2, sort_keys=True) + "\n",
                                       encoding="utf-8")
    print(json.dumps({"status": result["status"],
                      "test_sanity": result["test_sanity"],
                      "new_full_event_span_ceiling": len(new_possible),
                      "cohorts": result["cohorts"]}, indent=2))


if __name__ == "__main__":
    main()
