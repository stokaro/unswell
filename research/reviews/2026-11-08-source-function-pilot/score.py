"""Score a frozen source-function pilot after the separate complete-stream review."""

import json
from pathlib import Path

ROOT = Path(__file__).resolve().parent
PAGES = ("caddy", "restic", "gitea", "otelcol")


def read(path):
    return json.loads(path.read_text(encoding="utf-8"))


def score():
    labels = read(ROOT / "labels.json")
    review = read(ROOT / "review.json")
    expected = {event["id"] for event in labels["events"]}
    if len(expected) != 8:
        raise ValueError("frozen event denominator changed")
    got = set(review["recalled_event_ids"])
    if not got <= expected or len(got) != len(review["recalled_event_ids"]):
        raise ValueError("unknown or duplicate recalled event")
    rows = review["findings"]
    if len({(row["page"], row["index"]) for row in rows}) != len(rows):
        raise ValueError("duplicate review row")
    bound = []
    for page in PAGES:
        data = read(ROOT / "run" / page / "bound.json")
        bound.extend((page, index) for index, _ in enumerate(data["findings"]))
    if set(bound) != {(row["page"], row["index"]) for row in rows}:
        raise ValueError("review omits or invents a delivered finding")
    if any(row["disposition"] not in ("accepted", "uncertain", "rejected") for row in rows):
        raise ValueError("invalid disposition")
    accepted = sum(row["disposition"] == "accepted" for row in rows)
    by_page = {page: {"events": sum(event["page"] == page for event in labels["events"]),
                       "recalled": sum(event["page"] == page and event["id"] in got
                                       for event in labels["events"])} for page in PAGES}
    result = {"events": 8, "recalled": len(got), "by_page": by_page,
              "delivered": len(rows), "accepted": accepted,
              "uncertain": sum(row["disposition"] == "uncertain" for row in rows),
              "rejected": sum(row["disposition"] == "rejected" for row in rows),
              "accepted_fraction": accepted / len(rows) if rows else None,
              "supplied_edits": 0, "suggestion_safety": "not applicable",
              "invalid_pages": 0,
              "advance": len(got) >= 7 and len(rows) > 0 and
                         accepted / len(rows) >= 0.85 and
                         not review["protected_counterexamples_flagged"]}
    return result


if __name__ == "__main__":
    print(json.dumps(score(), indent=2, sort_keys=True))
