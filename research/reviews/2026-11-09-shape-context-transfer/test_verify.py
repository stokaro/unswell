import unittest

import verify


class PublicVerificationTest(unittest.TestCase):
    def setUp(self):
        self.freeze = verify.read("freeze.json")
        self.sources = verify.read("sources.json")
        self.summary = verify.read("candidate-summary.json")
        self.identities = verify.read("candidate-ids.json")
        self.review = verify.read("design-review.json")
        self.result = verify.read("result.json")

    def check(self):
        return verify.check_accounting(self.freeze, self.sources, self.summary,
                                       self.identities, self.review, self.result)

    def test_current_result(self):
        self.assertEqual(self.check(), {"retain": 10, "uncertain": 1})

    def test_rejects_false_qualification(self):
        self.result["qualified"] = True
        with self.assertRaisesRegex(ValueError, "stop claim"):
            self.check()

    def test_rejects_missing_review_row(self):
        self.review["rows"].pop()
        with self.assertRaisesRegex(ValueError, "omits"):
            self.check()

    def test_rejects_inflated_warranted_count(self):
        self.result["design_dispositions"]["warranted"] = 4
        with self.assertRaisesRegex(ValueError, "stop claim"):
            self.check()

    def test_rejects_moved_candidate(self):
        self.identities["design"][0]["id"] = "page-007/negation-frame/9238-9245"
        with self.assertRaisesRegex(ValueError, "source binding"):
            self.check()


if __name__ == "__main__":
    unittest.main()
