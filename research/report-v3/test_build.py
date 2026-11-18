"""Tests of the program that builds and checks the data of the research report."""
import copy
import json
import sys
import unittest
from pathlib import Path

HERE = Path(__file__).resolve().parent
sys.path.insert(0, str(HERE))
import build  # noqa: E402
import figures  # noqa: E402

TWO = "research/reviews/2026-11-17-two-results/results.json"
PAGE = "A model flagged 4.0% of 120 pages, as [the record](https://github.com/stokaro/unswell/blob/main/x.md) says in 2026.\n"
REFERENCE = {"id": "work", "title": "A work", "url": "https://example.org/work", "checked": "2026-11-18", "use": "read", "note": "Read."}
STYLES = """
:root {
	--u-page: #f3f4f3;
	--u-card: #ffffff;
	--u-chart-1: #2a78d6;
	--u-chart-2: #eb6834;
	--u-chart-3: #1baf7a;
}
@media (prefers-color-scheme: dark) {
	:root:not([data-theme='light']) {
		--u-page: #14181a;
		--u-card: #1b1f22;
		--u-chart-1: #3987e5;
		--u-chart-2: #d95926;
		--u-chart-3: #199e70;
	}
}
:root[data-theme='dark'] {
	--u-page: #14181a;
	--u-card: #1b1f22;
	--u-chart-1: #3987e5;
	--u-chart-2: #d95926;
	--u-chart-3: #199e70;
}
"""


def changed_reader(path, change):
    """The reader of the repository, with one record changed."""
    read = build.reader(build.ROOT)

    def changed(name):
        doc = read(name)
        if name != path:
            return doc
        doc = copy.deepcopy(doc)
        change(doc)
        return doc
    return changed


class FiguresTest(unittest.TestCase):
    def test_the_committed_data_are_what_the_records_give(self):
        build.main(build.ROOT, write=False)

    def test_a_changed_number_of_a_record_changes_the_figure(self):
        committed = json.loads((build.ROOT / build.OUT).read_text())

        def lose(doc):
            doc["copies"]["gpt-6-astra"]["harmless"]["count"] -= 1

        built = build.build(changed_reader(TWO, lose), figures.FIGURES, figures.RECORDS)
        self.assertNotEqual(built, committed)
        row = built["figures"]["compare"]["panels"][0]["rows"][0]["values"]["gpt"]
        self.assertEqual(row["text"], "154 of 155")

    def test_a_pointer_that_leaves_its_record_is_refused(self):
        def drop(doc):
            del doc["controls"][1]["own_family"]

        with self.assertRaisesRegex(build.Malformed, "leaves research/reviews/2026-11-17-two-results/results.json"):
            build.build(changed_reader(TWO, drop), figures.FIGURES, figures.RECORDS)

    def test_a_count_over_its_whole_is_no_share(self):
        def more(doc):
            doc["controls"][0]["own_family"] = 58

        with self.assertRaisesRegex(build.Malformed, "no share: 58 of 57"):
            build.build(changed_reader(TWO, more), figures.FIGURES, figures.RECORDS)

    def test_the_bars_of_a_record_are_counted_from_the_record(self):
        def miss(doc):
            doc["bars"][0]["holds"] = False

        built = build.build(changed_reader(TWO, miss), figures.FIGURES, figures.RECORDS)
        row = next(row for row in built["figures"]["bars"]["panels"][0]["rows"] if row["label"].startswith("Two results"))
        self.assertEqual(row["total"], "13 of 14")
        self.assertEqual(row["values"]["failed"]["value"], 1)

    def test_two_figures_cannot_share_a_name(self):
        with self.assertRaisesRegex(build.Malformed, "two figures are called timeline"):
            build.build(build.reader(build.ROOT), figures.FIGURES + [figures.FIGURES[0]], figures.RECORDS)


class ClaimsTest(unittest.TestCase):
    def setUp(self):
        self.read = build.reader(build.ROOT)

    def check(self, claim, page=PAGE):
        build.check_claims(build.ROOT, self.read, [claim], page)

    def test_a_claim_that_its_record_holds_passes(self):
        self.check({"id": "kept", "text": "120 pages", "source": "research/reviews/2026-11-18-late-denial/results.json",
                    "pointer": "/fresh_together/pages", "form": "int"})

    def test_a_claim_with_another_number_than_its_record_is_refused(self):
        with self.assertRaisesRegex(build.Malformed, "holds 120, and the claim says '4.0% of'"):
            self.check({"id": "kept", "text": "4.0% of", "source": "research/reviews/2026-11-18-late-denial/results.json",
                        "pointer": "/fresh_together/pages", "form": "int"})

    def test_a_claim_that_the_report_does_not_say_is_refused(self):
        with self.assertRaisesRegex(build.Malformed, "the report does not say '121 pages'"):
            self.check({"id": "kept", "text": "121 pages", "source": TWO, "pointer": "/bars_total", "form": "int"})

    def test_a_claim_needs_the_words_of_its_source(self):
        claim = {"id": "said", "text": "4.0%", "source": "docs/adr/0052-claims-follow-human-references.md",
                 "anchor": "the default rules flag 4.0% of the reworded sentences"}
        self.check(claim)
        with self.assertRaisesRegex(build.Malformed, "does not hold the words"):
            self.check({**claim, "anchor": "the default rules flag 4.0% of every sentence"})
        with self.assertRaisesRegex(build.Malformed, "the words of the source do not hold '120 pages'"):
            self.check({**claim, "text": "120 pages"})

    def test_a_claim_without_a_source_of_either_kind_is_refused(self):
        with self.assertRaisesRegex(build.Malformed, "names no pointer and no words"):
            self.check({"id": "bare", "text": "120 pages", "source": TWO})

    def test_two_claims_cannot_share_a_name(self):
        claim = {"id": "kept", "text": "120 pages", "source": "research/reviews/2026-11-18-late-denial/results.json",
                 "pointer": "/fresh_together/pages", "form": "int"}
        with self.assertRaisesRegex(build.Malformed, "without a name of its own"):
            build.check_claims(build.ROOT, self.read, [claim, claim], PAGE)

    def test_a_number_without_a_claim_is_found(self):
        self.assertEqual(build.uncovered_numbers(PAGE, []), ["120", "4.0%"])
        self.assertEqual(build.uncovered_numbers(PAGE, [{"text": "4.0% of 120 pages"}]), [])
        # A year, a decision, a release, and a number inside code or an address are no measurements.
        page = "ADR 0058 of November 13 to 15, 2026 turned `rule-9` off in alpha.10; see [issue 603](https://example.org/7)."
        self.assertEqual(build.uncovered_numbers(page, []), [])


class ReferencesTest(unittest.TestCase):
    PAGE = "See [A work](https://example.org/work) and [our record](https://github.com/stokaro/unswell/blob/main/x.md)."

    def test_a_cited_and_checked_reference_passes(self):
        build.check_references([REFERENCE], self.PAGE)

    def test_a_link_that_is_no_reference_is_refused(self):
        with self.assertRaisesRegex(build.Malformed, "links https://example.org/other"):
            build.check_references([REFERENCE], self.PAGE + " And [another](https://example.org/other).")

    def test_a_reference_that_is_not_cited_is_refused(self):
        with self.assertRaisesRegex(build.Malformed, "does not cite it by its title and its address"):
            build.check_references([REFERENCE], "No link here.")
        # A link under other words than the title of the work is no citation of it.
        with self.assertRaisesRegex(build.Malformed, "does not cite it by its title and its address"):
            build.check_references([REFERENCE], "See [this](https://example.org/work).")

    def test_a_reference_needs_its_check_and_its_use(self):
        with self.assertRaisesRegex(build.Malformed, "no checked"):
            build.check_references([{**REFERENCE, "checked": ""}], self.PAGE)
        with self.assertRaisesRegex(build.Malformed, "the date it was checked"):
            build.check_references([{**REFERENCE, "checked": "yesterday"}], self.PAGE)
        with self.assertRaisesRegex(build.Malformed, "how the project used it"):
            build.check_references([{**REFERENCE, "use": "liked"}], self.PAGE)

    def test_a_reference_is_listed_once(self):
        with self.assertRaisesRegex(build.Malformed, "is listed twice"):
            build.check_references([REFERENCE, {**REFERENCE, "id": "again"}], self.PAGE)


class PaletteTest(unittest.TestCase):
    def test_the_palette_of_the_site_passes(self):
        build.check_palette(STYLES)
        build.check_palette((build.ROOT / build.STYLES).read_text())

    def test_dark_blocks_that_differ_are_refused(self):
        with self.assertRaisesRegex(build.Malformed, "the two dark blocks of the style sheet differ"):
            build.check_palette(STYLES.replace("--u-chart-3: #199e70;\n\t}\n}", "--u-chart-3: #199e71;\n\t}\n}"))

    def test_a_color_too_close_to_the_surface_is_refused(self):
        with self.assertRaisesRegex(build.Malformed, "too little contrast"):
            build.check_palette(STYLES.replace("--u-chart-1: #2a78d6;", "--u-chart-1: #e8eef6;"))

    def test_two_colors_that_a_reader_cannot_tell_apart_are_refused(self):
        with self.assertRaisesRegex(build.Malformed, "too close for a reader with a color-vision deficiency"):
            build.check_palette(STYLES.replace("--u-chart-2: #eb6834;", "--u-chart-2: #2c7ad4;"))

    def test_a_missing_chart_color_is_refused(self):
        with self.assertRaisesRegex(build.Malformed, "lacks a chart color"):
            build.check_palette(STYLES.replace("\t--u-chart-3: #1baf7a;\n", ""))


class HistoryTest(unittest.TestCase):
    def setUp(self):
        self.history = json.loads((HERE / "data" / "history.json").read_text())

    def test_the_frozen_history_is_what_git_holds(self):
        build.check_history(build.ROOT, self.history)

    def test_a_changed_count_is_refused(self):
        self.history["weeks"][0]["commits"] += 1
        with self.assertRaisesRegex(build.Malformed, "not what git holds"):
            build.check_history(build.ROOT, self.history)

    def test_a_commit_that_the_clone_lacks_is_refused(self):
        self.history["head"] = "0" * 40
        with self.assertRaisesRegex(build.Malformed, "which this clone does not hold"):
            build.check_history(build.ROOT, self.history)

    def test_a_release_on_another_day_is_refused(self):
        self.history["releases"][0]["date"] = "2026-09-08"
        with self.assertRaisesRegex(build.Malformed, "is dated 2026-09-07 in git"):
            build.check_history(build.ROOT, self.history)


if __name__ == "__main__":
    unittest.main()
