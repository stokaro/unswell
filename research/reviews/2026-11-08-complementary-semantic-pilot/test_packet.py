"""Check frozen-packet and source-binding rejection paths."""

import importlib.util
import json
from pathlib import Path
import tempfile
import unittest
from unittest import mock

ROOT = Path(__file__).resolve().parent
SPEC = importlib.util.spec_from_file_location("semantic_pilot_run", ROOT / "run.py")
RUNNER = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(RUNNER)


class PacketTests(unittest.TestCase):
    def test_frozen_packet_verifies(self):
        self.assertEqual(RUNNER.verify()["calls_max"], 4)

    def test_tampered_file_hash_is_rejected(self):
        original = RUNNER.files()
        altered = dict(original)
        key = next(name for name in altered if name.endswith("requests/caddy.txt"))
        altered[key] = "0" * 64
        with mock.patch.object(RUNNER, "files", return_value=altered):
            with self.assertRaisesRegex(ValueError, "packet hash mismatch"):
                RUNNER.verify()

    def test_unbound_quote_is_rejected(self):
        response = {"page_id": "caddy", "source_sha256": RUNNER.payload("caddy")["source_sha256"],
                    "findings": [{"line": 1, "quote": "not in the source\n", "diagnosis": "d",
                                  "evidence": "e", "source_function": "f", "deletion_test": "t"}]}
        with self.assertRaisesRegex(ValueError, "quote not uniquely bound"):
            RUNNER.bind("caddy", response)

    def test_tool_trace_is_rejected(self):
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory)
            events = [
                {"type": "turn.started"},
                {"type": "item.completed", "item": {"type": "command_execution"}},
                {"type": "turn.completed"},
            ]
            (path / "events.jsonl").write_text("\n".join(json.dumps(e) for e in events), encoding="utf-8")
            with self.assertRaisesRegex(ValueError, "tool use"):
                RUNNER.trace(path, {})


if __name__ == "__main__":
    unittest.main()
