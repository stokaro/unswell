"""The figures of the research report, as pointers into the records.

This file holds no measured number. Each value names a record file and a JSON pointer into it, and build.py reads the
number from there. A label, a unit, and the limit of a scale are all that is written here.
"""

REVIEWS = "research/reviews/"
HUMAN = REVIEWS + "2026-11-12-human-reference/results.json"
STYLE = REVIEWS + "2026-11-15-style-guide-text/results.json"
PILOTS = REVIEWS + "2026-11-17-revision-pilots/results.json"
HABITS = REVIEWS + "2026-11-17-page-habits/results.json"
TWO = REVIEWS + "2026-11-17-two-results/results.json"
LATE = REVIEWS + "2026-11-18-late-denial/results.json"
HISTORY = "research/report-v3/data/history.json"

GPT = "gpt-6-astra"
QWEN = "private-host/Qwen3.6-35B-A3B"
QWEN27 = "private-host/Qwen3.8-27B"

# The records whose bars were written on an issue before the run, in the order of their dates. "registered" names the
# places in the file that hold one boolean for each bar; "bars" is a list of bars with a "holds" field.
RECORDS = [
    ("Human reference, index and fix", "Nov 12", HUMAN, {"registered": ["/fix/registered", "/fix_fresh/registered", "/fix_third/registered",
                                                                         "/fix_fourth/registered", "/page_index/registered"]}),
    ("Text under a style guide", "Nov 15", STYLE, {"registered": ["/registered"]}),
    ("A revision, checked blind", "Nov 15", REVIEWS + "2026-11-15-revision-check/results.json", {"registered": ["/registered"]}),
    ("Sentences of a revision", "Nov 15", REVIEWS + "2026-11-15-sentence-check/results.json", {"registered": ["/registered"]}),
    ("Every changed sentence", "Nov 16", REVIEWS + "2026-11-16-changed-sentences/results.json", {"registered": ["/registered"]}),
    ("Words left out", "Nov 16", REVIEWS + "2026-11-16-left-out-words/results.json", {"registered": ["/registered"]}),
    ("Habits of a page", "Nov 17", HABITS, {"bars": "/bars"}),
    ("Two results of the check", "Nov 17", TWO, {"bars": "/bars"}),
    ("A late denial", "Nov 18", LATE, {"bars": "/bars"}),
]


def ref(path, pointer):
    return {"ref": [path, pointer]}


def share(count, of):
    """A count of a whole; each side is a ref or a list of refs that are added."""
    return {"share": [count, of]}


def escape(key):
    """A key as one token of a JSON pointer."""
    return key.replace("~", "~0").replace("/", "~1")


def separation(label, key):
    return {"label": label, "values": {"before": ref(STYLE, "/pages/measured/separation_earlier/" + escape(key)),
                                       "after": ref(STYLE, "/pages/measured/separation/" + escape(key))}}


def picked(label, key):
    where = "/together" if key is None else "/alone/" + escape(key)
    of = ref(PILOTS, "/arms/closing-sentences/revised")
    return {"label": label, "values": {"before": share(ref(PILOTS, "/arms/closing-sentences/picked_as_it_is" + where), of),
                                       "after": share(ref(PILOTS, "/arms/closing-sentences/picked_revised" + where), of)}}


def copies(judge, kind):
    base = "/copies/" + escape(judge)
    if kind == "harmless":
        return share(ref(TWO, base + "/harmless/count"), ref(TWO, base + "/harmless/of"))
    pages = ("set-s", "fresh")
    return share([ref(TWO, f"{base}/damaged/{name}/nothing_lost") for name in pages], [ref(TWO, f"{base}/damaged/{name}/copies") for name in pages])


def control(index, field):
    return share(ref(TWO, f"/controls/{index}/{field}"), ref(TWO, f"/controls/{index}/revisions"))


def habit(label, key):
    late = "/fresh_together/four" if key is None else "/sets/" + escape(key + "/measured") + "/four"
    register = "/fresh_together/register" if key is None else "/sets/" + escape(key + "/measured") + "/register"
    return {"label": label, "values": {"denial": share(ref(LATE, late + "/count"), ref(LATE, late + "/of")),
                                       "register": share(ref(HABITS, register + "/count"), ref(HABITS, register + "/of"))}}


FIGURES = [
    {
        "id": "timeline",
        "title": "Commits on the main branch by week, with the releases of each week",
        "sources": [HISTORY],
        "panels": [{"form": "columns", "unit": "count", "history": HISTORY,
                    "series": [{"key": "commits", "label": "Commits"}]}],
    },
    {
        "id": "bars",
        "title": "Bars written before a run, by record: how many held",
        "sources": [record[2] for record in RECORDS],
        "panels": [{"form": "bars", "stacked": True, "unit": "count", "records": True,
                    "series": [{"key": "held", "label": "Held"}, {"key": "failed", "label": "Did not hold", "hatched": True}]}],
    },
    {
        "id": "defaults",
        "title": "Rules on documentation pages that people wrote: pages with a finding",
        "sources": [HUMAN],
        "panels": [{"form": "bars", "unit": "share", "max": 0.3, "reference": {"value": 0.05, "label": "one page in twenty"},
                    "series": [{"key": "pages", "label": "Pages with a finding"}],
                    "rows_from": {"counts": [HUMAN, "/default_load/rules"], "of": ref(HUMAN, "/default_load/pages"), "key": "pages", "code": True}}],
    },
    {
        "id": "separation",
        "title": "How well the index tells a model's text from blog posts of people",
        "sources": [STYLE],
        "panels": [{"form": "dumbbell", "unit": "auc", "min": 0.0, "max": 1.0, "reference": {"value": 0.5, "label": "chance"},
                    "series": [{"key": "before", "label": "Marks alone"}, {"key": "after", "label": "With the rhythm reading"}],
                    "rows": [separation("Posts of the style-guide set", "set-s"),
                             separation("Guided texts, first writer", "guided, " + GPT),
                             separation("Guided texts, second writer", "guided, " + QWEN),
                             separation("Plain texts, first writer", "plain, " + GPT),
                             separation("Plain texts, second writer", "plain, " + QWEN)]}],
    },
    {
        "id": "revision",
        "title": "A revision by the report: what the index reads and what blind judges pick",
        "sources": [PILOTS],
        "panels": [
            {"form": "bars", "unit": "index", "max": 100, "title": "Median index of the posts",
             "series": [{"key": "before", "label": "As published"}, {"key": "after", "label": "After the revision"}],
             "rows": [{"label": "Median index", "values": {"before": ref(PILOTS, "/arms/closing-sentences/before/median_index"),
                                                           "after": ref(PILOTS, "/arms/closing-sentences/after/median_index")}}]},
            {"form": "bars", "unit": "share", "max": 1.0, "title": "Posts that a judge picks as the model's text",
             "series": [{"key": "before", "label": "As published"}, {"key": "after", "label": "After the revision"}],
             "rows": [picked("Three judges together", None), picked("`gpt-6-astra`", GPT), picked("`Qwen3.8-27B`", QWEN27),
                      picked("`Qwen3.6-35B-A3B`", QWEN)]},
        ],
    },
    {
        "id": "compare",
        "title": "The check of a revision: what passes as nothing lost",
        "sources": [TWO],
        "panels": [
            {"form": "bars", "unit": "share", "max": 1.0, "title": "Copies of a page with one sentence changed",
             "series": [{"key": "gpt", "label": "Judge `gpt-6-astra`"}, {"key": "qwen", "label": "Judge `Qwen3.6-35B-A3B`"}],
             "rows": [{"label": "Harmless copies", "values": {"gpt": copies(GPT, "harmless"), "qwen": copies(QWEN, "harmless")}},
                      {"label": "Damaged copies", "values": {"gpt": copies(GPT, "damaged"), "qwen": copies(QWEN, "damaged")}}]},
            {"form": "bars", "unit": "share", "max": 1.0, "title": "Revisions by a model, by the model that revised",
             "series": [{"key": "gpt", "label": "Judge `gpt-6-astra`"}, {"key": "qwen", "label": "Judge `Qwen3.6-35B-A3B`"}],
             "rows": [{"label": "Revised by `Qwen3.6-35B-A3B`", "values": {"gpt": control(0, "other_family"), "qwen": control(0, "own_family")}},
                      {"label": "Revised by `gpt-6-astra`", "values": {"gpt": control(1, "own_family"), "qwen": control(1, "other_family")}}]},
        ],
    },
    {
        "id": "habits",
        "title": "Pages that each rule for a habit of a page reports, by who wrote them",
        "sources": [LATE, HABITS],
        "panels": [{"form": "bars", "unit": "share", "max": 1.0, "reference": {"value": 0.05, "label": "limit on pages of people"},
                    "series": [{"key": "denial", "label": "Paragraphs end on a denial"}, {"key": "register", "label": "One formal register"}],
                    "rows": [habit("Posts of the style-guide set", "set-s"), habit("Texts of two models under a style guide", None),
                             habit("Blog posts of people", "human-blog"), habit("Articles of people", "human-styled"),
                             habit("Documentation pages of people", "human-docs")]}],
    },
]
