"""Score the full frozen pilot only after complete-stream assistant review."""

import json
from pathlib import Path
import sys

import packet
import run

ROOT = Path(__file__).resolve().parent
RUN = ROOT / "run"


def score():
    run.verify()
    journal = packet.read(RUN / "run.json")
    if journal["status"] != "complete" or journal["complete"] is not True or \
            len(journal["calls"]) != 6 or any(c["status"] != "valid" for c in journal["calls"]):
        raise ValueError("incomplete pilot cannot be scored as complete")
    labels = packet.read(ROOT / "labels.json")
    review = packet.read(ROOT / "review.json")
    if review["format"] != "unswell-349-exhaustive-inventory-complete-stream-review-v1" or \
            "Codex assistant" not in review["reviewer"]:
        raise ValueError("reviewer identity or format changed")
    expected = {e["id"] for e in labels["events"]}
    if len(expected) != 4:
        raise ValueError("frozen event denominator changed")
    full = review["recalled_event_ids"]
    partial = review["partial_event_ids"]
    missed = review["missed_event_ids"]
    if len(full) != len(set(full)) or len(partial) != len(set(partial)) or \
            len(missed) != len(set(missed)) or \
            set(full) | set(partial) | set(missed) != expected or \
            set(full) & set(partial) or set(full) & set(missed) or set(partial) & set(missed):
        raise ValueError("review does not partition frozen events")
    delivered = []
    for slug, _ in packet.PAGES:
        data = packet.read(RUN / slug / "diagnosis" / "bound.json")
        delivered.extend((slug, index) for index, _ in enumerate(data["findings"]))
    rows = review["findings"]
    if type(review["protected_counterexamples_flagged"]) is not bool:
        raise ValueError("protected counterexample disposition missing")
    if len(rows) != len(delivered) or len({(row["page"], row["index"]) for row in rows}) != len(rows) or \
            {(row["page"], row["index"]) for row in rows} != set(delivered):
        raise ValueError("review omits or invents a delivered finding")
    if any(row["disposition"] not in ("accepted", "uncertain", "rejected") or
           not isinstance(row["rationale"], str) or not row["rationale"].strip() for row in rows):
        raise ValueError("review disposition or rationale missing")
    accepted = sum(row["disposition"] == "accepted" for row in rows)
    fraction = accepted / len(rows) if rows else None
    by_page = {slug: {"events": sum(e["page"] == page_id for e in labels["events"]),
                      "full": sum(e["page"] == page_id and e["id"] in full for e in labels["events"]),
                      "delivered": sum(row["page"] == slug for row in rows)}
               for slug, page_id in packet.PAGES}
    return {"format": "unswell-349-exhaustive-inventory-pilot-result-v2",
            "events": 4, "full_recalled": len(full), "partial": len(partial), "missed": len(missed),
            "by_page": by_page, "delivered": len(rows), "accepted": accepted,
            "uncertain": sum(row["disposition"] == "uncertain" for row in rows),
            "rejected": sum(row["disposition"] == "rejected" for row in rows),
            "accepted_fraction": fraction, "invalid_pages": 0,
            "protected_counterexamples_flagged": review["protected_counterexamples_flagged"],
            "supplied_edits": 0, "suggestion_safety": "not applicable",
            "advance": len(full) == 4 and fraction is not None and fraction >= 0.85 and
                       review["protected_counterexamples_flagged"] is False,
            "qualification": False,
            "qualification_note": "Directional three-page pilot only; #349 development and separate confirmation remain."}


if __name__ == "__main__":
    if sys.argv[1:] != ["score"]:
        raise SystemExit("usage: score.py score")
    print(json.dumps(score(), indent=2, sort_keys=True))
