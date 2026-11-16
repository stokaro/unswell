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


def copy(part, pages, number, kind):
    """A copy of a page: every damaged sentence is a suspect but the first removed one, and no harmless copy has a suspect."""
    row = {"part": part, "pages": pages, "unit": f"u{number}", "kind": kind, "compared": 60, "suspects": 0}
    if kind in verify.DAMAGE:
        row["named"] = (number, kind) != (0, "sentence")
        row["suspects"] = int(row["named"])
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


def revision(part, reviser, number):
    """A revision: the first is accepted, the second has a suspect that the judge cleared, the third loses a sentence."""
    parts = {0: checked(named=("page", "page")), 1: checked(suspects=2), 2: checked(named=("page", "page"), suspects=1, lost=1),
             3: checked(kept=48)}.get(number, checked())
    not_kept = {"changed": parts["statements"] - parts["kept"], "missing": 0, "unquoted": 0}
    return {"part": part, "pages": "set-s", "unit": f"u{number}", "reviser": reviser, "judge": verify.CHECK_JUDGE[reviser], **parts,
            "not_kept": not_kept, "lost_by": {"missing": 0, "negation": 0, "reworded": parts["lost"]},
            "em_dashes": [2, 3] if number == 4 else [2, 2]}


def record():
    copies = [copy(part, pages, number, kind) for part, groups in ((FITTING, ("set-s",)), (MEASURED, verify.PAGES))
              for pages in groups for number in range(10) for kind in verify.DAMAGE + verify.HARMLESS]
    damage = [damaged(part, pages, number, kind, judge) for part, groups in ((FITTING, ("set-s",)), (MEASURED, verify.PAGES))
              for pages in groups for number in range(10) for kind in verify.DAMAGE for judge in verify.JUDGES]
    revisions = [revision(part, reviser, number) for part in verify.REVISION_PARTS for reviser in verify.JUDGES for number in range(10)]
    return {"copies": copies, "damage": damage, "revisions": revisions}


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
        self.assertEqual((measured["sentence"]["named"], measured["swap"]["named"], measured["damaged"]["named_share"]), (9, 10, 0.98))
        self.assertEqual((measured["harmless"]["copies"], measured["harmless"]["with_a_suspect"]), (20, 0))
        cell = result["damage"][MEASURED]["fresh"][QWEN35]
        self.assertEqual((cell["copies"], cell["verdicts"], cell["reported"], cell["by_statements"], cell["by_either"]), (50, 50, 49, 1, 50))
        self.assertEqual((cell["first_verdicts"], cell["first_verdict_share"]), (50, 1.0))
        self.assertEqual((cell["sentence"]["reported"], cell["clause"]["reported"], cell["accepted"], cell["reported_share"]), (9, 10, 0, 0.98))
        self.assertEqual(cell["sound"], {"better": 0, "no difference": 50, "worse": 0})
        # Each kind of damage is read over the copies of both kinds of pages.
        self.assertEqual(result["kinds"][GPT]["sentence"], {"copies": 20, "reported": 18, "reported_share": 0.9})
        self.assertEqual(result["kinds"][GPT]["addnot"], {"copies": 20, "reported": 20, "reported_share": 1.0})
        mine = result["revisions"]["read"][GPT]
        self.assertEqual((mine["judge"], mine["pairs"], mine["accepted"], mine["kept"], mine["statements"]), (QWEN35, 10, 1, 498, 500))
        self.assertEqual((mine["kept_share"], mine["under_the_bar"], mine["not_kept"]["changed"]), (0.996, 1, 2))
        self.assertEqual((mine["with_a_suspect"], mine["suspects"], mine["cleared"], mine["with_a_lost_sentence"]), (2, 3, 2, 1))
        self.assertEqual((mine["lost"], mine["rejected_by_sentences_alone"], mine["add_em_dash"]),
                         ({"missing": 0, "negation": 0, "reworded": 1}, 1, 1))
        held = result["registered"]
        # H: 2. D1: 4. D2: one for each of five kinds and two judges. D3: 4. V: 4 and 4. R: 4.
        self.assertEqual((len(held), sum(held.values())), (32, 32))
        self.assertIn("32 of 32 registered bars hold.", (self.root / "RESULTS.md").read_text())

    def test_each_bar_can_fail(self):
        def missed(rows):
            for row in rows:
                if (row["part"], row["judge"], row["kind"]) == (MEASURED, GPT, "swap") and row["unit"] in ("u1", "u2"):
                    row.update(suspects=0, lost=0, reported=False)
        result = self.changed("damage", missed)
        # Four of twenty copies with a word turned around go unreported: that kind fails, and no other bar does.
        self.assertFalse(result["registered"][f"D2, swap reported, {GPT}"])
        self.assertTrue(result["registered"][f"D2, swap reported, {QWEN35}"])
        self.assertTrue(result["registered"][f"D2, clause reported, {GPT}"])
        self.assertTrue(result["registered"][f"D1, damage reported, set-s, {GPT}"])

        def many_missed(rows):
            for row in rows:
                if (row["part"], row["pages"], row["judge"]) == (MEASURED, "set-s", GPT) and row["unit"] in ("u1", "u2"):
                    row.update(suspects=0, lost=0, reported=False)
        result = self.changed("damage", many_missed)
        self.assertFalse(result["registered"][f"D1, damage reported, set-s, {GPT}"])
        self.assertTrue(result["registered"][f"D1, damage reported, fresh, {GPT}"])

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

    def test_a_retried_pair_counts_after_its_retry_and_is_shown(self):
        def retried(rows):
            for row in rows:
                if (row["part"], row["pages"], row["judge"]) == (MEASURED, "set-s", GPT) and row["unit"] in ("u1", "u2", "u3"):
                    row["retried"] = True
        result = self.changed("damage", retried)
        cell = result["damage"][MEASURED]["set-s"][GPT]
        # Fifteen copies needed the retry and got a verdict from it: the bar is read after it, and the first attempt is shown.
        self.assertEqual((cell["copies"], cell["verdicts"], cell["first_verdicts"], cell["first_verdict_share"]), (50, 50, 35, 0.7))
        self.assertTrue(result["registered"][f"V, verdicts on damaged copies, set-s, {GPT}"])
        self.assertIn("On the first attempt the shares were posts of the style-guide set, `gpt-6-astra`, 70.0%", (self.root / "RESULTS.md").read_text())

        def again(rows):
            for row in rows:
                if (row["part"], row["reviser"]) == ("read", GPT) and row["unit"] == "u5":
                    row["retried"] = True
        result = self.changed("revisions", again)
        self.assertEqual((result["revisions"]["read"][GPT]["verdicts"], result["revisions"]["read"][GPT]["first_verdicts"]), (10, 9))

        def suspected(rows):
            for row in rows:
                if (row["part"], row["pages"], row["kind"]) == (MEASURED, "fresh", "join") and row["unit"] in ("u1", "u2"):
                    row["suspects"] = 1
        result = self.changed("copies", suspected)
        self.assertFalse(result["registered"]["H, harmless copies, fresh"])

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
        self.assertIn("does not say whether", self.broken("copies", lambda rows: rows[0].pop("named")))
        self.assertIn("a suspect among none", self.broken("copies", lambda rows: rows[0].update(named=True)))

        def without(rows):
            rows[:] = [row for row in rows if (row["part"], row["pages"]) != (MEASURED, "fresh")]
        self.assertIn("are not complete", self.broken("copies", without))

    def test_a_malformed_damaged_copy_is_rejected(self):
        self.assertIn("damaged copy listed twice", self.broken("damage", lambda rows: rows.append(json.loads(json.dumps(rows[0])))))
        self.assertIn("unknown judge or verdict", self.broken("damage", lambda rows: rows[0].update(judge="someone")))
        self.assertIn("unknown judge or verdict", self.broken("damage", lambda rows: rows[0].update(verdict="maybe")))
        self.assertIn("unknown part", self.broken("damage", lambda rows: rows[0].update(kind="join")))
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

    def test_a_malformed_revision_is_rejected(self):
        self.assertIn("revision listed twice", self.broken("revisions", lambda rows: rows.append(json.loads(json.dumps(rows[0])))))
        self.assertIn("unknown part or reviser", self.broken("revisions", lambda rows: rows[0].update(part="later")))
        self.assertIn("unknown part or reviser", self.broken("revisions", lambda rows: rows[0].update(reviser="someone")))
        self.assertIn("not of the other family", self.broken("revisions", lambda rows: rows[0].update(judge=rows[0]["reviser"])))
        self.assertIn("statements do not add up", self.broken("revisions", lambda rows: rows[3]["not_kept"].update(changed=1)))
        self.assertIn("statements do not add up", self.broken("revisions", lambda rows: rows[0].pop("not_kept")))
        self.assertIn("lost sentences do not add up", self.broken("revisions", lambda rows: rows[2]["lost_by"].update(negation=1)))
        self.assertIn("lost sentences do not add up", self.broken("revisions", lambda rows: rows[2]["lost_by"].pop("reworded")))
        self.assertIn("verdict does not follow", self.broken("revisions", lambda rows: rows[2].update(verdict="accepted")))
        self.assertIn("verdict does not follow", self.broken("revisions", lambda rows: rows[3].update(verdict="accepted")))
        self.assertIn("without the em dashes", self.broken("revisions", lambda rows: rows[0].update(em_dashes=[1])))
        self.assertIn("with an em dash added", self.broken("revisions", lambda rows: rows[0].update(em_dashes=[0, 1])))

        def without(rows):
            rows[:] = [row for row in rows if (row["part"], row["reviser"]) != ("fresh", QWEN35)]
        self.assertIn("are not complete", self.broken("revisions", without))

    def test_every_kind_of_damage_has_to_be_measured(self):
        def without(rows):
            rows[:] = [row for row in rows if (row["part"], row["kind"]) != (MEASURED, "swap")]
        self.assertIn("no measured copy of the kind swap", self.broken("damage", without))

    def test_the_written_files_have_to_match_the_recount(self):
        results_file = self.root / "results.json"
        results_file.write_text(results_file.read_text().replace('"copies": 50', '"copies": 51', 1))
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
        (self.root / "data" / "revisions.jsonl.gz").unlink()
        with self.assertRaises(verify.Malformed) as caught:
            verify.verify(self.root)
        self.assertIn("missing data file", str(caught.exception))


if __name__ == "__main__":
    unittest.main()
