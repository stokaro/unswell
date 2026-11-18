"""Tests of the recount of the record of the fourth form of a closing denial."""
import gzip
import json
import shutil
import sys
import tempfile
import unittest
from pathlib import Path

HERE = Path(__file__).resolve().parent
sys.path.insert(0, str(HERE))
import verify  # noqa: E402


def copy_of_record(directory):
    root = Path(directory) / "record"
    shutil.copytree(HERE, root, ignore=shutil.ignore_patterns("__pycache__", "scripts"))
    return root


def rewrite(root, change):
    """Apply change to the rows of the data file and write the file again."""
    path = root / "data" / "pages.jsonl.gz"
    with gzip.open(path, "rt", encoding="utf-8") as handle:
        rows = [json.loads(line) for line in handle if line.strip()]
    rows = change(rows) or rows
    with open(path, "wb") as raw, gzip.GzipFile(fileobj=raw, mode="wb", mtime=0, filename="") as out:
        out.write("".join(json.dumps(row, sort_keys=True) + "\n" for row in rows).encode())


def pages_of(rows, corpus, part="measured", **fields):
    return [row for row in rows if row["corpus"] == corpus and row["part"] == part
            and all(row[key] == value for key, value in fields.items())]


def report(row, build, places=3):
    """Make one build report a page with the given number of closing denials."""
    row[build] = row["counted_" + build] = places


class RecordTest(unittest.TestCase):
    def test_the_committed_record_is_what_its_data_give(self):
        result = verify.main(HERE, write=False)
        self.assertEqual(result["bars_total"], 7)
        self.assertEqual(result["bars_held"], 7)
        self.assertEqual(result["not_kept"], 0)

    def test_the_bars_are_read_on_the_measured_parts(self):
        result = verify.results(HERE)
        by_name = {bar["bar"]: bar for bar in result["bars"]}
        self.assertEqual(by_name["P. the posts"]["value"], result["sets"]["set-s/measured"]["four"])
        self.assertEqual(by_name["H. quiet on human-docs"]["value"], result["sets"]["human-docs/measured"]["four"])
        self.assertEqual(result["people_measured"]["pages"], 77 + 205 + 61)
        self.assertEqual(by_name["H4. pages of people that only the fourth form reports"]["value"]["of"], 343)
        self.assertEqual(result["measured"]["pages"], 30 + 343 + 120)

    def test_a_changed_number_is_not_accepted(self):
        with tempfile.TemporaryDirectory() as directory:
            root = copy_of_record(directory)
            stored = json.loads((root / "results.json").read_text())
            stored["bars_held"] -= 1
            (root / "results.json").write_text(verify.encode(stored))
            with self.assertRaisesRegex(verify.Malformed, "results.json is not what the data give"):
                verify.main(root, write=False)

    def test_a_changed_data_file_is_not_the_one_the_record_names(self):
        with tempfile.TemporaryDirectory() as directory:
            root = copy_of_record(directory)
            rewrite(root, lambda rows: report(pages_of(rows, "public-today", "whole", four=0)[0], "four"))
            with self.assertRaises(verify.Malformed):
                verify.main(root, write=False)
            # After the record is written again the numbers follow the data, and the file list names the new file.
            before = (HERE / "files.sha256").read_text()
            verify.main(root, write=True)
            self.assertNotEqual((root / "files.sha256").read_text(), before)
            verify.main(root, write=False)


class BarsTest(unittest.TestCase):
    def setUp(self):
        self.directory = tempfile.TemporaryDirectory()
        self.addCleanup(self.directory.cleanup)
        self.root = copy_of_record(self.directory.name)

    def bars_after(self, change):
        rewrite(self.root, change)
        return {bar["bar"]: bar["holds"] for bar in verify.results(self.root)["bars"]}

    def test_a_rule_that_reports_too_many_pages_of_people_fails_its_bar(self):
        def loud(rows):
            for row in pages_of(rows, "human-blog", four=0)[:4]:
                report(row, "four")

        held = self.bars_after(loud)
        # Five of 77 blog posts are over one page in twenty, and four new pages are over the three that the form may add.
        self.assertFalse(held["H. quiet on human-blog"])
        self.assertFalse(held["H4. pages of people that only the fourth form reports"])
        self.assertTrue(held["H. quiet on human-styled"])

    def test_a_page_that_both_builds_report_is_not_added_by_the_form(self):
        def both(rows):
            for row in pages_of(rows, "human-styled", four=0)[:1]:
                report(row, "three")
                report(row, "four")
                row["kept"] = 3

        self.assertTrue(self.bars_after(both)["H4. pages of people that only the fourth form reports"])

    def test_too_few_posts_fail_the_bar_of_the_posts(self):
        def quiet(rows):
            for row in pages_of(rows, "set-s"):
                if row["four"] > 0 and row["three"] == 0:
                    row["four"] = 0
            for row in [row for row in pages_of(rows, "set-s") if row["four"] > 0][:4]:
                row["three"] = row["four"] = 0

        held = self.bars_after(quiet)
        self.assertFalse(held["P. the posts"])
        # A page that neither build reports has lost nothing.
        self.assertTrue(held["K. pages that lose their report"])

    def test_a_form_that_adds_nothing_fails_the_bar_of_the_gain(self):
        def same(rows):
            for row in pages_of(rows, "set-s"):
                row["counted_four"] = row["kept"] = row["counted_three"]
                row["four"] = row["three"]

        held = self.bars_after(same)
        self.assertFalse(held["G. closing denials that the fourth form adds on the posts"])
        self.assertTrue(held["K. pages that lose their report"])

    def test_a_page_that_loses_its_report_fails_the_bar(self):
        def lose(rows):
            row = next(row for row in rows if row["part"] == "measured" and row["three"] > 0)
            row["four"] = 0

        self.assertFalse(self.bars_after(lose)["K. pages that lose their report"])

    def test_a_page_of_a_fitting_part_moves_no_bar(self):
        before = {bar["bar"]: bar["value"] for bar in verify.results(self.root)["bars"]}

        def loud(rows):
            for row in pages_of(rows, "human-blog", "fitting", four=0)[:10]:
                report(row, "four")

        rewrite(self.root, loud)
        self.assertEqual({bar["bar"]: bar["value"] for bar in verify.results(self.root)["bars"]}, before)


class RowsTest(unittest.TestCase):
    def setUp(self):
        self.directory = tempfile.TemporaryDirectory()
        self.addCleanup(self.directory.cleanup)
        self.root = copy_of_record(self.directory.name)

    def refused(self, change, message):
        rewrite(self.root, change)
        with self.assertRaisesRegex(verify.Malformed, message):
            verify.results(self.root)

    def test_a_missing_page_is_refused(self):
        self.refused(lambda rows: rows[1:], "pages, and the record says")

    def test_a_page_listed_twice_is_refused(self):
        self.refused(lambda rows: rows + [dict(rows[0])], "listed twice")

    def test_a_page_of_an_unknown_set_is_refused(self):
        def move(rows):
            rows[0]["corpus"] = "human-forum"

        self.refused(move, "unknown set or part")

    def test_a_post_without_its_number_is_refused(self):
        def rename(rows):
            pages_of(rows, "set-s")[0]["id"] = "0123456789ab"

        self.refused(rename, "the id of a page is not of its set")

    def test_a_report_of_two_closing_denials_is_refused(self):
        def two(rows):
            report(pages_of(rows, "human-blog", four=0)[0], "four", places=2)

        self.refused(two, "2 listed, 2 counted")

    def test_a_report_that_lists_fewer_than_it_counts_is_refused(self):
        def fewer(rows):
            row = pages_of(rows, "human-blog", four=0)[0]
            row["four"], row["counted_four"] = 3, 5

        self.refused(fewer, "3 listed, 5 counted")

    def test_more_kept_closing_denials_than_counted_is_refused(self):
        def more(rows):
            rows[0]["kept"] = rows[0]["counted_three"] + 1

        self.refused(more, "that both builds count")

    def test_a_row_with_text_is_refused(self):
        def text(rows):
            rows[0]["sentence"] = "It does not rewrite a file."

        self.refused(text, "other fields")

    def test_a_missing_data_file_is_refused(self):
        (self.root / "data" / "pages.jsonl.gz").unlink()
        with self.assertRaisesRegex(verify.Malformed, "missing data file"):
            verify.results(self.root)


if __name__ == "__main__":
    unittest.main()
