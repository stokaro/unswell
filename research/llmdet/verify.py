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
    commit, sources, build = go_identity(args.probe)
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
        "go_source_commit": commit, "go_source_sha256": sources, "probe_build": build,
        "scope": "Numeric component controls only; tokenizer, full probability tables, text detection, and qualification not run.",
    }
    args.output.write_text(json.dumps(report, indent=2, allow_nan=False) + "\n")
    print(json.dumps({key: value for key, value in report.items() if key != "proxy_results"}))


def go_identity(probe):
    root = Path(__file__).resolve().parents[2]
    commit = subprocess.run(["git", "-C", str(root), "rev-parse", "HEAD"],
                            capture_output=True, text=True, check=True, timeout=10).stdout.strip()
    directories = ["llmdet", "internal/llmdetcommand", "cmd/llmdetprobe", "internal/jsoninput", "internal/commandio"]
    paths = [path for directory in directories for path in (root / "research/annotation" / directory).glob("*")
             if path.suffix in [".go", ".json"] and not path.name.endswith("_test.go")]
    sources = {}
    for path in sorted(paths):
        name = str(path.relative_to(root))
        saved = subprocess.run(["git", "-C", str(root), "show", f"{commit}:{name}"],
                               capture_output=True, check=True, timeout=10).stdout
        if saved != path.read_bytes():
            raise ValueError(f"uncommitted numerical source: {name}")
        sources[name] = digest(saved)
    metadata = subprocess.run(["go", "version", "-m", str(probe)],
                              capture_output=True, text=True, check=True, timeout=10).stdout
    build = {}
    for line in metadata.splitlines():
        fields = line.split()
        if len(fields) == 2 and fields[0] == "build" and "=" in fields[1]:
            key, value = fields[1].split("=", 1)
            build[key] = value
    if build.get("vcs.revision") != commit or build.get("vcs.modified") != "false" or build.get("CGO_ENABLED") != "0":
        raise ValueError("probe must be built without cgo from this clean committed revision")
    build["go_version"] = metadata.splitlines()[0].rsplit(": ", 1)[-1]
    return commit, sources, build


if __name__ == "__main__":
    main()
