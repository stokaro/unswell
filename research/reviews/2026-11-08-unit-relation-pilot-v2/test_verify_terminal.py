"""Prove that the terminal record rejects changed traces and outcomes."""

import json
from pathlib import Path
import shutil
import tempfile
import unittest
from unittest import mock

import packet
import run
import verify_terminal


class TerminalVerifierTest(unittest.TestCase):
    def copied(self):
        directory = tempfile.TemporaryDirectory()
        root = Path(directory.name) / "pilot"
        shutil.copytree(verify_terminal.ROOT, root)
        self.addCleanup(directory.cleanup)
        return root

    def checked(self, root):
        with mock.patch.object(packet, "ROOT", root), \
                mock.patch.object(packet, "PACKET", root / "packet"), \
                mock.patch.object(run, "ROOT", root), \
                mock.patch.object(run, "PACKET", root / "packet"), \
                mock.patch.object(run, "RUN", root / "run"), \
                mock.patch.object(verify_terminal, "ROOT", root), \
                mock.patch.object(verify_terminal, "RUN", root / "run"):
            return verify_terminal.verify()

    def test_published_outcome_change_is_rejected(self):
        root = self.copied()
        self.checked(root)
        path = root / "terminal.json"
        result = json.loads(path.read_text())
        result["registered_advance_possible_with_this_verdict"] = True
        path.write_text(json.dumps(result))
        with self.assertRaisesRegex(ValueError, "terminal result changed"):
            self.checked(root)

    def test_raw_verdict_change_is_rejected(self):
        root = self.copied()
        path = root / "run" / "fastfetch" / "verdict" / "response.json"
        response = json.loads(path.read_text())
        response["verdicts"][0]["decision"] = "diagnose"
        path.write_text(json.dumps(response))
        with self.assertRaisesRegex(ValueError, "response hash changed"):
            self.checked(root)

    def test_call_after_error_is_rejected(self):
        root = self.copied()
        (root / "run" / "mise").mkdir()
        with self.assertRaisesRegex(ValueError, "continued after first error"):
            self.checked(root)


if __name__ == "__main__":
    unittest.main()
