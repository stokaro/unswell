"""Tests of the recount: it accepts a well-formed record and rejects each kind of damage."""
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

GPT, QWEN35 = verify.GPT, verify.QWEN35
MEASURED = "measured"
FITTING = "fitting"
GROUPS = ((FITTING, ("set-s",)), (MEASURED, verify.PAGES))


def copy(part, pages, number, kind):
    """A copy of a page.

    Every damaged sentence is a suspect but the first removed one, and no model is asked about a removed sentence. No
    harmless copy has a suspect. The sentence of a trimmed copy is a suspect whose left-out words are named.
    """
    row = {"part": part, "pages": pages, "unit": f"u{number}", "kind": kind, "compared": 60, "suspects": 0}
    if kind in verify.DAMAGE:
        row["named"] = (number, kind) != (0, "sentence")
        row["suspects"] = int(row["named"])
        if row["named"]:
            row["asked"] = kind != "sentence"
    if kind == verify.TRIM:
        row.update(named=True, cut=True, suspects=1)
    return row


def checked(named=("revision", "page"), statements=50, kept=50, suspects=0, lost=0):
    """The parts of one verdict, with the verdict that follows from them."""
    row = {"sound": verify.sound_of(list(named)), "named": list(named), "statements": statements, "kept": kept, "suspects": suspects, "lost": lost}
    passes = row["sound"] == "better" and kept >= 0.98 * statements and lost == 0
    return {**row, "verdict": "accepted" if passes else "rejected"}


def damaged(part, pages, number, kind, judge):
    """A damaged copy under a judge: the sentences report all but the first removed sentence, which the statements report."""
    missed = (number, kind) == (0, "sentence")
    parts = checked(kept=49, suspects=0, lost=0) if missed else checked(suspects=1, lost=1)
    return {"part": part, "pages": pages, "unit": f"u{number}", "kind": kind, "judge": judge, **parts, "reported": not missed}


def trimmed(part, pages, number, judge):
    """A trimmed copy under a judge: the padding is cleared as padding, but on the first post, where the sentence is reported."""
    reported = (pages, number) == ("set-s", 0)
    parts = checked(suspects=1, lost=1) if reported else checked(suspects=1)
    return {"part": part, "pages": pages, "unit": f"u{number}", "kind": verify.TRIM, "judge": judge, **parts, "reported": reported,
            "padding": not reported}


def revision(part, reviser, number):
    """A revision: the first is accepted, the second has two suspects that were cleared, one of them as padding, the
    third loses a sentence, and the fourth loses statements."""
    parts = {0: checked(named=("page", "page")), 1: checked(suspects=2), 2: checked(named=("page", "page"), suspects=1, lost=1),
             3: checked(kept=48)}.get(number, checked())
    not_kept = {"changed": parts["statements"] - parts["kept"], "missing": 0, "unquoted": 0}
    return {"part": part, "pages": "set-s", "unit": f"u{number}", "reviser": reviser, "judge": verify.CHECK_JUDGE[reviser], **parts,
            "not_kept": not_kept, "lost_by": {"missing": 0, "negation": 0, "cut": 0, "reworded": parts["lost"]},
            "padding": int(number == 1), "unasked": 0, "em_dashes": [2, 3] if number == 4 else [2, 2]}


def record():
    cells = [(part, pages, number) for part, groups in GROUPS for pages in groups for number in range(10)]
    copies = [copy(*cell, kind) for cell in cells for kind in verify.DAMAGE + verify.HARMLESS + (verify.TRIM,)]
    damage = [damaged(*cell, kind, judge) for cell in cells for kind in verify.DAMAGE for judge in verify.JUDGES]
    trims = [trimmed(*cell, judge) for cell in cells for judge in verify.JUDGES]
    revisions = [revision(part, reviser, number) for part in verify.REVISION_PARTS for reviser in verify.JUDGES for number in range(10)]
    return {"copies": copies, "damage": damage, "trims": trims, "revisions": revisions}


def write(root, data):
    (root / "data").mkdir(exist_ok=True)
    for name, rows in data.items():
        with gzip.GzipFile(root / "data" / (name + ".jsonl.gz"), "wb", mtime=0) as handle:
            for row in rows:
                handle.write((json.dumps(row, sort_keys=True) + "\n").encode())


class RecountTest(unittest.TestCase):
    def setUp(self):
        self.root = Path(tempfile.mkdtemp())
        self.addCleanup(shutil.rmtree, self.root)
        self.data = record()
        write(self.root, self.data)
        verify.verify(self.root, write=True)

    def broken(self, name, change):
        """Change the rows of one data file, name them in files.sha256 again, and return the failure of the recount."""
        rows = json.loads(json.dumps(self.data[name]))
        change(rows)
        write(self.root, {**self.data, name: rows})
        (self.root / "files.sha256").write_text(verify.digests(self.root))
        with self.assertRaises(verify.Malformed) as caught:
            verify.verify(self.root)
        return str(caught.exception)

    def changed(self, name, change):
        """Change the rows of one data file and return the recount of the changed record."""
        rows = json.loads(json.dumps(self.data[name]))
        change(rows)
        write(self.root, {**self.data, name: rows})
        return verify.verify(self.root, write=True)

    def test_a_well_formed_record_recounts(self):
        result = verify.verify(self.root)
        measured = result["copies"][MEASURED]["set-s"]
        self.assertEqual((measured["sentence"]["named"], measured["swap"]["named"], measured["damaged"]["named_share"]), (9, 10, 0.983))
        # No model is asked about a removed sentence, and one is asked about every other changed sentence.
        self.assertEqual((measured["sentence"]["unasked"], measured["tail"]["unasked"], measured["damaged"]["unasked"]), (9, 0, 9))
        self.assertEqual((measured["harmless"]["copies"], measured["harmless"]["with_a_suspect"]), (20, 0))
        self.assertEqual((measured["trim"]["copies"], measured["trim"]["named"], measured["trim"]["cut"]), (10, 10, 10))
        cell = result["damage"][MEASURED]["fresh"][QWEN35]
        self.assertEqual((cell["copies"], cell["verdicts"], cell["reported"], cell["by_statements"], cell["by_either"]), (60, 60, 59, 1, 60))
        self.assertEqual((cell["first_verdicts"], cell["first_verdict_share"]), (60, 1.0))
        self.assertEqual((cell["sentence"]["reported"], cell["tail"]["reported"], cell["accepted"], cell["reported_share"]), (9, 10, 0, 0.983))
        self.assertEqual(cell["sound"], {"better": 0, "no difference": 60, "worse": 0})
        # Each kind of damage is read over the copies of both kinds of pages.
        self.assertEqual(result["kinds"][GPT]["sentence"], {"copies": 20, "reported": 18, "reported_share": 0.9})
        self.assertEqual(result["kinds"][GPT]["tail"], {"copies": 20, "reported": 20, "reported_share": 1.0})
        trims = result["trims"][MEASURED]["set-s"][GPT]
        self.assertEqual((trims["copies"], trims["verdicts"], trims["reported"], trims["padding"], trims["accepted"]), (10, 10, 1, 9, 0))
        self.assertEqual(result["trimmed"][GPT], {"copies": 20, "verdicts": 20, "reported": 1, "padding": 19, "accepted": 0, "reported_share": 0.05})
        mine = result["revisions"]["read"][GPT]
        self.assertEqual((mine["judge"], mine["pairs"], mine["accepted"], mine["kept"], mine["statements"]), (QWEN35, 10, 1, 498, 500))
        self.assertEqual((mine["kept_share"], mine["under_the_bar"], mine["not_kept"]["changed"]), (0.996, 1, 2))
        self.assertEqual((mine["with_a_suspect"], mine["suspects"], mine["cleared"], mine["with_a_lost_sentence"]), (2, 3, 2, 1))
        self.assertEqual((mine["lost"], mine["rejected_by_sentences_alone"], mine["add_em_dash"], mine["padding"], mine["unasked"]),
                         ({"missing": 0, "negation": 0, "cut": 0, "reworded": 1}, 1, 1, 1, 0))
        held = result["registered"]
        # S: 2. H: 2. D1: 4. D2: one for each of six kinds and two judges. D3: 4. T: 2. V: 4 and 4. R: 4.
        self.assertEqual((len(held), sum(held.values())), (38, 38))
        self.assertIn("38 of 38 registered bars hold.", (self.root / "RESULTS.md").read_text())

    def test_each_bar_on_the_copies_can_fail(self):
        def unnamed(rows):
            for row in rows:
                if (row["part"], row["pages"], row["kind"]) == (MEASURED, "fresh", "tail") and row["unit"] in ("u1", "u2"):
                    row.update(named=False, suspects=0)
                    row.pop("asked")
        result = self.changed("copies", unnamed)
        # Three of sixty changed sentences are no suspects on the fresh texts.
        self.assertFalse(result["registered"]["S, the changed sentence is a suspect, fresh"])
        self.assertTrue(result["registered"]["S, the changed sentence is a suspect, set-s"])

        def suspected(rows):
            for row in rows:
                if (row["part"], row["pages"], row["kind"]) == (MEASURED, "fresh", "join") and row["unit"] in ("u1", "u2"):
                    row["suspects"] = 1
        result = self.changed("copies", suspected)
        self.assertFalse(result["registered"]["H, harmless copies, fresh"])
        self.assertTrue(result["registered"]["H, harmless copies, set-s"])

    def test_each_bar_on_the_damaged_copies_can_fail(self):
        def missed(rows):
            for row in rows:
                if (row["part"], row["judge"], row["kind"]) == (MEASURED, GPT, "clause") and row["unit"] in ("u1", "u2", "u3"):
                    row.update(suspects=0, lost=0, reported=False)
        result = self.changed("damage", missed)
        # Six of twenty copies with a clause cut off go unreported: that kind fails, and no other bar does.
        self.assertFalse(result["registered"][f"D2, clause reported, {GPT}"])
        self.assertTrue(result["registered"][f"D2, clause reported, {QWEN35}"])
        self.assertTrue(result["registered"][f"D2, tail reported, {GPT}"])
        self.assertTrue(result["registered"][f"D1, damage reported, set-s, {GPT}"])

        def many_missed(rows):
            for row in rows:
                if (row["part"], row["pages"], row["judge"]) == (MEASURED, "set-s", GPT) and row["unit"] in ("u1", "u2"):
                    row.update(suspects=0, lost=0, reported=False)
        result = self.changed("damage", many_missed)
        # Thirteen of sixty copies of the posts go unreported: the bar on the pages fails, and each kind, read over
        # both kinds of pages, still holds.
        self.assertFalse(result["registered"][f"D1, damage reported, set-s, {GPT}"])
        self.assertTrue(result["registered"][f"D1, damage reported, fresh, {GPT}"])
        self.assertTrue(result["registered"][f"D2, swap reported, {GPT}"])

        def accepted(rows):
            for row in rows:
                if (row["part"], row["pages"], row["judge"], row["kind"]) == (MEASURED, "fresh", QWEN35, "clause") and row["unit"] in ("u1", "u2", "u3"):
                    row.update(checked(named=("page", "page")), reported=False)
        result = self.changed("damage", accepted)
        self.assertFalse(result["registered"][f"D3, damaged copies accepted, fresh, {QWEN35}"])
        self.assertTrue(result["registered"][f"D3, damaged copies accepted, set-s, {QWEN35}"])

        def silent(rows):
            for row in rows:
                if (row["part"], row["pages"], row["judge"]) == (MEASURED, "set-s", QWEN35) and row["unit"] in ("u3", "u4"):
                    for name in ("sound", "named", "statements", "kept", "suspects", "lost", "reported"):
                        row.pop(name)
                    row["verdict"] = "no verdict"
        result = self.changed("damage", silent)
        self.assertFalse(result["registered"][f"V, verdicts on damaged copies, set-s, {QWEN35}"])

    def test_the_bar_on_the_trimmed_copies_can_fail(self):
        def reported(rows):
            for row in rows:
                if (row["part"], row["pages"], row["judge"]) == (MEASURED, "fresh", QWEN35) and row["unit"] in ("u1", "u2"):
                    row.update(checked(suspects=1, lost=1), reported=True, padding=False)
        result = self.changed("trims", reported)
        # Three of twenty sentences that only lost their padding are reported with one judge.
        self.assertFalse(result["registered"][f"T, trimmed copies reported, {QWEN35}"])
        self.assertTrue(result["registered"][f"T, trimmed copies reported, {GPT}"])
        self.assertEqual(result["trimmed"][QWEN35]["reported"], 3)

    def test_a_retried_pair_counts_after_its_retry_and_is_shown(self):
        def retried(rows):
            for row in rows:
                if (row["part"], row["pages"], row["judge"]) == (MEASURED, "set-s", GPT) and row["unit"] in ("u1", "u2", "u3"):
                    row["retried"] = True
        result = self.changed("damage", retried)
        cell = result["damage"][MEASURED]["set-s"][GPT]
        # Eighteen copies needed the retry and got a verdict from it: the bar is read after it, and the first attempt is shown.
        self.assertEqual((cell["copies"], cell["verdicts"], cell["first_verdicts"], cell["first_verdict_share"]), (60, 60, 42, 0.7))
        self.assertTrue(result["registered"][f"V, verdicts on damaged copies, set-s, {GPT}"])
        self.assertIn("On the first attempt the shares were posts of the style-guide set, `gpt-6-astra`, 70.0%", (self.root / "RESULTS.md").read_text())

        def again(rows):
            for row in rows:
                if (row["part"], row["reviser"]) == ("read", GPT) and row["unit"] == "u5":
                    row["retried"] = True
        result = self.changed("revisions", again)
        self.assertEqual((result["revisions"]["read"][GPT]["verdicts"], result["revisions"]["read"][GPT]["first_verdicts"]), (10, 9))

        def lossy(rows):
            for row in rows:
                if (row["part"], row["reviser"]) == ("fresh", GPT):
                    row.update(kept=47, verdict="rejected", not_kept={"changed": 0, "missing": 0, "unquoted": 3})
        result = self.changed("revisions", lossy)
        self.assertFalse(result["registered"][f"R, statements kept, fresh, {GPT}"])
        self.assertTrue(result["registered"][f"R, statements kept, read, {GPT}"])

    def test_a_malformed_copy_is_rejected(self):
        self.assertIn("copy listed twice", self.broken("copies", lambda rows: rows.append(json.loads(json.dumps(rows[0])))))
        self.assertIn("unknown part", self.broken("copies", lambda rows: rows[0].update(part="later")))
        self.assertIn("unknown part", self.broken("copies", lambda rows: rows[0].update(kind="paragraph")))
        self.assertIn("is not a count", self.broken("copies", lambda rows: rows[0].update(suspects=-1)))
        self.assertIn("is not a count", self.broken("copies", lambda rows: rows[0].update(compared=True)))
        self.assertIn("does not say whether the sentence", self.broken("copies", lambda rows: rows[0].pop("named")))
        self.assertIn("a suspect among none", self.broken("copies", lambda rows: rows[0].update(named=True, asked=True)))
        self.assertIn("whether a judge is asked", self.broken("copies", lambda rows: rows[1].pop("asked")))
        self.assertIn("whether the left-out words are named", self.broken("copies", lambda rows: rows[8].pop("cut")))

        def without(rows):
            rows[:] = [row for row in rows if (row["part"], row["pages"]) != (MEASURED, "fresh")]
        self.assertIn("are not complete", self.broken("copies", without))

    def test_a_malformed_damaged_copy_is_rejected(self):
        self.assertIn("damaged copy listed twice", self.broken("damage", lambda rows: rows.append(json.loads(json.dumps(rows[0])))))
        self.assertIn("unknown judge or verdict", self.broken("damage", lambda rows: rows[0].update(judge="someone")))
        self.assertIn("unknown judge or verdict", self.broken("damage", lambda rows: rows[0].update(verdict="maybe")))
        self.assertIn("unknown part", self.broken("damage", lambda rows: rows[0].update(kind="trim")))
        self.assertIn("sound does not follow", self.broken("damage", lambda rows: rows[0].update(sound="better")))
        self.assertIn("malformed answers", self.broken("damage", lambda rows: rows[0].update(named=["page"])))
        self.assertIn("counts do not add up", self.broken("damage", lambda rows: rows[0].update(kept=51)))
        self.assertIn("counts do not add up", self.broken("damage", lambda rows: rows[2].update(lost=2)))
        self.assertIn("verdict does not follow", self.broken("damage", lambda rows: rows[2].update(verdict="accepted")))
        self.assertIn("report does not follow", self.broken("damage", lambda rows: rows[0].update(reported=True)))
        self.assertIn("report does not follow", self.broken("damage", lambda rows: rows[2].pop("reported")))
        self.assertIn("whether it was run again", self.broken("damage", lambda rows: rows[0].update(retried="yes")))

        def one_judge(rows):
            rows[:] = [row for row in rows if (row["part"], row["judge"]) != (MEASURED, GPT)]
        self.assertIn("are not complete", self.broken("damage", one_judge))

    def test_a_malformed_trimmed_copy_is_rejected(self):
        self.assertIn("trimmed copy listed twice", self.broken("trims", lambda rows: rows.append(json.loads(json.dumps(rows[0])))))
        self.assertIn("unknown part", self.broken("trims", lambda rows: rows[0].update(kind="tail")))
        self.assertIn("what became of its sentence", self.broken("trims", lambda rows: rows[2].pop("padding")))
        self.assertIn("report does not follow", self.broken("trims", lambda rows: rows[2].update(reported=True)))
        self.assertIn("report does not follow", self.broken("trims", lambda rows: rows[0].update(padding=True)))

        def none(rows):
            rows[:] = [row for row in rows if (row["part"], row["judge"]) != (MEASURED, GPT)]
        self.assertIn("no measured trimmed copy", self.broken("trims", none))

    def test_a_malformed_revision_is_rejected(self):
        self.assertIn("revision listed twice", self.broken("revisions", lambda rows: rows.append(json.loads(json.dumps(rows[0])))))
        self.assertIn("unknown part or reviser", self.broken("revisions", lambda rows: rows[0].update(part="later")))
        self.assertIn("unknown part or reviser", self.broken("revisions", lambda rows: rows[0].update(reviser="someone")))
        self.assertIn("not of the other family", self.broken("revisions", lambda rows: rows[0].update(judge=rows[0]["reviser"])))
        self.assertIn("statements do not add up", self.broken("revisions", lambda rows: rows[3]["not_kept"].update(changed=1)))
        self.assertIn("statements do not add up", self.broken("revisions", lambda rows: rows[0].pop("not_kept")))
        self.assertIn("lost sentences do not add up", self.broken("revisions", lambda rows: rows[2]["lost_by"].update(negation=1)))
        self.assertIn("lost sentences do not add up", self.broken("revisions", lambda rows: rows[2]["lost_by"].pop("cut")))
        self.assertIn("padding or unasked sentences do not add up", self.broken("revisions", lambda rows: rows[1].update(padding=3)))
        self.assertIn("padding or unasked sentences do not add up", self.broken("revisions", lambda rows: rows[2].update(unasked=2)))
        self.assertIn("is not a count", self.broken("revisions", lambda rows: rows[0].pop("padding")))
        self.assertIn("verdict does not follow", self.broken("revisions", lambda rows: rows[2].update(verdict="accepted")))
        self.assertIn("verdict does not follow", self.broken("revisions", lambda rows: rows[3].update(verdict="accepted")))
        self.assertIn("without the em dashes", self.broken("revisions", lambda rows: rows[0].update(em_dashes=[1])))
        self.assertIn("with an em dash added", self.broken("revisions", lambda rows: rows[0].update(em_dashes=[0, 1])))

        def without(rows):
            rows[:] = [row for row in rows if (row["part"], row["reviser"]) != ("fresh", QWEN35)]
        self.assertIn("are not complete", self.broken("revisions", without))

    def test_every_kind_of_damage_has_to_be_measured(self):
        def without(rows):
            rows[:] = [row for row in rows if (row["part"], row["kind"]) != (MEASURED, "tail")]
        self.assertIn("no measured copy of the kind tail", self.broken("damage", without))

    def test_the_written_files_have_to_match_the_recount(self):
        results_file = self.root / "results.json"
        results_file.write_text(results_file.read_text().replace('"copies": 60', '"copies": 61', 1))
        with self.assertRaises(verify.Malformed) as caught:
            verify.verify(self.root)
        self.assertIn("results.json differs", str(caught.exception))
        verify.verify(self.root, write=True)
        page_file = self.root / "RESULTS.md"
        page_file.write_text(page_file.read_text().replace("holds", "holds well", 1))
        with self.assertRaises(verify.Malformed) as caught:
            verify.verify(self.root)
        self.assertIn("RESULTS.md differs", str(caught.exception))

    def test_a_changed_data_file_is_an_error(self):
        rows = json.loads(json.dumps(self.data["copies"]))
        rows[0]["compared"] = 61
        write(self.root, {"copies": rows})
        with self.assertRaises(verify.Malformed) as caught:
            verify.verify(self.root)
        self.assertIn("not the one that files.sha256 names", str(caught.exception))

    def test_a_missing_data_file_is_an_error(self):
        (self.root / "data" / "trims.jsonl.gz").unlink()
        with self.assertRaises(verify.Malformed) as caught:
            verify.verify(self.root)
        self.assertIn("missing data file", str(caught.exception))


if __name__ == "__main__":
    unittest.main()
