"""Negative checks for the optional independent source-pair verifier."""

import base64
import copy
import unittest

import replay


def source():
    raw = (b'<doc id="P01-1001" editor="A"><title><text>Title</text></title>'
           b'<abstract><text>Caf\xc3\xa9</text><edit crr="can retry">may retry</edit>'
           b'<edit crr="">very</edit><edit crr="again"></edit></abstract>'
           b'<introduction><text>Body.</text></introduction></doc>')
    return {"path": "original/P01-1001-A.xml", "sha256": replay.sha256(raw),
            "git_blob": replay.git_blob(raw), "xml_base64": base64.b64encode(raw).decode()}


def report(files):
    expected, groups, counts = replay.expected_report(files)
    for file in expected:
        code = file.pop("issue_code")
        file["issues"] = [{"code": code, "reason": "source failure"}] if code else []
    return {"version": replay.VERSION, "repository": "https://github.com/chemicaltree/tetra",
            "revision": replay.REVISION, "license": "CC-BY-4.0", "files": expected,
            "rendering": replay.RENDERING, "target": replay.TARGET,
            "groups": groups, "counts": counts, "complete": not counts["quarantined_files"],
            "task_labels_assigned": False, "product_qualified": False}


class VerificationTests(unittest.TestCase):
    def test_unicode_space_is_not_xml_indentation(self):
        file = source()
        raw = base64.b64decode(file["xml_base64"]).replace(b"</text></title>", b"</text>\xc2\xa0</title>")
        file["xml_base64"] = base64.b64encode(raw).decode()
        file["sha256"], file["git_blob"] = replay.sha256(raw), replay.git_blob(raw)
        expected = replay.expected_file(file)
        self.assertEqual(expected["status"], "invalid_xml")
        self.assertEqual(expected["issue_code"], "unsupported_xml")

    def test_negative_changes_to_bound_output(self):
        files = [source()]
        original = report(files)
        replay.verify(original, files)
        for name in ("span", "source", "group", "lost file", "lost edit", "label"):
            with self.subTest(name=name):
                value = copy.deepcopy(original)
                if name == "span":
                    value["files"][0]["sections"][1]["edits"][0]["original_span"]["start"] -= 1
                elif name == "source":
                    value["files"][0]["sha256"] = "0" * 64
                elif name == "group":
                    value["groups"][0]["files"] = []
                elif name == "lost file":
                    value["files"] = []
                elif name == "lost edit":
                    value["files"][0]["sections"][1]["edits"].pop()
                elif name == "label":
                    value["task_labels_assigned"] = True
                with self.assertRaises(ValueError):
                    replay.verify(value, files)

    def test_stray_characters_quarantine_peers(self):
        first = source()
        raw = base64.b64decode(first["xml_base64"])
        raw = raw.replace(b'editor="A"', b'editor="B"').replace(b"</text></title>", b"</text>,</title>")
        second = {"path": "original/P01-1001-B.xml", "sha256": replay.sha256(raw),
                  "git_blob": replay.git_blob(raw), "xml_base64": base64.b64encode(raw).decode()}
        files = [first, second]
        value = report(files)
        self.assertEqual(value["counts"]["quarantined_files"], 2)
        self.assertEqual(value["files"][1]["issues"][0]["code"], "unsupported_xml")
        replay.verify(value, files)
        value["files"][0]["source_eligible"] = True
        with self.assertRaises(ValueError):
            replay.verify(value, files)


if __name__ == "__main__":
    unittest.main()
