#!/usr/bin/env python3
"""Select real documentation pages for final test 2 (`real-test-2`) and build a corpus manifest.

usage: acquire_real3.py <unswell repo> <workspace>

Selection is fixed before any model exists. For four languages and two creation
bands, repositories are taken in order of stars. A repository is skipped when
any earlier Unswell research material names it. Each repository contributes its
README and at most two Markdown pages under docs/, chosen by path order.
"""
import base64, hashlib, json, re, subprocess, sys
from pathlib import Path

REPO, W = Path(sys.argv[1]), Path(sys.argv[2])
LANGUAGES = ["Go", "Rust", "Python", "TypeScript"]
BANDS = {"recent": "created:>=2025-07-01", "older": "created:<2021-01-01 pushed:>=2026-01-01"}
PER_CELL = 40
MIN_BYTES, MAX_BYTES, MIN_WORDS = 3000, 40000, 300
ROOT = W / "real3"
TAKEN = set()
for _log in W.glob("*-selection-log.json"):
    TAKEN |= {e["repository"].lower() for e in json.loads(_log.read_text())}


def gh(path, raw=False):
    args = ["gh", "api", "-H", "Accept: application/vnd.github.raw" if raw else "Accept: application/vnd.github+json", path]
    result = subprocess.run(args, capture_output=True)
    if result.returncode != 0:
        raise RuntimeError(result.stderr.decode()[:200])
    return result.stdout if raw else json.loads(result.stdout)


def excluded_repositories():
    names = set()
    text = subprocess.run(["git", "-C", str(REPO), "grep", "-ohIE",
                           r"github\.com/[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+|\"(repository|repo)\": *\"[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+\"",
                           "origin/main", "--", "research", "docs", "e2e"], capture_output=True).stdout.decode()
    for match in re.findall(r"([A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+)\"?$", text, re.M):
        names.add(match.lower().removesuffix(".git").rstrip("."))
    extra = json.loads((REPO / "artifacts/planning/origin-new-corpus-exclusions-2026-11-10.json").read_text())
    names.update(n.lower() for n in extra["excluded_repositories"])
    return names


def prose_words(text):
    text = re.sub(r"```.*?```", " ", text, flags=re.S)
    text = re.sub(r"<[^>]+>|\[([^\]]*)\]\([^)]*\)", r"\1", text)
    return len(re.findall(r"[A-Za-z]{2,}", text))


def usable(data):
    if not MIN_BYTES <= len(data) <= MAX_BYTES:
        return False
    try:
        text = data.decode("utf-8")
    except UnicodeDecodeError:
        return False
    ascii_share = sum(ch < 128 for ch in data) / len(data)
    return ascii_share >= 0.97 and prose_words(text) >= MIN_WORDS


excluded = excluded_repositories()
print("repositories named in earlier material:", len(excluded), flush=True)
sources, log = [], []
for language in LANGUAGES:
    for band, clause in BANDS.items():
        query = f"language:{language} {clause} stars:>500 archived:false fork:false".replace(" ", "+")
        found = []
        for page in (1, 2, 3, 4, 5, 6, 7, 8, 9, 10):
            found += gh(f"search/repositories?q={query}&sort=stars&order=desc&per_page=100&page={page}")["items"]
        taken = 0
        for item in found:
            if taken >= PER_CELL:
                break
            name = item["full_name"]
            if name.lower() in TAKEN:
                continue
            if name.lower() in excluded or item.get("private"):
                log.append({"repository": name, "cell": [language, band], "decision": "excluded: named earlier"})
                continue
            try:
                commit = gh(f"repos/{name}/commits/{item['default_branch']}")["sha"]
                tree = gh(f"repos/{name}/git/trees/{commit}?recursive=1")["tree"]
            except RuntimeError as error:
                log.append({"repository": name, "decision": f"error: {error}"})
                continue
            paths = [t["path"] for t in tree if t["type"] == "blob"]
            readme = [p for p in paths if p.lower() == "readme.md"]
            docs = sorted(p for p in paths if re.match(r"^(docs?|documentation)/[^/]+(/[^/]+)?\.md$", p, re.I)
                          and not re.search(r"(^|/)(zh|ja|ko|ru|fr|de|es|pt|it|tr|i18n|translations?|changelog|license|code_of_conduct)([-_./]|$)", p, re.I))
            chosen = []
            for path in readme + docs:
                if len(chosen) >= 3 or (path not in readme and len([c for c in chosen if c[0] not in readme]) >= 2):
                    break
                try:
                    data = gh(f"repos/{name}/contents/{path}?ref={commit}", raw=True)
                except RuntimeError:
                    continue
                if usable(data):
                    chosen.append((path, data))
            if not chosen:
                log.append({"repository": name, "cell": [language, band], "decision": "no usable page"})
                continue
            taken += 1
            folder = name.replace("/", "__")
            for path, data in chosen:
                target = ROOT / folder / path
                target.parent.mkdir(parents=True, exist_ok=True)
                target.write_bytes(data)
                sources.append({
                    "id": "", "path": f"{folder}/{path}", "sha256": hashlib.sha256(data).hexdigest(),
                    "bytes": len(data), "format": "markdown", "prose_language": "en", "repository": name,
                    "document": f"{name}/{path}", "authors": [], "templates": [], "related": [],
                    "generation_tasks": [],
                    "reference": f"https://github.com/{name}/blob/{commit}/{path}",
                    "topic": (item.get("description") or "software documentation")[:120],
                    "purpose": "Technical documentation page for the final test",
                    "author_language": "", "author_language_basis": "",
                    "role": "readme" if path.lower() == "readme.md" else "documentation", "role_regions": [],
                    "origin": {"label": "unknown", "scope": "repository",
                               "evidence": f"Public repository snapshot at commit {commit}; authorship not recorded.",
                               "generation_record": ""}})
            log.append({"repository": name, "cell": [language, band], "commit": commit, "stars": item["stargazers_count"],
                        "created": item["created_at"][:10], "decision": "selected", "pages": [p for p, _ in chosen]})
            print(f"{language:10s} {band:6s} {name} pages={len(chosen)}", flush=True)
for number, source in enumerate(sources, 1):
    source["id"] = f"d{number:06d}"
template = json.loads((REPO / "research/annotation/corpus/testdata/ptah-manifest.json").read_text())
manifest = {key: template[key] for key in ("version", "weights", "extraction_policy", "unit_kinds")}
manifest.update({"id": "real-pages-final-test-v3", "seed": "real-pages-final-test-v3", "sources": sources})
(W / "real3-manifest.json").write_text(json.dumps(manifest, indent=1, ensure_ascii=False))
(W / "real3-selection-log.json").write_text(json.dumps(log, indent=1, ensure_ascii=False))
selected = [entry for entry in log if entry["decision"] == "selected"]
print(f"selected repositories={len(selected)} pages={len(sources)} skipped={len(log) - len(selected)}")
