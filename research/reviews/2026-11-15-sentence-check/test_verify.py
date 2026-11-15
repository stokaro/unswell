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
FIRST = "fitting, first copies"


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
            "not_kept": not_kept, "lost_by": {"missing": parts["lost"], "negation": 0}}


def record():
    copies = [copy(part, pages, number, kind) for part, groups in ((FIRST, ("set-s",)), ("fitting, second copies", ("set-s",)),
                                                                   (MEASURED, verify.PAGES))
              for pages in groups for number in range(10) for kind in verify.DAMAGE + verify.HARMLESS]
    damage = [damaged(part, pages, number, kind, judge) for part, groups in ((FIRST, ("set-s",)), (MEASURED, verify.PAGES))
              for pages in groups for number in range(10) for kind in verify.DAMAGE for judge in verify.JUDGES]
    revisions = [revision(part, reviser, number) for part in verify.REVISION_PARTS for reviser in verify.JUDGES for number in range(10)]
    variants = []
    for name, caught in (("all words", 3), ("own words", 5)):
        variants += [{"variant": name, "kind": kind, "unit": f"u{number}", "suspects": int(number < caught), "named": number < caught}
                     for kind in verify.DAMAGE for number in range(5)]
        variants += [{"variant": name, "kind": "revision", "reviser": reviser, "unit": f"u{number}", "suspects": 2 * (number == 0)}
                     for reviser in verify.JUDGES for number in range(5)]
    return {"copies": copies, "damage": damage, "revisions": revisions, "variants": variants}


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
        self.assertEqual((measured["sentence"]["named"], measured["negation"]["named"], measured["damaged"]["named_share"]), (9, 10, 0.95))
        self.assertEqual((measured["harmless"]["copies"], measured["harmless"]["with_a_suspect"]), (20, 0))
        cell = result["damage"][MEASURED]["fresh"][QWEN35]
        self.assertEqual((cell["copies"], cell["verdicts"], cell["reported"], cell["by_statements"], cell["by_either"]), (20, 20, 19, 1, 20))
        self.assertEqual((cell["sentence"]["reported"], cell["negation"]["reported"], cell["accepted"], cell["reported_share"]), (9, 10, 0, 0.95))
        self.assertEqual(cell["sound"], {"better": 0, "no difference": 20, "worse": 0})
        mine = result["revisions"]["read"][GPT]
        self.assertEqual((mine["judge"], mine["pairs"], mine["accepted"], mine["kept"], mine["statements"]), (QWEN35, 10, 1, 498, 500))
        self.assertEqual((mine["kept_share"], mine["under_the_bar"], mine["not_kept"]["changed"]), (0.996, 1, 2))
        self.assertEqual((mine["with_a_suspect"], mine["suspects"], mine["cleared"], mine["with_a_lost_sentence"]), (2, 3, 2, 1))
        self.assertEqual((mine["lost"], mine["rejected_by_sentences_alone"]), ({"missing": 1, "negation": 0}, 1))
        first, second = result["variants"]
        self.assertEqual((first["variant"], first["damaged"]["sentence"]["named"], second["damaged"]["negation"]["named"]), ("all words", 3, 5))
        self.assertEqual(first["revisions"][GPT], {"revisions": 5, "with_a_suspect": 1, "suspects": 2})
        held = result["registered"]
        self.assertEqual((len(held), sum(held.values())), (22, 22))
        self.assertIn("22 of 22 registered bars hold.", (self.root / "RESULTS.md").read_text())

    def test_each_bar_can_fail(self):
        def missed(rows):
            for row in rows:
                if (row["part"], row["pages"], row["judge"], row["kind"]) == (MEASURED, "set-s", GPT, "negation") and row["unit"] in ("u1", "u2"):
                    row.update(suspects=0, lost=0, reported=False)
        result = self.changed("damage", missed)
        self.assertFalse(result["registered"][f"D1, damage reported, set-s, {GPT}"])
        self.assertTrue(result["registered"][f"D1, damage reported, fresh, {GPT}"])

        def accepted(rows):
            for row in rows:
                if (row["part"], row["pages"], row["judge"]) == (MEASURED, "fresh", QWEN35) and row["unit"] in ("u1", "u2"):
                    row.update(checked(named=("page", "page")), reported=False)
        result = self.changed("damage", accepted)
        self.assertFalse(result["registered"][f"D2, damaged copies accepted, fresh, {QWEN35}"])

        def silent(rows):
            for row in rows:
                if (row["part"], row["pages"], row["judge"]) == (MEASURED, "set-s", QWEN35) and row["unit"] in ("u3", "u4"):
                    for name in ("sound", "named", "statements", "kept", "suspects", "lost", "reported"):
                        row.pop(name)
                    row["verdict"] = "no verdict"
        result = self.changed("damage", silent)
        self.assertFalse(result["registered"][f"V, verdicts on damaged copies, set-s, {QWEN35}"])

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
        self.assertIn("verdict does not follow", self.broken("revisions", lambda rows: rows[2].update(verdict="accepted")))
        self.assertIn("verdict does not follow", self.broken("revisions", lambda rows: rows[3].update(verdict="accepted")))

        def without(rows):
            rows[:] = [row for row in rows if (row["part"], row["reviser"]) != ("fresh", QWEN35)]
        self.assertIn("are not complete", self.broken("revisions", without))

    def test_a_malformed_variant_is_rejected(self):
        self.assertIn("variant row listed twice", self.broken("variants", lambda rows: rows.append(json.loads(json.dumps(rows[0])))))
        self.assertIn("malformed variant row", self.broken("variants", lambda rows: rows[0].update(kind="join")))
        self.assertIn("does not say whether", self.broken("variants", lambda rows: rows[0].pop("named")))
        revision_row = lambda rows: next(row for row in rows if row["kind"] == "revision")
        self.assertIn("unknown reviser", self.broken("variants", lambda rows: revision_row(rows).update(reviser="someone")))

    def test_the_written_files_have_to_match_the_recount(self):
        results_file = self.root / "results.json"
        results_file.write_text(results_file.read_text().replace('"copies": 20', '"copies": 21', 1))
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
        (self.root / "data" / "variants.jsonl.gz").unlink()
        with self.assertRaises(verify.Malformed) as caught:
            verify.verify(self.root)
        self.assertIn("missing data file", str(caught.exception))


if __name__ == "__main__":
    unittest.main()
