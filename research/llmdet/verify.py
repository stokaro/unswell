"""Compare a built Go probe with frozen numerical reference outputs."""

import argparse
import hashlib
import json
import math
import subprocess
import tempfile
from pathlib import Path


def digest(data):
    return hashlib.sha256(data).hexdigest()


def run(probe, kind, pack, vectors):
    completed = subprocess.run(
        [str(probe), "--kind", kind, "--model", str(pack)],
        input=json.dumps(vectors, allow_nan=False).encode(), capture_output=True,
        check=True, timeout=60,
    )
    result = json.loads(completed.stdout)
    assert result["version"] == "unswell-llmdet-probe-v1"
    assert result["kind"] == kind
    assert result["model_sha256"] == digest(pack.read_bytes())
    assert len(result["results"]) == len(vectors)
    return result["results"]


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--probe", type=Path, required=True)
    parser.add_argument("--reference", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    reference = json.loads((args.reference / "reference.json").read_bytes())
    for name, expected in reference["artifacts"].items():
        assert Path(name).name == name
        assert digest((args.reference / name).read_bytes()) == expected
    errors = {"proxy": 0.0, "raw": 0.0, "responses": 0.0}

    def compare(actual, expected, field):
        assert math.isfinite(actual) and math.isfinite(expected)
        assert math.isclose(actual, expected, abs_tol=reference["absolute_tolerance"],
                            rel_tol=reference["relative_tolerance"]), (field, actual, expected)
        errors[field] = max(errors[field], abs(actual - expected))

    proxies = json.loads((args.reference / "proxy-controls.json").read_bytes())
    proxy_results = []
    with tempfile.TemporaryDirectory(prefix="unswell-llmdet-proxy-") as directory:
        pack = Path(directory) / "proxy.json"
        for case in proxies:
            pack.write_text(json.dumps(case["spec"], allow_nan=False))
            result = run(args.probe, "proxy", pack, [case["tokens"]])[0]
            compare(result["reference_score"], case["reference_score"], "proxy")
            proxy_results.append({"id": case["id"], "result": result})
    vectors = json.loads((args.reference / "classifier-controls.json").read_bytes())
    predictions = run(args.probe, "ensemble", args.reference / "ensemble.json",
                      [row["values"] for row in vectors])
    for actual, expected in zip(predictions, vectors, strict=True):
        assert actual["classes"] == reference["classes"]
        for field in ["raw", "responses"]:
            assert len(actual[field]) == len(expected[field]) == len(reference["classes"])
            for value, target in zip(actual[field], expected[field], strict=True):
                compare(value, target, field)
    report = {
        "version": "unswell-llmdet-component-parity-v1", "status": "passed",
        "reference_sha256": digest((args.reference / "reference.json").read_bytes()),
        "probe_sha256": digest(args.probe.read_bytes()), "script_sha256": digest(Path(__file__).read_bytes()),
        "proxy_controls": len(proxies), "classifier_controls": len(vectors),
        "absolute_tolerance": reference["absolute_tolerance"], "relative_tolerance": reference["relative_tolerance"],
        "maximum_absolute_error": errors, "proxy_results": proxy_results,
        "go_source_sha256": go_sources(),
        "scope": "Numeric component controls only; tokenizer, full probability tables, text detection, and qualification not run.",
    }
    args.output.write_text(json.dumps(report, indent=2, allow_nan=False) + "\n")
    print(json.dumps({key: value for key, value in report.items() if key != "proxy_results"}))


def go_sources():
    root = Path(__file__).resolve().parents[2]
    directories = ["llmdet", "internal/llmdetcommand", "cmd/llmdetprobe", "internal/jsoninput", "internal/commandio"]
    paths = [path for directory in directories for path in (root / "research/annotation" / directory).glob("*")
             if path.suffix in [".go", ".json"] and not path.name.endswith("_test.go")]
    return {str(path.relative_to(root)): digest(path.read_bytes()) for path in sorted(paths)}


if __name__ == "__main__":
    main()
