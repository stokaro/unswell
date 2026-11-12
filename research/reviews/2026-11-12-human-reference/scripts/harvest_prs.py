#!/usr/bin/env python3
"""Collect documentation pull requests with an editorial title, merged before November 2022.

usage: harvest_prs.py <human directory> [--per-query N] [--workers N]

A pull request is kept when it changes at most three files, all Markdown, and
its patch removes or rewrites between 1 and 40 lines of one file of at most
40,000 bytes of mostly ASCII text. The base file is fetched at the base commit
and must contain every removed line where the patch says it is. One pull
request is kept per repository. The reference is the words of the base file
that the pull request removed or replaced.

Writes raw/prs/candidates.jsonl (search results), raw/prs/kept.jsonl (documents
with text and reference spans), and prints the counts of every filter.
"""
import argparse, collections, concurrent.futures, difflib, hashlib, json, re, subprocess, sys, time
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))
from common import byte_span, write_jsonl

QUERIES = ["copy edit", "copyedit", "reword", "rephrase", "wording", "concise", "wordy", "verbose", "redundant",
           "tighten", "readability", "simplify docs"]
BEFORE = "2022-11-01"
MARKDOWN = (".md", ".mdx", ".markdown")
EXAMINE = 1200

parser = argparse.ArgumentParser()
parser.add_argument("root")
parser.add_argument("--per-query", type=int, default=300)
parser.add_argument("--workers", type=int, default=4)
options = parser.parse_args()
RAW = Path(options.root) / "raw" / "prs"
RAW.mkdir(parents=True, exist_ok=True)


def api(path, *fields, raw=False):
    args = ["gh", "api", "-X", "GET", path]
    if raw:
        args += ["-H", "Accept: application/vnd.github.raw"]
    for field in fields:
        args += ["-f", field]
    for attempt in range(4):
        done = subprocess.run(args, capture_output=True)
        if done.returncode == 0:
            return done.stdout if raw else json.loads(done.stdout)
        message = done.stderr.decode("utf-8", "replace")
        if "rate limit" in message.lower() or "secondary" in message.lower():
            time.sleep(30 * (attempt + 1))
            continue
        if "404" in message or "Not Found" in message:
            return None
        time.sleep(3)
    return None


candidates_path = RAW / "candidates.jsonl"
if not candidates_path.exists():
    seen, rows = set(), []
    for query in QUERIES:
        for page in range(1, options.per_query // 100 + 1):
            found = api("search/issues", f'q=is:pr is:merged merged:<{BEFORE} in:title "{query}"', "per_page=100", f"page={page}")
            time.sleep(2.2)
            if not found or not found.get("items"):
                break
            for item in found["items"]:
                repo = item["repository_url"].split("/repos/", 1)[1]
                key = (repo, item["number"])
                if key in seen:
                    continue
                seen.add(key)
                rows.append({"repo": repo, "number": item["number"], "title": item["title"], "query": query,
                             "merged": (item.get("pull_request") or {}).get("merged_at") or item.get("closed_at")})
            if len(found["items"]) < 100:
                break
        print("searched", query, "candidates so far", len(rows), flush=True)
    write_jsonl(candidates_path, rows)
candidates = [json.loads(line) for line in open(candidates_path)]
print("candidates", len(candidates), collections.Counter(c["query"] for c in candidates).most_common())


def hunks(patch):
    """Yield (old start line, old lines, new lines) for every hunk of a unified patch."""
    current = None
    for line in patch.split("\n"):
        header = re.match(r"@@ -(\d+)(?:,\d+)? \+\d+(?:,\d+)? @@", line)
        if header:
            if current:
                yield current
            current = [int(header.group(1)), [], []]
        elif current is not None and not line.startswith("\\"):
            if line.startswith("-"):
                current[1].append(("-", line[1:]))
                current[2].append(None)
            elif line.startswith("+"):
                current[2].append(line[1:])
            else:
                current[1].append((" ", line[1:]))
                current[2].append(line[1:])
    if current:
        yield current


def reference(text, patch):
    """Return byte spans of the words that the patch removed or replaced, or None if it does not apply."""
    lines = text.split("\n")
    offsets, position = [], 0
    for line in lines:
        offsets.append(position)
        position += len(line) + 1
    spans, removed = [], 0
    for start, old, new in hunks(patch):
        old_text = [content for _, content in old]
        if lines[start - 1:start - 1 + len(old_text)] != old_text:
            return None, 0
        new_text = [content for content in new if content is not None]
        old_tokens = []
        for index, (mark, content) in enumerate(old):
            base = offsets[start - 1 + index]
            for match in re.finditer(r"\S+", content):
                old_tokens.append((match.group(), base + match.start(), base + match.end()))
        new_tokens = re.findall(r"\S+", "\n".join(new_text))
        matcher = difflib.SequenceMatcher(a=[t[0] for t in old_tokens], b=new_tokens, autojunk=False)
        for tag, a0, a1, _, _ in matcher.get_opcodes():
            if tag in ("replace", "delete") and a1 > a0:
                spans.append(list(byte_span(text, old_tokens[a0][1], old_tokens[a1 - 1][2])))
        removed += sum(mark == "-" for mark, _ in old)
    return spans, removed


def examine(candidate):
    repo, number = candidate["repo"], candidate["number"]
    files = api(f"repos/{repo}/pulls/{number}/files", "per_page=20")
    if files is None:
        return "pull request not readable", None
    if not 1 <= len(files) <= 3:
        return "more than three files", None
    if not all(f["filename"].lower().endswith(MARKDOWN) for f in files):
        return "not only Markdown", None
    usable = [f for f in files if f.get("status") == "modified" and f.get("patch") and 1 <= f.get("deletions", 0) <= 40]
    if not usable:
        return "no file with 1 to 40 removed lines", None
    pull = api(f"repos/{repo}/pulls/{number}")
    if pull is None or not pull.get("base", {}).get("sha"):
        return "no base commit", None
    changed = max(usable, key=lambda f: f["deletions"])
    for sha in (pull["base"]["sha"], (pull.get("merge_commit_sha") or "") + "^"):
        if not sha.strip("^"):
            continue
        content = api(f"repos/{repo}/contents/{changed['filename']}", f"ref={sha}", raw=True)
        if content is None:
            continue
        try:
            text = content.decode("utf-8")
        except UnicodeDecodeError:
            return "not UTF-8", None
        if len(content) > 40000:
            return "file over 40,000 bytes", None
        if sum(ch < 128 for ch in content) < 0.9 * len(content):
            return "not mostly ASCII", None
        spans, removed = reference(text, changed["patch"])
        if spans is None:
            continue
        if not spans:
            return "no removed or replaced words", None
        return "kept", {"id": "pr-" + hashlib.sha256(f"{repo}#{number}".encode()).hexdigest()[:12], "set": "prs",
                        "format": "markdown", "text": text, "editor": pull["user"]["login"] if pull.get("user") else "",
                        "kind": candidate["query"], "generator": "", "repo": repo, "number": number, "path": changed["filename"],
                        "base": sha, "title": candidate["title"], "merged": candidate["merged"],
                        "reference": [{"span": span, "category": "pull request"} for span in spans]}
    return "patch does not apply to the base file", None


kept_path = RAW / "kept.jsonl"
done = {}
progress = RAW / "examined.jsonl"
if progress.exists():
    for line in open(progress):
        row = json.loads(line)
        done[(row["repo"], row["number"])] = row
# The search returns more candidates than one hour of API calls can examine.
# The first EXAMINE candidates in hash order are examined, so no query is favored.
candidates.sort(key=lambda c: hashlib.sha256(f"{c['repo']}#{c['number']}".encode()).hexdigest())
candidates = candidates[:EXAMINE]
selected = {(c["repo"], c["number"]) for c in candidates}
done = {key: row for key, row in done.items() if key in selected}
todo = [c for c in candidates if (c["repo"], c["number"]) not in done]
with open(progress, "a") as log, concurrent.futures.ThreadPoolExecutor(options.workers) as pool:
    for candidate, (outcome, document) in zip(todo, pool.map(examine, todo)):
        row = {"repo": candidate["repo"], "number": candidate["number"], "outcome": outcome, "document": document}
        done[(candidate["repo"], candidate["number"])] = row
        log.write(json.dumps(row, ensure_ascii=False) + "\n")
        log.flush()
        if len(done) % 100 == 0:
            print("examined", len(done), "kept", sum(1 for r in done.values() if r["document"]), flush=True)
print("examined", len(done), collections.Counter(row["outcome"] for row in done.values()).most_common())
by_repo = {}
for row in sorted(done.values(), key=lambda r: hashlib.sha256(f"{r['repo']}#{r['number']}".encode()).hexdigest()):
    if row["document"] and row["repo"].lower() not in by_repo:
        by_repo[row["repo"].lower()] = row["document"]
write_jsonl(kept_path, list(by_repo.values()))
print("kept, one per repository:", len(by_repo))
