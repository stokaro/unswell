"""Shared helpers for the measurement against human editors.

Offsets are bytes of the UTF-8 text, the unit that unswell reports.
"""
import json
import re

BOUNDARY = re.compile(r'[.!?]+["”’)\]]*(?=\s+["“‘(\[]?[A-Z0-9])|\n+')


def sentences(text):
    """Split text into sentences with one fixed rule and return byte spans.

    A sentence ends after final punctuation followed by white space and an
    upper-case letter or digit, and at every line break. Pieces without a
    letter or digit are dropped.
    """
    spans, start = [], 0
    for match in BOUNDARY.finditer(text):
        spans.append((start, match.end()))
        start = match.end()
    spans.append((start, len(text)))
    result = []
    for begin, end in spans:
        piece = text[begin:end]
        stripped = piece.strip()
        if not stripped or not re.search(r"[A-Za-z0-9]", stripped):
            continue
        lead = len(piece) - len(piece.lstrip())
        begin, end = begin + lead, begin + lead + len(stripped)
        result.append((len(text[:begin].encode()), len(text[:end].encode())))
    return result


def byte_span(text, start, end):
    """Convert character offsets to byte offsets."""
    return len(text[:start].encode()), len(text[:end].encode())


def overlaps(span, others):
    return any(span[0] < end and start < span[1] for start, end in others)


def read_jsonl(path):
    with open(path) as handle:
        return [json.loads(line) for line in handle if line.strip()]


def write_jsonl(path, rows):
    with open(path, "w") as handle:
        for row in rows:
            handle.write(json.dumps(row, ensure_ascii=False, sort_keys=True) + "\n")
