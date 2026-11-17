"""Tests of the recount of the record of three pilots of revising a post."""
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


class RecordTest(unittest.TestCase):
    def test_the_committed_record_is_what_its_data_give(self):
        result = verify.main(HERE, write=False)
        first = result["arms"]["closing-sentences"]
        # The revision takes the closing denials out, and the judges pick about as many posts as before.
        self.assertLess(first["after"]["closing_sentences"], first["before"]["closing_sentences"])
        self.assertEqual(first["picked_revised"]["alone"][verify.GPT], first["revised"])
        self.assertEqual(result["rounds"]["named_post"], result["rounds"]["judgments"])

    def test_a_changed_number_is_not_accepted(self):
        with tempfile.TemporaryDirectory() as directory:
            root = copy_of_record(directory)
            stored = json.loads((root / "results.json").read_text())
            stored["rounds"]["named_post"] -= 1
            (root / "results.json").write_text(verify.encode(stored))
            with self.assertRaisesRegex(verify.Malformed, "results.json is not what the data give"):
                verify.main(root, write=False)

    def test_a_changed_answer_changes_the_count_and_the_file_list(self):
        with tempfile.TemporaryDirectory() as directory:
            root = copy_of_record(directory)
            before = verify.results(root)["arms"]["closing-sentences"]["picked_revised"]["alone"][verify.GPT]

            def spare(rows):
                page = next(row["page"] for row in rows if row["arm"] == "closing-sentences" and row["judge"] == verify.GPT)
                for row in rows:
                    if row["arm"] == "closing-sentences" and row["judge"] == verify.GPT and row["page"] == page:
                        row["named"] = "human"

            rewrite(root, "answers", spare)
            self.assertEqual(verify.results(root)["arms"]["closing-sentences"]["picked_revised"]["alone"][verify.GPT], before - 1)
            with self.assertRaises(verify.Malformed):
                verify.main(root, write=False)


class CountsTest(unittest.TestCase):
    def answers(self, named):
        return [{"arm": verify.AS_IT_IS, "judge": judge, "page": "s-01", "pair": pair, "named": named[judge][pair - 1]}
                for judge in named for pair in (1, 2, 3)]

    def test_a_post_is_picked_when_its_votes_outnumber_those_of_the_human_posts(self):
        named = {verify.GPT: ["post", "post", "post"], verify.QWEN27: ["post", "human", "human"], verify.QWEN35: ["human", "human", "same"]}
        found = verify.picks(self.answers(named), verify.AS_IT_IS, verify.JUDGES, ["s-01"])
        # Four votes for the post and four for the human posts: the three together do not pick it.
        self.assertEqual(found["together"], 0)
        self.assertEqual(found["alone"], {verify.GPT: 1, verify.QWEN27: 0, verify.QWEN35: 0})

    def test_a_tie_of_one_judge_is_not_a_pick(self):
        named = {verify.GPT: ["post", "human", "same"]}
        self.assertEqual(verify.picks(self.answers(named), verify.AS_IT_IS, (verify.GPT,), ["s-01"])["alone"][verify.GPT], 0)

    def test_a_missing_answer_is_refused(self):
        named = {verify.GPT: ["post", "post", "post"]}
        with self.assertRaisesRegex(verify.Malformed, "2 answers for 1 posts"):
            verify.picks(self.answers(named)[:2], verify.AS_IT_IS, (verify.GPT,), ["s-01"])


class RowsTest(unittest.TestCase):
    def setUp(self):
        self.directory = tempfile.TemporaryDirectory()
        self.addCleanup(self.directory.cleanup)
        self.root = copy_of_record(self.directory.name)

    def refused(self, name, change, message):
        rewrite(self.root, name, change)
        with self.assertRaisesRegex(verify.Malformed, message):
            verify.results(self.root)

    def test_a_post_with_a_name_is_refused(self):
        self.refused("answers", lambda rows: rows[0].update(page="a-post-about-migrations"), "no name")

    def test_an_answer_with_a_text_is_refused(self):
        self.refused("answers", lambda rows: rows[0].update(reason="Its sentences have one length."), "what a judge named")

    def test_an_answer_of_a_judge_that_was_not_asked_is_refused(self):
        def ask(rows):
            row = next(row for row in rows if row["arm"] == "manner")
            row["judge"] = verify.GPT

        self.refused("answers", ask, "not asked")

    def test_an_answer_listed_twice_is_refused(self):
        self.refused("answers", lambda rows: rows + [dict(rows[0])], "listed twice")

    def test_fewer_closing_sentences_than_the_rule_lists_are_refused(self):
        def two(rows):
            next(row for row in rows if row["revised"])["after"]["closing_sentences"] = 2

        self.refused("revisions", two, "none or more than two")

    def test_a_revision_without_its_after_is_refused(self):
        def drop(rows):
            del next(row for row in rows if row["revised"])["after"]

        self.refused("revisions", drop, "fields of a revision")

    def test_a_judgment_of_a_sixth_round_is_refused(self):
        self.refused("rounds", lambda rows: rows[0].update(round=5), "unknown round or pair")

    def test_a_post_with_a_missing_round_is_refused(self):
        self.refused("rounds", lambda rows: rows[1:], "does not have 5 rounds")

    def test_a_missing_data_file_is_refused(self):
        (self.root / "data" / "rounds.jsonl.gz").unlink()
        with self.assertRaisesRegex(verify.Malformed, "missing data file"):
            verify.results(self.root)


if __name__ == "__main__":
    unittest.main()
