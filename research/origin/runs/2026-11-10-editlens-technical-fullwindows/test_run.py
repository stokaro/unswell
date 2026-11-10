"""Check the frozen full-window aggregation and failure paths."""

from __future__ import annotations

import importlib.util
import unittest
from pathlib import Path


RUNNER = Path(__file__).with_name("run.py")
SPEC = importlib.util.spec_from_file_location("fullwindow_runner", RUNNER)
assert SPEC is not None and SPEC.loader is not None
run = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(run)


def fixture() -> tuple[list[dict], list[dict]]:
    docs = []
    windows = []
    for source_index in range(38):
        task = f"task-{source_index:02d}"
        repo = f"repo-{source_index % 20:02d}"
        source_key = f"source/{task}"
        docs.append({"kind": "source", "key": source_key, "task_id": task,
                     "repository": repo, "cleaned_words": 100, "windows": 2})
        windows.extend(({"key": source_key, "index": index, "score": score})
                       for index, score in enumerate((0.1, 0.3)))
        for family in run.FAMILIES:
            for operation in run.OPERATIONS:
                for prompt in ("neutral", "plain"):
                    key = "/".join(("response", task, family, operation, prompt))
                    docs.append({"kind": "response", "key": key, "task_id": task,
                                 "repository": repo, "family": family,
                                 "operation": operation, "prompt": prompt,
                                 "cleaned_words": 100, "windows": 2})
                    windows.extend(({"key": key, "index": index, "score": score})
                                   for index, score in enumerate((0.4, 0.6)))
    return docs, windows


class FullWindowTests(unittest.TestCase):
    def test_document_mean_and_paired_accounting(self) -> None:
        docs, windows = fixture()
        result, pairs = run.summarize(docs, windows, "identity", "selection")
        self.assertEqual((result["sources"], result["responses"]), (38, 456))
        self.assertEqual((result["source_windows"], result["response_windows"]),
                         (76, 912))
        self.assertEqual(len(pairs), 456)
        self.assertAlmostEqual(pairs[0]["source_score"], 0.2)
        self.assertAlmostEqual(pairs[0]["response_score"], 0.5)
        self.assertAlmostEqual(pairs[0]["delta"], 0.3)
        for operation in run.OPERATIONS:
            self.assertEqual(result["operations"][operation]["overall"]["pairs"], 228)
            self.assertTrue(result["operations"][operation]["direction_pass"])

    def test_missing_window_rejected(self) -> None:
        docs, windows = fixture()
        windows.pop()
        with self.assertRaisesRegex(ValueError, "Missing or repeated window"):
            run.summarize(docs, windows, "identity", "selection")

    def test_duplicate_window_rejected(self) -> None:
        docs, windows = fixture()
        windows.append(dict(windows[-1]))
        with self.assertRaisesRegex(ValueError, "Missing or repeated window"):
            run.summarize(docs, windows, "identity", "selection")

    def test_document_key_rejects_unknown_kind(self) -> None:
        with self.assertRaisesRegex(ValueError, "Unknown document kind"):
            run.doc_key({"kind": "other"})


if __name__ == "__main__":
    unittest.main()
