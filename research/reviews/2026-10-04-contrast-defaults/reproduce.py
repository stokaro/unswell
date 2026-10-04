"""Reproduce the complete-source comparison with explicit offline CLI binaries."""
import argparse
import gzip
import hashlib
import json
from pathlib import Path
import subprocess


def read_report(here, binding):
    raw = (here / binding["report"]).read_bytes()
    assert hashlib.sha256(raw).hexdigest() == binding["gzip_sha256"]
    content = gzip.decompress(raw)
    assert hashlib.sha256(content).hexdigest() == binding["uncompressed_sha256"]
    return json.loads(content)


def finding_key(finding):
    locations = [finding["primary"], *finding["related"]]
    return (finding["rule_id"], tuple(
        (loc["path"], loc["span"]["start"], loc["span"]["end"]) for loc in locations))


def normalized(finding):
    return {key: value for key, value in finding.items()
            if key not in {"id", "fingerprint", "rule_version"}}


def verify_findings(actual, expected):
    assert actual["status"] == "complete" and actual["manifest"]["complete"]
    assert not actual["errors"] and not actual.get("abstentions")
    assert len(actual["assessments"]) == len(expected["assessments"])
    assert all(row["status"] == "available" for row in actual["assessments"])
    assert actual["assessments"] == expected["assessments"]
    for field in ("ruleset_hash", "feature_contract", "scoring_profile", "nlp"):
        assert actual["manifest"][field] == expected["manifest"][field], field
    assert [(d["name"], d["source_hash"]) for d in actual["documents"]] == [
        (d["name"], d["source_hash"]) for d in expected["documents"]]
    found = {finding_key(f): normalized(f) for f in actual["findings"]}
    wanted = {finding_key(f): normalized(f) for f in expected["findings"]}
    assert len(found) == len(actual["findings"])
    assert found == wanted, "Diagnostic content differs from the recorded comparison"


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--before", type=Path, required=True)
    parser.add_argument("--after", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    here = Path(__file__).resolve().parent
    identity = json.loads((here / "measurement.json").read_bytes())
    reports = {(profile, variant): read_report(here, binding)
               for profile, variants in identity["reports"].items()
               for variant, binding in variants.items() if variant in {"before", "after", "explicit"}}
    manifest = json.loads((here / "source-manifest.json").read_bytes())
    source = reports["technical", "before"]
    expected_hashes = {"inputs/" + row["path"]: row["source_sha256"]
                       for row in manifest["pages"]}
    assert len(source["documents"]) == len(expected_hashes) == 36
    for doc in source["documents"]:
        assert hashlib.sha256(doc["source"].encode()).hexdigest() == expected_hashes[doc["name"]]

    before = args.before.resolve(strict=True)
    after = args.after.resolve(strict=True)
    output = args.output.resolve()
    output.mkdir(exist_ok=False)
    for doc in source["documents"]:
        path = (output / doc["name"]).resolve()
        assert path.is_relative_to(output / "inputs")
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_bytes(doc["source"].encode())
    for profile in ("technical", "strict"):
        config = output / (profile + ".yaml")
        config.write_text("version: 1\nextends: [builtin:" + profile + "-v1]\n")
        for variant, binary in (("before", before), ("after", after), ("explicit", after)):
            selected_config = config
            if variant == "explicit":
                selected_config = output / (profile + "-explicit.yaml")
                selected_config.write_text(config.read_text() + "rules:\n  syntax.paired-contrast-density: {enabled: true}\n")
            target = output / (variant + "-" + profile + ".json")
            command = [str(binary), "check", *[d["name"] for d in source["documents"]],
                       "--project-root", str(output), "--config", str(selected_config),
                       "--include-source", "--no-gate", "--jobs", "1", "--timeout", "5m",
                       "--report", "json:" + str(target)]
            result = subprocess.run(command, cwd=output, capture_output=True, timeout=360)
            (output / (variant + "-" + profile + ".log")).write_bytes(result.stdout + result.stderr)
            assert result.returncode == 0, (variant, profile, result.returncode)
            actual = json.loads(target.read_bytes())
            verify_findings(actual, reports[profile, variant])
            print(variant, profile, len(actual["findings"]), "findings: verified", flush=True)


if __name__ == "__main__":
    main()
