"""Tests of the recount of the record of the two results of the check of a revision."""
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

GPT, QWEN = verify.GPT, verify.QWEN35


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


def lose(row, check="sentences"):
    """Make a row one whose pair lost something."""
    row.update(nothing_lost=False, failed=[check], accepted=False)


def keep(row):
    """Make a row one whose pair lost nothing."""
    row.update(nothing_lost=True, failed=[], accepted=row["sound"] == "better")


class RecordTest(unittest.TestCase):
    def test_the_committed_record_is_what_its_data_give(self):
        result = verify.main(HERE, write=False)
        self.assertEqual(result["bars_total"], 14)
        self.assertEqual(result["bars_held"], sum(1 for bar in result["bars"] if bar["holds"]))
        self.assertEqual([bar["bar"] for bar in result["bars"]], ["G1"] * 2 + ["G2"] * 4 + ["G3"] * 2 + ["G4"] * 6)
        # The run of the copies with one judge was stopped, so three bars of that judge are not read. No bar that is read fails.
        not_read = [(bar["bar"], bar["judge"], bar["read_on"]) for bar in result["bars"] if bar["holds"] is None]
        self.assertEqual(not_read, [("G1", GPT, "harmless and trimmed copies"), ("G2", GPT, "damaged copies of fresh"), ("G4", GPT, "the copies")])
        self.assertEqual(result["bars_read"], 11)
        self.assertEqual(result["bars_held"], 11)
        self.assertEqual(result["stopped"], [GPT])

    def test_a_changed_number_is_not_accepted(self):
        with tempfile.TemporaryDirectory() as directory:
            root = copy_of_record(directory)
            stored = json.loads((root / "results.json").read_text())
            stored["bars_held"] += 1
            (root / "results.json").write_text(verify.encode(stored))
            with self.assertRaisesRegex(verify.Malformed, "results.json is not what the data give"):
                verify.main(root, write=False)

    def test_a_changed_data_file_is_not_the_one_the_record_names(self):
        with tempfile.TemporaryDirectory() as directory:
            root = copy_of_record(directory)
            rewrite(root, "fitting", lambda rows: rows[0].update(retried=not rows[0]["retried"]))
            with self.assertRaisesRegex(verify.Malformed, "files.sha256 is not what the data give"):
                verify.main(root, write=False)


class BarsTest(unittest.TestCase):
    def setUp(self):
        self.directory = tempfile.TemporaryDirectory()
        self.addCleanup(self.directory.cleanup)
        self.root = copy_of_record(self.directory.name)

    def bars(self):
        return {(bar["bar"], bar["judge"], bar["read_on"]): bar["holds"] for bar in verify.results(self.root)["bars"]}

    def test_harmless_copies_that_lose_something_fail_the_bar_of_their_judge(self):
        def spoil(rows):
            mine = [row for row in rows if row["judge"] == QWEN and row["kind"] in verify.HARMLESS and row["verdict"]]
            for row in mine[:60]:
                lose(row, "statements")

        rewrite(self.root, "copies", spoil)
        held = self.bars()
        self.assertIs(held[("G1", QWEN, "harmless and trimmed copies")], False)
        # The other judge did not see every harmless copy, so its bar is still not read.
        self.assertIsNone(held[("G1", GPT, "harmless and trimmed copies")])

    def test_a_bar_of_a_stopped_run_is_not_read_whatever_its_copies_say(self):
        def spoil(rows):
            for row in rows:
                if row["judge"] == GPT and row["kind"] in verify.HARMLESS and row["verdict"]:
                    lose(row, "statements")

        rewrite(self.root, "copies", spoil)
        held = self.bars()
        self.assertIsNone(held[("G1", GPT, "harmless and trimmed copies")])
        self.assertIsNone(held[("G4", GPT, "the copies")])
        # The damaged copies of the posts were all judged, so that bar is read.
        self.assertIs(held[("G2", GPT, "damaged copies of set-s")], True)

    def test_a_harmless_copy_without_a_verdict_does_not_pass(self):
        def silence(rows):
            mine = [row for row in rows if row["judge"] == QWEN and row["kind"] in verify.HARMLESS]
            for row in mine[:30]:
                for key in ("nothing_lost", "failed", "sound", "accepted"):
                    row.pop(key, None)
                row["verdict"] = False

        rewrite(self.root, "copies", silence)
        result = verify.results(self.root)
        self.assertLessEqual(result["copies"][QWEN]["harmless"]["count"], 125)
        self.assertEqual(result["copies"][QWEN]["harmless"]["of"], 155)
        held = {(bar["bar"], bar["judge"], bar["read_on"]): bar["holds"] for bar in result["bars"]}
        self.assertFalse(held[("G1", QWEN, "harmless and trimmed copies")])
        self.assertFalse(held[("G4", QWEN, "the copies")])

    def test_three_damaged_copies_that_lose_nothing_fail_the_bar_of_their_pages(self):
        def pass_three(rows):
            mine = [row for row in rows if row["judge"] == QWEN and row["kind"] in verify.DAMAGE and row["pages"] == "fresh" and row["verdict"]]
            for row in mine:
                lose(row)
            for row in mine[:3]:
                keep(row)

        rewrite(self.root, "copies", pass_three)
        held = self.bars()
        self.assertIs(held[("G2", QWEN, "damaged copies of fresh")], False)
        self.assertIs(held[("G2", QWEN, "damaged copies of set-s")], True)

    def test_a_result_that_differs_between_the_runs_fails_the_bar_of_repeats(self):
        def differ(rows):
            first = [row for row in rows if row["judge"] == QWEN and row["run"] == 1 and row["verdict"]]
            second = {row["page"]: row for row in rows if row["judge"] == QWEN and row["run"] == 2 and row["verdict"]}
            changed = 0
            for row in first:
                other = second.get(row["page"])
                if other and changed < 12:
                    keep(row)
                    lose(other)
                    changed += 1

        rewrite(self.root, "revisions", differ)
        result = verify.results(self.root)
        self.assertLess(result["revisions"][QWEN]["repeat"]["share"], 0.85)
        held = {(bar["bar"], bar["judge"]): bar["holds"] for bar in result["bars"] if bar["bar"] == "G3"}
        self.assertFalse(held[("G3", QWEN)])

    def test_the_fitting_half_does_not_move_a_bar(self):
        before = self.bars()

        def spoil(rows):
            for row in rows:
                if row["verdict"]:
                    lose(row)

        rewrite(self.root, "fitting", spoil)
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

    def judged(self, rows):
        return next(row for row in rows if row["verdict"])

    def test_a_copy_of_an_unknown_kind_is_refused(self):
        self.refused("copies", lambda rows: rows[0].update(kind="reorder"), "unknown judge, pages, or kind")

    def test_a_copy_listed_twice_is_refused(self):
        self.refused("copies", lambda rows: rows + [dict(rows[0])], "listed twice")

    def test_a_missing_copy_of_a_run_that_was_not_stopped_is_refused(self):
        def drop(rows):
            rows.remove(next(row for row in rows if row["judge"] == QWEN))

        self.refused("copies", drop, "373 copies, and the record says 374")

    def test_only_a_stopped_run_may_have_fewer_copies_and_none_may_have_more(self):
        self.assertFalse(verify.judged_all(GPT, "copies", 198, 374))
        self.assertTrue(verify.judged_all(GPT, "copies", 374, 374))
        with self.assertRaisesRegex(verify.Malformed, "375 copies"):
            verify.judged_all(GPT, "copies", 375, 374)
        with self.assertRaisesRegex(verify.Malformed, "373 copies"):
            verify.judged_all(QWEN, "copies", 373, 374)

    def test_a_pair_that_lost_nothing_and_failed_a_check_is_refused(self):
        self.refused("copies", lambda rows: self.judged(rows).update(nothing_lost=True, failed=["dashes"]),
                     "nothing is lost exactly when no check failed")

    def test_an_accepted_pair_that_lost_something_is_refused(self):
        def accept(rows):
            row = self.judged(rows)
            lose(row)
            row["accepted"] = True

        self.refused("copies", accept, "accepted exactly when")

    def test_a_check_without_a_verdict_and_with_a_result_is_refused(self):
        self.refused("copies", lambda rows: self.judged(rows).update(verdict=False), "without a verdict has no result")

    def test_an_unknown_check_is_refused(self):
        self.refused("copies", lambda rows: self.judged(rows).update(nothing_lost=False, failed=["tone"], accepted=False),
                     "sound or failed checks")

    def test_a_revision_judged_by_the_family_of_its_reviser_is_refused(self):
        self.refused("revisions", lambda rows: rows[0].update(reviser=rows[0]["judge"]), "other family")

    def test_a_revision_of_a_third_run_is_refused(self):
        self.refused("revisions", lambda rows: rows[0].update(run=3), "unknown pages or run")

    def test_more_kept_statements_than_statements_are_refused(self):
        self.refused("revisions", lambda rows: self.judged(rows).update(kept=10_000), "statements kept")

    def test_a_fitting_pair_of_an_unknown_build_is_refused(self):
        self.refused("fitting", lambda rows: rows[0].update(build="next"), "unknown kind of pair, judge, or build")

    def test_a_missing_data_file_is_refused(self):
        (self.root / "data" / "revisions.jsonl.gz").unlink()
        with self.assertRaisesRegex(verify.Malformed, "missing data file"):
            verify.results(self.root)


class ControlTest(unittest.TestCase):
    def setUp(self):
        self.directory = tempfile.TemporaryDirectory()
        self.addCleanup(self.directory.cleanup)
        self.root = copy_of_record(self.directory.name)

    def test_the_two_judges_are_read_on_the_same_revisions(self):
        found = verify.results(HERE)["same_revisions"]
        self.assertEqual(found["revisions"], 57)
        self.assertEqual((found["reviser"], found["judge_of_own_family"], found["judge_of_other_family"]), (QWEN, QWEN, GPT))
        # The count of the other judge is that of its first run of the measurement.
        self.assertEqual(found["other_family"], verify.results(HERE)["revisions"][GPT]["runs"]["1"]["nothing_lost"])
        self.assertLessEqual(found["both"], min(found["own_family"], found["other_family"]))
        self.assertEqual(sum(part["revisions"] for part in found["by_pages"].values()), 57)

    def test_an_agreement_is_counted_on_each_revision(self):
        def turn(rows):
            for row in rows:
                if row["verdict"]:
                    keep(row)

        rewrite(self.root, "same", turn)
        found = verify.results(self.root)["same_revisions"]
        self.assertEqual(found["own_family"], 57)
        self.assertEqual(found["both"], found["other_family"])
        self.assertEqual(found["agree"]["count"], found["other_family"])

    def test_the_control_does_not_move_a_bar(self):
        before = verify.results(self.root)["bars"]

        def spoil(rows):
            for row in rows:
                if row["verdict"]:
                    lose(row)

        rewrite(self.root, "same", spoil)
        self.assertEqual(verify.results(self.root)["bars"], before)

    def test_a_control_row_of_another_judge_is_refused(self):
        rewrite(self.root, "same", lambda rows: rows[0].update(judge=GPT))
        with self.assertRaisesRegex(verify.Malformed, "judge of its own family"):
            verify.results(self.root)

    def test_a_control_without_one_of_the_revisions_is_refused(self):
        rewrite(self.root, "same", lambda rows: rows[1:])
        with self.assertRaisesRegex(verify.Malformed, "each once"):
            verify.results(self.root)


class CountsTest(unittest.TestCase):
    def row(self, page, run, lost, sound="no difference", verdict=True):
        row = {"page": page, "run": run, "verdict": verdict}
        if verdict:
            row.update(nothing_lost=not lost, sound=sound)
        return row

    def test_a_repeat_is_read_only_on_pairs_with_a_verdict_in_both_runs(self):
        rows = [self.row("a", 1, False), self.row("a", 2, False), self.row("b", 1, False), self.row("b", 2, True),
                self.row("c", 1, True), self.row("c", 2, True, verdict=False), self.row("d", 1, True)]
        same, both = verify.repeats(rows, lambda row: row["nothing_lost"])
        self.assertEqual(same, {"count": 1, "of": 2, "share": 0.5})
        self.assertEqual(len(both), 2)

    def test_the_blind_comparison_has_its_own_repeat(self):
        rows = [self.row("a", 1, False, "better"), self.row("a", 2, False, "no difference"),
                self.row("b", 1, True, "worse"), self.row("b", 2, False, "worse")]
        self.assertEqual(verify.repeats(rows, lambda row: row["sound"])[0]["count"], 1)
        self.assertEqual(verify.repeats(rows, lambda row: row["nothing_lost"])[0]["count"], 1)


if __name__ == "__main__":
    unittest.main()
