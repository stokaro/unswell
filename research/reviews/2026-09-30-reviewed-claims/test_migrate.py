"""Negative checks for the explicit local migration review boundary."""

import copy
import importlib.util
import json
import os
from pathlib import Path
import tempfile
import unittest

SPEC = importlib.util.spec_from_file_location("migration", Path(__file__).with_name("migrate.py"))
M = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(M)


class MigrationTests(unittest.TestCase):
    def setUp(self):
        self.document = {"source": "The café has two claims.\n\nAn antecedent remains."}
        self.units = [{"block": 0, "span": {"start": 0, "end": 25}, "excluded": False},
                      {"block": 1, "span": {"start": 27, "end": 49}, "excluded": False}]
        self.reference = {"span": {"start": 0, "end": 25}, "quote": "The café has two claims."}
        self.parent = {"id": "parent-1", "diagnostic": "Grammar and rhetoric.", "targets": [self.reference]}
        self.claim = {"key": "rhetoric", "category": "wording", "diagnostic": "Review this framing.",
                      "reason": "Preserve the independently declared rhetorical claim.",
                      "suggestion": "Revise the framing while retaining the fact.", "targets": [self.reference], "support": []}
        self.decision = {"parent": "parent-1", "parent_sha256": M.digest(self.parent), "status": "migrated",
                         "reason": "The grammar and rhetoric are independently actionable.", "claims": [self.claim]}

    def test_parent_and_explicit_claims_survive_independently(self):
        before = copy.deepcopy(self.parent)
        other = copy.deepcopy(self.claim)
        other["key"] = "grammar"
        self.decision["claims"].append(other)
        claims = M.declarations(self.document, self.parent, self.decision, self.units, "run")
        self.assertEqual([c["origin"]["key"] for c in claims], ["rhetoric", "grammar"])
        self.assertEqual(self.parent, before)
        self.assertEqual(claims[0]["targets"][0]["span"], self.reference["span"])

    def test_approval_binds_the_whole_original_criticism(self):
        self.parent["diagnostic"] = "Grammar only."
        with self.assertRaisesRegex(ValueError, "complete unchanged original"):
            M.declarations(self.document, self.parent, self.decision, self.units, "run")

    def test_support_can_name_a_retained_antecedent_without_becoming_a_target(self):
        self.claim["support"] = [{"span": {"start": 27, "end": 49}, "quote": "An antecedent remains."}]
        claims = M.declarations(self.document, self.parent, self.decision, self.units, "run")
        self.assertEqual(claims[0]["support"][0]["block"], 1)
        self.assertEqual(claims[0]["targets"][0]["block"], 0)
        self.claim["targets"] = self.claim["support"]
        with self.assertRaisesRegex(ValueError, "escapes the original"):
            M.declarations(self.document, self.parent, self.decision, self.units, "run")

    def test_invented_quoted_bytes_and_split_unicode_fail(self):
        for reference in [{"span": {"start": 0, "end": 25}, "quote": "Invented"},
                          {"span": {"start": 7, "end": 8}, "quote": "é"}]:
            with self.subTest(reference=reference):
                self.claim["targets"] = [reference]
                with self.assertRaises((ValueError, UnicodeDecodeError)):
                    M.declarations(self.document, self.parent, self.decision, self.units, "run")

    def test_ambiguous_or_ineligible_source_owners_fail(self):
        for units in [self.units + [copy.deepcopy(self.units[0])],
                      [{**self.units[0], "excluded": True}]]:
            with self.subTest(units=units):
                with self.assertRaisesRegex(ValueError, "one eligible"):
                    M.declarations(self.document, self.parent, self.decision, units, "run")

    def test_unsupported_statuses_and_silent_claim_removal_fail(self):
        for status, claims in [("accepted", [self.claim]), ("migrated", []), ("rejected", [self.claim]),
                               ("uncertain", [self.claim]), ("migrated", [self.claim, self.claim])]:
            with self.subTest(status=status, claims=claims):
                decision = {**self.decision, "status": status, "claims": claims}
                with self.assertRaises(ValueError):
                    M.declarations(self.document, self.parent, decision, self.units, "run")

    def test_missing_or_unknown_fields_fail(self):
        for field in self.claim:
            broken = copy.deepcopy(self.decision)
            del broken["claims"][0][field]
            with self.subTest(field=field), self.assertRaisesRegex(ValueError, "fields"):
                M.declarations(self.document, self.parent, broken, self.units, "run")
        self.decision["approved"] = True
        with self.assertRaisesRegex(ValueError, "fields"):
            M.declarations(self.document, self.parent, self.decision, self.units, "run")

    def test_unknown_repeated_and_changed_page_review_fails(self):
        manifest = {"_file_sha256": "a" * 64,
                    "pages": [{"page": "p1", "request_sha256": "b" * 64, "candidates": 2}]}
        page = {"page": "p1", "request_sha256": "b" * 64, "decisions": [self.decision]}
        review = {"version": M.VERSION, "packet_manifest_sha256": "a" * 64,
                  "reviewer": {"id": "assistant", "basis": "Constructed control"}, "pages": [page]}
        self.assertEqual(list(M.index_review(review, manifest)), ["p1"])
        for changes in [{"page": "unknown"}, {"request_sha256": "c" * 64},
                        {"decisions": [self.decision, self.decision]}]:
            with self.subTest(changes=changes), self.assertRaises(ValueError):
                M.index_review({**review, "pages": [{**page, **changes}]}, manifest)
        with self.assertRaises(ValueError):
            M.index_review({**review, "pages": [page, page]}, manifest)

    def test_duplicate_json_fields_fail_before_review(self):
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / "review.json"
            path.write_text('{"version":"a","version":"b"}')
            with self.assertRaisesRegex(ValueError, "Duplicate JSON"):
                M.read(path)

    def test_local_evidence_is_not_overwritten(self):
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / "evidence.json"
            M.store(path, {"original": self.parent})
            self.assertEqual(json.loads(path.read_text())["original"], self.parent)
            with self.assertRaises(FileExistsError):
                M.store(path, {"original": "changed"})


@unittest.skipUnless(os.environ.get("UNSWELL_REVIEWCLAIMS_BINARY"), "Build the existing Go consumer for the isolated integration check")
class GoReplayTests(unittest.TestCase):
    def setUp(self):
        self.directory = tempfile.TemporaryDirectory()
        self.addCleanup(self.directory.cleanup)
        self.root = Path(self.directory.name)
        self.packet = self.root / "packet"
        self.packet.mkdir()
        (self.packet / "requests").mkdir()
        self.binary = Path(os.environ["UNSWELL_REVIEWCLAIMS_BINARY"])
        source = "The café retries after timeout.\n\nFour settings decides where `parseURL` sends requests.\n"
        raw = source.encode()
        quote = "Four settings decides where `parseURL` sends requests."
        start = raw.index(quote.encode())
        target = {"span": {"start": start, "end": start + len(quote.encode())}, "quote": quote, "unit": "u1"}
        parent = {"id": "parent-1", "origin": "model", "category": "other_editorial", "targets": [target],
                  "diagnostic": "The count framing and agreement are independent criticisms.",
                  "rationale": "One correction must not erase the other criticism.",
                  "suggestion": "Revise the framing and correct agreement while preserving the request behavior."}
        request = {"document": {"page": "p1", "format": "markdown", "source": source, "source_sha256": M.sha(raw)},
                   "candidates": [parent]}
        request_data = json.dumps(request, ensure_ascii=False, indent=2).encode()
        (self.packet / "requests/p1.json").write_bytes(request_data)
        manifest = {"pages": [{"page": "p1", "source_sha256": M.sha(raw), "candidates": 1,
                               "request_sha256": M.sha(request_data)}]}
        manifest_data = json.dumps(manifest, indent=2).encode()
        (self.packet / "manifest.json").write_bytes(manifest_data)
        M.store(self.packet / "freeze.json", {"manifest.json": M.sha(manifest_data), "requests/p1.json": M.sha(request_data)})

        def reference(phrase):
            offset = raw.index(phrase.encode())
            return {"span": {"start": offset, "end": offset + len(phrase.encode())}, "quote": phrase}

        self.reference = reference
        claims = [{"key": key, "category": key, "diagnostic": "Review the independently declared " + key + " criticism.",
                   "reason": "Preserve this original criticism separately from the other claim.",
                   "suggestion": "Revise only the declared target while preserving request behavior.",
                   "targets": [reference(phrase)], "support": [reference("The café retries after timeout.")]}
                  for key, phrase in [("grammar", "decides"), ("rhetoric", "Four settings")]]
        self.review = {"version": M.VERSION, "packet_manifest_sha256": M.sha(manifest_data),
                       "reviewer": {"id": "constructed assistant", "basis": "Constructed source-role control, not a quality label"},
                       "pages": [{"page": "p1", "request_sha256": M.sha(request_data), "decisions": [{
                           "parent": "parent-1", "parent_sha256": M.digest(parent), "status": "migrated",
                           "reason": "Keep grammar and rhetoric separate, the antecedent retained, and the identifier untouched.",
                           "claims": claims}]}]}

    def execute(self, name, review):
        path = self.root / (name + "-review.json")
        M.store(path, review)
        return M.run(self.packet, self.binary, path, self.root / name)

    def test_real_go_replay_keeps_independent_claims_and_stable_ids(self):
        first = self.execute("first", self.review)
        self.assertTrue(first["complete"])
        self.assertEqual(first["original_candidates"], 1)
        self.assertEqual(first["bound_child_claims"], 2)
        changed = copy.deepcopy(self.review)
        changed["reviewer"]["basis"] = "A later record preserves these same explicit original declarations."
        second = self.execute("second", changed)
        a = M.read(self.root / "first/p1.json")
        b = M.read(self.root / "second/p1.json")
        self.assertEqual(a["account"]["claims"], b["account"]["claims"])
        self.assertNotEqual(first["review_sha256"], second["review_sha256"])
        self.assertNotEqual(a["account"]["claims"][0]["id"], a["account"]["claims"][1]["id"])
        self.assertFalse(a["account"]["editorial_qualified"])
        self.assertEqual(a["parents"][0]["original"]["targets"][0]["quote"],
                         "Four settings decides where `parseURL` sends requests.")

    def test_unreviewed_and_uncertain_parents_cannot_pass(self):
        omitted = {**self.review, "pages": []}
        summary = self.execute("missing", omitted)
        self.assertFalse(summary["complete"])
        self.assertEqual(summary["parent_dispositions"], {"unreviewed": 1})
        self.assertEqual(summary["unaccounted"], 0)
        uncertain = copy.deepcopy(self.review)
        uncertain["pages"][0]["decisions"][0].update(status="uncertain", claims=[])
        summary = self.execute("uncertain", uncertain)
        self.assertFalse(summary["complete"])
        self.assertEqual(summary["parent_dispositions"], {"uncertain": 1})

    def test_go_rejects_protected_targets_and_mislabeled_context(self):
        for name, edit in [("protected", lambda c: c.update(targets=[self.reference("parseURL")])),
                           ("overlapping-context", lambda c: c.update(support=c["targets"]))]:
            with self.subTest(name=name):
                broken = copy.deepcopy(self.review)
                edit(broken["pages"][0]["decisions"][0]["claims"][0])
                with self.assertRaisesRegex(ValueError, "existing Go binding"):
                    self.execute(name, broken)
                self.assertFalse((self.root / name / "summary.json").exists())

    def test_changed_frozen_source_fails_before_binding(self):
        path = self.packet / "requests/p1.json"
        path.write_bytes(path.read_bytes().replace(b"parseURL", b"parseUrl"))
        with self.assertRaisesRegex(ValueError, "Frozen candidate request changed"):
            self.execute("tampered", self.review)
        self.assertFalse((self.root / "tampered/summary.json").exists())

    def test_go_rejects_unordered_explicit_references_without_a_summary(self):
        broken = copy.deepcopy(self.review)
        broken["pages"][0]["decisions"][0]["claims"][0]["targets"] = [
            self.reference("where"), self.reference("decides")]
        with self.assertRaisesRegex(ValueError, "references must be ordered and disjoint"):
            self.execute("unordered", broken)
        self.assertFalse((self.root / "unordered/summary.json").exists())


if __name__ == "__main__":
    unittest.main()
