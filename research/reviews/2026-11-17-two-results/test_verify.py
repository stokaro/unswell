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
        # Every bar is read on all its copies, and a bar holds or it does not.
        self.assertTrue(all(isinstance(bar["holds"], bool) for bar in result["bars"]))
        # The run of the copies with one judge was made in three parts, and the other judge made its run in one.
        self.assertEqual([part["copies"] for part in result["copies"][GPT]["parts"]], [198, 55, 121])
        self.assertEqual(sum(part["copies"] for part in result["copies"][GPT]["parts"]), result["copies"][GPT]["copies"])
        self.assertEqual([part["copies"] for part in result["copies"][QWEN]["parts"]], [374])

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
        before = self.bars()
        rewrite(self.root, "copies", spoil)
        held = self.bars()
        self.assertIs(held[("G1", QWEN, "harmless and trimmed copies")], False)
        # The bar of the other judge is read on its own copies.
        self.assertIs(held[("G1", GPT, "harmless and trimmed copies")], before[("G1", GPT, "harmless and trimmed copies")])

    def test_a_bar_is_read_on_every_part_of_a_run(self):
        def spoil(rows):
            # The harmless copies of the later parts alone are enough to fail the bar of the whole run.
            for row in rows:
                if row["judge"] == GPT and row["part"] > 1 and row["kind"] in verify.HARMLESS and row["verdict"]:
                    lose(row, "statements")

        rewrite(self.root, "copies", spoil)
        result = verify.results(self.root)
        later = result["copies"][GPT]["parts"][1:]
        self.assertEqual(sum(part["harmless"]["count"] for part in later), 0)
        self.assertGreater(sum(part["harmless"]["of"] for part in later), 0.15 * verify.COPIES["harmless"])
        held = {(bar["bar"], bar["judge"], bar["read_on"]): bar["holds"] for bar in result["bars"]}
        self.assertIs(held[("G1", GPT, "harmless and trimmed copies")], False)

    def test_a_copy_without_a_verdict_after_its_retry_counts_against_the_bar_of_verdicts(self):
        def silence(rows):
            mine = [row for row in rows if row["judge"] == GPT and row["retried"]]
            for row in mine[:20]:
                for key in ("nothing_lost", "failed", "sound", "accepted"):
                    row.pop(key, None)
                row["verdict"] = False

        rewrite(self.root, "copies", silence)
        result = verify.results(self.root)
        self.assertEqual(result["copies"][GPT]["verdicts"]["count"], 354)
        held = {(bar["bar"], bar["judge"], bar["read_on"]): bar["holds"] for bar in result["bars"]}
        self.assertIs(held[("G4", GPT, "the copies")], False)

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

    def test_a_missing_copy_is_refused_for_either_judge(self):
        for judge in (QWEN, GPT):
            with self.subTest(judge=judge):
                root = copy_of_record(tempfile.mkdtemp(dir=self.directory.name))
                rewrite(root, "copies", lambda rows: rows.remove(next(row for row in rows if row["judge"] == judge)))
                with self.assertRaisesRegex(verify.Malformed, "373 copies, and the record says 374"):
                    verify.results(root)

    def test_a_copy_of_a_part_that_its_run_did_not_have_is_refused(self):
        def later(rows):
            next(row for row in rows if row["judge"] == QWEN)["part"] = 2

        self.refused("copies", later, "a part that the run of this judge did not have")

    def test_a_copy_without_its_part_is_refused(self):
        def forget(rows):
            del rows[0]["part"]

        self.refused("copies", forget, "a part that the run of this judge did not have")

    def test_a_copy_moved_to_the_other_part_is_refused(self):
        def move(rows):
            next(row for row in rows if row["judge"] == GPT and row["part"] == 2)["part"] = 1

        self.refused("copies", move, "199 copies in part 1, and the record says 198")

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
        result = verify.results(HERE)
        self.assertEqual([found["reviser"] for found in result["controls"]], [QWEN, GPT])
        for found, other in zip(result["controls"], (GPT, QWEN)):
            with self.subTest(reviser=found["reviser"]):
                self.assertEqual((found["judge_of_own_family"], found["judge_of_other_family"]), (found["reviser"], other))
                self.assertEqual(found["revisions"] + found["without_verdict"], 57)
                # The count of the other judge is that of its first run of the measurement.
                if not found["without_verdict"]:
                    self.assertEqual(found["other_family"], result["revisions"][other]["runs"]["1"]["nothing_lost"])
                self.assertLessEqual(found["both"], min(found["own_family"], found["other_family"]))
                self.assertEqual(sum(part["revisions"] for part in found["by_pages"].values()), found["revisions"])

    def test_an_agreement_is_counted_on_each_revision(self):
        def turn(rows):
            for row in rows:
                if row["verdict"]:
                    keep(row)

        rewrite(self.root, "same", turn)
        for found in verify.results(self.root)["controls"]:
            with self.subTest(reviser=found["reviser"]):
                self.assertEqual(found["own_family"], found["revisions"])
                self.assertEqual(found["both"], found["other_family"])
                self.assertEqual(found["agree"]["count"], found["other_family"])

    def test_one_control_does_not_move_the_other(self):
        before = verify.results(self.root)["controls"]

        def spoil(rows):
            for row in rows:
                if row["judge"] == GPT and row["verdict"]:
                    lose(row)

        rewrite(self.root, "same", spoil)
        after = verify.results(self.root)["controls"]
        self.assertEqual(after[0], before[0])
        self.assertEqual(after[1]["own_family"], 0)

    def test_the_control_does_not_move_a_bar(self):
        before = verify.results(self.root)["bars"]

        def spoil(rows):
            for row in rows:
                if row["verdict"]:
                    lose(row)

        rewrite(self.root, "same", spoil)
        self.assertEqual(verify.results(self.root)["bars"], before)

    def test_a_control_row_under_the_judge_of_the_other_family_is_refused(self):
        def other(rows):
            rows[0]["judge"] = GPT if rows[0]["judge"] == QWEN else QWEN

        rewrite(self.root, "same", other)
        with self.assertRaisesRegex(verify.Malformed, "judge of its own family"):
            verify.results(self.root)

    def test_a_control_without_one_of_the_revisions_is_refused(self):
        for judge in (QWEN, GPT):
            with self.subTest(judge=judge):
                root = copy_of_record(tempfile.mkdtemp(dir=self.directory.name))
                rewrite(root, "same", lambda rows: rows.remove(next(row for row in rows if row["judge"] == judge)))
                with self.assertRaisesRegex(verify.Malformed, "each once"):
                    verify.results(root)

    def test_a_control_row_listed_twice_is_refused(self):
        rewrite(self.root, "same", lambda rows: rows + [dict(rows[0])])
        with self.assertRaisesRegex(verify.Malformed, "listed twice"):
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
