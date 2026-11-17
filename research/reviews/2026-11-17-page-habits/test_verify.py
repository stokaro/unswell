"""Tests of the recount of the record of two habits of a page."""
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


def rewrite(root, name, change):
    """Apply change to the rows of one data file and write the file again."""
    path = root / "data" / (name + ".jsonl.gz")
    with gzip.open(path, "rt", encoding="utf-8") as handle:
        rows = [json.loads(line) for line in handle if line.strip()]
    rows = change(rows) or rows
    with open(path, "wb") as raw, gzip.GzipFile(fileobj=raw, mode="wb", mtime=0, filename="") as out:
        out.write("".join(json.dumps(row, sort_keys=True) + "\n" for row in rows).encode())


def first(rows, **fields):
    return next(row for row in rows if all(row[key] == value for key, value in fields.items()))


class RecordTest(unittest.TestCase):
    def test_the_committed_record_is_what_its_data_give(self):
        result = verify.main(HERE, write=False)
        self.assertEqual(result["bars_total"], 10)
        self.assertEqual(result["bars_held"], 10)
        self.assertTrue(all(bar["holds"] for bar in result["bars"]))

    def test_the_bars_are_read_on_the_measured_parts(self):
        result = verify.results(HERE)
        posts = result["sets"]["set-s/measured"]
        by_name = {bar["bar"]: bar for bar in result["bars"]}
        self.assertEqual(by_name["posts: either rule"]["value"], posts["either"])
        self.assertEqual(by_name["posts: syntax.closing-denial-density"]["value"], posts["denial"])
        self.assertEqual(by_name["quiet on human-docs: syntax.uniform-register"]["value"], result["sets"]["human-docs/measured"]["register"])
        self.assertEqual(posts["pages"], 30)
        # The model texts of the two writers are read together.
        self.assertEqual(result["fresh_together"]["pages"], 120)

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

            def quiet(rows):
                first(rows, corpus="public-today", register=1)["register"] = 0

            rewrite(root, "pages", quiet)
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

    def bars(self):
        return {bar["bar"]: bar["holds"] for bar in verify.results(self.root)["bars"]}

    def test_a_rule_that_reports_too_many_pages_of_people_fails_its_bar(self):
        def loud(rows):
            for row in [row for row in rows if row["corpus"] == "human-blog" and row["part"] == "measured"][:4]:
                row["register"] = 1

        rewrite(self.root, "pages", loud)
        held = self.bars()
        # Four of 77 pages are over one in twenty; three would not be.
        self.assertFalse(held["quiet on human-blog: syntax.uniform-register"])
        self.assertTrue(held["quiet on human-blog: syntax.closing-denial-density"])
        self.assertTrue(held["quiet on human-styled: syntax.uniform-register"])

    def test_a_rule_that_is_quiet_on_the_posts_fails_its_bar(self):
        def quiet(rows):
            for row in rows:
                if row["corpus"] == "set-s" and row["part"] == "measured":
                    row["denial"] = 0

        rewrite(self.root, "pages", quiet)
        held = self.bars()
        self.assertFalse(held["posts: syntax.closing-denial-density"])
        self.assertTrue(held["posts: syntax.uniform-register"])

    def test_the_fitting_parts_do_not_move_a_bar(self):
        before = self.bars()

        def loud(rows):
            for row in rows:
                if row["part"] == "fitting":
                    row["register"], row["denial"] = 1, 3

        rewrite(self.root, "pages", loud)
        self.assertEqual(self.bars(), before)


class RowsTest(unittest.TestCase):
    def setUp(self):
        self.directory = tempfile.TemporaryDirectory()
        self.addCleanup(self.directory.cleanup)
        self.root = copy_of_record(self.directory.name)

    def refused(self, name, change, message):
        rewrite(self.root, name, change)
        with self.assertRaisesRegex(verify.Malformed, message):
            verify.results(self.root)

    def test_a_page_of_an_unknown_set_is_refused(self):
        self.refused("pages", lambda rows: rows[0].update(corpus="elsewhere"), "unknown set or part")

    def test_a_page_in_a_part_that_its_set_does_not_have_is_refused(self):
        self.refused("pages", lambda rows: first(rows, corpus="panel").update(part="measured"), "unknown set or part")

    def test_a_page_listed_twice_is_refused(self):
        self.refused("pages", lambda rows: rows + [dict(rows[0])], "listed twice")

    def test_a_missing_page_is_refused(self):
        self.refused("pages", lambda rows: [row for row in rows if row is not first(rows, corpus="set-s", part="measured")],
                     "set-s measured has 29 pages")

    def test_a_post_with_a_name_is_refused(self):
        self.refused("pages", lambda rows: first(rows, corpus="set-s").update(id="a-post-about-migrations"), "not of its set")

    def test_a_page_with_a_text_is_refused(self):
        self.refused("pages", lambda rows: rows[0].update(text="The words of a page."), "other fields")

    def test_fewer_denials_than_the_rule_reports_are_refused(self):
        self.refused("pages", lambda rows: rows[0].update(denial=2), "places of the denial rule")

    def test_a_register_finding_counted_twice_is_refused(self):
        self.refused("pages", lambda rows: rows[0].update(register=2), "reports a page once")

    def test_a_pair_without_its_after_is_refused(self):
        self.refused("pairs", lambda rows: rows[0].update(denial=[1]), "a before and an after")

    def test_a_pair_of_an_unknown_part_is_refused(self):
        self.refused("pairs", lambda rows: rows[0].update(part="reserve"), "unknown part")

    def test_a_missing_data_file_is_refused(self):
        (self.root / "data" / "pairs.jsonl.gz").unlink()
        with self.assertRaisesRegex(verify.Malformed, "missing data file"):
            verify.results(self.root)


class CountsTest(unittest.TestCase):
    def test_a_file_that_gains_or_loses_a_finding_is_counted_once(self):
        pairs = [{"part": "development", "id": "a", "denial": [1, 0], "register": [0, 0]},
                 {"part": "development", "id": "b", "denial": [1, 1], "register": [0, 1]},
                 {"part": "test", "id": "c", "denial": [0, 1], "register": [1, 0]}]
        counts = verify.pair_counts(pairs)
        self.assertEqual(counts["development"]["denial"], {"before": 2, "after": 1, "lost": 1, "gained": 0})
        self.assertEqual(counts["development"]["register"], {"before": 0, "after": 1, "lost": 0, "gained": 1})
        self.assertEqual(counts["test"]["denial"], {"before": 0, "after": 1, "lost": 0, "gained": 1})

    def test_the_interval_of_a_share_holds_the_share(self):
        low, high = verify.wilson(13, 30)
        self.assertLess(low, 13 / 30)
        self.assertGreater(high, 13 / 30)
        self.assertEqual(verify.wilson(0, 0), [0.0, 0.0])
        self.assertEqual(verify.share(0, 77)["interval"][0], 0.0)


if __name__ == "__main__":
    unittest.main()
