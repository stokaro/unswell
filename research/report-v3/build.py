#!/usr/bin/env python3
"""Build and check the data of the research report on the site.

usage: build.py [--write]

The report is docs/site/src/content/docs/research-report.mdx. Its figures are drawn by the site from
docs/site/src/data/research-report.json, and this program makes that file from the records of the repository:
figures.py names, for each plotted value, a record file and a pointer into it. Without --write the program builds
the file again and fails when the committed one differs.

It also checks the prose of the report:

- every number that claims.json lists is the number that its source holds, and the report says it;
- every number in the prose of the report is covered by a claim (years, issue numbers, and the like are let through);
- every external link of the report is in references.json with the date it was checked and how the project used
  the work, and every reference is cited;
- the chart colors of the site are told apart and readable in both themes;
- the frozen commit history of the timeline is the history that git holds.
"""
import json
import re
import subprocess
import sys
from pathlib import Path

HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[1]
OUT = "docs/site/src/data/research-report.json"
PAGE = "docs/site/src/content/docs/research-report.mdx"
STYLES = "docs/site/src/styles/unswell.css"
SCHEMA = "unswell-docs-research-report-v1"
USES = ("ran", "adapted", "compared", "corpus", "read")
MONTHS = ("Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec")
# The chart colors of the site: at least this far apart for a reader with a color-vision deficiency, and this
# contrast with the surfaces that a chart can stand on. A lighter color on a light surface needs the labels and the
# table that every figure of the report has.
CVD_LEAST = 8.0
CONTRAST_LEAST = {"light": 2.5, "dark": 3.0}


class Malformed(Exception):
    """The report, its data, or a record under it is not what the report says it is."""


def fail(message):
    raise Malformed(message)


# ---- reading the records ----

def reader(root):
    """A function that reads a JSON file of the repository once."""
    cache = {}

    def read(path):
        if path not in cache:
            file = root / path
            if not file.exists():
                fail(f"missing record: {path}")
            cache[path] = json.loads(file.read_text())
        return cache[path]
    return read


def resolve(doc, pointer, where=""):
    """The value at a JSON pointer."""
    value = doc
    for token in pointer.split("/")[1:]:
        key = token.replace("~1", "/").replace("~0", "~")
        if isinstance(value, list):
            if not key.isdigit() or int(key) >= len(value):
                fail(f"pointer {pointer} leaves {where}")
            value = value[int(key)]
        elif isinstance(value, dict) and key in value:
            value = value[key]
        else:
            fail(f"pointer {pointer} leaves {where}")
    return value


def number(read, spec):
    """A number from a ref, or the sum of a list of refs."""
    if isinstance(spec, list):
        return sum(number(read, part) for part in spec)
    path, pointer = spec["ref"]
    value = resolve(read(path), pointer, path)
    if isinstance(value, bool) or not isinstance(value, (int, float)):
        fail(f"{path}{pointer} is no number")
    return value


def cell(read, spec, unit):
    """A plotted value: the number for the geometry and the words for the label and the table."""
    if "share" in spec:
        count, of = number(read, spec["share"][0]), number(read, spec["share"][1])
        if of <= 0 or not 0 <= count <= of:
            fail(f"no share: {count} of {of}")
        return {"value": round(count / of, 4), "text": f"{count} of {of}", "detail": f"{100 * count / of:.1f}%"}
    value = number(read, spec)
    if unit == "auc":
        return {"value": value, "text": f"{value:.2f}"}
    return {"value": value, "text": f"{value:,}"}


def booleans(value):
    """Every boolean under a value: one for each registered bar."""
    if isinstance(value, bool):
        return [value]
    if isinstance(value, dict):
        return [flag for part in value.values() for flag in booleans(part)]
    return []


def record_rows(read, records):
    """How many of the bars of each record held."""
    rows = []
    for label, date, path, where in records:
        doc = read(path)
        if "bars" in where:
            flags = [bar["holds"] for bar in resolve(doc, where["bars"], path)]
        else:
            flags = [flag for pointer in where["registered"] for flag in booleans(resolve(doc, pointer, path))]
        if not flags or any(not isinstance(flag, bool) for flag in flags):
            fail(f"no bars in {path}")
        held = sum(flags)
        rows.append({"label": f"{label}, {date}", "values": {
            "held": {"value": held, "text": str(held)},
            "failed": {"value": len(flags) - held, "text": str(len(flags) - held)}}, "total": f"{held} of {len(flags)}"})
    return rows


def history_rows(read, path):
    """Commits by week, and the releases of each week."""
    history = read(path)
    rows = []
    for week in history["weeks"]:
        year, month, day = (int(part) for part in week["monday"].split("-"))
        tags = [release["tag"].removeprefix("v0.1.0-") for release in history["releases"] if release_week(release["date"]) == week["week"]]
        rows.append({"label": f"{MONTHS[month - 1]} {day}", "note": week["week"],
                     "values": {"commits": {"value": week["commits"], "text": str(week["commits"])}},
                     "events": tags})
    return rows


def release_week(date):
    import datetime
    year, week, _ = datetime.date.fromisoformat(date).isocalendar()
    return f"{year}-W{week:02d}"


def counted_rows(read, spec, unit):
    """One row for each key of a mapping of counts, largest first."""
    path, pointer = spec["counts"]
    counts = resolve(read(path), pointer, path)
    of = number(read, spec["of"])
    rows = []
    for name, count in sorted(counts.items(), key=lambda item: (-item[1], item[0])):
        label = f"`{name}`" if spec.get("code") else name
        rows.append({"label": label, "values": {spec["key"]: cell(read, {"share": [{"ref": [path, pointer + "/" + name]}, spec["of"]]}, unit)}})
    if not rows or of <= 0:
        fail(f"no counts at {path}{pointer}")
    return rows


def build_panel(read, spec, records):
    panel = {key: spec[key] for key in ("form", "unit", "series") if key in spec}
    for key in ("title", "min", "max", "reference", "stacked"):
        if key in spec:
            panel[key] = spec[key]
    if spec.get("records"):
        panel["rows"] = record_rows(read, records)
    elif "history" in spec:
        panel["rows"] = history_rows(read, spec["history"])
    elif "rows_from" in spec:
        panel["rows"] = counted_rows(read, spec["rows_from"], spec["unit"])
    else:
        panel["rows"] = [{"label": row["label"], "values": {key: cell(read, value, spec["unit"]) for key, value in row["values"].items()}}
                         for row in spec["rows"]]
    keys = {series["key"] for series in panel["series"]}
    for row in panel["rows"]:
        if set(row["values"]) != keys:
            fail(f"a row of {spec.get('title', 'a panel')} has other series than the panel names")
    return panel


def build(read, figures, records):
    """The file that the site draws the figures from."""
    built = {}
    for spec in figures:
        if spec["id"] in built:
            fail(f"two figures are called {spec['id']}")
        built[spec["id"]] = {"title": spec["title"], "sources": spec["sources"],
                             "panels": [build_panel(read, panel, records) for panel in spec["panels"]]}
    return {"schema": SCHEMA, "figures": built}


def encode(value):
    return json.dumps(value, indent=2, ensure_ascii=False) + "\n"


# ---- the prose of the report ----

def prose(page):
    """The words of the report that a reader sees, without code, addresses, and components."""
    text = re.sub(r"\A---\n.*?\n---\n", "", page, flags=re.S)
    text = re.sub(r"```.*?```", " ", text, flags=re.S)
    text = re.sub(r"^(import|export) .*$", " ", text, flags=re.M)
    text = re.sub(r"\{/\*.*?\*/\}", " ", text, flags=re.S)
    text = re.sub(r"`[^`\n]*`", " ", text)
    text = re.sub(r"\]\([^)\s]*\)", "]", text)
    text = re.sub(r"<[^>\n]*>", " ", text)
    return text


# Numbers that are no measurement: years, dates, versions, and the numbers of issues, decisions, figures, and sections.
LET_THROUGH = [
    r"\b20\d\d\b",
    r"\b(January|February|March|April|May|June|July|August|September|October|November|December) \d{1,2}(?: (?:to|and) \d{1,2})?\b",
    r"\bADR \d{4}\b", r"\bADRs? \d{4}(?:(?:, | and | to )\d{4})*", r"#\d+", r"\bissue \d+\b", r"\balpha\.\d+\b", r"\bv\d+\.\d+\.\d+\b",
    r"\b(Figure|Table|Section) \d+\b", r"^\s*\d+\. ", r"\[\d+\]", r"\bversion \d+\b",
]


def uncovered_numbers(page, claims):
    """The numbers of the prose that no claim covers."""
    # The list of references holds titles and years, and no measurement.
    text = prose(page).split("\n## References", 1)[0]
    for claim in sorted(claims, key=lambda claim: -len(claim["text"])):
        text = text.replace(claim["text"], " ")
    for pattern in LET_THROUGH:
        text = re.sub(pattern, " ", text, flags=re.M)
    return sorted(set(re.findall(r"\d[\d,]*(?:\.\d+)?%?", text)))


def formatted(value, form):
    """A number of a record in the words of the report."""
    if form == "int":
        return f"{value:,}"
    if form == "percent":
        return f"{100 * value:.1f}%"
    if form == "percent0":
        return f"{100 * value:.0f}%"
    if form in ("fixed2", "fixed3"):
        return f"{value:.{form[-1]}f}"
    if form == "of":
        return f"{value['count']} of {value['of']}"
    if form == "held":
        flags = booleans(value)
        return f"{sum(flags)} of {len(flags)}"
    if form == "length":
        return str(len(value))
    fail(f"unknown form of a claim: {form}")


def check_claims(root, read, claims, page):
    """Each claim is what its source holds, and the report says it."""
    seen = set()
    for claim in claims:
        name = claim.get("id")
        if not name or name in seen:
            fail(f"a claim without a name of its own: {name}")
        seen.add(name)
        if claim["text"] not in page:
            fail(f"claim {name}: the report does not say {claim['text']!r}")
        if "pointer" in claim:
            value = resolve(read(claim["source"]), claim["pointer"], claim["source"])
            said = formatted(value, claim["form"])
            if said not in claim["text"]:
                fail(f"claim {name}: {claim['source']} holds {said}, and the claim says {claim['text']!r}")
        elif "anchor" in claim:
            file = root / claim["source"]
            if not file.exists():
                fail(f"claim {name}: missing source {claim['source']}")
            flat = " ".join(file.read_text().split())
            if " ".join(claim["anchor"].split()) not in flat:
                fail(f"claim {name}: {claim['source']} does not hold the words {claim['anchor']!r}")
            for part in claim.get("numbers", [claim["text"]]):
                if part not in claim["anchor"]:
                    fail(f"claim {name}: the words of the source do not hold {part!r}")
        else:
            fail(f"claim {name} names no pointer and no words of its source")


def check_references(references, page):
    """Every external link of the report is a checked reference, and every reference is cited."""
    urls = {}
    for entry in references:
        for field in ("id", "title", "url", "checked", "use", "note"):
            if not entry.get(field):
                fail(f"reference {entry.get('id')}: no {field}")
        if entry["use"] not in USES:
            fail(f"reference {entry['id']}: how the project used it is one of {', '.join(USES)}")
        if not re.fullmatch(r"\d{4}-\d{2}-\d{2}", entry["checked"]):
            fail(f"reference {entry['id']}: the date it was checked")
        if entry["url"] in urls:
            fail(f"reference {entry['id']}: {entry['url']} is listed twice")
        if f"[{entry['title']}]({entry['url']})" not in page:
            fail(f"reference {entry['id']}: the report does not cite it by its title and its address")
        urls[entry["url"]] = entry["id"]
    linked = set(re.findall(r"\]\((https?://[^)\s]+)\)", page))
    ours = ("https://github.com/stokaro/", "https://unswell.dev", "https://play.unswell.dev")
    for url in sorted(linked):
        if not url.startswith(ours) and url not in urls:
            fail(f"the report links {url}, which references.json does not list")


# ---- the colors of the charts ----

PROTAN = ((0.152286, 1.052583, -0.204868), (0.114503, 0.786281, 0.099216), (-0.003882, -0.048116, 1.051998))
DEUTAN = ((0.367322, 0.860646, -0.227968), (0.280085, 0.672501, 0.047413), (-0.011820, 0.042940, 0.968881))


def linear(hex_color):
    channels = [int(hex_color[index:index + 2], 16) / 255 for index in (1, 3, 5)]
    return [c / 12.92 if c <= 0.04045 else ((c + 0.055) / 1.055) ** 2.4 for c in channels]


def contrast(first, second):
    """The WCAG contrast of two colors."""
    lights = sorted((0.2126 * r + 0.7152 * g + 0.0722 * b for r, g, b in (linear(first), linear(second))), reverse=True)
    return (lights[0] + 0.05) / (lights[1] + 0.05)


def oklab(rgb):
    r, g, b = rgb
    l = (0.4122214708 * r + 0.5363325363 * g + 0.0514459929 * b) ** (1 / 3)
    m = (0.2119034982 * r + 0.6806995451 * g + 0.1073969566 * b) ** (1 / 3)
    s = (0.0883024619 * r + 0.2817188376 * g + 0.6299787005 * b) ** (1 / 3)
    return (0.2104542553 * l + 0.7936177850 * m - 0.0040720468 * s, 1.9779984951 * l - 2.4285922050 * m + 0.4505937099 * s,
            0.0259040371 * l + 0.7827717662 * m - 0.8086757660 * s)


def apart(first, second, matrix):
    """The distance of two colors as a reader with one color-vision deficiency sees them."""
    seen = []
    for color in (first, second):
        rgb = linear(color)
        seen.append(oklab([max(0.0, sum(row[i] * rgb[i] for i in range(3))) for row in matrix]))
    return 100 * sum((a - b) ** 2 for a, b in zip(*seen)) ** 0.5


def theme_tokens(styles):
    """The custom properties of the light block and of the two dark blocks of the style sheet."""
    blocks = re.findall(r"(:root(?:\[data-theme='dark'\]|:not\(\[data-theme='light'\]\))?)\s*\{(.*?)\}", styles, flags=re.S)
    themes = {}
    for selector, body in blocks:
        tokens = dict(re.findall(r"(--u-[a-z0-9-]+):\s*(#[0-9a-fA-F]{6})\s*;", body))
        if "--u-page" not in tokens:
            continue
        name = "light" if selector == ":root" else "dark-system" if "not" in selector else "dark-chosen"
        themes.setdefault(name, tokens)
    return themes


def check_palette(styles):
    themes = theme_tokens(styles)
    if set(themes) != {"light", "dark-system", "dark-chosen"}:
        fail(f"the style sheet has these themes: {sorted(themes)}")
    if themes["dark-system"] != themes["dark-chosen"]:
        fail("the two dark blocks of the style sheet differ")
    for name in ("light", "dark-chosen"):
        tokens = themes[name]
        colors = [tokens.get(f"--u-chart-{index}") for index in (1, 2, 3)]
        if None in colors:
            fail(f"the {name} theme lacks a chart color")
        least = CONTRAST_LEAST["light" if name == "light" else "dark"]
        for color in colors:
            for surface in ("--u-page", "--u-card"):
                if contrast(color, tokens[surface]) < least:
                    fail(f"chart color {color} has too little contrast with {surface} in the {name} theme")
        for first, second in zip(colors, colors[1:]):
            if min(apart(first, second, PROTAN), apart(first, second, DEUTAN)) < CVD_LEAST:
                fail(f"chart colors {first} and {second} are too close for a reader with a color-vision deficiency")


# ---- the history of the timeline ----

def check_history(root, history):
    """The frozen counts are what git holds for the frozen commit."""
    def git(*arguments):
        done = subprocess.run(["git", "-C", str(root), *arguments], capture_output=True, text=True)
        return done.returncode, done.stdout.strip()

    code, _ = git("cat-file", "-e", history["head"] + "^{commit}")
    if code != 0:
        fail(f"the timeline is frozen at commit {history['head'][:8]}, which this clone does not hold")
    _, weeks = git("log", "--format=%ad", "--date=format:%G-W%V", history["head"])
    counts = {}
    for week in weeks.split():
        counts[week] = counts.get(week, 0) + 1
    if counts != {week["week"]: week["commits"] for week in history["weeks"]} or sum(counts.values()) != history["commits"]:
        fail("the commits by week of the timeline are not what git holds")
    for release in history["releases"]:
        code, date = git("for-each-ref", "--format=%(creatordate:short)", "refs/tags/" + release["tag"])
        if code == 0 and date and date != release["date"]:
            fail(f"release {release['tag']} is dated {date} in git and {release['date']} in the timeline")


def main(root=ROOT, write="--write" in sys.argv[1:]):
    sys.path.insert(0, str(HERE))
    import figures
    read = reader(root)
    built = encode(build(read, figures.FIGURES, figures.RECORDS))
    out = root / OUT
    if write:
        out.write_text(built)
    elif not out.exists() or out.read_text() != built:
        fail(f"{OUT} is not what the records give; run build.py --write and read the difference")
    page = (root / PAGE).read_text()
    claims = json.loads((HERE / "claims.json").read_text())
    check_claims(root, read, claims, page)
    left = uncovered_numbers(page, claims)
    if left:
        fail(f"numbers of the report that no claim covers: {', '.join(left)}")
    check_references(json.loads((HERE / "references.json").read_text()), page)
    check_palette((root / STYLES).read_text())
    check_history(root, read("research/report-v3/data/history.json"))
    print(f"research report: {len(figures.FIGURES)} figures, {len(claims)} claims, and its references are what the records give")


if __name__ == "__main__":
    try:
        main()
    except Malformed as error:
        sys.exit(f"build: {error}")
