"""Constructed blackbox checks against the actual Go claim-accounting command."""

import copy
import importlib.util
import json
import os
from pathlib import Path
import subprocess
import tempfile
import unittest

SPEC = importlib.util.spec_from_file_location("pipeline", Path(__file__).with_name("pipeline.py"))
P = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(P)
M = P.M


@unittest.skipUnless(os.environ.get("UNSWELL_REVIEWCLAIMS_BINARY"), "Build the existing Go consumer for the isolated integration check")
class PipelineTests(unittest.TestCase):
    def setUp(self):
        self.directory = tempfile.TemporaryDirectory()
        self.addCleanup(self.directory.cleanup)
        self.root = Path(self.directory.name)
        self.binary = Path(os.environ["UNSWELL_REVIEWCLAIMS_BINARY"])
        self.packet = self.root / "packet"
        (self.packet / "requests").mkdir(parents=True)
        source = "The café retries only after timeout.\n\nFour settings decides where `parseURL` sends requests.\n"
        raw = source.encode()

        def reference(phrase):
            offset = raw.index(phrase.encode())
            return {"span": {"start": offset, "end": offset + len(phrase.encode())}, "quote": phrase}

        self.reference = reference
        parent = {"id": "parent-1", "origin": "model", "category": "other_editorial",
                  "targets": [reference("Four settings decides where `parseURL` sends requests.")],
                  "diagnostic": "The agreement and rhetorical count framing are independent criticisms.",
                  "rationale": "Correcting agreement does not make the count framing useful.",
                  "suggestion": "Preserve the request behavior while checking both criticisms."}
        request = {"document": {"page": "p1", "format": "markdown", "source": source, "source_sha256": M.sha(raw)},
                   "candidates": [parent]}
        request_data = json.dumps(request, ensure_ascii=False, indent=2).encode()
        (self.packet / "requests/p1.json").write_bytes(request_data)
        manifest_data = json.dumps({"pages": [{"page": "p1", "source_sha256": M.sha(raw), "candidates": 1,
                                              "request_sha256": M.sha(request_data)}]}).encode()
        (self.packet / "manifest.json").write_bytes(manifest_data)
        M.store(self.packet / "freeze.json", {"manifest.json": M.sha(manifest_data), "requests/p1.json": M.sha(request_data)})
        claims = [{"key": key, "category": key, "diagnostic": "Review this " + key + " criticism.",
                   "reason": "Keep this original criticism separate from the other claim.",
                   "suggestion": "Revise the declared target while keeping the request condition.",
                   "targets": [reference(phrase)], "support": [reference("The café retries only after timeout.")]}
                  for key, phrase in [("grammar", "decides"), ("rhetoric", "Four settings decides where")]]
        review = {"version": M.VERSION, "packet_manifest_sha256": M.sha(manifest_data),
                  "reviewer": {"id": "constructed assistant", "basis": "Constructed claim-role test; not an editorial corpus label"},
                  "pages": [{"page": "p1", "request_sha256": M.sha(request_data), "decisions": [{
                      "parent": "parent-1", "parent_sha256": M.digest(parent), "status": "migrated",
                      "reason": "Agreement, framing, and supporting context have separate roles.", "claims": claims}]}]}
        self.review_path = self.root / "review.json"
        M.store(self.review_path, review)
        self.prepared = self.root / "prepared"
        self.preparation = P.prepare(self.packet, self.binary, self.review_path, self.prepared)
        self.stages = M.read(self.prepared / "stages.json")
        self.approvals = M.read(self.prepared / "approvals.json")
        self.inventory = M.read(self.prepared / "migration/p1.json")["account"]["claims"]
        self.ids = [c["id"] for c in self.inventory]
        self.grammar_edit = {"id": "agreement", "target": self.inventory[0]["targets"][0], "replacement": "decide"}

    def retained(self):
        result = copy.deepcopy(self.stages)
        for stage in result["pages"][0]["stages"]:
            for decision in stage["decisions"]:
                decision.update(status="retained", reason="Preserve this constructed criticism for the next step.")
        return result

    def resolution(self, stage, claim_id=None, edit=None):
        edit = edit or self.grammar_edit
        claim_id = claim_id or self.ids[0]
        # Match the documented Go Edit field order for this ASCII-only fixture.
        digest = M.sha(json.dumps(edit, ensure_ascii=False, separators=(",", ":")).encode())
        return {"stage_id": stage, "kind": "resolution", "claims": [claim_id], "edit_digest": digest,
                "reviewer": "constructed edit reviewer", "reason": "Only the agreement defect is fixed; the framing remains.",
                "meaning_preserved": True}

    def corrected(self):
        stages, approvals = self.retained(), copy.deepcopy(self.approvals)
        for stage in stages["pages"][0]["stages"][1:]:
            stage["decisions"][0].update(status="resolved", edit_id="agreement", reason="Correct agreement without resolving framing.")
            stage["edits"] = [self.grammar_edit]
            approvals["pages"][0]["approvals"].append(self.resolution(stage["id"]))
        return stages, approvals

    def execute(self, stages, approvals=None, name="replay", bind_approvals=True):
        stage_path, approval_path = self.root / (name + "-stages.json"), self.root / (name + "-approvals.json")
        M.store(stage_path, stages)
        approvals = copy.deepcopy(approvals if approvals is not None else self.approvals)
        if bind_approvals:
            approvals["stages_sha256"] = M.sha(stage_path.read_bytes())
        M.store(approval_path, approvals)
        return P.replay(self.packet, self.binary, self.review_path, stage_path, approval_path, self.root / name)

    def fail(self, stages, pattern, approvals=None, bind_approvals=True):
        with self.assertRaisesRegex(ValueError, pattern):
            self.execute(stages, approvals, bind_approvals=bind_approvals)
        self.assertFalse((self.root / "replay/summary.json").exists())

    def test_preparation_is_pending_and_template_replay_exits_two(self):
        self.assertFalse(self.preparation["complete"])
        self.assertTrue(self.preparation["prepared"])
        result = self.execute(self.stages)
        self.assertFalse(result["complete"])
        self.assertEqual(result["stage_dispositions"], {s: {"uncertain": 2} for s in P.STAGES})
        self.assertEqual(result["selected_displayed_claims"], 0)
        command = ["python3", "-B", str(Path(P.__file__)), "replay", "--packet", str(self.packet),
                   "--binary", str(self.binary), "--review", str(self.review_path),
                   "--stages", str(self.root / "replay-stages.json"), "--approvals", str(self.root / "replay-approvals.json"),
                   "--output", str(self.root / "cli")]
        process = subprocess.run(command, capture_output=True, timeout=30, check=False)
        self.assertEqual(process.returncode, 2, process.stderr.decode())
        self.assertFalse(json.loads(process.stdout)["complete"])

    def test_all_original_claims_and_roles_survive_every_stage(self):
        result = self.execute(self.retained())
        self.assertTrue(result["complete"])
        self.assertEqual(result["selected_displayed_claims"], 2)
        self.assertFalse(result["editorial_qualified"])
        self.assertFalse(result["full_event_metrics_recomputed"])
        page = M.read(self.root / "replay/p1.json")
        self.assertEqual(page["account"]["claims"], self.inventory)
        for account in page["account"]["stages"]:
            self.assertEqual(account["claims"], self.inventory)
            self.assertEqual(account["displayed_claims"], self.ids)
        self.assertIn("`parseURL`", page["parents_by_stage"][0]["parents"][0]["original"]["targets"][0]["quote"])
        self.assertEqual(self.inventory[1]["support"][0]["quote"], "The café retries only after timeout.")
        self.assertEqual((self.root / "replay/p1.json").stat().st_mode & 0o777, 0o600)

    def test_grammar_edit_leaves_rhetoric_and_mixed_parent_visible(self):
        stages, approvals = self.corrected()
        result = self.execute(stages, approvals)
        self.assertTrue(result["complete"])
        self.assertEqual(result["selected_parent_dispositions"], {"mixed": 1})
        self.assertEqual(result["selected_displayed_claims"], 1)
        page = M.read(self.root / "replay/p1.json")
        self.assertEqual(page["account"]["stages"][-1]["displayed_claims"], [self.ids[1]])
        self.assertEqual([c["status"] for c in page["parents_by_stage"][-1]["parents"][0]["children"]], ["resolved", "retained"])

    def test_one_approved_edit_cannot_resolve_a_second_criticism(self):
        stages, approvals = self.corrected()
        stages["pages"][0]["stages"][-1]["decisions"][1].update(status="resolved", edit_id="agreement")
        self.fail(stages, "claim-specific review", approvals)

    def test_resolutions_require_a_separate_approval_file_and_exact_edit(self):
        stages, approvals = self.corrected()
        approvals["pages"][0]["approvals"][0]["edit_digest"] = "0" * 64
        self.fail(stages, "claim-specific review", approvals)

    def test_supporting_context_cannot_become_an_edit_target(self):
        stages, approvals = self.corrected()
        for stage in stages["pages"][0]["stages"][1:]:
            stage["edits"][0] = {"id": "agreement", "target": self.inventory[0]["support"][0], "replacement": "An antecedent changed."}
        self.fail(stages, "original targets", approvals)

    def test_protected_identifier_edit_fails_in_existing_go_extraction(self):
        stages, approvals = self.corrected()
        target = {"block": self.inventory[0]["targets"][0]["block"], **self.reference("parseURL")}
        stages["pages"][0]["stages"][1]["edits"][0] = {"id": "agreement", "target": target, "replacement": "parseUrl"}
        self.fail(stages, "protected", approvals)

    def test_missing_claim_and_invented_identity_fail(self):
        for edit in [lambda s: s["decisions"].pop(), lambda s: s["decisions"][0].update(claim_id="invented")]:
            with self.subTest(edit=edit):
                stages = self.retained()
                edit(stages["pages"][0]["stages"][0])
                with self.assertRaisesRegex(ValueError, "decision"):
                    self.execute(stages, name="missing" if len(stages["pages"][0]["stages"][0]["decisions"]) == 1 else "invented")

    def test_earlier_uncertainty_cannot_be_hidden_by_complete_selection(self):
        stages = self.retained()
        stages["pages"][0]["stages"][0]["decisions"][1].update(status="uncertain")
        result = self.execute(stages)
        self.assertFalse(result["complete"])
        self.assertEqual(result["stage_dispositions"]["selection"], {"retained": 2})
        self.assertEqual(result["stage_dispositions"]["audit"], {"retained": 1, "uncertain": 1})

    def test_missing_and_reordered_stages_fail(self):
        for name, operation in [("missing", lambda stages: stages.pop()),
                                ("reordered", lambda stages: stages.reverse())]:
            with self.subTest(name=name):
                stages = self.retained()
                operation(stages["pages"][0]["stages"])
                with self.assertRaisesRegex(ValueError, "required|appear as"):
                    self.execute(stages, name=name)

    def test_missing_page_changed_inventory_and_changed_request_fail(self):
        for name, operation in [("missing", lambda value: value.update(pages=[])),
                                ("claims", lambda value: value["pages"][0].update(claims_sha256="0" * 64)),
                                ("request", lambda value: value["pages"][0].update(request_sha256="0" * 64))]:
            with self.subTest(name=name):
                stages = self.retained()
                operation(stages)
                with self.assertRaisesRegex(ValueError, "page|inventory"):
                    self.execute(stages, name=name)

    def test_stage_cannot_embed_its_own_approvals(self):
        stages = self.retained()
        stages["pages"][0]["stages"][0]["approvals"] = []
        self.fail(stages, "fields")
        with self.assertRaisesRegex(ValueError, "own approvals"):
            P.replay(self.packet, self.binary, self.review_path, self.prepared / "stages.json",
                     self.prepared / "stages.json", self.root / "same-file")

    def test_duplicate_display_needs_a_separate_same_defect_verdict(self):
        stages = self.retained()
        stages["pages"][0]["stages"][-1]["duplicates"] = [{"claims": self.ids, "representative": self.ids[0],
                                                            "reason": "A deliberately incorrect overlap-only grouping."}]
        self.fail(stages, "same-defect review")

    def test_reviewed_same_defect_group_keeps_every_original(self):
        review = M.read(self.review_path)
        duplicate = copy.deepcopy(review["pages"][0]["decisions"][0]["claims"][0])
        duplicate["key"] = "second-observation-of-agreement"
        review["pages"][0]["decisions"][0]["claims"].append(duplicate)
        review_path = self.root / "duplicate-review.json"
        M.store(review_path, review)
        prepared = self.root / "duplicate-prepared"
        P.prepare(self.packet, self.binary, review_path, prepared)
        stages, approvals = M.read(prepared / "stages.json"), M.read(prepared / "approvals.json")
        claims = M.read(prepared / "migration/p1.json")["account"]["claims"]
        for stage in stages["pages"][0]["stages"]:
            for decision in stage["decisions"]:
                decision.update(status="retained", reason="Retain this constructed observation.")
        members = [claims[0]["id"], claims[2]["id"]]
        stages["pages"][0]["stages"][-1]["duplicates"] = [{"claims": members, "representative": members[0],
                                                            "reason": "Both observations name this same agreement defect."}]
        approvals["pages"][0]["approvals"] = [{"stage_id": "selection", "kind": "same_defect", "claims": members,
                                               "edit_digest": "", "reviewer": "constructed reviewer",
                                               "reason": "The agreement criticisms match; rhetoric remains independent.",
                                               "meaning_preserved": False}]
        M.store(prepared / "reviewed-stages.json", stages)
        approvals["stages_sha256"] = M.sha((prepared / "reviewed-stages.json").read_bytes())
        M.store(prepared / "reviewed-approvals.json", approvals)
        result = P.replay(self.packet, self.binary, review_path, prepared / "reviewed-stages.json",
                          prepared / "reviewed-approvals.json", self.root / "duplicates")
        self.assertEqual(result["bound_child_claims"], 3)
        self.assertEqual(result["selected_displayed_claims"], 2)
        page = M.read(self.root / "duplicates/p1.json")
        self.assertEqual(page["account"]["claims"], claims)
        self.assertEqual(page["account"]["stages"][-1]["displayed_claims"], [claims[0]["id"], claims[1]["id"]])

    def test_rejected_child_keeps_its_original_and_explicit_reason(self):
        stages = self.retained()
        stages["pages"][0]["stages"][-1]["decisions"][1].update(status="rejected", reason="A constructed selection decline.")
        result = self.execute(stages)
        self.assertTrue(result["complete"])
        self.assertEqual(result["selected_parent_dispositions"], {"mixed": 1})
        page = M.read(self.root / "replay/p1.json")
        self.assertEqual(page["account"]["claims"], self.inventory)
        child = page["parents_by_stage"][-1]["parents"][0]["children"][1]
        self.assertEqual(child["reason"], "A constructed selection decline.")
        self.assertFalse(child["displayed"])

    def test_scope_declined_parent_remains_distinct_from_editorial_rejection(self):
        review = M.read(self.review_path)
        review["pages"][0]["decisions"][0].update(status="rejected", claims=[], reason="Constructed migration scope decline.")
        path = self.root / "scope-review.json"
        M.store(path, review)
        prepared = self.root / "scope-prepared"
        P.prepare(self.packet, self.binary, path, prepared)
        result = P.replay(self.packet, self.binary, path, prepared / "stages.json", prepared / "approvals.json", self.root / "scope")
        self.assertTrue(result["complete"])
        self.assertEqual(result["selected_parent_dispositions"], {"migration_rejected": 1})
        page = M.read(self.root / "scope/p1.json")
        parent = page["parents_by_stage"][-1]["parents"][0]
        self.assertEqual(parent["original"]["id"], "parent-1")
        self.assertEqual(parent["migration"]["reason"], "Constructed migration scope decline.")
        self.assertFalse(result["editorial_qualified"])

    def test_changed_stage_file_cannot_reuse_old_approvals(self):
        self.fail(self.retained(), "exact supplied stage file", bind_approvals=False)

    def test_unreviewed_migration_cannot_start_downstream_accounting(self):
        review = M.read(self.review_path)
        review["pages"] = []
        path = self.root / "unreviewed.json"
        M.store(path, review)
        with self.assertRaisesRegex(ValueError, "explicit migration decision"):
            P.prepare(self.packet, self.binary, path, self.root / "unreviewed")
        self.assertFalse((self.root / "unreviewed/preparation.json").exists())

    def test_outputs_are_exclusive_and_duplicate_json_keys_fail(self):
        with self.assertRaises(FileExistsError):
            P.prepare(self.packet, self.binary, self.review_path, self.prepared)
        path = self.root / "duplicated.json"
        path.write_text('{"pages":[],"pages":[]}')
        with self.assertRaisesRegex(ValueError, "Duplicate JSON"):
            P.read_hashed(path)


if __name__ == "__main__":
    unittest.main()
