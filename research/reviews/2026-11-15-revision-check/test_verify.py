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

GPT, QWEN35, QWEN27 = verify.GPT, verify.QWEN35, verify.QWEN27


def answers(unit, votes, **labels):
    """The answers of the three judges on the three pairs of one post; votes are those of the judges in order."""
    return [{**labels, "unit": unit, "against": f"h{pair}", "judge": judge, "vote": vote}
            for pair in range(3) for judge, vote in zip(verify.JUDGES, votes)]


def revision(reviser, number, named=("revision", "page"), kept=50, dashes=(0, 0), third="page", kind="set-s", lost=0):
    """A revised page with the verdict that follows from its parts."""
    check = {"judge": verify.CHECK_JUDGE[reviser], "sound": verify.sound_of(list(named)), "named": list(named), "statements": 50, "kept": kept,
             "not_kept": {"changed": 50 - kept, "missing": 0, "unquoted": 0}, "lost": {"numbers": lost, "code": 0, "links": 0}}
    check["verdict"] = "accepted" if verify.accepted(check, list(dashes)) else "rejected"
    row = {"reviser": reviser, "set": kind, "id": f"s-{number:02d}" if kind == "set-s" else f"g{number}", "outcome": "revised",
           "index": [80, 30], "em_dashes": list(dashes), "third": third, "check": check}
    if kind == "guided":
        row["writer"] = GPT
    return row


def record():
    """A small record with every kind of row."""
    published, revisions, fitting = [], [], []
    for number in range(6):
        unit = f"s-{number:02d}"
        for human in verify.HUMAN:
            # Before a revision the strict judge picks every post and the panel four of six.
            published += answers(unit, ("post", "post" if number < 4 else "human", "human"), reviser="none", human=human)
            # After a revision by the first reviser the panel picks three of six, and after the second two of six.
            published += answers(unit, ("post" if number < 3 else "human", "post" if number < 3 else "human", "human"), reviser=GPT, human=human)
            published += answers(unit, ("post" if number < 2 else "human", "human", "human"), reviser=QWEN35, human=human)
        # Without the check every revision is published: the strict judge still picks five of six.
        published += answers(unit, ("post" if number < 5 else "human", "post" if number < 5 else "human", "human"), reviser=GPT, human="blog",
                             kept="every revision")
    # The first reviser: twelve revisions, eleven of them accepted; the third model names the page on ten of the accepted.
    for number in range(11):
        revisions.append(revision(GPT, number, named=("page", "page"), third="page" if number < 10 else "revision"))
    revisions.append(revision(GPT, 11, third="revision"))
    revisions.append({"reviser": GPT, "set": "set-s", "id": "s-90", "outcome": "no mark listed"})
    # The second reviser: one accepted, others rejected for each reason, one check without a verdict.
    revisions.append(revision(QWEN35, 0, named=("page", "page")))
    revisions.append(revision(QWEN35, 1, named=("revision", "revision"), third="revision"))
    revisions.append(revision(QWEN35, 2, named=("page", "page"), kept=48, third="same"))
    revisions.append(revision(QWEN35, 3, named=("page", "page"), lost=1, third="revision", kind="guided"))
    revisions.append(revision(QWEN35, 4, named=("page", "page"), dashes=(0, 2), third="revision", kind="guided"))
    revisions.append({"reviser": QWEN35, "set": "set-s", "id": "s-91", "outcome": "no revision"})
    damage = [{"unit": f"s-{number:02d}", "kind": kind, "judge": QWEN35, "verdict": "rejected", "statements": 40,
               "not_kept": 0 if (number, kind) == (0, "sentence") else 1, "sound": "no difference", "named": ["revision", "page"]}
              for number in range(5) for kind in ("sentence", "negation")]
    known = [{"id": f"k{number}", "sound": "better" if number < 4 else "no difference", "named": ["page", "page"] if number < 4 else ["revision", "page"]}
             for number in range(5)]
    for number in range(4):
        unit = f"s-{number + 40:02d}"
        fitting += answers(unit, ("post", "post", "human"), experiment="against human text", variant="as it is", reviser="none", human="blog")
        for judge, vote in zip(verify.JUDGES, ("page", "revision", "same")):
            fitting.append({"experiment": "page against revision", "variant": "report", "reviser": GPT, "unit": unit, "judge": judge,
                            "vote": "page" if number == 0 else vote})
        fitting.append({"experiment": "revision", "variant": "report", "reviser": GPT, "unit": unit, "index": [70 + number, 30], "em_dashes": [0, number % 2]})
        fitting.append({"experiment": "both orders", "pairs": "revisions", "wording": "wording", "judge": QWEN35, "unit": unit,
                        "named": ["revision", "page"] if number else ["page", "page"]})
        fitting.append({"experiment": "sections", "pairs": "revisions", "judge": QWEN35, "unit": unit, "votes": [1, 0, -1] if number else [1, 0]})
        fitting.append({"experiment": "damage", "unit": unit, "kind": "sentence", "judge": GPT, "verdict": "rejected", "not_kept": 1})
    return {"published": published, "revisions": revisions, "damage": damage, "known": known, "fitting": fitting}


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

    def damaged(self, name, change):
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
        before, after = result["published"]["none"]["blog"], result["published"][GPT]["blog"]
        self.assertEqual((before["panel"]["post"], before["judges"][GPT]["post"], before["pairs"]), (4, 6, 18))
        self.assertEqual((after["panel"]["post"], after["panel"]["rate"], after["judges"][GPT]["post"]), (3, 0.5, 3))
        # One judge of three picks the post on two of six: the panel follows the two that pick the human text.
        self.assertEqual(result["published"][QWEN35]["styled"]["panel"]["post"], 0)
        # The rows without the check are counted apart and do not change the published ones.
        self.assertEqual(result["published"]["every revision"][GPT]["panel"]["post"], 5)
        self.assertNotIn(QWEN35, result["published"]["every revision"])
        first, second = result["revisions"][GPT], result["revisions"][QWEN35]
        self.assertEqual((first["pages"], first["revised"], first["no_mark_listed"], first["check"]["accepted"]), (13, 12, 1, 11))
        self.assertEqual(first["check"]["second_text_both_times"], 1)
        self.assertEqual(first["set_s"], {"revised": 12, "index_lower": 12, "median_index": [80, 30], "high": [12, 0]})
        self.assertEqual(second["set_s"]["revised"], 3)
        self.assertEqual((second["check"]["accepted"], second["check"]["rejected"], second["no_revision"]), (1, 4, 1))
        self.assertEqual((second["check"]["under_the_bar"], second["check"]["lost_something"], second["add_em_dash"]), (1, 1, 1))
        self.assertEqual(second["check"]["not_kept"], {"changed": 2, "missing": 0, "unquoted": 0})
        self.assertEqual(result["revisions"]["together"]["third_on_accepted"]["page"], 11)
        damage = result["damage"][QWEN35]
        self.assertEqual(list(result["damage"]), [QWEN35])
        self.assertEqual((damage["noticed"], damage["sentence"]["noticed"], damage["noticed_share"]), (9, 4, 0.9))
        self.assertEqual(damage["sound"], {"better": 0, "no difference": 10, "worse": 0})
        self.assertEqual(damage["better"], {"sentence": 0, "negation": 0})
        self.assertEqual((result["known"]["better"], result["known"]["orders"]["second text both times"]), (4, 1))
        fitting = result["fitting"]
        self.assertEqual(fitting["page_against_revision"][0]["panel"], {"page": 1, "revision": 0, "tie": 3})
        self.assertEqual(fitting["revisions"][0]["add_em_dash"], 2)
        self.assertEqual((fitting["sections"][0]["sections"], fitting["sections"][0]["better"], fitting["sections"][0]["no_difference"]), (11, 1, 3))
        self.assertEqual(fitting["both_orders"][0]["second text both times"], 3)

    def test_each_bar_is_read(self):
        held = verify.verify(self.root)["registered"]
        self.assertTrue(held[f"A1, panel, {GPT}, blog"])
        self.assertTrue(held[f"A2, {GPT} alone, {GPT}, blog"])
        self.assertTrue(held[f"B1, verdicts, {GPT}"])
        self.assertTrue(held["B2, accepted revisions are better"])
        self.assertTrue(held["B3, rejected revisions are less often better"])
        # One revision of five of the second reviser adds an em dash: more than the bar allows.
        self.assertFalse(held[f"C1, em dashes, {QWEN35}"])
        self.assertTrue(held[f"C1, em dashes, {GPT}"])
        self.assertTrue(held[f"C2, statements kept, {QWEN35}"])
        self.assertTrue(held["C3, damage is noticed"])
        self.assertTrue(held["D, a known difference is found"])

    def test_the_strict_judge_can_fail_the_bar_that_the_panel_passes(self):
        def still_told(rows):
            for row in rows:
                if row["reviser"] == GPT and row["judge"] == GPT:
                    row["vote"] = "post"
        result = self.changed("published", still_told)
        # The strict judge picks every post; the other two pick the human text, so the panel picks none.
        self.assertTrue(result["registered"][f"A1, panel, {GPT}, blog"])
        self.assertFalse(result["registered"][f"A2, {GPT} alone, {GPT}, blog"])
        self.assertIn("**does not hold**", (self.root / "RESULTS.md").read_text())

    def test_a_bar_on_accepted_revisions_needs_enough_of_them(self):
        def fewer(rows):
            rows[:] = [row for row in rows if not (row["reviser"] == GPT and row.get("check", {}).get("verdict") == "accepted" and row["id"] > "s-03")]
        result = self.changed("revisions", fewer)
        self.assertEqual(result["revisions"]["together"]["accepted"], 5)
        self.assertIsNone(result["registered"]["B2, accepted revisions are better"])
        self.assertIsNone(result["registered"]["B3, rejected revisions are less often better"])
        self.assertIn("does not apply", (self.root / "RESULTS.md").read_text())

    def test_other_bars_can_fail(self):
        result = self.changed("damage", lambda rows: [row.update(not_kept=0) for row in rows[:4]])
        self.assertFalse(result["registered"]["C3, damage is noticed"])
        result = self.changed("known", lambda rows: [row.update(sound="no difference", named=["revision", "page"]) for row in rows[:3]])
        self.assertFalse(result["registered"]["D, a known difference is found"])

    def test_a_second_judge_of_damage_has_no_bar(self):
        def second(rows):
            for row in rows[:4]:
                row.update(not_kept=0)
            rows.extend({**row, "judge": GPT, "not_kept": 1} for row in json.loads(json.dumps(rows)))
        result = self.changed("damage", second)
        self.assertEqual(list(result["damage"]), [GPT, QWEN35])
        self.assertEqual((result["damage"][GPT]["noticed"], result["damage"][QWEN35]["noticed"]), (10, 6))
        self.assertFalse(result["registered"]["C3, damage is noticed"])
        self.assertIn("was added after the bars were written", (self.root / "RESULTS.md").read_text())

        def no_verdicts(rows):
            for row in rows:
                if row["reviser"] == GPT and "check" in row and row["id"] < "s-03":
                    row["check"] = {"judge": row["check"]["judge"], "verdict": "no verdict"}
        result = self.changed("revisions", no_verdicts)
        self.assertFalse(result["registered"][f"B1, verdicts, {GPT}"])

    def test_a_verdict_has_to_follow_from_its_parts(self):
        accept = lambda rows: next(r for r in rows if r["reviser"] == QWEN35 and r["id"] == "s-02")["check"].update(verdict="accepted")
        self.assertIn("verdict does not follow", self.damaged("revisions", accept))
        better = lambda rows: next(r for r in rows if r["reviser"] == GPT and r["id"] == "s-11")["check"].update(sound="better")
        self.assertIn("sound does not follow", self.damaged("revisions", better))
        judge = lambda rows: rows[0]["check"].update(judge=GPT)
        self.assertIn("judge is not of the other family", self.damaged("revisions", judge))

    def test_a_revision_has_to_be_well_formed(self):
        self.assertIn("listed twice", self.damaged("revisions", lambda rows: rows.append(json.loads(json.dumps(rows[0])))))
        self.assertIn("known reviser or set", self.damaged("revisions", lambda rows: rows[0].update(reviser="someone")))
        self.assertIn("unknown outcome", self.damaged("revisions", lambda rows: rows[0].update(outcome="improved")))
        self.assertIn("more than its outcome", self.damaged("revisions", lambda rows: rows[12].update(index=[1, 1])))
        self.assertIn("before and after", self.damaged("revisions", lambda rows: rows[0].update(em_dashes=[0])))
        self.assertIn("third model", self.damaged("revisions", lambda rows: rows[0].pop("third")))
        self.assertIn("without its statements", self.damaged("revisions", lambda rows: rows[0]["check"].update(kept=51)))
        self.assertIn("do not add up", self.damaged("revisions", lambda rows: rows[0]["check"]["not_kept"].update(unquoted=1)))
        self.assertIn("what was lost", self.damaged("revisions", lambda rows: rows[0]["check"]["lost"].pop("code")))
        self.assertIn("named in each order", self.damaged("revisions", lambda rows: rows[0]["check"].update(named=["page"])))

    def test_every_pair_needs_one_answer_of_each_judge(self):
        self.assertIn("one answer of each judge", self.damaged("published", lambda rows: rows.pop(0)))
        self.assertIn("malformed answer", self.damaged("published", lambda rows: rows[0].update(vote="maybe")))
        self.assertIn("known reviser or kind", self.damaged("published", lambda rows: rows[0].update(human="forum")))
        self.assertIn("known reviser or kind", self.damaged("published", lambda rows: rows[0].update(kept="some revisions")))

        def without(rows):
            rows[:] = [row for row in rows if not (row["reviser"] == QWEN35 and row["human"] == "styled")]
        self.assertIn("no published answers", self.damaged("published", without))

    def test_the_other_files_have_to_be_well_formed(self):
        self.assertIn("malformed damaged copy", self.damaged("damage", lambda rows: rows[0].update(kind="paragraph")))
        self.assertIn("without its statements", self.damaged("damage", lambda rows: rows[0].pop("not_kept")))
        self.assertIn("without a known judge", self.damaged("damage", lambda rows: rows[0].update(judge="someone")))
        self.assertIn("damaged copy listed twice", self.damaged("damage", lambda rows: rows.append(json.loads(json.dumps(rows[0])))))
        self.assertIn("judge of the bar", self.damaged("damage", lambda rows: [row.update(judge=GPT) for row in rows]))
        self.assertIn("sound does not follow", self.damaged("damage", lambda rows: rows[0].update(sound="better")))
        self.assertIn("verdict does not follow", self.damaged("damage", lambda rows: rows[0].update(verdict="accepted")))
        missed = lambda rows: rows[1].update(verdict="accepted", sound="better", named=["page", "page"])
        self.assertIn("verdict does not follow", self.damaged("damage", missed))
        self.assertIn("unknown sound", self.damaged("known", lambda rows: rows[0].update(sound="louder")))
        self.assertIn("does not follow from the two answers", self.damaged("known", lambda rows: rows[0].update(named=["revision", "page"])))
        self.assertIn("unknown experiment", self.damaged("fitting", lambda rows: rows[0].update(experiment="guess")))
        versus = lambda rows: next(r for r in rows if r["experiment"] == "page against revision").update(vote="post")
        self.assertIn("page and its revision", self.damaged("fitting", versus))
        orders = lambda rows: next(r for r in rows if r["experiment"] == "both orders").update(named=["page", "sure"])
        self.assertIn("named in each order", self.damaged("fitting", orders))

    def test_an_unreadable_answer_is_no_vote(self):
        def unreadable(rows):
            for row in rows:
                if row["reviser"] == "none" and row["human"] == "blog" and row["judge"] == GPT:
                    row["vote"] = "unreadable"
        result = self.changed("published", unreadable)
        entry = result["published"]["none"]["blog"]
        self.assertEqual((entry["unreadable"], entry["judges"][GPT]["post"], entry["panel"]["tie"]), (18, 0, 4))

    def test_a_changed_data_file_is_noticed(self):
        rows = json.loads(json.dumps(self.data["known"]))
        rows[4].update(sound="better", named=["page", "page"])
        write(self.root, {"known": rows})
        with self.assertRaises(verify.Malformed) as caught:
            verify.verify(self.root)
        self.assertIn("files.sha256", str(caught.exception))

    def test_results_and_the_page_cannot_be_edited_by_hand(self):
        results = self.root / "results.json"
        results.write_text(results.read_text().replace('"panel": 0.55', '"panel": 0.95'))
        with self.assertRaises(verify.Malformed) as caught:
            verify.verify(self.root)
        self.assertIn("results.json differs", str(caught.exception))
        verify.verify(self.root, write=True)
        page_file = self.root / "RESULTS.md"
        page_file.write_text(page_file.read_text().replace("holds", "holds well", 1))
        with self.assertRaises(verify.Malformed) as caught:
            verify.verify(self.root)
        self.assertIn("RESULTS.md differs", str(caught.exception))

    def test_a_missing_data_file_is_an_error(self):
        (self.root / "data" / "damage.jsonl.gz").unlink()
        with self.assertRaises(verify.Malformed) as caught:
            verify.verify(self.root)
        self.assertIn("missing data file", str(caught.exception))

    def test_small_measures(self):
        self.assertEqual(verify.sound_of(["page", "page"]), "better")
        self.assertEqual(verify.sound_of(["revision", "revision"]), "worse")
        self.assertEqual(verify.sound_of(["revision", "page"]), "no difference")
        self.assertEqual(verify.sound_of(["page", "same"]), "no difference")
        self.assertEqual(verify.median([4, 1, 2, 3]), 2.5)
        self.assertEqual(verify.wilson(8, 10), [0.49, 0.943])


if __name__ == "__main__":
    unittest.main()
