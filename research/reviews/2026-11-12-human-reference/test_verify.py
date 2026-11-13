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

    def test_filtered_opener_load_names_its_candidates(self):
        row = {"part": "code", "pages": 50, "candidates": 4, "failed": 1, "findings": 0}
        write_rows(self.root / "data/opener-load.jsonl.gz", [row])
        self.seal()
        self.assertEqual(self.openers()["load"]["code"]["candidates"], 4)
        # More failures than texts that were run cannot be.
        self.rewrite("data/opener-load.jsonl.gz", lambda rows: rows[0].update(failed=5))
        self.seal_digests_only()
        self.rejects("invalid counts in the opener load")

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
def vocabulary_row(name, part, findings, text):
    """An answer of ten words; the editor changed the first two."""
    return {"id": name, "part": part, "text_sha256": text * 64, "words": 10, "changed": 2, "reference": [[0, 9]],
            "findings": findings}


def vocabulary_finding(start, phrase, words=1):
    """A finding of one or two words; only a finding at the start of the answer lies in the editor's span."""
    spans = [[start, start + 4]] if words == 1 else [[start, start + 4], [start + 5, start + 8]]
    return {"span": [start, spans[-1][1]], "phrase": phrase, "words": spans}


class VocabularyCase(RecordCase):
    def setUp(self):
        super().setUp()
        # On the selection part the second phrase has four findings and the editor changed none, so it leaves.
        selection = [vocabulary_finding(0, 0), *(vocabulary_finding(start, 1, 2) for start in (20, 30, 40, 50))]
        confirmation = [vocabulary_finding(0, 0), vocabulary_finding(20, 0)]
        write_rows(self.root / "data/vocabulary.jsonl.gz",
                   [vocabulary_row("v1", "selection", selection, "c"), vocabulary_row("v2", "confirmation", confirmation, "d")])
        write_rows(self.root / "data/vocabulary-phrases.jsonl.gz", [{"phrases": ["alpha", "beta gamma"], "dropped": ["beta gamma"]}])
        self.seal()

    def vocabulary(self):
        return verify.verify(self.root)["vocabulary"]

    def test_findings_are_recounted_from_coordinates(self):
        result = self.vocabulary()
        parts = result["parts"]
        self.assertEqual((parts["selection_all"]["findings"], parts["selection"]["findings"]), (5, 1))
        confirmed = parts["confirmation"]
        # Two words under two findings; the editor changed the one inside the reference span.
        self.assertEqual((confirmed["findings"], confirmed["findings_changed"], confirmed["removed_words"]), (2, 1, 2))
        self.assertAlmostEqual(confirmed["precision"], 0.5)
        self.assertAlmostEqual(confirmed["chance_document"], 0.2)
        self.assertEqual(result["list"], {"phrases": 2, "dropped": ["beta gamma"], "kept": 1})
        self.assertEqual(result["rule"], {"words_above_chance": True, "findings_changed_enough": False, "ships": False})

    def test_shipping_cannot_be_declared(self):
        results = json.loads((self.root / "results.json").read_text())
        results["vocabulary"]["rule"]["ships"] = True
        (self.root / "results.json").write_text(json.dumps(results))
        verify.write_digests(self.root)
        self.rejects("results differ from the data files")

    def test_dropped_phrases_follow_from_the_selection_part(self):
        self.rewrite("data/vocabulary-phrases.jsonl.gz", lambda rows: rows[0].update(dropped=[]))
        self.seal_digests_only()
        self.rejects("do not follow from the selection part")

    def test_confirmation_has_no_dropped_phrase(self):
        self.rewrite("data/vocabulary.jsonl.gz", lambda rows: rows[1]["findings"][1].update(phrase=1))
        self.seal_digests_only()
        self.rejects("has a finding of a dropped phrase")

    def test_vocabulary_answers_are_not_those_of_the_other_results(self):
        self.rewrite("data/vocabulary.jsonl.gz", lambda rows: rows[0].update(text_sha256="a" * 64))
        self.seal_digests_only()
        self.rejects("is a text of the other results")

    def test_finding_words_stay_inside_the_finding(self):
        self.rewrite("data/vocabulary.jsonl.gz", lambda rows: rows[1]["findings"][0].update(words=[[0, 6]]))
        self.seal_digests_only()
        self.rejects("out of order or hold words outside their span")

    def test_changed_finding_word_is_rejected(self):
        self.rewrite("data/vocabulary.jsonl.gz", lambda rows: rows[1]["findings"][0].update(span=[10, 14], words=[[10, 14]]))
        self.seal_digests_only()
        self.rejects("results differ from the data files")

    def test_dropped_phrases_belong_to_the_list(self):
        self.rewrite("data/vocabulary-phrases.jsonl.gz", lambda rows: rows[0].update(dropped=["delta"]))
        self.seal_digests_only()
        self.rejects("are not distinct phrases of the list")

    def test_vocabulary_part_must_be_known(self):
        self.rewrite("data/vocabulary.jsonl.gz", lambda rows: rows[0].update(part="test"))
        self.seal_digests_only()
        self.rejects("invalid or repeated vocabulary row")

    def test_vocabulary_load_is_reported(self):
        write_rows(self.root / "data/vocabulary-load.jsonl.gz",
                   [{"pages": 4, "failed": 0, "pages_with_finding": 2, "findings": 3, "words": 1000}])
        self.seal()
        self.assertEqual(self.vocabulary()["load"]["findings"], 3)
        self.rewrite("data/vocabulary-load.jsonl.gz", lambda rows: rows[0].update(pages_with_finding=9))
        self.seal_digests_only()
        self.rejects("load of the vocabulary test is inconsistent")


def deslop_pair(name, part, before, after):
    return {"id": name, "repository": "owner/" + name, "commit": "c" * 40, "path": "README.md", "part": part, "web_edit": False,
            "english": True, "words_before": 100, "words_after": 100, "before": before, "after": after}


class DeslopCase(RecordCase):
    def setUp(self):
        super().setUp()
        emoji = "format.decorative-emoji"
        write_rows(self.root / "data/deslop-pairs.jsonl.gz",
                   [deslop_pair("p1", "test", {emoji: 4, "hype.vague-praise": 1}, {"hype.vague-praise": 1}),
                    deslop_pair("p2", "test", {emoji: 2}, {emoji: 3}),
                    deslop_pair("p3", "development", {"hype.vague-praise": 2}, {})])
        self.seal()

    def deslop(self):
        return verify.verify(self.root)["deslop"]

    def test_findings_are_compared_per_thousand_words(self):
        test = self.deslop()["parts"]["test"]
        emoji = test["format.decorative-emoji"]
        self.assertEqual((emoji["fires"], emoji["lower"], emoji["higher"], emoji["findings_before"], emoji["findings_after"]),
                         (2, 1, 1, 6, 3))
        # Without the new rules the first pair does not differ and the second is silent.
        self.assertEqual((test["alpha9_defaults"]["fires"], test["alpha9_defaults"]["same"]), (1, 1))
        self.assertFalse(self.deslop()["rule"]["format.decorative-emoji"])

    def test_shorter_text_is_compared_by_rate(self):
        # Three findings in fifty words are more per word than four in a hundred.
        self.rewrite("data/deslop-pairs.jsonl.gz", lambda rows: rows[0].update(words_after=50, after={"format.decorative-emoji": 3}))
        self.seal()
        self.assertEqual(self.deslop()["parts"]["test"]["format.decorative-emoji"]["higher"], 2)

    def test_default_cannot_be_declared(self):
        results = json.loads((self.root / "results.json").read_text())
        results["deslop"]["rule"]["format.decorative-emoji"] = True
        (self.root / "results.json").write_text(json.dumps(results))
        verify.write_digests(self.root)
        self.rejects("results differ from the data files")

    def frequency_pairs(self, flagged_human_pages):
        dash = "format.em-dash-frequency"
        pairs = [deslop_pair(f"d{n}", "test", {dash: 9}, {dash: 4} if n < 3 else {}) for n in range(40)]
        pairs.append(deslop_pair("d40", "test", {dash: 3, "hype.vague-praise": 1}, {dash: 8, "hype.vague-praise": 1}))
        write_rows(self.root / "data/deslop-pairs.jsonl.gz", pairs)
        write_rows(self.root / "data/frequency-load.jsonl.gz",
                   [{"corpus": "people", "pages": 100, "pages_with_finding": flagged_human_pages, "findings": 9, "words": 50_000}])
        self.seal()
        return self.deslop()

    def test_document_rule_is_measured_by_the_em_dashes_it_lists(self):
        deslop = self.frequency_pairs(5)
        own = deslop["parts"]["test"]["format.em-dash-frequency"]
        self.assertEqual((own["lower"], own["higher"], own["findings_before"], own["findings_after"]), (40, 1, 363, 20))
        # In a set it is one finding per document: the last pair keeps one finding of each rule, so the set is unchanged there.
        defaults = deslop["parts"]["test"]["current_defaults"]
        self.assertEqual((defaults["lower"], defaults["higher"], defaults["same"]), (37, 0, 4))
        self.assertEqual(deslop["parts"]["test"]["with_decorative_emoji"]["fires"], 1)
        self.assertTrue(deslop["rule"]["format.em-dash-frequency"])

    def test_document_rule_fails_above_the_limit_on_human_pages(self):
        deslop = self.frequency_pairs(6)
        self.assertFalse(deslop["rule"]["frequency_within_human_limit"])
        self.assertFalse(deslop["rule"]["format.em-dash-frequency"])

    def test_document_rule_fails_without_the_human_pages(self):
        self.frequency_pairs(5)
        (self.root / "data/frequency-load.jsonl.gz").unlink()
        self.seal()
        self.assertFalse(self.deslop()["rule"]["format.em-dash-frequency"])

    def test_load_row_must_be_consistent(self):
        self.frequency_pairs(5)
        self.rewrite("data/frequency-load.jsonl.gz", lambda rows: rows[0].update(pages_with_finding=101))
        self.seal_digests_only()
        self.rejects("invalid em dash load row")

    def test_english_table_leaves_out_other_languages(self):
        self.rewrite("data/deslop-pairs.jsonl.gz", lambda rows: rows[1].update(english=False))
        self.seal()
        deslop = self.deslop()
        self.assertEqual(deslop["parts"]["test"]["format.decorative-emoji"]["fires"], 2)
        emoji = deslop["english"]["test"]["format.decorative-emoji"]
        self.assertEqual((emoji["pairs"], emoji["fires"], emoji["lower"], emoji["higher"]), (1, 1, 1, 0))
        # One pair cannot meet the bar: the lower end of its interval is below the floor.
        self.assertFalse(deslop["rule"]["english"]["format.decorative-emoji"])

    def test_language_mark_must_be_a_boolean(self):
        self.rewrite("data/deslop-pairs.jsonl.gz", lambda rows: rows[0].update(english="yes"))
        self.seal_digests_only()
        self.rejects("invalid counts in de-slop pair")

    def test_pair_needs_a_known_part(self):
        self.rewrite("data/deslop-pairs.jsonl.gz", lambda rows: rows[0].update(part="extra"))
        self.seal_digests_only()
        self.rejects("invalid or repeated de-slop pair")

    def test_pair_counts_must_be_positive(self):
        self.rewrite("data/deslop-pairs.jsonl.gz", lambda rows: rows[0].update(words_after=0))
        self.seal_digests_only()
        self.rejects("invalid counts in de-slop pair")


def review_row(document, protocol, sentences, both, flagged=(10, 10), hits=(6, 6), findings=(2, 2)):
    later = (None, None) if protocol == "list" else (90, 1)
    edit = None if protocol == "list" else {"sentences": 10, "sentences_changed": 3, "flagged_before": 2, "changed_by_commit": 1,
                                             "flagged_after": 1, "on_new_text": 1, "flagged_both_times": 0}
    return {"document": document, "reviewer": "model-a", "protocol": protocol, "words": 100, "changed": 20,
            "findings": list(findings), "flagged": list(flagged), "flagged_changed": list(hits), "sentences": list(sentences),
            "both": both, "after_words": later[0], "after_findings": later[1], "edit": edit}


class ReviewRepeatCase(RecordCase):
    def setUp(self):
        super().setUp()
        table = []
        for number in range(12):
            name = f"d{number:02}"
            table += [review_row(name, "verdicts", (4, 4), 4), review_row(name, "strict", (2, 2), 2, flagged=(5, 5), hits=(3, 3)),
                      review_row(name, "list", (4, 4), 2)]
        write_rows(self.root / "data/review-repeat.jsonl.gz", table)
        self.seal()

    def repeat(self):
        return verify.verify(self.root)["review_repeat"]["model-a"]

    def change(self, protocol, **fields):
        self.rewrite("data/review-repeat.jsonl.gz", lambda rows: [row.update(fields) for row in rows if row["protocol"] == protocol])
        self.seal()
        return self.repeat()

    def test_repeat_rate_and_rule(self):
        entry = self.repeat()
        self.assertEqual((entry["verdicts"]["repeat_rate"], entry["strict"]["repeat_rate"], entry["list"]["repeat_rate"]), (1.0, 1.0, 0.5))
        self.assertEqual(entry["verdicts"]["difference_interval"], [0.5, 0.5])
        # Six of ten flagged words changed where two in ten words of the document did.
        self.assertAlmostEqual(entry["verdicts"]["precision"] - entry["verdicts"]["chance"], 0.4)
        self.assertEqual((entry["verdicts"]["changed_words_flagged"], entry["strict"]["changed_words_flagged"]), (72, 36))
        self.assertTrue(entry["rule"]["ships"])
        # Two findings in 100 words before, one in 90 after: lower in every pair.
        move = entry["verdicts"]["movement"]
        self.assertEqual((move["lower"], move["higher"]), (12, 0))
        self.assertEqual((move["edit"]["flagged_before"], move["edit"]["changed_by_commit"], move["edit"]["on_new_text"]), (24, 12, 12))

    def test_no_gain_does_not_ship(self):
        entry = self.change("verdicts", both=2)
        self.assertEqual(entry["verdicts"]["difference"], 0.0)
        self.assertFalse(entry["rule"]["repeats_more"])
        self.assertFalse(entry["rule"]["ships"])

    def test_flags_at_chance_do_not_ship(self):
        entry = self.change("verdicts", flagged_changed=[2, 2])
        self.assertFalse(entry["rule"]["on_changed_words"])
        self.assertFalse(entry["rule"]["ships"])

    def test_finding_less_than_the_list_form_does_not_ship(self):
        # The list form now flags more changed words than the verdict form does.
        entry = self.change("list", flagged=[20, 20], flagged_changed=[12, 12])
        self.assertFalse(entry["rule"]["finds_no_less"])
        self.assertFalse(entry["rule"]["ships"])

    def test_strict_form_does_not_decide(self):
        entry = self.change("strict", both=0, flagged_changed=[0, 0])
        self.assertTrue(entry["rule"]["ships"])

    def test_result_cannot_be_declared(self):
        self.change("verdicts", both=2)
        results = json.loads((self.root / "results.json").read_text())
        results["review_repeat"]["model-a"]["rule"]["ships"] = True
        (self.root / "results.json").write_text(json.dumps(results))
        verify.write_digests(self.root)
        self.rejects("results differ from the data files")

    def test_shared_sentences_cannot_exceed_a_run(self):
        self.rewrite("data/review-repeat.jsonl.gz", lambda rows: rows[0].update(both=5))
        self.seal_digests_only()
        self.rejects("inconsistent review row")

    def test_commit_cannot_change_more_flagged_sentences_than_were_flagged(self):
        self.rewrite("data/review-repeat.jsonl.gz", lambda rows: rows[0]["edit"].update(changed_by_commit=3))
        self.seal_digests_only()
        self.rejects("inconsistent review row")

    def test_protocols_must_cover_the_same_texts(self):
        self.rewrite("data/review-repeat.jsonl.gz", lambda rows: rows[2].update(document="other"))
        self.seal_digests_only()
        self.rejects("do not pair the protocols")

    def test_a_protocol_cannot_be_missing(self):
        self.rewrite("data/review-repeat.jsonl.gz", lambda rows: [row.update(protocol="verdicts", reviewer="model-b")
                                                                   for row in rows if row["protocol"] == "strict"])
        self.seal_digests_only()
        self.rejects("do not pair the protocols")

    def test_list_rows_carry_no_text_after_the_commit(self):
        self.rewrite("data/review-repeat.jsonl.gz", lambda rows: rows[2].update(after_words=90, after_findings=1))
        self.seal_digests_only()
        self.rejects("inconsistent review row")


class DefaultsCase(RecordCase):
    def setUp(self):
        super().setUp()
        pages = [{"page": f"p{n:02}", "words": 300, "rules": []} for n in range(40)]
        pages[0]["rules"] = ["filler.wordy-phrase", "format.em-dash-frequency"]
        pages[1]["rules"] = ["filler.wordy-phrase"]
        pages[2]["rules"] = ["filler.wordy-phrase", "format.em-dash-frequency"]
        write_rows(self.root / "data/default-load.jsonl.gz", pages)
        dash = "format.em-dash-frequency"
        pairs = [deslop_pair(f"m{n:02}", "test", {dash: 9, "filler.wordy-phrase": 1}, {dash: 3, "filler.wordy-phrase": 2})
                 for n in range(30)]
        for pair in pairs:
            del pair["part"], pair["web_edit"], pair["english"]
        pairs[1]["repository"] = pairs[0]["repository"]
        write_rows(self.root / "data/manual-pairs.jsonl.gz", pairs)
        self.seal()

    def results(self):
        return verify.verify(self.root)

    def test_rules_over_the_limit_ship_off(self):
        load = self.results()["default_load"]
        # Three pages in forty are over one in twenty; two in forty are not.
        self.assertEqual((load["rules"]["filler.wordy-phrase"], load["rules"]["format.em-dash-frequency"]), (3, 2))
        self.assertEqual(load["over_limit"], ["filler.wordy-phrase"])
        self.assertTrue(load["defaults_within_limit"])
        self.assertEqual((load["with_any_finding"], load["with_default_finding"]), (3, 2))

    def test_default_rule_over_the_limit_is_reported(self):
        self.rewrite("data/default-load.jsonl.gz", lambda rows: rows[1].update(rules=["filler.wordy-phrase", "format.em-dash-frequency"]))
        self.seal()
        load = self.results()["default_load"]
        self.assertEqual(load["over_limit"], ["filler.wordy-phrase", "format.em-dash-frequency"])
        self.assertFalse(load["defaults_within_limit"])

    def test_manual_reading(self):
        manual = self.results()["manual"]
        self.assertEqual((manual["pairs"], manual["repositories"]), (30, 29))
        dash = manual["all"]["format.em-dash-frequency"]
        self.assertEqual((dash["lower"], dash["higher"]), (30, 0))
        self.assertTrue(manual["confirmed"]["format.em-dash-frequency"])
        # The wordy-phrase rule rises in every pair and is counted with the rules that ship off.
        self.assertEqual(manual["all"]["strict_only_rules"]["higher"], 30)
        self.assertEqual(manual["one_per_repository"]["format.em-dash-frequency"]["lower"], 29)
        self.assertFalse(manual["confirmed"]["format.decorative-emoji"])

    def test_confirmation_cannot_be_declared(self):
        results = json.loads((self.root / "results.json").read_text())
        results["manual"]["confirmed"]["format.decorative-emoji"] = True
        (self.root / "results.json").write_text(json.dumps(results))
        verify.write_digests(self.root)
        self.rejects("results differ from the data files")

    def test_load_page_cannot_repeat(self):
        self.rewrite("data/default-load.jsonl.gz", lambda rows: rows[1].update(page="p00"))
        self.seal_digests_only()
        self.rejects("invalid load row")

    def test_manual_pair_needs_positive_counts(self):
        self.rewrite("data/manual-pairs.jsonl.gz", lambda rows: rows[0].update(words_after=0))
        self.seal_digests_only()
        self.rejects("invalid counts in manual pair")


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
