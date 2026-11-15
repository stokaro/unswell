#!/usr/bin/env python3
"""Collect blog posts that people wrote before November 2022 from public repositories.

usage: fetch_human.py <out jsonl> [per source]

Each source is read at its last commit before 2022-11-01, so the text is the text of that day. From each source
the posts are tried in the order of a hash of their path, and the first ones with enough English prose are kept.
"""
import hashlib, json, re, subprocess, sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))
import prepare as lib

SOURCES = [
    ("ariga/atlas", "doc/website/blog"), ("ent/ent", "doc/website/blog"), ("supabase/supabase", "apps/www/_blog"),
    ("vitessio/website", "content/en/blog"), ("tikv/website", "content/blog"), ("prometheus/docs", "content/blog"),
    ("grpc/grpc.io", "content/en/blog"), ("golang/website", "_content/blog"), ("linkerd/website", "linkerd.io/content/blog"),
    ("kubernetes/website", "content/en/blog/_posts"), ("nats-io/nats-site", "content/blog"),
]
UNTIL = "2022-10-31T23:59:59Z"
MIN_WORDS = 500


def api(path, raw=False):
    args = ["gh", "api", path]
    if raw:
        args += ["-H", "Accept: application/vnd.github.raw"]
    done = subprocess.run(args, capture_output=True)
    if done.returncode != 0:
        return None
    return done.stdout.decode("utf-8", "replace") if raw else json.loads(done.stdout)


def posts(repo, directory, sha):
    """List the Markdown posts of a blog directory, one level of post folders included."""
    found = []
    for entry in api(f"repos/{repo}/contents/{directory}?ref={sha}") or []:
        name = entry["name"]
        if entry["type"] == "file" and re.search(r"\.(md|mdx|markdown)$", name) and not name.startswith("_"):
            found.append(entry["path"])
        elif entry["type"] == "dir":
            for inner in api(f"repos/{repo}/contents/{entry['path']}?ref={sha}") or []:
                if inner["type"] == "file" and re.fullmatch(r"index\.(md|mdx)", inner["name"]):
                    found.append(inner["path"])
    return found


def english(text):
    letters = [c for c in text if c.isalpha()]
    return bool(letters) and sum(c.isascii() for c in letters) / len(letters) > 0.97


out, per_source = Path(sys.argv[1]), int(sys.argv[2]) if len(sys.argv) > 2 else 12
rows = []
for repo, directory in SOURCES:
    sha = api(f"repos/{repo}/commits?until={UNTIL}&per_page=1")[0]["sha"]
    paths = sorted(posts(repo, directory, sha), key=lambda p: hashlib.sha256((repo + "/" + p).encode()).hexdigest())
    kept = 0
    for path in paths:
        if kept >= per_source:
            break
        raw = api(f"repos/{repo}/contents/{path}?ref={sha}", raw=True)
        if not raw:
            continue
        text = lib.clean(raw)
        if lib.words(text) < MIN_WORDS or not english(text):
            continue
        key = f"{repo}/{path}"
        rows.append({"id": hashlib.sha256(key.encode()).hexdigest()[:16], "repo": repo, "path": path, "commit": sha,
                     "title": lib.front_matter_title(raw), "words": lib.words(text), "split": lib.bucket(key),
                     "text": text})
        kept += 1
    print(f"{repo}: {len(paths)} posts listed, {kept} kept", flush=True)
out.write_text("".join(json.dumps(row, ensure_ascii=False) + "\n" for row in rows))
print(len(rows), "posts")
