"""Tests of the recount of the record of the rules on the edited sentences of TETRA."""
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
    path = root / "data" / "sentences.jsonl.gz"
    with gzip.open(path, "rt", encoding="utf-8") as handle:
        rows = [json.loads(line) for line in handle if line.strip()]
    rows = change(rows) or rows
    with open(path, "wb") as raw, gzip.GzipFile(fileobj=raw, mode="wb", mtime=0, filename="") as out:
        out.write("".join(json.dumps(row, sort_keys=True) + "\n" for row in rows).encode())


class RecordTest(unittest.TestCase):
    def setUp(self):
        scratch = tempfile.TemporaryDirectory()
        self.addCleanup(scratch.cleanup)
        self.root = copy_of_record(scratch.name)

    def refused(self, change, message):
        rewrite(self.root, change)
        with self.assertRaisesRegex(verify.Malformed, message):
            verify.main(self.root, write=False)

    def test_the_committed_record_is_what_its_data_give(self):
        result = verify.main(HERE, write=False)
        self.assertFalse(result["bar"]["holds"])
        self.assertEqual(result["sentences"], 3876)

    def test_a_changed_mark_changes_the_record(self):
        def flag(rows):
            next(row for row in rows if row["edited"] and not row["default"])["default"] = True

        self.refused(flag, "results.json is not what the data give")

    def test_a_changed_data_file_needs_its_digest(self):
        def flag(rows):
            next(row for row in rows if row["edited"] and not row["default"])["default"] = True

        rewrite(self.root, flag)
        result = verify.results(self.root)
        (self.root / "results.json").write_text(verify.encode(result))
        (self.root / "RESULTS.md").write_text(verify.render(result))
        with self.assertRaisesRegex(verify.Malformed, "files.sha256 is not what the data give"):
            verify.main(self.root, write=False)

    def test_the_bar_holds_when_the_rules_follow_the_editors(self):
        def follow(rows):
            for row in rows:
                row["default"] = row["edited"]

        rewrite(self.root, follow)
        result = verify.results(self.root)
        self.assertTrue(result["bar"]["holds"])
        self.assertEqual(result["bar"]["value"], 1.0)

    def test_a_difference_just_under_the_bar_fails(self):
        def nearly(rows):
            # Flag 4.9% of the edited sentences and none of the others.
            edited = [row for row in rows if row["edited"]]
            for row in rows:
                row["default"] = False
            for row in edited[:int(0.049 * len(edited))]:
                row["default"] = True

        rewrite(self.root, nearly)
        self.assertFalse(verify.results(self.root)["bar"]["holds"])

    def test_a_row_with_another_field_is_refused(self):
        def add(rows):
            rows[0]["text"] = "a sentence"

        self.refused(add, "a sentence row has other fields")

    def test_a_sentence_listed_twice_is_refused(self):
        self.refused(lambda rows: rows + [dict(rows[0])], "listed twice")

    def test_a_redundancy_edit_needs_an_edit(self):
        def mark(rows):
            next(row for row in rows if not row["edited"])["redundancy"] = True

        self.refused(mark, "has a redundancy edit and no edit")

    def test_unsorted_rules_are_refused(self):
        def swap(rows):
            next(row for row in rows if len(row["every"]) > 1)["every"].reverse()

        self.refused(swap, "no sorted list of rule names")

    def test_a_paper_without_one_of_its_editors_is_refused(self):
        def drop(rows):
            paper = rows[0]["paper"]
            return [row for row in rows if not (row["paper"] == paper and row["editor"] == "C")]

        self.refused(drop, "lacks an editor")

    def test_a_missing_paper_is_refused(self):
        def drop(rows):
            paper = rows[0]["paper"]
            return [row for row in rows if row["paper"] != paper]

        self.refused(drop, "the data hold 55 papers")

    def test_a_count_that_the_page_does_not_state_is_refused(self):
        page = self.root / "README.md"
        page.write_text(page.read_text().replace("10 of 1,572", "11 of 1,572"))
        with self.assertRaisesRegex(verify.Malformed, "the README does not state '10 of 1,572'"):
            verify.main(self.root, write=False)


if __name__ == "__main__":
    unittest.main()
