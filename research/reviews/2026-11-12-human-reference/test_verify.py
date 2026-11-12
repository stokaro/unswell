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
        # The top model against two others, with one abstaining: two votes each, and the tie goes to the top model.
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

    def test_edited_results_page_is_rejected(self):
        import render
        results = json.loads((self.root / "results.json").read_text())
        (self.root / "RESULTS.md").write_text(render.render(results))
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


if __name__ == "__main__":
    unittest.main()
