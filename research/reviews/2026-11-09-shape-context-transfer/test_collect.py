import importlib.util
from pathlib import Path
import re
import unittest

HERE = Path(__file__).resolve().parent
SPEC = importlib.util.spec_from_file_location("shape_collect", HERE / "collect.py")
COLLECT = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(COLLECT)


class CollectorTest(unittest.TestCase):
    def test_eligible_utf8_offset_and_protected_segment(self):
        source = "é rather than code rather than prose"
        raw = source.encode("utf-8")
        first = raw.index(b"rather")
        second = raw.rindex(b"rather")
        unit = {"ordinal": 0,
                "raw_span": {"start": 0, "end": len(raw)},
                "raw_slice": source,
                "binding": {"segments": [
                    {"start": 0, "end": first + len(b"rather than")},
                    {"start": second, "end": len(raw)},
                ]}}
        excluded = [{"span": {"start": first + len(b"rather than"),
                               "end": second}}]
        shapes = {"contrast": re.compile(r"rather than")}
        found = COLLECT.extract_candidates(source, [unit], excluded, shapes)
        self.assertEqual([(x["start"], x["end"]) for x in found],
                         [(first, first + len(b"rather than")),
                          (second, second + len(b"rather than"))])

    def test_rejects_eligible_protected_overlap(self):
        source = "rather than"
        unit = {"ordinal": 0, "raw_span": {"start": 0, "end": len(source)},
                "raw_slice": source,
                "binding": {"segments": [{"start": 0, "end": len(source)}]}}
        excluded = [{"span": {"start": 2, "end": 4}}]
        with self.assertRaisesRegex(ValueError, "intersects protected"):
            COLLECT.extract_candidates(source, [unit], excluded,
                                       {"contrast": re.compile("rather")})

    def test_rejects_wrong_raw_slice(self):
        unit = {"ordinal": 0, "raw_span": {"start": 0, "end": 3},
                "raw_slice": "bad", "binding": {"segments": []}}
        with self.assertRaisesRegex(ValueError, "raw unit mismatch"):
            COLLECT.extract_candidates("abc", [unit], [], {})


if __name__ == "__main__":
    unittest.main()
