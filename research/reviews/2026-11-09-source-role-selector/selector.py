"""Conservative source-role admission for archived authorial-narration proposals."""

from __future__ import annotations

import re
from dataclasses import dataclass


@dataclass(frozen=True)
class Decision:
    action: str
    reason: str


def _normalize(value: str) -> str:
    return " ".join(value.split()).lower()


def _paragraph(source: bytes, start: int, end: int) -> tuple[str, str]:
    before = source.rfind(b"\n\n", 0, start)
    after = source.find(b"\n\n", end)
    paragraph = source[before + 2 if before >= 0 else 0 : after if after >= 0 else len(source)]
    following = source[after + 2 : after + 202] if after >= 0 else b""
    return paragraph.decode("utf-8"), following.decode("utf-8", errors="replace")


def classify_target(source: bytes, start: int, end: int) -> Decision:
    """Classify one original byte target using its complete source paragraph."""
    if not (0 <= start < end <= len(source)):
        raise ValueError("target span is outside the original source")
    target = source[start:end].decode("utf-8")
    paragraph, following = _paragraph(source, start, end)
    text = _normalize(target)
    block = _normalize(paragraph)

    # Role guards take precedence over lexical resemblance to an introduction.
    if re.search(r"\b(?:refer to|defined as|denoted by|we call)\b", text) and re.search(r"\bas\b", text):
        return Decision("withhold", "term definition")
    if re.search(r"\b(?:while|although|though)\b.*\b(?:value|benefit|useful|debug tool)\b", text):
        return Decision("withhold", "technical concession")
    if re.search(r"\b(?:please be careful|at your own risk|only do this if)\b", text):
        return Decision("withhold", "safety qualification")
    if re.search(r"\b(?:unprecedented|compared to any|best|fastest)\b", text):
        return Decision("withhold", "separate comparative claim")
    if re.search(r"\[[^]]+\]\([^)]+\)", paragraph) or "`" in paragraph:
        return Decision("withhold", "linked or protected task context")
    if re.search(r"\b(?:navigate|open up|enter|run|follow|create|visit)\b", block) and re.search(
        r"\b(?:directory|file|terminal|command|issue|api key)\b", block
    ):
        return Decision("withhold", "concrete reader action")
    if text.endswith("then run") and following.lstrip().startswith("```"):
        return Decision("withhold", "command introduction")
    if re.search(r"\d", text):
        return Decision("withhold", "numeric technical content")
    if re.search(r"(?:🎉|\bawesome\b|\bso happy\b)", paragraph, re.IGNORECASE):
        return Decision("abstain", "deliberate author voice")

    if re.search(r"\b(?:i|we) hope\b", text) and re.search(r"\b(?:guide|document|page|process)\b", text):
        return Decision("admit", "document outcome wish")
    if re.search(r"\b(?:let's|we will|now it is time|now it's time|next,? we will)\b", text) and re.search(
        r"\b(?:begin|explore|look|dig|start)\b", text
    ):
        return Decision("admit", "generic navigation")
    if re.search(r"\bto champion\b.*\bwe have tried\b", text):
        return Decision("admit", "authorial process")
    if re.search(r"\bthis is the place to start\b", text) and not re.search(r"\b(?:at|in)\s+https?://", text):
        return Decision("admit", "unnamed navigation")
    if re.search(r"\b(?:this page|this file|this document)\b.*\b(?:explains|describes|says)\b", text):
        return Decision("admit", "document announcement")
    return Decision("abstain", "no bounded role decision")


def classify_spans(source: bytes, spans: list[tuple[int, int]]) -> Decision:
    """Require every target in a proposal to support the same admission."""
    if not spans:
        raise ValueError("proposal has no targets")
    decisions = [classify_target(source, start, end) for start, end in spans]
    if any(item.action == "withhold" for item in decisions):
        return next(item for item in decisions if item.action == "withhold")
    if all(item.action == "admit" for item in decisions):
        return Decision("admit", "+".join(item.reason for item in decisions))
    return Decision("abstain", "at least one target lacks an admission role")
