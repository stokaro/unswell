"""Shared text preparation for the measurement of text written under a style guide.

Every text that a judge or the tool reads goes through clean(): a human post, a post of the style-guide set, and a
post that a model wrote. The same rules for all of them keep the comparison about the prose.
"""
import hashlib
import re

FENCE = re.compile(r"^\s*(```|~~~)")


def front_matter_title(text):
    """Return the title of the front matter, or of the first heading."""
    match = re.match(r"^(---|\+\+\+)\s*\n(.*?)\n\1\s*\n", text, re.S)
    if match:
        title = re.search(r"^title\s*[:=]\s*(.+)$", match.group(2), re.M)
        if title:
            return title.group(1).strip().strip("\"'")
    heading = re.search(r"^#\s+(.+)$", text, re.M)
    return heading.group(1).strip() if heading else ""


def clean(text):
    """Reduce a Markdown or MDX post to its prose: headings, paragraphs, and lists.

    Front matter, imports, components, images, shortcodes, and bylines go. A fenced code block becomes one line
    "[code]". Each paragraph and each list item becomes one line, so that the way a source wraps its lines does
    not tell the sources apart.
    """
    text = text.replace("\r\n", "\n")
    # Typographic quotes come from an editor or a site generator, not from the writer.
    for mark, plain in (("\u2018", "'"), ("\u2019", "'"), ("\u201c", '"'), ("\u201d", '"'), ("\u00a0", " ")):
        text = text.replace(mark, plain)
    text = re.sub(r"^(---|\+\+\+)\s*\n.*?\n\1\s*\n", "", text, count=1, flags=re.S)
    lines, fenced = [], False
    for line in text.split("\n"):
        if FENCE.match(line):
            if not fenced:
                lines.append("[code]")
            fenced = not fenced
            continue
        if fenced:
            continue
        stripped = line.strip()
        if re.match(r"^(import|export)\s", stripped) or stripped.startswith(":::") or stripped.startswith("<!--"):
            continue
        if re.match(r"^\{\{[<%].*[>%]\}\}$", stripped) or re.match(r"^</?[A-Za-z][^>]*>$", stripped):
            continue
        if BYLINE.match(stripped):
            continue
        line = re.sub(r"!\[[^\]]*\]\([^)]*\)", "", line)
        line = re.sub(r"\{\{.*?\}\}", "", line)
        line = re.sub(r"\[([^\]]+)\]\([^)]*\)", r"\1", line)
        lines.append(line.rstrip())
    text = "\n".join(lines)
    # A tag can run over several lines: an embedded frame with its attributes.
    text = re.sub(r"</?[A-Za-z][^<>]{0,600}>", "", text)
    blocks = [reflow(block) for block in re.split(r"\n\s*\n", text)]
    return "\n\n".join(block for block in blocks if block.strip()).strip() + "\n"


BYLINE = re.compile(r"^[*_]{0,2}(authors?|by|posted by|written by|guest post by)\b[*_:]{0,3}\s.{0,160}$", re.I)
STARTS_LINE = re.compile(r"^\s*([-*+]\s|\d+[.)]\s|#{1,6}\s|\||>|\[code\]$)")


def reflow(block):
    """Join the wrapped lines of one block: a paragraph or a list item becomes one line."""
    joined = []
    for line in block.split("\n"):
        if not line.strip():
            continue
        if joined and not STARTS_LINE.match(line) and joined[-1].strip() != "[code]" and not joined[-1].lstrip().startswith(("#", "|")):
            joined[-1] += " " + line.strip()
        else:
            joined.append(line.rstrip())
    return "\n".join(joined)


def words(text):
    """Count the words of the prose, without the code placeholders."""
    return len([token for token in text.split() if token != "[code]"])


def excerpt(text, limit=450):
    """Return the opening of a cleaned post: whole blocks until the limit of words is reached."""
    blocks, total = [], 0
    for block in text.strip().split("\n\n"):
        blocks.append(block)
        total += words(block)
        if total >= limit:
            break
    return "\n\n".join(blocks).strip() + "\n"


def bucket(key, share=0.5):
    """Assign a key to the fitting half or the measured half by its hash."""
    value = int(hashlib.sha256(("split:" + key).encode()).hexdigest()[:8], 16) / 0xFFFFFFFF
    return "dev" if value < share else "test"
