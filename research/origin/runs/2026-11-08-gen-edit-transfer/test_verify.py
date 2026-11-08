"""A saved artifact change must fail the repository research check."""

import importlib.util
import shutil
import tempfile
import unittest
from pathlib import Path


HERE = Path(__file__).resolve().parent
spec = importlib.util.spec_from_file_location("gen_verify", HERE / "verify.py")
assert spec is not None and spec.loader is not None
gen_verify = importlib.util.module_from_spec(spec)
spec.loader.exec_module(gen_verify)


class VerifyTests(unittest.TestCase):
    def test_saved_result_change_is_rejected(self):
        with tempfile.TemporaryDirectory() as temporary:
            directory = Path(temporary)
            for name in ("freeze.json", "decision.json", "result.json", "PROTOCOL.md",
                         "run.py", "acquire.py", "closed_class.txt"):
                shutil.copyfile(HERE / name, directory / name)
            gen_verify.verify(directory)
            with (directory / "result.json").open("a") as stream:
                stream.write("\n")
            with self.assertRaisesRegex(ValueError, "Saved result changed"):
                gen_verify.verify(directory)


if __name__ == "__main__":
    unittest.main()
