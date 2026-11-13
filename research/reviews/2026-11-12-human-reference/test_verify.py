"""Negative tests for the human-reference record check.

Each test builds a small valid record, breaks one property, and expects the
check to reject it.
"""
import gzip
import json
import tempfile
import unittest
from pathlib import Path

import verify

ORDER = ["Human-edited", "AI-edited", "AI-generated"]


def document(name, split="test", editor="e1", systems=None, text="a" * 64):
    return {"id": name, "set": "lamp", "split": split, "editor": editor, "kind": "fiction", "generator": "model",
            "text_sha256": text, "units": [[0, 10], [11, 20], [21, 30]], "reference": [[2, 5]],
            "systems": systems if systems is not None else {"tool": [[3, 4], [22, 25]]}}


def write_rows(path, rows):
    path.parent.mkdir(parents=True, exist_ok=True)
    with gzip.GzipFile(path, "wb", mtime=0) as handle:
        for row in rows:
            handle.write((json.dumps(row, sort_keys=True) + "\n").encode())


def read_rows(path):
    return list(verify.rows(path))


class RecordCase(unittest.TestCase):
    def setUp(self):
        self.directory = tempfile.TemporaryDirectory()
        self.addCleanup(self.directory.cleanup)
        self.root = Path(self.directory.name)
        ceiling = [document("c1", "ceiling", "e1", {}, "b" * 64), document("c2", "ceiling", "e2", {}, "b" * 64)]
        ceiling[1]["reference"] = [[12, 15]]
        write_rows(self.root / "data/sentences.jsonl.gz", [document("d1"), *ceiling])
        writers = [{"writer": name, "order": ORDER, "setting": "full"} for name in ("w1", "w2", "w3")]
        write_rows(self.root / "data/preferences.jsonl.gz",
                   [{"triplet": "t1", "writers": writers, "judges": {"judge": list(reversed(ORDER))}}])
        write_rows(self.root / "data/expert-pairs.jsonl.gz",
                   [{"id": "p1", "judges": {"judge": {"better": "raw", "position": "A"}}}])
        write_rows(self.root / "data/readers.jsonl.gz",
                   [{"reader": "r1", "pair": "a", "arm": "reviewed", "reviewer": "m", "outcome": "revised"},
                    {"reader": "r1", "pair": "b", "arm": "control", "reviewer": "m", "outcome": "original"}])
        # The top model against the second and one more, with one abstaining: three votes each, and the tie goes to the top model.
        tie = {"fable": "revised", "claude": "original", "astra": "original", "qwen": "same"}
        # Three one-vote models together outweigh the top model.
        outvoted = {"fable": "revised", "claude": "original", "astra": "original", "qwen": "original"}
        write_rows(self.root / "data/panel.jsonl.gz",
                   [{"comparison": "first", "pair": "a", "arm": "reviewed", "reviewer": "m", "judges": tie},
                    {"comparison": "trim", "pair": "t", "arm": "reviewed", "reviewer": "m", "judges": outvoted},
                    {"comparison": "product", "pair": "p", "arm": "product", "reviewer": "m", "judges": tie}])
        self.seal()

    def seal(self):
        (self.root / "results.json").write_text(json.dumps(verify.summarize(self.root)))
        verify.write_digests(self.root)

    def rewrite(self, name, change):
        path = self.root / name
        rows = read_rows(path)
        change(rows)
        write_rows(path, rows)

    def rejects(self, pattern):
        with self.assertRaisesRegex(ValueError, pattern):
            verify.verify(self.root)

    def test_valid_record_passes(self):
        results = verify.verify(self.root)
        tool = results["sentences"]["systems"]["tool"]["lamp/test"]
        # One of three sentences was edited; the tool flags it and one unedited sentence.
        self.assertEqual((tool["edited_flagged"], tool["unedited_flagged"], tool["tpr"], tool["fpr"]), (1, 1, 1.0, 0.5))
        # The two editors changed different sentences, so neither predicts the other.
        self.assertEqual(results["sentences"]["ceiling"]["edited_flagged"], 0)
        self.assertFalse(results["judges"]["judge"]["admitted"])

    def test_changed_file_is_rejected(self):
        write_rows(self.root / "data/sentences.jsonl.gz", [document("d2")])
        self.rejects("digest mismatch")

    def test_stale_results_are_rejected(self):
        self.rewrite("data/sentences.jsonl.gz", lambda rows: rows[0]["systems"]["tool"].append([12, 14]))
        verify.write_digests(self.root)
        self.rejects("results differ from the data files")

    def test_changed_published_number_is_rejected(self):
        results = json.loads((self.root / "results.json").read_text())
        results["sentences"]["systems"]["tool"]["lamp/test"]["j"] = 0.9
        (self.root / "results.json").write_text(json.dumps(results))
        verify.write_digests(self.root)
        self.rejects("results differ from the data files")

    def test_admission_cannot_be_declared(self):
        results = json.loads((self.root / "results.json").read_text())
        results["judges"]["judge"]["admitted"] = True
        (self.root / "results.json").write_text(json.dumps(results))
        verify.write_digests(self.root)
        self.rejects("results differ from the data files")

    def write_pages(self):
        """Write every generated page that the results call for."""
        import render
        results = json.loads((self.root / "results.json").read_text())
        for page, function in render.PAGES.items():
            if page == "RESULTS.md" or render.has_words(results):
                (self.root / page).write_text(function(results))

    def test_edited_results_page_is_rejected(self):
        self.write_pages()
        verify.verify(self.root)
        page = (self.root / "RESULTS.md").read_text().replace("+50.0", "+90.0")
        (self.root / "RESULTS.md").write_text(page)
        self.rejects("RESULTS.md differs from the rendering")

    def test_text_is_rejected(self):
        self.rewrite("data/sentences.jsonl.gz", lambda rows: rows[2].update(text="A paragraph of the source."))
        self.seal_digests_only()
        self.rejects("unexpected document fields")

    def seal_digests_only(self):
        verify.write_digests(self.root)

    def test_invalid_span_is_rejected(self):
        self.rewrite("data/sentences.jsonl.gz", lambda rows: rows[2]["reference"].append([9, 4]))
        self.seal_digests_only()
        self.rejects("invalid span")

    def test_unknown_split_is_rejected(self):
        self.rewrite("data/sentences.jsonl.gz", lambda rows: rows[2].update(split="train"))
        self.seal_digests_only()
        self.rejects("unknown set or split")

    def test_ranking_must_order_three_versions(self):
        self.rewrite("data/preferences.jsonl.gz", lambda rows: rows[0]["judges"].update(judge=["Human-edited", "Human-edited", "AI-edited"]))
        self.seal_digests_only()
        self.rejects("not an order of the three versions")

    def test_triplet_needs_three_writers(self):
        self.rewrite("data/preferences.jsonl.gz", lambda rows: rows[0]["writers"].pop())
        self.seal_digests_only()
        self.rejects("does not have three writers")

    def test_invalid_expert_choice_is_rejected(self):
        self.rewrite("data/expert-pairs.jsonl.gz", lambda rows: rows[0]["judges"]["judge"].update(better="both"))
        self.seal_digests_only()
        self.rejects("invalid choice")

    def test_reader_cannot_answer_a_pair_twice(self):
        self.rewrite("data/readers.jsonl.gz", lambda rows: rows.append(dict(rows[0], outcome="original")))
        self.seal_digests_only()
        self.rejects("answered pair a twice")

    def test_invalid_reader_answer_is_rejected(self):
        self.rewrite("data/readers.jsonl.gz", lambda rows: rows[0].update(outcome="left"))
        self.seal_digests_only()
        self.rejects("invalid reader answer")

    def test_reader_text_is_rejected(self):
        self.rewrite("data/readers.jsonl.gz", lambda rows: rows[0].update(comment="The left one reads better."))
        self.seal_digests_only()
        self.rejects("unexpected reader fields")

    def test_consensus_weighs_the_top_model(self):
        results = verify.verify(self.root)
        self.assertEqual(results["panel"]["first"]["arms"]["reviewed"]["revised"], 1)
        self.assertEqual(results["panel"]["trim"]["arms"]["reviewed"]["original"], 1)
        self.assertEqual(results["panel"]["first"]["opposed"], 1)
        # One pair cannot meet a rule that needs fifteen decisive answers.
        self.assertFalse(results["panel"]["first"]["reviewed_meets_rule"])

    def test_one_vote_per_family_is_counted_apart(self):
        results = verify.verify(self.root)
        # In the tie row the three families give three different answers; in the other two families prefer the original.
        self.assertEqual(results["panel"]["first"]["one_vote_per_family"]["reviewed"]["same"], 1)
        self.assertEqual(results["panel"]["trim"]["one_vote_per_family"]["reviewed"]["original"], 1)
        self.assertFalse(results["panel"]["trim"]["one_vote_per_family"]["reviewed"]["meets_rule"])

    def test_product_pairs_have_one_arm(self):
        results = verify.verify(self.root)
        self.assertEqual(results["panel"]["product"]["arms"]["product"]["revised"], 1)
        self.assertFalse(results["panel"]["product"]["product_meets_rule"])
        self.rewrite("data/panel.jsonl.gz", lambda rows: rows[2].update(arm="reviewed"))
        self.seal_digests_only()
        self.rejects("invalid panel row")

    def test_consensus_outcome_cannot_be_declared(self):
        results = json.loads((self.root / "results.json").read_text())
        results["panel"]["trim"]["reviewed_meets_rule"] = True
        (self.root / "results.json").write_text(json.dumps(results))
        verify.write_digests(self.root)
        self.rejects("results differ from the data files")

    def test_panel_needs_every_judge(self):
        self.rewrite("data/panel.jsonl.gz", lambda rows: rows[0]["judges"].pop("qwen"))
        self.seal_digests_only()
        self.rejects("does not carry one valid answer of every judge")

    def test_panel_answer_must_be_valid(self):
        self.rewrite("data/panel.jsonl.gz", lambda rows: rows[0]["judges"].update(claude="better"))
        self.seal_digests_only()
        self.rejects("does not carry one valid answer of every judge")

    def test_panel_pair_cannot_repeat(self):
        self.rewrite("data/panel.jsonl.gz", lambda rows: rows.append(dict(rows[1])))
        self.seal_digests_only()
        self.rejects("appears twice in the panel answers")

    def test_editor_cannot_edit_a_paragraph_twice(self):
        self.rewrite("data/sentences.jsonl.gz", lambda rows: rows[2].update(editor="e1"))
        self.seal_digests_only()
        self.rejects("edited one paragraph twice")


def trim_row(model, removed, distance):
    """One trim of the first sentence of d1: four words, the last one fixed, the second one changed by the editor."""
    return {"id": "d1", "model": model, "sentences": 3, "proposed": 2, "rejected": 1, "distance": distance,
            "trims": [{"span": [0, 10], "words": [[0, 1], [2, 5], [6, 8], [9, 10]], "fixed": [3], "removed": removed}]}


class TrimCase(RecordCase):
    def setUp(self):
        super().setUp()

        def with_trims(rows):
            rows[0]["systems"].update({"trim-claude": [[0, 10]], "trim-astra": [[0, 10]]})

        self.rewrite("data/sentences.jsonl.gz", with_trims)
        write_rows(self.root / "data/word-counts.jsonl.gz",
                   [{"id": "d1", "words": 9, "changed": 1, "systems": {"review-claude": [4, 1]}}])
        # The first model removes the word the editor changed; the second removes a word the editor kept.
        write_rows(self.root / "data/trims.jsonl.gz",
                   [trim_row("astra", [2], {"before": 3, "after": 4, "control": 3.4}),
                    trim_row("claude", [1], {"before": 3, "after": 2, "control": 3.4})])
        self.seal()

    def trims(self):
        return verify.verify(self.root)["trims"]["sets"]["lamp/test"]

    def test_trims_are_recounted_from_coordinates(self):
        entry = self.trims()
        claude, astra = entry["models"]["claude"], entry["models"]["astra"]
        self.assertEqual((claude["removed_words"], claude["precision"], astra["precision"]), (1, 1.0, 0.0))
        # Three words may be removed and the editor changed one of them; the document has nine words and one changed.
        self.assertAlmostEqual(claude["chance_sentences"], 1 / 3)
        self.assertAlmostEqual(claude["chance_document"], 1 / 9)
        self.assertEqual((claude["net"], astra["net"]), (-1.0, 1.0))
        self.assertAlmostEqual(claude["differences"]["net_minus_control"], -1.4)
        review = entry["comparisons"]["review-claude"]
        self.assertEqual((review["removed_words"], review["precision"]), (4, 0.25))

    def test_rule_needs_both_models(self):
        self.assertEqual(self.trims()["rule"], {"agrees_above_chance": False, "moves_toward_editor": False})
        self.rewrite("data/trims.jsonl.gz",
                     lambda rows: rows[0].update(trim_row("astra", [1], {"before": 3, "after": 2, "control": 3.4})))
        self.seal()
        self.assertEqual(self.trims()["rule"], {"agrees_above_chance": True, "moves_toward_editor": True})

    def test_rule_outcome_cannot_be_declared(self):
        results = json.loads((self.root / "results.json").read_text())
        results["trims"]["sets"]["lamp/test"]["rule"]["agrees_above_chance"] = True
        (self.root / "results.json").write_text(json.dumps(results))
        verify.write_digests(self.root)
        self.rejects("results differ from the data files")

    def test_changed_removed_word_is_rejected(self):
        self.rewrite("data/trims.jsonl.gz", lambda rows: rows[0]["trims"][0].update(removed=[1]))
        self.seal_digests_only()
        self.rejects("results differ from the data files")

    def test_trim_cannot_remove_a_fixed_word(self):
        self.rewrite("data/trims.jsonl.gz", lambda rows: rows[0]["trims"][0].update(removed=[3]))
        self.seal_digests_only()
        self.rejects("removes a word that the check never removes")

    def test_trim_must_be_a_finding_of_its_model(self):
        self.rewrite("data/trims.jsonl.gz", lambda rows: rows[0]["trims"][0].update(span=[0, 12]))
        self.seal_digests_only()
        self.rejects("is not a finding of trim-astra")

    def test_trim_words_stay_inside_the_sentence(self):
        self.rewrite("data/trims.jsonl.gz", lambda rows: rows[0]["trims"][0]["words"].__setitem__(3, [9, 14]))
        self.seal_digests_only()
        self.rejects("out of order or outside the sentence")

    def test_trim_needs_a_removed_word(self):
        self.rewrite("data/trims.jsonl.gz", lambda rows: rows[0]["trims"][0].update(removed=[]))
        self.seal_digests_only()
        self.rejects("invalid word indexes")

    def test_distance_is_required_where_the_editors_text_exists(self):
        self.rewrite("data/trims.jsonl.gz", lambda rows: rows[0].pop("distance"))
        self.seal_digests_only()
        self.rejects("must carry a distance exactly when")

    def test_distance_cannot_move_further_than_the_removed_words(self):
        self.rewrite("data/trims.jsonl.gz", lambda rows: rows[1]["distance"].update(after=0))
        self.seal_digests_only()
        self.rejects("moves further than the removed words allow")

    def test_every_document_needs_a_row_of_every_model(self):
        self.rewrite("data/trims.jsonl.gz", lambda rows: rows.pop())
        self.seal_digests_only()
        self.rejects("do not cover every development and test document with every model")

    def test_word_counts_cover_the_documents(self):
        self.rewrite("data/word-counts.jsonl.gz", lambda rows: rows.pop())
        self.seal_digests_only()
        self.rejects("word counts do not cover")

    def test_word_counts_must_be_consistent(self):
        self.rewrite("data/word-counts.jsonl.gz", lambda rows: rows[0].update(changed=12))
        self.seal_digests_only()
        self.rejects("invalid word counts")


class OpenerCase(TrimCase):
    """A second document, an expert-edited answer, opens with a flagged word that the editor changed."""

    def setUp(self):
        super().setUp()
        answer = document("b1", systems={"rules-opener": [[2, 6]], "trim-claude": [], "trim-astra": []})
        answer["set"] = "beemo"
        self.rewrite("data/sentences.jsonl.gz", lambda rows: rows.append(answer))
        self.rewrite("data/word-counts.jsonl.gz", lambda rows: rows.append({"id": "b1", "words": 9, "changed": 1, "systems": {}}))
        self.rewrite("data/trims.jsonl.gz", lambda rows: rows.extend(
            {"id": "b1", "model": model, "sentences": 3, "proposed": 0, "rejected": 0, "trims": []}
            for model in ("astra", "claude")))
        write_rows(self.root / "data/openers.jsonl.gz", [{"id": "b1", "span": [2, 6], "words": [[2, 5]]}])
        write_rows(self.root / "data/rule-words.jsonl.gz",
                   [{"id": "b1", "rules": {"some.rule": [4, 1]}}, {"id": "d1", "rules": {}}])
        self.seal()

    def openers(self):
        return verify.verify(self.root)["openers"]

    def test_opener_is_counted_from_its_words(self):
        entry = self.openers()["sets"]["beemo/test"]
        self.assertEqual((entry["findings"], entry["changed"]), (1, 1))
        rule = verify.verify(self.root)["rule_words"]["beemo/test"]["some.rule"]
        self.assertEqual((rule["removed_words"], rule["precision"]), (4, 0.25))

    def test_one_finding_does_not_list_the_opener(self):
        # One finding of one cannot put the lower end of the interval above the floor.
        self.assertFalse(self.openers()["listed"])

    def test_listing_cannot_be_declared(self):
        results = json.loads((self.root / "results.json").read_text())
        results["openers"]["listed"] = True
        (self.root / "results.json").write_text(json.dumps(results))
        verify.write_digests(self.root)
        self.rejects("results differ from the data files")

    def test_opener_load_is_reported(self):
        write_rows(self.root / "data/opener-load.jsonl.gz", [{"part": "real", "pages": 5, "failed": 0, "findings": 1}])
        self.seal()
        self.assertEqual(self.openers()["load"], {"real": {"pages": 5, "failed": 0, "findings": 1}})

    def test_opener_load_must_be_consistent(self):
        write_rows(self.root / "data/opener-load.jsonl.gz", [{"part": "real", "pages": 5, "failed": 6, "findings": 0}])
        self.seal_digests_only()
        self.rejects("invalid counts in the opener load")

    def test_opener_load_part_cannot_repeat(self):
        row = {"part": "real", "pages": 5, "failed": 0, "findings": 0}
        write_rows(self.root / "data/opener-load.jsonl.gz", [row, row])
        self.seal_digests_only()
        self.rejects("invalid or repeated opener load row")

    def test_opener_must_be_a_finding_of_the_changed_rule(self):
        self.rewrite("data/openers.jsonl.gz", lambda rows: rows[0].update(span=[2, 7]))
        self.seal_digests_only()
        self.rejects("is repeated or is not a finding of rules-opener")

    def test_opener_cannot_repeat(self):
        self.rewrite("data/openers.jsonl.gz", lambda rows: rows.append(dict(rows[0])))
        self.seal_digests_only()
        self.rejects("is repeated or is not a finding of rules-opener")

    def test_opener_words_stay_inside_the_finding(self):
        self.rewrite("data/openers.jsonl.gz", lambda rows: rows[0].update(words=[[2, 9]]))
        self.seal_digests_only()
        self.rejects("needs words inside its span")

    def test_changed_opener_word_is_rejected(self):
        # The word [3, 6] is not inside the editor's span [2, 5], so the count changes.
        self.rewrite("data/openers.jsonl.gz", lambda rows: rows[0].update(words=[[3, 6]]))
        self.seal_digests_only()
        self.rejects("results differ from the data files")

    def test_rule_words_cover_the_documents(self):
        self.rewrite("data/rule-words.jsonl.gz", lambda rows: rows.pop())
        self.seal_digests_only()
        self.rejects("rule word rows do not cover")

    def test_rule_words_must_be_consistent(self):
        self.rewrite("data/rule-words.jsonl.gz", lambda rows: rows[0]["rules"].update({"some.rule": [4, 6]}))
        self.seal_digests_only()
        self.rejects("invalid counts of some.rule")


class WordsPageTest(OpenerCase):
    """The page of the word-level results is checked like the first page."""

    def seal(self):
        super().seal()
        self.write_pages()

    def test_pages_match_the_results(self):
        verify.verify(self.root)

    def test_edited_words_page_is_rejected(self):
        page = self.root / "RESULTS-words.md"
        page.write_text(page.read_text().replace("1 of 1", "2 of 2"))
        self.rejects("RESULTS-words.md differs from the rendering of the results")

    def test_missing_words_page_is_rejected(self):
        (self.root / "RESULTS-words.md").unlink()
        self.rejects("RESULTS-words.md is missing")


class SummationTest(unittest.TestCase):
    """Published numbers must not depend on the Python version that recounts them."""

    def test_column_sums_are_exactly_rounded(self):
        # Ten times 0.1 is 1.0 when rounded once; adding one by one gives 0.9999999999999999 before Python 3.12.
        table = [(0.1,) * len(verify.SUMS)] * 10
        self.assertEqual(verify.totals(table), [1.0] * len(verify.SUMS))


class OpenerRuleTest(unittest.TestCase):
    """The listing rule, on counts."""

    def listed(self, findings, changed, elsewhere=0):
        test = {"findings": findings, "changed": changed, "interval": verify.wilson(changed, findings)}
        return (bool(findings) and changed >= verify.OPENER_SHARE * findings
                and test["interval"][0] > verify.OPENER_FLOOR and elsewhere <= verify.OPENER_LOAD)

    def test_thresholds(self):
        self.assertTrue(self.listed(39, 39))
        self.assertTrue(self.listed(39, 36))
        # 33 of 39 is 84.6%, below the share.
        self.assertFalse(self.listed(39, 33))
        # Ten of ten leaves the lower end of the interval at 72%; eight of eight leaves it below the floor.
        self.assertTrue(self.listed(10, 10))
        self.assertFalse(self.listed(8, 8))
        self.assertFalse(self.listed(39, 39, elsewhere=2))


if __name__ == "__main__":
    unittest.main()
