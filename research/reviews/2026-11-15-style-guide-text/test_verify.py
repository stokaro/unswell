"""Tests of the recount: it accepts a well-formed record and rejects each kind of damage."""
import gzip
import json
import math
import shutil
import sys
import tempfile
import unittest
from pathlib import Path

HERE = Path(__file__).resolve().parent
sys.path.insert(0, str(HERE))
import verify  # noqa: E402

GPT, QWEN = verify.WRITERS


def page(kind, half, identifier, even=0, openers=0, dashes=0, words=400, writer=None):
    """A page row whose readings follow from its counts."""
    signals = {name: [0, 0] for name in list(verify.MARKS) + list(verify.RHYTHM)}
    signals["em-dash"] = [dashes, 0]
    signals["plain-sentence-run"] = [1 if even else 0, even]
    signals["subject-opener-run"] = [1 if openers else 0, openers]
    marks = verify.index_of(verify.MARKS_BIAS + verify.MARKS["em-dash"] * math.log1p(1000 * dashes / words))
    rhythm = verify.index_of(verify.RHYTHM_BIAS + sum(
        weight * math.log1p(1000 * signals[name][1] / words) for name, weight in verify.RHYTHM.items()))
    row = {"set": kind, "half": half, "id": identifier, "index": max(marks, rhythm), "marks": marks, "rhythm": rhythm, "earlier": marks}
    if kind == "set-s":
        return row
    row.update(words=words, signals=signals)
    if writer:
        row.update(writer=writer, guide="A")
    return row


def control_page(kind, identifier, even=0, paragraphs=0.9, sentences=None):
    """A page of the control whose readings follow from its counts."""
    counted = page("human-blog", "measured", identifier, even=even)
    row = {name: counted[name] for name in ("id", "index", "marks", "rhythm", "words", "signals")}
    # Twelve sentences in two stretches: eight plain ones, four of them in runs of two.
    default = [[1, 12, 1], [1, 12, 1], [1, 30, 0], [1, 12, 1], [1, 30, 0], [1, 12, 1]] * 2
    for at in range(6, 12):
        default[at] = [2] + default[at][1:]
    row.update(kind=kind, repo="example/" + kind, path=f"content/{identifier}.md", commit="c" * 40, guide="a style guide",
               paragraphs=paragraphs, sentences=default if sentences is None else sentences)
    return row


def votes(comparison, unit, picked, **more):
    """One answer of each judge on one pair."""
    return [{"comparison": comparison, "unit": unit, "judge": judge, "vote": vote, **more} for judge, vote in zip(verify.JUDGES, picked)]


def record():
    """A small record with every kind of row."""
    pages, judgments, revisions = [], [], []
    for half in verify.HALVES:
        # A page is in one half: the fitting half has its own ids.
        mark, offset = ("", 0) if half == "measured" else ("f", 10)
        for number in range(6):
            pages.append(page("human-blog", half, f"{mark}h{number}", even=2 * number))
            pages.append(page("readme", half, f"{mark}r{number}", dashes=number // 3))
            pages.append(page("set-s", half, f"s-{number + offset:02d}", even=20 + number))
            for writer in verify.WRITERS:
                pages.append(page("guided", half, f"{mark}h{number}", even=30 + number, openers=10, writer=writer))
                pages.append(page("plain", half, f"{mark}h{number}", even=18 + number, writer=writer))
    for number in range(6):
        for writer in verify.WRITERS:
            judgments += votes("plain", f"h{number}", ("model", "model", "same"), writer=writer)
            judgments += votes("guided", f"h{number}", ("model", "model", "human"), writer=writer, guide="AB"[number % 2])
        judgments += votes("cross", f"h{number}", ("model", "human", "model"), writer=GPT, against=f"h{(number + 1) % 6}")
        for partner in range(3):
            judgments += votes("set", f"s-{number:02d}", ("model", "model" if number < 4 else "human", "human"), against=f"h{partner}")
        for kind, writer in (("set-s", None), ("guided", GPT), ("guided", QWEN)):
            identifier = f"s-{number:02d}" if kind == "set-s" else f"h{number}"
            revision = {"reviser": GPT, "set": kind, "id": identifier, "outcome": "revised", "before": [80, 23, 80], "after": [40 + number, 23, 40 + number],
                        "em_dashes_added": 1}
            extra = {"set": kind}
            if writer:
                revision.update(writer=writer, words=[400, 404])
                extra["writer"] = writer
            revisions.append(revision)
            judgments += votes("revision", identifier, ("page", "page", "revision" if number else "same"), reviser=GPT, **extra)
    revisions.append({"reviser": GPT, "set": "set-s", "id": "s-99", "outcome": "no mark listed"})
    earlier = [{"group": "human", "id": f"p{n}", "side": "page", "outside_main": True, "index": 23 + n, "marks": 23 + n, "rhythm": 5} for n in range(8)]
    for number in range(5):
        for side, value in (("before", 70 + number), ("after", 60)):
            earlier.append({"group": "reserve", "id": f"d{number}", "side": side, "outside_main": True, "index": value, "marks": value, "rhythm": 10})
    control = [control_page("articles", f"a{number}", even=2 * number) for number in range(6)]
    control += [control_page("articles", f"list{number}", paragraphs=0.4) for number in range(2)]
    control += [control_page("documentation", f"d{number}", even=number) for number in range(4)]
    return {"pages": pages, "judgments": judgments, "revisions": revisions, "earlier": earlier, "control": control}


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
        write(self.root, {name: rows})
        (self.root / "files.sha256").write_text(verify.digests(self.root))
        with self.assertRaises(verify.Malformed) as caught:
            verify.verify(self.root)
        return str(caught.exception)

    def test_a_well_formed_record_recounts(self):
        result = verify.verify(self.root)
        self.assertEqual(result["judges"]["plain"]["panel"]["rate"], 1.0)
        # One judge of three names the human text on every guided pair: the panel still picks the model's text.
        self.assertEqual(result["judges"]["guided"]["panel"]["model"], 12)
        # Four of the six posts of the set win the sum of their nine votes; on the other two the votes for the human text win.
        self.assertEqual((result["judges"]["set"]["panel"]["model"], result["judges"]["set"]["panel"]["other"]), (4, 2))
        self.assertEqual(result["judges"]["guided"]["guides"]["A"]["units"], 6)
        self.assertEqual(result["pages"]["measured"]["separation"]["set-s"], 1.0)
        self.assertEqual(result["earlier"]["pairs"], 5)
        self.assertEqual(result["revisions"][GPT]["groups"]["all"]["revised"], 18)
        self.assertEqual(result["revisions"][GPT]["no_mark_listed"], 1)
        self.assertTrue(all(result["registered"].values()), result["registered"])

    def test_a_tie_counts_against_the_rate(self):
        units = {"a": [(verify.JUDGES[0], 1), (verify.JUDGES[1], -1), (verify.JUDGES[2], 0)], "b": [(verify.JUDGES[0], 1)]}
        counted = verify.panel(units)
        self.assertEqual((counted["model"], counted["other"], counted["tie"], counted["rate"], counted["decisive"]), (1, 0, 1, 0.5, 1.0))
        self.assertEqual(verify.panel(units, (verify.JUDGES[1],))["other"], 1)

    def test_levels_and_measures(self):
        self.assertEqual([verify.level(value) for value in (0, 39, 40, 69, 70, 100)], ["low", "low", "elevated", "elevated", "high", "high"])
        self.assertEqual(verify.auc([3, 2], [2, 1]), 0.875)
        self.assertIsNone(verify.auc([], [1]))
        self.assertEqual(verify.wilson(8, 10), [0.49, 0.943])
        self.assertEqual(verify.index_of(verify.MARKS_BIAS), 23)

    def test_a_bar_that_fails_is_reported_and_not_an_error(self):
        def louder(rows):
            for row in rows:
                if row["set"] == "human-blog" and row["half"] == "measured":
                    row.update(page("human-blog", "measured", row["id"], even=40))
        rows = json.loads(json.dumps(self.data["pages"]))
        louder(rows)
        write(self.root, {"pages": rows})
        result = verify.verify(self.root, write=True)
        self.assertFalse(result["registered"]["quiet on human pages"])
        self.assertFalse(result["registered"]["separation, set-s"])
        self.assertIn("**does not hold**", (self.root / "RESULTS.md").read_text())

    def test_an_index_has_to_be_the_higher_reading(self):
        self.assertIn("higher of its two readings", self.damaged("pages", lambda rows: rows[0].update(index=rows[0]["index"] + 1)))

    def test_readings_have_to_follow_from_the_counts(self):
        def more(rows):
            row = next(r for r in rows if r["set"] == "guided")
            row["signals"]["plain-sentence-run"][1] += 9
        self.assertIn("do not follow from the counts", self.damaged("pages", more))

    def test_the_earlier_index_is_the_first_reading(self):
        def other(rows):
            row = next(r for r in rows if r["set"] == "readme")
            row["earlier"] += 1
        self.assertIn("earlier index", self.damaged("pages", other))

    def test_a_page_is_listed_once_and_in_a_known_set(self):
        self.assertIn("listed twice", self.damaged("pages", lambda rows: rows.append(dict(rows[0]))))
        self.assertIn("unknown set", self.damaged("pages", lambda rows: rows[0].update(set="forum")))
        self.assertIn("known writer", self.damaged("pages", lambda rows: next(r for r in rows if r["set"] == "plain").update(writer="someone")))

    def test_a_post_of_the_style_guide_set_carries_only_its_readings(self):
        def named(rows):
            next(r for r in rows if r["set"] == "set-s").update(words=900)
        self.assertIn("more than its readings", self.damaged("pages", named))

    def test_every_pair_needs_one_answer_of_each_judge(self):
        def fewer(rows):
            rows.remove(next(r for r in rows if r["comparison"] == "guided"))
        self.assertIn("one answer of each judge", self.damaged("judgments", fewer))
        self.assertIn("unknown judge", self.damaged("judgments", lambda rows: rows[0].update(judge="someone")))
        self.assertIn("malformed answer", self.damaged("judgments", lambda rows: next(r for r in rows if r["comparison"] == "plain").update(vote="maybe")))

    def test_an_unreadable_answer_is_no_vote(self):
        rows = json.loads(json.dumps(self.data["judgments"]))
        for row in rows:
            if row["comparison"] == "plain" and row["judge"] == verify.JUDGES[0]:
                row["vote"] = "unreadable"
        write(self.root, {"judgments": rows})
        result = verify.verify(self.root, write=True)
        self.assertEqual(result["judges"]["plain"]["unreadable"], 12)
        self.assertEqual(result["judges"]["plain"]["judges"][verify.JUDGES[0]]["model"], 0)
        self.assertEqual(result["judges"]["plain"]["panel"]["model"], 12)

    def test_revisions_and_their_answers_have_to_match(self):
        def unjudged(rows):
            rows[:] = [r for r in rows if not (r["comparison"] == "revision" and r["unit"] == "h0" and r.get("writer") == GPT)]
        self.assertIn("revised pages and judged pages differ", self.damaged("judgments", unjudged))
        self.assertIn("index before and after", self.damaged("revisions", lambda rows: rows[0].pop("before")))
        self.assertIn("malformed revision", self.damaged("revisions", lambda rows: rows[0].update(outcome="improved")))

    def test_an_earlier_pair_needs_both_versions(self):
        def half(rows):
            rows.remove(next(r for r in rows if r["side"] == "after"))
        self.assertIn("both versions", self.damaged("earlier", half))
        self.assertIn("higher reading", self.damaged("earlier", lambda rows: rows[0].update(index=99)))

    def test_a_changed_data_file_is_noticed(self):
        rows = json.loads(json.dumps(self.data["earlier"]))
        rows[0]["index"] = rows[0]["marks"] = 24
        write(self.root, {"earlier": rows})
        with self.assertRaises(verify.Malformed) as caught:
            verify.verify(self.root)
        self.assertIn("files.sha256", str(caught.exception))

    def test_results_and_the_page_cannot_be_edited_by_hand(self):
        results = self.root / "results.json"
        results.write_text(results.read_text().replace('"quiet": 0.05', '"quiet": 0.5'))
        with self.assertRaises(verify.Malformed) as caught:
            verify.verify(self.root)
        self.assertIn("results.json differs", str(caught.exception))
        verify.verify(self.root, write=True)
        page_file = self.root / "RESULTS.md"
        page_file.write_text(page_file.read_text().replace("holds", "holds well", 1))
        with self.assertRaises(verify.Malformed) as caught:
            verify.verify(self.root)
        self.assertIn("RESULTS.md differs", str(caught.exception))

    def test_the_control_is_recounted(self):
        control = verify.verify(self.root)["control"]
        articles = control["groups"]["matched articles"]
        # Two articles are mostly lists: they are counted with all articles and not with the matched ones.
        self.assertEqual((articles["pages"], control["groups"]["all articles"]["pages"], control["groups"]["matched documentation"]["pages"]), (6, 8, 4))
        self.assertEqual((articles["rhythm_high"], articles["rhythm_high_share"], articles["auc_index"]), (0, 0.0, 1.0))
        self.assertEqual(control["sources"]["example/articles"]["matched"], 6)
        self.assertEqual(control["set_s"]["posts"], 6)
        figures = control["sentences"]["matched articles"]
        self.assertEqual((figures["pages"], figures["words_per_sentence"], figures["plain"], figures["in_runs"]), (6, 18.0, 0.667, 0.333))
        self.assertEqual(control["sentences"]["set_s"], verify.SET_S_SENTENCES)

    def test_runs_and_what_chance_gives(self):
        sentences = [[1, 10, 1], [1, 10, 1], [1, 30, 0], [2, 10, 1], [3, 10, 1], [3, 10, 1], [3, 10, 1]]
        # The lone plain sentence of the second stretch is in no run: a run does not cross from one stretch to the next.
        self.assertEqual(verify.in_runs(sentences), 5)
        # Two plain sentences among three: two of the three placements put them side by side.
        self.assertAlmostEqual(verify.by_chance([[1, 10, 1], [1, 10, 1], [1, 30, 0]]), 4 / 3)
        self.assertEqual(verify.by_chance([[1, 10, 1], [2, 10, 1]]), 0.0)
        self.assertEqual(verify.median([3, 1, 2]), 2)
        self.assertEqual(verify.median([4, 1, 2, 3]), 2.5)

    def test_a_control_bar_that_fails_is_reported(self):
        rows = json.loads(json.dumps(self.data["control"]))
        for row in rows:
            if row["kind"] == "articles" and row["paragraphs"] >= verify.MATCH:
                row.update({name: value for name, value in control_page("articles", row["id"], even=60).items() if name != "sentences"})
        write(self.root, {"control": rows})
        result = verify.verify(self.root, write=True)
        self.assertFalse(result["registered"]["control, second reading quiet on people's articles"])
        self.assertFalse(result["registered"]["control, the style-guide set apart from people's articles"])
        self.assertIn("**does not hold**", (self.root / "RESULTS.md").read_text())

    def test_a_control_page_has_to_be_well_formed(self):
        self.assertIn("listed twice", self.damaged("control", lambda rows: rows.append(dict(rows[0]))))
        self.assertIn("its kind or its source", self.damaged("control", lambda rows: rows[0].update(kind="forum")))
        self.assertIn("its kind or its source", self.damaged("control", lambda rows: rows[0].pop("commit")))
        self.assertIn("higher of its two readings", self.damaged("control", lambda rows: rows[0].update(index=rows[0]["index"] + 1)))
        self.assertIn("do not follow from the counts", self.damaged("control", lambda rows: rows[1]["signals"]["plain-sentence-run"].__setitem__(1, 99)))
        self.assertIn("share of prose in paragraphs", self.damaged("control", lambda rows: rows[0].update(paragraphs=1.5)))
        # A plain sentence has five to twenty words.
        self.assertIn("malformed sentence", self.damaged("control", lambda rows: rows[0]["sentences"].append([2, 30, 1])))
        self.assertIn("malformed sentence", self.damaged("control", lambda rows: rows[0]["sentences"].append([2, 12])))
        self.assertIn("out of order", self.damaged("control", lambda rows: rows[0]["sentences"].append([1, 12, 0])))

    def test_a_missing_data_file_is_an_error(self):
        (self.root / "data" / "revisions.jsonl.gz").unlink()
        with self.assertRaises(verify.Malformed) as caught:
            verify.verify(self.root)
        self.assertIn("missing data file", str(caught.exception))


class RecordTest(unittest.TestCase):
    def test_the_record_recounts(self):
        result = verify.verify(HERE)
        self.assertEqual(json.loads((HERE / "results.json").read_text()), result)
        self.assertEqual(result["readings"]["rhythm"]["weights"], verify.RHYTHM)

    def test_the_record_names_no_post_of_the_style_guide_set(self):
        for row in verify.read(HERE, "pages"):
            if row["set"] == "set-s":
                self.assertRegex(row["id"], r"^s-\d\d$")
                self.assertEqual(set(row), {"set", "half", "id", "index", "marks", "rhythm", "earlier"})

    def test_the_control_holds_pages_of_public_repositories_only(self):
        rows = verify.read(HERE, "control")
        self.assertEqual(len(rows), 339)
        for row in rows:
            self.assertRegex(row["repo"], r"^[A-Za-z0-9.-]+/[A-Za-z0-9._-]+$")
            self.assertRegex(row["commit"], r"^[0-9a-f]{40}$")
            self.assertEqual(set(row), {"id", "kind", "repo", "path", "commit", "guide", "paragraphs", "words", "signals", "index", "marks",
                                        "rhythm", "sentences"})


if __name__ == "__main__":
    unittest.main()
