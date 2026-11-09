#!/usr/bin/env python3
"""Offline source-function candidate using the engine's prepared source spans."""

import argparse
import hashlib
import json
import re
from pathlib import Path


WORD = re.compile(r"[a-z][a-z0-9-]*", re.IGNORECASE)
STOP = {
    "about", "after", "along", "also", "and", "are", "before", "can",
    "for", "from", "has", "have", "into", "its", "may", "more", "not",
    "of", "our", "that", "the", "their", "then", "these", "this", "those",
    "very", "was", "were", "with", "your", "important", "it", "is", "be",
}

TABLE_ANNOUNCE = re.compile(
    r"\b(?:is|are) (?:shown|listed|presented|summarized) in the following table\b",
    re.IGNORECASE,
)
LIMITATION_ANNOUNCE = re.compile(
    r"\b(?:has|have) (?:its|their) own limitation(?:s)?\.?\Z",
    re.IGNORECASE,
)
EASE_ANNOUNCE = re.compile(
    r"\bIt is (?:very |quite )?(?:simple|easy) to (?:use|install|configure)\b",
    re.IGNORECASE,
)
COMMAND_NARRATION = re.compile(
    r"\bThis will install ([A-Za-z0-9_.+ -]+?) to your system\.?\Z",
    re.IGNORECASE,
)
QUALITY = [
    ("rank", re.compile(r"\bbest-of-class\b|\bthe best\b.{0,55}\bin the industry\b", re.I)),
    ("generation", re.compile(r"\bnext[- ]generation\b", re.I)),
    ("unsupported_praise", re.compile(r"\b(?:many )?outstanding advantages\b|\bwidely praised\b", re.I)),
    ("stacked_praise", re.compile(r"\b(?:extremely )?powerful and flexible\b", re.I)),
    ("ease", re.compile(r"\bquite trivial\b|\ba wealth of\b", re.I)),
    ("vague_degree", re.compile(r"\badvanced\b|\brich\b", re.I)),
]


def words(text: str) -> set[str]:
    return {token.lower() for token in WORD.findall(text) if token.lower() not in STOP}


def byte_span(source: bytes, sentence: dict, match: re.Match[str]) -> tuple[int, int]:
    start = sentence["span"]["start"]
    raw = source[start:sentence["span"]["end"]].decode("utf-8")
    return (start + len(raw[:match.start()].encode("utf-8")),
            start + len(raw[:match.end()].encode("utf-8")))


def intersects_excluded(start: int, end: int, excluded: list[dict]) -> bool:
    return any(start < region["span"]["end"] and end > region["span"]["start"]
               for region in excluded)


def previous_command_names(source: bytes, sentence_start: int) -> set[str]:
    preceding = source[max(0, sentence_start - 350):sentence_start].decode("utf-8", "ignore")
    matches = re.findall(r"(?:apt-get|apt|brew|dnf|yum) install ([^\n`]+)", preceding, re.I)
    if not matches:
        return set()
    return words(matches[-1])


def jaccard(a: set[str], b: set[str]) -> float:
    return len(a & b) / len(a | b) if a and b else 0.0


def sentence_text(source: bytes, sentence: dict) -> str:
    return source[sentence["span"]["start"]:sentence["span"]["end"]].decode("utf-8")


def candidate_for_sentence(source: bytes, sentence: dict, previous: list[dict],
                           following: dict | None, excluded: list[dict]) -> dict | None:
    text = sentence_text(source, sentence)
    stripped = text.strip()
    start, end = sentence["span"]["start"], sentence["span"]["end"]
    if not stripped or stripped.startswith("#"):
        return None
    if len(WORD.findall(text)) < 4:
        return None

    def diagnosis(operator: str, target_start: int, target_end: int,
                  message: str, retain: str, deletion: bool = False) -> dict | None:
        if intersects_excluded(target_start, target_end, excluded):
            return None
        result = {"operator": operator, "start_byte": target_start,
                  "end_byte": target_end, "quote": source[target_start:target_end].decode("utf-8"),
                  "diagnosis": message, "content_to_retain": retain,
                  "edit_status": "proposed_unverified" if deletion else "unresolved"}
        if deletion:
            result["edit"] = {"operation": "delete", "start_byte": target_start,
                              "end_byte": target_end}
        return result

    if TABLE_ANNOUNCE.search(text) and len(WORD.findall(text)) <= 22:
        nearby = source[end:min(len(source), end + 900)]
        if b"\n|" in nearby or b"\n:" in nearby or b"\ndemo " in nearby:
            return diagnosis("empty_table_announcement", start, end,
                             "The sentence only announces the nearby table.",
                             "Keep the table and any separate instruction for using it.", True)

    if LIMITATION_ANNOUNCE.search(stripped) and following:
        next_text = sentence_text(source, following)
        if len(words(next_text)) >= 5 and following["span"]["start"] - end < 160:
            return diagnosis("empty_limitation_announcement", start, end,
                             "The next sentence names the limitation; this one only announces it.",
                             "Keep the specific limitation and its conditions.", True)

    if EASE_ANNOUNCE.search(text) and len(WORD.findall(text)) <= 12:
        nearby = source[end:min(len(source), end + 500)].decode("utf-8", "ignore")
        if re.search(r"(?m)^\s*1[.)]", nearby):
            return diagnosis("unsupported_ease_announcement", start, end,
                             "A bare ease claim precedes concrete steps.",
                             "Keep the complete steps and prerequisites.", True)

    command = COMMAND_NARRATION.search(stripped)
    if command:
        named = words(command.group(1)) - {"and"}
        if named and named <= previous_command_names(source, start):
            return diagnosis("duplicate_command_narration", start, end,
                             "The sentence repeats packages named in the preceding install command.",
                             "Keep the exact command and package names.", True)

    if not re.search(r"\]\(|https?://", text) and len(words(text)) >= 5:
        current = words(text)
        for index, earlier in enumerate(previous):
            if start - earlier["span"]["end"] > 5000:
                continue
            earlier_text = sentence_text(source, earlier)
            if re.search(r"\]\(|https?://", earlier_text):
                continue
            prior = words(earlier_text)
            if earlier_text.lstrip().lower().startswith("it ") and index:
                antecedent = sentence_text(source, previous[index - 1])
                if earlier["span"]["start"] - previous[index - 1]["span"]["end"] < 10:
                    prior |= words(antecedent)
            if len(prior) >= 5 and jaccard(current, prior) >= 0.78:
                return diagnosis("repeated_proposition", start, end,
                                 "This repeats an earlier claim without adding a condition or step.",
                                 "Keep the earlier complete claim and any distinct warning.", True)

    if re.search(r"\b(?:if|unless|must|should|cannot|invalid|unsafe|risk)\b", text, re.I):
        return None
    for quality_type, pattern in QUALITY:
        match = pattern.search(text)
        if not match:
            continue
        if quality_type == "stacked_praise" and match.group().lower() == "powerful and flexible":
            next_text = sentence_text(source, following) if following else ""
            if next_text.lstrip().lower().startswith(("however", "but")):
                continue
        target_start, target_end = byte_span(source, sentence, match)
        return diagnosis("unsupported_quality_" + quality_type, target_start, target_end,
                         "The evaluative wording supplies no checkable property at this location.",
                         "Keep the named product, capability, scope, and any comparison evidence.")
    return None


def run_page(source: bytes, report: dict, number: int) -> tuple[list[dict], int]:
    documents = report["documents"]
    if report["status"] != "complete" or report["errors"] or len(documents) != 1:
        raise ValueError(f"page {number}: incomplete CLI report")
    document = documents[0]
    if document["source_hash"] != hashlib.sha256(source).hexdigest() or document["bytes"] != len(source):
        raise ValueError(f"page {number}: source mismatch")
    excluded = document["excluded"]
    sentences = sorted((a for a in report["assessments"] if a["scope"] == "sentence"),
                       key=lambda a: (a["span"]["start"], a["span"]["end"]))
    spans = set()
    unique = []
    for sentence in sentences:
        span = sentence["span"]
        key = (span["start"], span["end"])
        if key in spans:
            continue
        if not 0 <= key[0] < key[1] <= len(source):
            raise ValueError(f"page {number}: invalid sentence span")
        spans.add(key)
        unique.append(sentence)
    findings = []
    for index, sentence in enumerate(unique):
        following = unique[index + 1] if index + 1 < len(unique) else None
        result = candidate_for_sentence(source, sentence, unique[:index], following, excluded)
        if result:
            result["page"] = number
            result["source_sha256"] = document["source_hash"]
            findings.append(result)
    return findings, len(unique)


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--sources-dir", type=Path, required=True)
    parser.add_argument("--baseline-dir", type=Path, required=True)
    parser.add_argument("--audit", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    if args.output.exists():
        raise ValueError("candidate output already exists")
    audit = json.loads(args.audit.read_bytes())
    if len(audit["reports"]) != 9:
        raise ValueError("expected nine pinned reports")
    results = []
    pages = []
    for number, expected in enumerate(audit["reports"], 1):
        if expected["page"] != number:
            raise ValueError("report order changed")
        source = (args.sources_dir / f"page-{number:03d}.source").read_bytes()
        report_bytes = (args.baseline_dir / f"page-{number:03d}.json").read_bytes()
        if hashlib.sha256(report_bytes).hexdigest() != expected["sha256"]:
            raise ValueError(f"page {number}: report changed")
        findings, sentences = run_page(source, json.loads(report_bytes), number)
        results.extend(findings)
        pages.append({"page": number, "source_sha256": hashlib.sha256(source).hexdigest(),
                      "eligible_sentences": sentences, "diagnoses": len(findings)})
    for index, result in enumerate(results, 1):
        result["id"] = f"candidate-{index:03d}"
    output = {"format": "unswell-349-source-function-candidate-v1",
              "baseline_audit_sha256": hashlib.sha256(args.audit.read_bytes()).hexdigest(),
              "pages": pages, "diagnoses": results}
    args.output.write_text(json.dumps(output, indent=2, ensure_ascii=False) + "\n")
    print(f"analyzed nine pages; emitted {len(results)} diagnoses")


if __name__ == "__main__":
    main()
