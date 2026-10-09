"""Bind partially observed editorial decisions without inventing negatives.

This research adapter does not call an annotator, fit a model, or grade source
correctness. Source review remains necessary. Intent names are nonexclusive;
choosing one purpose never rules out another purpose.
"""

INTENTS = (
    "clarity", "empty_framing", "fluency", "needless_complexity",
    "needless_repetition", "unjustified_intensifiers", "vague_claims", "wordiness",
)
INTENT_STATUSES = {"necessary", "not_necessary", "optional", "unknown"}
NECESSITY_STATUSES = {"necessary", "retain", "optional", "unknown"}


def _decision(value, statuses, description):
    if not isinstance(value, dict) or set(value) - {"status", "reason"}:
        raise ValueError(description + " needs an explicit decision object")
    status = value.get("status", "unknown")
    reason = value.get("reason")
    if not isinstance(status, str) or status not in statuses:
        raise ValueError("Unsupported " + description + " decision")
    if reason is not None and not isinstance(reason, str):
        raise ValueError(description + " reason must be text")
    if status != "unknown" and (reason is None or not reason.strip()):
        raise ValueError("An observed " + description + " decision needs its own reason")
    return status, reason


def normalize_review(response, *, source, source_start=0, protected=()):
    """Return explicit decisions and verified absolute UTF-8 edit spans.

    Unmentioned intent dimensions remain unknown. A revised source needs its
    own response; neither a successful repair nor text equality supplies labels.
    Protected ranges use the original document's absolute byte coordinates.
    """
    if not isinstance(response, dict) or set(response) - {"necessity", "intents", "edits"}:
        raise ValueError("Unexpected annotation response fields")
    if not isinstance(source, str) or type(source_start) is not int or source_start < 0:
        raise ValueError("Source and its absolute starting byte are required")
    status, necessity_reason = _decision(response.get("necessity", {}), NECESSITY_STATUSES, "necessity")
    if not isinstance(protected, (list, tuple)):
        raise ValueError("Protected ranges must be a sequence of absolute byte spans")
    for span in protected:
        if (not isinstance(span, dict) or set(span) != {"start", "end"}
                or type(span["start"]) is not int or type(span["end"]) is not int
                or not 0 <= span["start"] < span["end"]):
            raise ValueError("Invalid protected absolute byte span")
    decisions = {name: {"status": "unknown", "reason": None, "origin": "unobserved"}
                 for name in INTENTS}
    reported = response.get("intents", {})
    if not isinstance(reported, dict) or set(reported) - set(INTENTS):
        raise ValueError("Unsupported intent dimension")
    for name, decision in reported.items():
        selected, reason = _decision(decision, INTENT_STATUSES, "intent")
        decisions[name] = {"status": selected, "reason": reason, "origin": "explicit_dimension"}
    raw = source.encode("utf-8")
    edits = response.get("edits", [])
    if not isinstance(edits, list):
        raise ValueError("Edits must be a list")
    bound = []
    for edit in edits:
        allowed = {"quote", "occurrence", "replacement", "intents", "status", "reason"}
        if not isinstance(edit, dict) or set(edit) - allowed:
            raise ValueError("Unexpected edit fields")
        quote = edit.get("quote")
        names = edit.get("intents")
        purpose = edit.get("status")
        reason = edit.get("reason")
        if not isinstance(quote, str) or not quote:
            raise ValueError("An edit needs an exact nonempty source quote")
        if (not isinstance(names, list) or not names or not all(isinstance(name, str) for name in names)
                or len(set(names)) != len(names)
                or not set(names) <= set(INTENTS)):
            raise ValueError("An edit needs one or more nonexclusive intent names")
        if (not isinstance(purpose, str) or purpose not in {"necessary", "optional"}
                or not isinstance(reason, str) or not reason.strip()):
            raise ValueError("An edit needs a separate necessity and reason")
        if purpose == "necessary" and status in {"retain", "optional"}:
            raise ValueError("A mandatory edit contradicts the unit's necessity decision")
        replacement = edit.get("replacement")
        if replacement is not None and (not isinstance(replacement, str) or replacement == quote):
            raise ValueError("A supplied replacement must change the quoted text")
        needle = quote.encode("utf-8")
        locations = []
        at = raw.find(needle)
        while at >= 0:
            locations.append(at)
            at = raw.find(needle, at + 1)
        occurrence = edit.get("occurrence")
        if occurrence is None:
            if len(locations) != 1:
                raise ValueError("Missing or repeated quote needs exact disambiguation")
            occurrence = 0
        if type(occurrence) is not int or not 0 <= occurrence < len(locations):
            raise ValueError("Invalid quote occurrence")
        start = source_start + locations[occurrence]
        end = start + len(needle)
        if any(start < span["end"] and span["start"] < end for span in protected):
            raise ValueError("Edit crosses protected original source")
        if any(start < other["span"]["end"] and other["span"]["start"] < end for other in bound):
            raise ValueError("Overlapping edits must be represented together")
        for name in names:
            previous = decisions[name]
            if (purpose == "necessary" and previous["origin"] == "explicit_dimension"
                    and previous["status"] in {"not_necessary", "optional"}):
                raise ValueError("Mandatory edit conflicts with an explicit intent decision")
            if previous["status"] == "unknown":
                decisions[name] = {"status": purpose, "reason": reason, "origin": "named_edit"}
            elif (purpose == "necessary" and previous["status"] == "optional"
                  and previous["origin"] == "named_edit"):
                decisions[name] = {"status": purpose, "reason": reason, "origin": "named_edit"}
        bound.append({"span": {"start": start, "end": end}, "quote": quote,
                      "replacement": replacement, "intents": list(names), "status": purpose,
                      "reason": reason, "occurrence": occurrence})
    if status == "necessary" and not any(edit["status"] == "necessary" for edit in bound):
        raise ValueError("A necessary source review needs a localized mandatory target")
    for name, decision in decisions.items():
        if decision["status"] == "necessary" and not any(name in edit["intents"] and
                                                          edit["status"] == "necessary" for edit in bound):
            raise ValueError("A necessary intent needs a localized mandatory target")
    labels = {name: {"necessary": 1, "not_necessary": 0}.get(row["status"])
              for name, row in decisions.items()}
    return {"necessity": {"status": status, "reason": necessity_reason,
                          "loss_label": {"necessary": 1, "retain": 0}.get(status)},
            "intent_decisions": decisions, "intent_loss_labels": labels,
            "intent_loss_mask": {name: value is not None for name, value in labels.items()},
            "edits": bound, "automatic_after_label": False,
            "unmarked_tokens_are_background": False}
