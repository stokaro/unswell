#!/usr/bin/env python3
"""Collect pages that people wrote under a documentation style guide before November 2022.

usage: fetch_styled.py <out jsonl> [per source] [--list] [--articles]

Each source is a public repository whose project publishes a style guide for its writers. A source is read at its
last commit before 2022-11-01. Its Markdown pages are tried in the order of a hash of their path, and the first
ones with enough English prose are kept. With --list the script only counts the pages of each source.

A page goes through expand() and then through the same clean() as every other text of the measurement. expand()
puts the visible words in the place of a macro that stands inside a sentence, so that the sentence keeps its
words: a glossary link, a cross reference, a tooltip.
"""
import hashlib, json, re, subprocess, sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))
import prepare as lib

SOURCES = [
    ("gitlabhq/gitlabhq", ["doc/user", "doc/ci", "doc/topics"], "GitLab documentation style guide"),
    ("mdn/content", ["files/en-us/learn"], "MDN writing guidelines"),
    ("kubernetes/website", ["content/en/docs/concepts", "content/en/docs/tutorials"], "Kubernetes documentation style guide"),
    ("GoogleChrome/web.dev", ["src/site/content/en/blog"], "web.dev handbook, Google developer documentation style guide"),
    ("dotnet/docs", ["docs/standard", "docs/fundamentals"], "Microsoft Writing Style Guide"),
    ("alphagov/gds-way", ["source"], "GOV.UK content style guide"),
]
# Articles and tutorials, the genre of the posts that the measurement is about. Read with --articles.
ARTICLES = [
    ("GoogleChrome/web.dev", ["src/site/content/en/blog"], "web.dev handbook, Google developer documentation style guide"),
    ("GoogleChrome/developer.chrome.com", ["site/en/blog"], "developer.chrome.com handbook, Google developer documentation style guide"),
    ("GoogleCloudPlatform/community", ["tutorials"], "Google Cloud community tutorial style guide"),
    ("istio/istio.io", ["content/en/blog"], "Istio documentation style guide"),
]
UNTIL = "2022-10-31T23:59:59Z"
MIN_WORDS = 500
PAGE = re.compile(r"\.(md|mdx|markdown)(\.erb)?$")
INLINE_MACROS = "glossary|cssxref|domxref|jsxref|htmlelement|svgelement|httpheader|htmlattrxref|httpmethod|httpstatus|mathmlelement"


def api(path, raw=False):
    args = ["gh", "api", path]
    if raw:
        args += ["-H", "Accept: application/vnd.github.raw"]
    done = subprocess.run(args, capture_output=True)
    if done.returncode != 0:
        return None
    return done.stdout.decode("utf-8", "replace") if raw else json.loads(done.stdout)


def subtree(repo, sha, directory):
    """Return the tree of a directory at a commit, or None."""
    tree = sha
    for part in directory.split("/"):
        listing = api(f"repos/{repo}/git/trees/{tree}")
        entry = next((e for e in (listing or {}).get("tree", []) if e["path"] == part and e["type"] == "tree"), None)
        if not entry:
            return None
        tree = entry["sha"]
    return tree


def pages(repo, sha, directories):
    found, cut = [], False
    for directory in directories:
        tree = subtree(repo, sha, directory)
        if not tree:
            print(f"  {repo}: no directory {directory}", flush=True)
            continue
        listing = api(f"repos/{repo}/git/trees/{tree}?recursive=1") or {}
        cut = cut or bool(listing.get("truncated"))
        for entry in listing.get("tree", []):
            name = entry["path"].rsplit("/", 1)[-1]
            if entry["type"] == "blob" and PAGE.search(name) and not name.startswith("_"):
                found.append(directory + "/" + entry["path"])
    return found, cut


def expand(text):
    """Put the visible words in the place of a macro that stands inside a sentence."""
    shown = lambda match: match.group(2) or match.group(1)
    text = re.sub(r"\{\{\s*(?:" + INLINE_MACROS + r")\(\s*[\"']([^\"']*)[\"'](?:\s*,\s*[\"']([^\"']*)[\"'])?[^}]*\)\s*\}\}", shown, text,
                  flags=re.I)
    text = re.sub(r"\{\{<\s*glossary_tooltip\s+[^>]*?text=\"([^\"]*)\"[^>]*>\}\}", r"\1", text)
    text = re.sub(r"\{\{<\s*glossary_tooltip\s+[^>]*?term_id=\"([^\"]*)\"[^>]*>\}\}", lambda m: m.group(1).replace("-", " "), text)
    text = re.sub(r"\{%.*?%\}", "", text, flags=re.S)
    text = re.sub(r"<%.*?%>", "", text, flags=re.S)
    text = re.sub(r"\[!INCLUDE\s*\[[^\]]*\]\([^)]*\)\]", "", text)
    text = re.sub(r"<xref:([^>?]+)[^>]*>", lambda m: m.group(1).replace("%2A", "").rstrip("*").split(".")[-1], text)
    text = re.sub(r"\[!(NOTE|TIP|IMPORTANT|WARNING|CAUTION)\]", "", text)
    return text


def paragraph_share(text):
    """Share of the prose words of a cleaned page that stand in paragraphs, where a run of sentences can form."""
    inside = total = 0
    for line in text.split("\n"):
        count = lib.words(line)
        total += count
        if count and not lib.STARTS_LINE.match(line):
            inside += count
    return round(inside / total, 4) if total else 0.0


def strip_shortcodes(text):
    """Remove a shortcode that the preparation left: one that runs over several lines, or an empty pair of braces."""
    text = re.sub(r"\{\{[<%].*?[>%]\}\}", "", text, flags=re.S)
    return re.sub(r"\{\{\s*\}\}", "", text)


def english(text):
    letters = [c for c in text if c.isalpha()]
    return bool(letters) and sum(c.isascii() for c in letters) / len(letters) > 0.97


args = [a for a in sys.argv[1:] if not a.startswith("--")]
out, per_source = Path(args[0]), int(args[1]) if len(args) > 1 else 60
only_list = "--list" in sys.argv
rows = []
for repo, directories, guide in (ARTICLES if "--articles" in sys.argv else SOURCES):
    commits = api(f"repos/{repo}/commits?until={UNTIL}&per_page=1")
    if not commits:
        print(f"{repo}: not readable", flush=True)
        continue
    sha = commits[0]["sha"]
    listed, cut = pages(repo, sha, directories)
    paths = sorted(listed, key=lambda p: hashlib.sha256((repo + "/" + p).encode()).hexdigest())
    if only_list:
        print(f"{repo}: commit {sha[:10]} of {commits[0]['commit']['committer']['date'][:10]}, {len(paths)} pages listed"
              + (", listing cut short" if cut else ""), flush=True)
        continue
    kept, tried = 0, 0
    for path in paths:
        if kept >= per_source:
            break
        tried += 1
        raw = api(f"repos/{repo}/contents/{path}?ref={sha}", raw=True)
        if not raw:
            continue
        text = strip_shortcodes(lib.clean(expand(raw)))
        if lib.words(text) < MIN_WORDS or not english(text):
            continue
        key = f"{repo}/{path}"
        rows.append({"id": hashlib.sha256(key.encode()).hexdigest()[:16], "repo": repo, "path": path, "commit": sha, "guide": guide,
                     "title": lib.front_matter_title(raw), "words": lib.words(text), "paragraphs": paragraph_share(text),
                     "split": lib.bucket(key), "text": text})
        kept += 1
    print(f"{repo}: {len(paths)} pages listed, {tried} tried, {kept} kept", flush=True)
if not only_list:
    out.write_text("".join(json.dumps(row, ensure_ascii=False) + "\n" for row in rows))
    print(len(rows), "pages")
