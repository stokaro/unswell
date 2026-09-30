"""Recompute published constructed-control results without model dependencies."""

import hashlib
import json
import math
from pathlib import Path
import tarfile

ROOT = Path(__file__).resolve().parent


def require(condition, message):
    if not condition:
        raise ValueError(message)


def read(path):
    return json.loads(path.read_text())


def finite(value):
    return isinstance(value, (int, float)) and not isinstance(value, bool) and math.isfinite(value)


def lens_screen(controls, rows, protocol):
    predictions = {row["id"]: row["result"] for row in rows}
    expected = {"control/" + group["id"] + "/" + role for group in controls["groups"] for role in ["identity", "useful", "unsafe"]}
    expected.update("control/" + group["id"] + "/" + role for group in controls["neutral"] for role in ["identity", "neutral"])
    expected.update("control/case/" + role for role in ["identity", "useful", "unsafe"])
    require(len(predictions) == len(rows) and set(predictions) == expected, "Missing or duplicate constructed predictions")
    for output in predictions.values():
        require(output["available"] is True, "An unavailable control cannot pass this complete preflight")
        require(finite(output["score"]) and output["all_input_tokens_retained"] is True, "Invalid or truncated available score")
    margin = protocol["fixed_gain_margin"]
    gains = []
    for group in controls["groups"]:
        identity, useful, unsafe = (predictions["control/" + group["id"] + "/" + role]["score"] for role in ["identity", "useful", "unsafe"])
        gains.append((useful - identity, unsafe - identity, useful > unsafe))
    neutral = [predictions["control/" + group["id"] + "/neutral"]["score"] - predictions["control/" + group["id"] + "/identity"]["score"] for group in controls["neutral"]]
    useful_count = sum(g[0] > margin for g in gains)
    unsafe_count = sum(g[1] > margin for g in gains)
    neutral_count = sum(g > margin for g in neutral)
    rank_count = sum(g[2] for g in gains)
    gate = protocol["preflight"]
    benefit = useful_count >= gate["minimum_useful_gain_count"] and neutral_count <= gate["maximum_neutral_gain_count"]
    ranking = rank_count >= gate["minimum_useful_above_unsafe_count"]
    safety = unsafe_count <= gate["maximum_unsafe_gain_count_for_standalone_safety"]
    return {
        "required_groups": len(controls["groups"]),
        "available_groups": len(controls["groups"]),
        "useful_gain_count": useful_count,
        "unsafe_gain_count": unsafe_count,
        "neutral_gain_count": neutral_count,
        "useful_above_unsafe_count": rank_count,
        "benefit_screen_pass": benefit,
        "relative_safety_ranking_pass": ranking,
        "standalone_safety_screen_pass": safety,
        "combined_screen_pass": benefit and ranking and safety,
        "case_control_identical_outputs": predictions["control/case/useful"] == predictions["control/case/unsafe"],
        "product_qualified": False,
    }


def support(output, cutoff):
    require(output["available"] is True and output["all_input_tokens_retained"] is True, "Unavailable or truncated grounding control")
    probability = output["support_score"]
    require(finite(probability) and 0 <= probability <= 1, "Invalid support score")
    label = int(probability > cutoff)
    require(output["published_label"] == label, "Label differs from the fixed published cutoff")
    return bool(label)


def supported_pair(pair, cutoff):
    # Evaluate both sides even when the first fails so neither escapes validation.
    forward = support(pair["forward"], cutoff)
    backward = support(pair["backward"], cutoff)
    result = forward and backward
    require(pair["pair_supported"] == result, "Pair flag does not match both direction scores")
    return result


def mini_screen(controls, record, protocol):
    require(record["complete"] is True, "Incomplete MiniCheck execution")
    groups, neutral = record["groups"], record["neutral"]
    expected_groups = {g["id"] for g in controls["groups"]}
    expected_neutral = {g["id"] for g in controls["neutral"]}
    require(len(groups) == len(expected_groups) and {g["id"] for g in groups} == expected_groups, "Missing MiniCheck groups")
    require(len(neutral) == len(expected_neutral) and {g["id"] for g in neutral} == expected_neutral, "Missing neutral grounding groups")
    cutoff = protocol["published_label_cutoff"]
    safe_count = sum(supported_pair(g["useful"], cutoff) for g in groups)
    unsafe_count = sum(supported_pair(g["unsafe"], cutoff) for g in groups)
    neutral_count = sum(supported_pair(g, cutoff) for g in neutral)
    support(record["case"]["useful"], cutoff)
    unsafe_case = support(record["case"]["unsafe"], cutoff)
    return {
        "safe_retained": safe_count,
        "unsafe_retained": unsafe_count,
        "neutral_retained": neutral_count,
        "safe_required": len(expected_groups),
        "unsafe_required": len(expected_groups),
        "neutral_required": len(expected_neutral),
        "screen_pass": safe_count >= protocol["minimum_safe_pairs_retained"] and unsafe_count <= protocol["maximum_unsafe_pairs_retained"],
        "case_sensitive_identifier_error_accepted": unsafe_case,
    }


def verify(root=ROOT):
    manifest = read(root / "manifest.json")
    for name, info in manifest["files"].items():
        data = (root / name).read_bytes()
        require(len(data) == info["bytes"] and hashlib.sha256(data).hexdigest() == info["sha256"], "Changed public artifact: " + name)
    with tarfile.open(root / "runner-source.tar.gz", "r:gz") as archive:
        members = archive.getmembers()
        require(len(members) == len(manifest["runner_members"]), "Archive member count changed")
        require({m.name for m in members} == set(manifest["runner_members"]), "Archive source identity changed")
        for member in members:
            require(member.isfile() and member.size < 1024 * 1024 and ".." not in Path(member.name).parts and not member.name.startswith("/"), "Unsafe archive member")
            data = archive.extractfile(member).read()
            info = manifest["runner_members"][member.name]
            require(len(data) == info["bytes"] and hashlib.sha256(data).hexdigest() == info["sha256"], "Archived runner changed")
    controls, protocol, records, summary = (read(root / name) for name in ["testdata/controls.json", "protocol.json", "records.json", "summary.json"])
    require(hashlib.sha256((root / "testdata/controls.json").read_bytes()).hexdigest() == protocol["controls_sha256"], "Controls do not match the frozen projection")
    require(len(controls["groups"]) == 12 and len(controls["neutral"]) == 4, "Changed constructed denominators")
    require(len({g["id"] for g in controls["groups"] + controls["neutral"]}) == 16, "Duplicate control identities")
    lens = records["lens"]
    require(lens["model_identity"]["strict_load"] and lens["model_identity"]["all_parameters_equal_checkpoint"], "Incomplete LENS model restoration")
    smoke = lens["smoke"]
    require(abs(lens["smoke_observed"]["score"] - smoke["expected_score"]) <= smoke["absolute_tolerance"], "LENS numerical reference failed")
    require(lens["execution"]["complete"] and lens["execution"]["jobs_completed"] == lens["execution"]["jobs_required"] == 331 and lens["execution"]["outbound_requests"] == 0, "Incomplete LENS run")
    require(controls["case_control"]["useful"].lower() == controls["case_control"]["unsafe"].lower(), "Case collision example changed")
    recalculated_lens = lens_screen(controls, lens["control_predictions"], protocol["lens"])
    require(recalculated_lens == summary["lens"], "LENS summary differs from predictions")
    mini = records["minicheck"]
    require(mini["model_identity"]["strict_parameter_load"] is True, "Incomplete MiniCheck restoration")
    require(len(mini["smoke"]) == 2 and all(abs(s["support_score"] - s["expected_score"]) <= 0.001 for s in mini["smoke"]), "MiniCheck numerical reference failed")
    recalculated_mini = mini_screen(controls, mini, protocol["minicheck"])
    require(recalculated_mini == summary["minicheck"], "MiniCheck summary differs from predictions")
    # Only arithmetic consistency is public for the actual technical-page transfer.
    transfer = lens["private_pair_transfer"]
    require(transfer["required"] == 142 and transfer["available"] == 138 and sum(transfer["unavailable_reasons"].values()) == 4, "Changed private aggregate denominator")
    require(sum(r["required"] for r in transfer["per_page"].values()) == 142 and sum(r["available"] for r in transfer["per_page"].values()) == 138, "Inconsistent aggregate coverage")
    require(records["original_recall_objective_achieved"] is False and summary["product_qualified"] is False and summary["original_recall_objective_achieved"] is False, "Unsupported qualification claim")
    return {"public_constructed_results_recomputed": True, "lens": recalculated_lens, "minicheck": recalculated_mini, "private_task_predictions_recomputed": False, "product_qualified": False}


if __name__ == "__main__":
    print(json.dumps(verify(), indent=2))
