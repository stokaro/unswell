#!/usr/bin/env python3
"""Tell whether a Markdown text is English prose, by its share of common English function words."""
import re

FUNCTION = frozenset("""the of and to a in is for that with on as this it be are by or from an at not you your can will if
when which we all has have but its into use used using each one more than then so no only also may must should does do
how what any other these there their they our was were been out up about after before between both such through under
over per via while where who new see make set run get add""".split())
LETTERS = re.compile(r"[^\W\d_]+", re.UNICODE)


def prose(text):
    text = re.sub(r"(?ms)^(```|~~~).*?^\1[^\n]*$", " ", text)
    text = re.sub(r"`[^`\n]*`", " ", text)
    return re.sub(r"https?://\S+|\]\([^)]*\)|<[^>\n]*>", " ", text)


def share(text):
    """Share of the words of the prose that are English function words, and the number of words."""
    words = [word.lower() for word in LETTERS.findall(prose(text))]
    return (sum(word in FUNCTION for word in words) / len(words) if words else 0.0), len(words)


def english(text, floor=0.2):
    value, words = share(text)
    return words >= 30 and value >= floor
