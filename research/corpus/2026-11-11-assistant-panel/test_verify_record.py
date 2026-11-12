"""Negative tests for the corpus record check.

Each test builds a small valid record, breaks one property, and expects the
check to reject it.
"""
import gzip
import json
import tempfile
import unittest
from pathlib import Path

import verify_record

PANEL = {
    "astra": {"model": "model-a", "effort": "medium", "roles": ["annotator", "reviewer", "caller"]},
    "claude": {"model": "model-b", "effort": "default", "roles": ["annotator", "reviewer", "caller"]},
    "qwen": {"model": "model-c", "effort": "default", "roles": ["auditor", "reviewer", "caller"]},
}
ROLES = {"dev": "development", "test": "final test 1"}


def document(name, repository):
    return {"doc": name, "repository": repository, "path": "README.md", "sha256": "0" * 64, "reference": "",
            "origin": "real", "family": "", "operation": "",
            "items": [{"id": "i001", "kind": "sentence", "segments": [[0, 10]]},
                      {"id": "i002", "kind": "sentence", "segments": [[11, 30]]}]}


def label(name, annotator, items):
    return {"doc": name, "annotator": annotator, "model": PANEL[annotator]["model"], "effort": "default",
            "rubric": "rubric-v1", "prompt_sha256": "1" * 64, "unbound_flags": 0,
            "flags": [{"item": item, "label": "needs_revision", "categories": ["wordiness"]} for item in items]}


def review(identifier, source_set, name, verdicts):
    return {"id": identifier, "set": source_set, "doc": name, "rule": "filler.wordy-phrase", "group": "",
            "phrase": "in order to", "reviews": verdicts, "severity": "warning", "span": [11, 22]}


def flag(identifier, caller, verdicts):
    return {"id": identifier, "set": "dev", "doc": "dev-1", "caller": caller,
            "categories": ["vague_claims"], "span": [11, 22], "reviews": verdicts}


def generated(name):
    return {**document(name, "owner/dev"), "origin": "generated", "family": "family", "operation": "generate"}


def command_flag(identifier, span, verdicts):
    return {"id": identifier, "set": "dev", "doc": "gen-1", "caller": "claude",
            "categories": ["wordiness"], "span": span, "reviews": verdicts}


def command_round(root, reference="astra", independent=True):
    """Publish a round for one run whose counts match the record at root."""
    sets = verify_record.load_items(root)
    marks = verify_record.load_labels(root, sets, PANEL)
    run = {"reviewer": "claude", "confirmer": "qwen", "reference": reference,
           "independent_reference": independent, "documents": 1}
    counts = verify_record.command_counts(root, sets, marks, "dev", run)
    pooled = counts if independent else {}
    separates, keeps = verify_record.command_decision(verify_record.collections.Counter(pooled))
    return {"set": "dev", "runs": [{**run, **{field: counts[field] for field in verify_record.COMMAND_COUNTS}}],
            "pooled_independent": {**{field: pooled.get(field, 0) for field in verify_record.COMMAND_COUNTS},
                                   "separates": separates, "keeps_two_thirds_of_coverage": keeps}}


def write_rows(path, rows):
    path.parent.mkdir(parents=True, exist_ok=True)
    with gzip.GzipFile(path, "wb", mtime=0) as handle:
        for row in rows:
            handle.write((json.dumps(row, sort_keys=True) + "\n").encode())


def read_rows(path):
    return list(verify_record.rows(path))


class RecordCase(unittest.TestCase):
    def setUp(self):
        self.directory = tempfile.TemporaryDirectory()
        self.addCleanup(self.directory.cleanup)
        self.root = Path(self.directory.name)
        write_rows(self.root / "items/dev.jsonl.gz", [document("dev-1", "owner/dev"), generated("gen-1")])
        write_rows(self.root / "items/test.jsonl.gz", [document("test-1", "owner/test")])
        write_rows(self.root / "labels/dev.jsonl.gz",
                   [label("dev-1", "astra", ["i001", "i002"]), label("dev-1", "claude", ["i002"]),
                    label("gen-1", "astra", ["i001"]), label("gen-1", "claude", [])])
        # The reviewer flags both items of the generated document. The confirming
        # model keeps the flag on the item that the reference annotator marked.
        write_rows(self.root / "contextual/command-claude.jsonl.gz",
                   [command_flag("k1", [2, 8], {"qwen": "yes"}), command_flag("k2", [12, 20], {"qwen": "no"})])
        (self.root / "rounds").mkdir()
        (self.root / "rounds/review-command.json").write_text(json.dumps(command_round(self.root)))
        write_rows(self.root / "reviews/round.jsonl.gz",
                   [review("f1", "dev", "dev-1", {"astra": "yes", "claude": "yes"}),
                    review("f2", "test", "test-1", {"claude": "no", "qwen": "no"})])
        write_rows(self.root / "contextual/claude.jsonl.gz", [flag("c1", "claude", {"astra": "yes", "qwen": "no"})])
        self.seal()

    def seal(self, summary=None):
        """Write a summary that matches the data, or the given one, and the digests."""
        if summary is None:
            sets = verify_record.load_items(self.root)
            marks = verify_record.load_labels(self.root, sets, PANEL)
            rounds = verify_record.load_reviews(self.root, sets, PANEL)
            summary = verify_record.summarize(sets, marks, rounds, ROLES, PANEL)
        (self.root / "summary.json").write_text(json.dumps(summary))
        verify_record.write_digests(self.root)

    def rewrite(self, name, change):
        path = self.root / name
        rows = read_rows(path)
        change(rows)
        write_rows(path, rows)

    def rejects(self, pattern):
        with self.assertRaisesRegex(ValueError, pattern):
            verify_record.verify(self.root)

    def test_valid_record_passes(self):
        summary = verify_record.verify(self.root)
        self.assertEqual(summary["sets"][0]["annotated"]["both"], 1)
        self.assertEqual(summary["rounds"][0]["rules"]["filler.wordy-phrase"]["both_accept"], 1)
        published = json.loads((self.root / "rounds/review-command.json").read_text())
        run = published["runs"][0]
        self.assertEqual([run[field] for field in verify_record.COMMAND_COUNTS], [2, 1, 1, 1, 1, 1, 1, 1, 0])
        self.assertTrue(published["pooled_independent"]["separates"])

    def test_wrong_command_count_is_rejected(self):
        published = json.loads((self.root / "rounds/review-command.json").read_text())
        published["runs"][0]["marked_touched_after"] = 0
        (self.root / "rounds/review-command.json").write_text(json.dumps(published))
        verify_record.write_digests(self.root)
        self.rejects("review command round differs from the data files: claude with qwen, marked_touched_after")

    def test_wrong_command_decision_is_rejected(self):
        published = json.loads((self.root / "rounds/review-command.json").read_text())
        published["pooled_independent"]["separates"] = False
        (self.root / "rounds/review-command.json").write_text(json.dumps(published))
        verify_record.write_digests(self.root)
        self.rejects("review command decision differs")

    def test_reference_that_reviews_is_not_independent(self):
        published = command_round(self.root, reference="claude")
        (self.root / "rounds/review-command.json").write_text(json.dumps(published))
        verify_record.write_digests(self.root)
        self.rejects("the reference takes part in it")

    def test_command_flag_outside_the_selection_is_rejected(self):
        self.rewrite("contextual/command-claude.jsonl.gz", lambda rows: rows[0].update(doc="dev-1"))
        verify_record.write_digests(self.root)
        self.rejects("lie outside the selected documents")

    def test_changed_file_is_rejected(self):
        write_rows(self.root / "items/dev.jsonl.gz", [document("dev-1", "owner/other")])
        self.rejects("digest mismatch")

    def test_unlisted_file_is_rejected(self):
        (self.root / "audit").mkdir()
        (self.root / "audit/extra.json").write_text("[]")
        self.rejects("data files differ")

    def test_page_text_is_rejected(self):
        self.rewrite("items/dev.jsonl.gz", lambda rows: rows[0].update(text="A sentence from the page."))
        verify_record.write_digests(self.root)
        self.rejects("unexpected document fields")

    def test_quote_in_review_is_rejected(self):
        quote = "This long sentence was copied from the page under review."
        self.rewrite("reviews/round.jsonl.gz", lambda rows: rows[0].update(phrase=quote))
        verify_record.write_digests(self.root)
        self.rejects("carries text")

    def test_unbound_flag_is_rejected(self):
        self.rewrite("labels/dev.jsonl.gz", lambda rows: rows[0]["flags"][0].update(item="i999"))
        verify_record.write_digests(self.root)
        self.rejects("unknown item")

    def test_reviewer_outside_panel_is_rejected(self):
        self.rewrite("reviews/round.jsonl.gz", lambda rows: rows[0]["reviews"].update(human="yes"))
        verify_record.write_digests(self.root)
        self.rejects("outside the panel")

    def test_flag_with_quote_is_rejected(self):
        self.rewrite("contextual/claude.jsonl.gz", lambda rows: rows[0].update(quote="A sentence from the page."))
        verify_record.write_digests(self.root)
        self.rejects("unexpected flag fields")

    def test_flag_judged_by_its_caller_is_rejected(self):
        self.rewrite("contextual/claude.jsonl.gz", lambda rows: rows[0]["reviews"].update(claude="yes"))
        verify_record.write_digests(self.root)
        self.rejects("judged by its own caller")

    def test_human_role_is_rejected(self):
        summary = json.loads((self.root / "summary.json").read_text())
        summary["panel"]["astra"]["roles"].append("human")
        self.seal(summary)
        self.rejects("cannot hold a human role")

    def test_shared_repository_is_rejected(self):
        write_rows(self.root / "items/test.jsonl.gz", [document("test-1", "Owner/Dev")])
        self.seal()
        self.rejects("appears in dev and test")

    def test_wrong_summary_is_rejected(self):
        summary = json.loads((self.root / "summary.json").read_text())
        summary["rounds"][0]["rules"]["filler.wordy-phrase"]["both_accept"] = 2
        self.seal(summary)
        self.rejects("summary differs")


if __name__ == "__main__":
    unittest.main()
