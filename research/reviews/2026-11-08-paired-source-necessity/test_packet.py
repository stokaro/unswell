"""Negative controls for the frozen paired research packet."""

import json
import pathlib
import shutil
import tempfile
import unittest

import run


class PacketTest(unittest.TestCase):
    def test_frozen_packet_and_blind_positions(self):
        protocol = run.verify()
        pairs = run.selected()
        self.assertEqual(protocol["calls_max"], 8)
        self.assertEqual(len(pairs), 8)
        self.assertEqual(sum(key["A_old_status"] == "accepted"
                             for _, key in pairs), 4)
        for payload, _ in pairs:
            request = (run.PACKET / "requests" / (payload["pair_id"] + ".txt")).read_bytes()
            self.assertEqual(request, run.request(payload))
            self.assertNotIn(b'"old_status"', request)

    def test_changed_request_is_rejected(self):
        with tempfile.TemporaryDirectory() as temp:
            packet = pathlib.Path(temp) / "packet"
            shutil.copytree(run.PACKET, packet)
            target = packet / "requests" / "bat.txt"
            target.write_bytes(target.read_bytes() + b"changed")
            original = run.PACKET
            try:
                run.PACKET = packet
                with self.assertRaisesRegex(ValueError, "request or answer key"):
                    run.verify()
            finally:
                run.PACKET = original

    def test_changed_protocol_and_refrozen_hash_are_rejected(self):
        with tempfile.TemporaryDirectory() as temp:
            packet = pathlib.Path(temp) / "packet"
            shutil.copytree(run.PACKET, packet)
            protocol = run.read(packet / "protocol.json")
            protocol["advance"] = "Any result passes"
            (packet / "protocol.json").write_bytes(run.encoded(protocol))
            original = run.PACKET
            try:
                run.PACKET = packet
                (packet / "freeze.json").write_bytes(run.encoded(run.files()))
                with self.assertRaisesRegex(ValueError, "frozen protocol"):
                    run.verify()
            finally:
                run.PACKET = original

    def test_changed_target_offset_is_rejected(self):
        payload, _ = run.selected()[0]
        response = {"pair_id": payload["pair_id"],
                    "source_sha256": payload["source_sha256"]}
        for slot in ("A", "B"):
            candidate = payload["candidates"][slot]
            response.update({slot + "_id": candidate["item_id"],
                             slot + "_start": candidate["target_start_byte"],
                             slot + "_end": candidate["target_end_byte"],
                             slot + "_decision": "retain",
                             slot + "_basis": "The source preserves this detail."})
        run.check_response(response, payload)
        response["B_start"] += 1
        with self.assertRaisesRegex(ValueError, "target binding"):
            run.check_response(response, payload)

    def test_tool_trace_is_rejected(self):
        payload, _ = run.selected()[0]
        with tempfile.TemporaryDirectory() as temp:
            target = pathlib.Path(temp)
            (target / "events.jsonl").write_text(json.dumps({
                "type": "item.completed", "item": {"type": "command_execution"}}) + "\n")
            (target / "response.json").write_text("{}\n")
            with self.assertRaisesRegex(ValueError, "tool use"):
                run.check_trace(target, payload)


if __name__ == "__main__":
    unittest.main()
