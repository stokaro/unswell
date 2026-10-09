"""Regression checks for partial labels, UTF-8 coordinates, and source boundaries."""
import unittest

from annotation_contract import INTENTS, normalize_review


def required_edit(quote, intents):
    return {"necessity": {"status": "necessary", "reason": "Remove redundant framing."},
            "edits": [{"quote": quote, "intents": intents, "status": "necessary",
                       "reason": "The phrase adds no technical condition.", "replacement": ""}]}


class ContractRegressionTests(unittest.TestCase):
    def test_primary_purpose_does_not_invent_other_negatives(self):
        result = normalize_review(required_edit("in order to", ["wordiness"]), source="Use it in order to test.")
        self.assertEqual(1, result["intent_loss_labels"]["wordiness"])
        for name in set(INTENTS) - {"wordiness"}:
            self.assertIsNone(result["intent_loss_labels"][name])
            self.assertFalse(result["intent_loss_mask"][name])

    def test_one_edit_can_have_multiple_purposes(self):
        result = normalize_review(required_edit("This page explains", ["wordiness", "empty_framing"]),
                                  source="This page explains the flags.")
        self.assertEqual(1, result["intent_loss_labels"]["wordiness"])
        self.assertEqual(1, result["intent_loss_labels"]["empty_framing"])
        self.assertEqual(1, len(result["edits"]))

    def test_explicit_negative_requires_a_reason(self):
        with self.assertRaises(ValueError):
            normalize_review({"intents": {"wordiness": {"status": "not_necessary"}}}, source="Required condition.")
        result = normalize_review({"intents": {"wordiness": {"status": "not_necessary",
                                                              "reason": "Every clause states a required condition."}}},
                                  source="Required condition.")
        self.assertEqual(0, result["intent_loss_labels"]["wordiness"])
        self.assertIsNone(result["necessity"]["loss_label"])

    def test_edited_text_is_not_automatically_clean(self):
        result = normalize_review({}, source="The successfully repaired text.")
        self.assertIsNone(result["necessity"]["loss_label"])
        self.assertTrue(all(value is None for value in result["intent_loss_labels"].values()))
        self.assertFalse(result["automatic_after_label"])

    def test_unit_retention_does_not_invent_each_intent_negative(self):
        result = normalize_review({"necessity": {"status": "retain", "reason": "No mandatory edit found."}},
                                  source="A condition that should stay.")
        self.assertEqual(0, result["necessity"]["loss_label"])
        self.assertTrue(all(value is None for value in result["intent_loss_labels"].values()))
        with self.assertRaises(ValueError):
            normalize_review({"necessity": {"status": "retain", "reason": None}}, source="A condition.")

    def test_optional_edits_supply_no_negative_or_positive_loss(self):
        response = {"necessity": {"status": "optional", "reason": "Either form is acceptable."},
                    "edits": [{"quote": "however", "intents": ["wordiness"], "status": "optional",
                               "reason": "Deletion is a preference.", "replacement": ""}]}
        result = normalize_review(response, source="It works; however, it needs a token.")
        self.assertIsNone(result["necessity"]["loss_label"])
        self.assertIsNone(result["intent_loss_labels"]["wordiness"])

    def test_repeated_unicode_quote_uses_original_byte_coordinates(self):
        response = required_edit("é", ["fluency"])
        with self.assertRaises(ValueError):
            normalize_review(response, source="é then é", source_start=100)
        response["edits"][0]["occurrence"] = 1
        result = normalize_review(response, source="é then é", source_start=100)
        self.assertEqual({"start": 108, "end": 110}, result["edits"][0]["span"])

    def test_protected_source_cannot_be_edited(self):
        with self.assertRaises(ValueError):
            normalize_review(required_edit("token", ["wordiness"]), source="The `token` field.",
                             protected=[{"start": 4, "end": 11}])

    def test_conflicting_positive_and_negative_are_rejected(self):
        response = required_edit("in order to", ["wordiness"])
        response["intents"] = {"wordiness": {"status": "not_necessary", "reason": "Retain it."}}
        with self.assertRaises(ValueError):
            normalize_review(response, source="Use it in order to test.")

    def test_required_and_optional_edits_of_one_intent_can_coexist(self):
        response = required_edit("These is", ["fluency"])
        response["edits"][0]["replacement"] = "These are"
        other = {"quote": "however", "intents": ["fluency"], "status": "optional",
                 "reason": "Either transition is acceptable.", "replacement": "still"}
        for first in (True, False):
            with self.subTest(optional_first=first):
                ordered = dict(response, edits=([other] + response["edits"] if first
                                               else response["edits"] + [other]))
                result = normalize_review(ordered, source="These is useful; however, they need a token.")
                self.assertEqual(1, result["intent_loss_labels"]["fluency"])
                self.assertEqual(2, len(result["edits"]))

    def test_unlocalized_mandatory_decisions_are_rejected(self):
        for response in ({"necessity": {"status": "necessary", "reason": "Improve it."}},
                         {"intents": {"clarity": {"status": "necessary", "reason": "Unclear."}}}):
            with self.subTest(response=response), self.assertRaises(ValueError):
                normalize_review(response, source="An unsupported mandatory decision.")

    def test_overlapping_targets_and_unchanged_replacements_are_rejected(self):
        response = required_edit("in order to", ["wordiness"])
        response["edits"].append({"quote": "order", "intents": ["fluency"],
                                   "status": "necessary", "replacement": None,
                                   "reason": "An overlapping target."})
        with self.assertRaises(ValueError):
            normalize_review(response, source="Use it in order to test.")
        unchanged = required_edit("test", ["wordiness"])
        unchanged["edits"][0]["replacement"] = "test"
        with self.assertRaises(ValueError):
            normalize_review(unchanged, source="test")

    def test_invalid_model_output_fails_as_validation_error(self):
        for response in ({"necessity": {"status": []}},
                         {"necessity": {"status": "unknown", "reason": {}}},
                         {"intents": {"clarity": {"status": False}}},
                         {"edits": [{"quote": "test", "intents": ["clarity"],
                                     "status": {}, "reason": "Invalid status."}]},
                         {"authorship_probability": 0.9}):
            with self.subTest(response=response), self.assertRaises(ValueError):
                normalize_review(response, source="test")

    def test_invalid_protected_coordinates_cannot_disable_protection(self):
        for spans in ([{"start": 7, "end": 4}], [{"start": False, "end": 8}],
                      [{"start": "4", "end": 8}], {"start": 4, "end": 8}):
            with self.subTest(spans=spans), self.assertRaises(ValueError):
                normalize_review({}, source="A `code` span.", protected=spans)


if __name__ == "__main__":
    unittest.main()
