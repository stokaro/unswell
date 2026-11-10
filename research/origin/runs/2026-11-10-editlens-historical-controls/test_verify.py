"""Check the independent historical-control verifier against corrupted evidence."""

from __future__ import annotations

import copy
import importlib.util
import json
from pathlib import Path
import unittest


HERE = Path(__file__).resolve().parent


def load(name: str):
    spec = importlib.util.spec_from_file_location(name, HERE / (name + ".py"))
    assert spec is not None and spec.loader is not None
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


run = load("run")
verify = load("verify")


def fixture():
    frame = json.loads((HERE / "frame.json").read_text())
    docs = [{**row, "removed_header": False, "score": None, "flagged": None}
            for row in frame["selected"]]
    windows = [{"source_id": row["source_id"], "index": index,
                "tokens": 3, "score": 0.1}
               for row in frame["selected"]
               for index in range(row["windows_100"])]
    selection = "synthetic-selection"
    result = run.summarize(docs, windows, selection)
    freeze = {"model_sha256": run.MODEL_SHA256,
              "frame_sha256": run.sha(HERE / "frame.json"),
              "selection_sha256": selection}
    return frame, freeze, docs, windows, result


class HistoricalVerifierTests(unittest.TestCase):
    def test_complete_zero_flag_evidence(self) -> None:
        frame, freeze, docs, windows, result = fixture()
        self.assertEqual(verify.verify_payload(frame, freeze, docs, windows, result),
                         {"documents": 462, "repositories": 153,
                          "windows": 3735, "flagged": 0,
                          "control_transfer_pass": True})

    def test_rejects_changed_window_score(self) -> None:
        frame, freeze, docs, windows, result = fixture()
        windows[0]["score"] = 0.9
        with self.assertRaisesRegex(ValueError, "document mean"):
            verify.verify_payload(frame, freeze, docs, windows, result)

    def test_rejects_missing_window(self) -> None:
        frame, freeze, docs, windows, result = fixture()
        windows.pop()
        with self.assertRaisesRegex(ValueError, "Control coverage changed"):
            verify.verify_payload(frame, freeze, docs, windows, result)

    def test_rejects_forged_source_identity(self) -> None:
        frame, freeze, docs, windows, result = fixture()
        docs[0]["source_sha256"] = "0" * 64
        with self.assertRaisesRegex(ValueError, "source metadata"):
            verify.verify_payload(frame, freeze, docs, windows, result)

    def test_rejects_forged_gate(self) -> None:
        frame, freeze, docs, windows, result = fixture()
        result["control_transfer_pass"] = False
        with self.assertRaisesRegex(ValueError, "control_transfer_pass"):
            verify.verify_payload(frame, freeze, docs, windows, result)


if __name__ == "__main__":
    unittest.main()
