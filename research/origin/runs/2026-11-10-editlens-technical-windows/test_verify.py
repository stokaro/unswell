"""Negative controls for the technical-window evidence verifier."""

import copy
import importlib.util
import json
from pathlib import Path
import re
import unittest


HERE = Path(__file__).resolve().parent


def load(name):
    spec = importlib.util.spec_from_file_location(name, HERE / f"{name}.py")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


class WindowEvidenceTest(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.audit = load("audit")
        cls.verify = load("verify")
        cls.result = json.loads((HERE / "result.json").read_text())

    def test_word_windows_cover_words_and_fold_short_tail(self):
        text = " ".join(f"word{index}" for index in range(224))
        windows = self.audit.word_windows(text, re.compile(r"word\d+"), 100)
        self.assertEqual([len(re.findall(r"word\d+", window)) for window in windows],
                         [100, 124])
        self.assertEqual("".join(windows), text)
        self.assertEqual(self.audit.word_windows("only a few words", re.compile(r"\w+"), 100), [])

    def test_reject_corrupt_summary(self):
        altered = copy.deepcopy(self.result)
        altered["full_coverage"]["100"]["source_complete"] -= 1
        with self.assertRaisesRegex(ValueError, "aggregate changed"):
            self.verify.check(altered)

    def test_reject_duplicate_response(self):
        altered = copy.deepcopy(self.result)
        response = next(index for index, row in enumerate(altered["rows"])
                        if row["kind"] == "response")
        altered["rows"][response + 1] = altered["rows"][response]
        with self.assertRaisesRegex(ValueError, "Duplicate response"):
            self.verify.check(altered)

    def test_reject_false_complete(self):
        altered = copy.deepcopy(self.result)
        altered["rows"][0]["full_coverage"]["100"]["over_cap"] = 1
        with self.assertRaisesRegex(ValueError, "False complete"):
            self.verify.check(altered)

    def test_reject_raw_text(self):
        altered = copy.deepcopy(self.result)
        altered["rows"][0]["text"] = "private source"
        with self.assertRaisesRegex(ValueError, "Raw text"):
            self.verify.check(altered)


if __name__ == "__main__":
    unittest.main()
