"""Run pinned numerical components; this does not reproduce the full detector."""

import argparse
import ast
import hashlib
import json
import math
import platform
import random
from pathlib import Path

import lightgbm
import numpy as np
import scipy

SOURCE_SHA256 = "e1595b595b138cd35e82e71cb62e0a70f829ba6e174027f47db66264a2d8f268"
CLASSIFIER_SHA256 = "13ad64811768abf3aa88f4ca846fee397e8bce797253135d84bc322ebe907517"
CLASSES = ["Human_write", "GPT-2", "OPT", "UniLM", "LLaMA", "BART", "T5", "Bloom", "GPT-neo"]


def digest(data):
    return hashlib.sha256(data).hexdigest()


def write(path, data):
    path.write_text(json.dumps(data, indent=2, ensure_ascii=False, allow_nan=False) + "\n")


def row(context, continuations, probabilities):
    return {"context": context, "continuations": continuations, "probabilities": probabilities}


def proxy_cases():
    four = row([1, 2, 3], [4], [0.5])
    three = row([2, 3], [4], [0.25])
    two = row([3], [4], [0.125])
    zero = row([1, 2, 3], [4], [0.0])
    rows = [
        ("four-gram", [1, 2, 3, 4], [four]),
        ("three-gram", [1, 2, 3, 4], [three]),
        ("two-gram", [1, 2, 3, 4], [two]),
        ("highest-context-first", [1, 2, 3, 4], [two, three, four]),
        ("residual-before-backoff", [1, 2, 3, 4], [two, row([1, 2, 3], [5, 6], [0.25, 0.25])]),
        ("retained-zero", [1, 2, 3, 4], [zero]),
        ("saturated-residual", [1, 2, 3, 4], [row([1, 2, 3], [5], [1.0])]),
        ("no-context", [1, 2, 3, 4], []),
        ("too-short", [1, 2, 3], [two]),
        ("empty-input", [], [four]),
        ("mixed-orders", [1, 2, 3, 4, 5, 6], [four, row([3, 4], [5], [0.25]), row([5], [6], [0.125])]),
        ("partial-coverage", [1, 2, 3, 4, 5, 6, 7], [four]),
        ("zero-in-denominator", [1, 2, 3, 4, 5], [zero, row([4], [5], [0.5])]),
        ("token-zero", [0, 0, 0, 0, 5], [row([0], [0], [0.5])]),
        ("uniform-row", [1, 2, 3, 4], [row([1, 2, 3], [], [])]),
    ]
    return [{"id": name, "tokens": tokens,
             "spec": {"version": "unswell-llmdet-proxy-v1", "vocab_size": 16, "rows": entries}}
            for name, tokens, entries in rows]


def original_proxy(source):
    module = ast.parse(source)
    function = next(node for node in module.body if isinstance(node, ast.FunctionDef) and node.name == "perplexity")
    selected = ast.Module(body=[function], type_ignores=[])
    scope = {"math": math, "tqdm": lambda values: values}
    # Only the hash-pinned, reviewed numerical function executes; loading,
    # tokenization, downloads, classifier setup, and other imports do not run.
    exec(compile(selected, "pinned-detector.py", "exec"), scope)
    return scope["perplexity"]


def measure_proxy(function, case):
    tables = [{} for _ in range(6)]
    for entry in case["spec"]["rows"]:
        index = (len(entry["context"]) - 1) * 2
        key = tuple(entry["context"])
        tables[index][key] = np.array(entry["continuations"], dtype=np.int64)
        tables[index + 1][key] = np.array(entry["probabilities"], dtype=np.float64)
    return function([case["tokens"]], tables, case["spec"]["vocab_size"])[0]


def export_tree(tree, index):
    splits = [None] * (tree["num_leaves"] - 1)
    leaves = [None] * tree["num_leaves"]

    def visit(node):
        if "leaf_index" in node:
            leaf = node["leaf_index"]
            assert leaves[leaf] is None
            leaves[leaf] = node["leaf_value"]
            return -leaf - 1
        assert node["decision_type"] == "<=" and node["missing_type"] == "None"
        assert node["default_left"] is True
        split = node["split_index"]
        assert splits[split] is None
        splits[split] = {"feature": node["split_feature"], "threshold": node["threshold"],
                         "left": visit(node["left_child"]), "right": visit(node["right_child"])}
        return split

    assert visit(tree["tree_structure"]) == 0
    assert all(value is not None for value in splits + leaves)
    return {"class": index % len(CLASSES), "comparison": "numeric-le", "splits": splits, "leaves": leaves}


def classifier_vectors(spec):
    vectors = [{"id": f"constant-{value}", "values": [value] * 11}
               for value in [0.0, -1.0, 1.0, 10.0, 20.0, 35.0, -1e300, 1e300]]
    rng = random.Random(5101)
    for index in range(128):
        vectors.append({"id": f"seeded-{index}", "values": [rng.uniform(0, 35) for _ in range(11)]})
    for index, tree in enumerate(spec["trees"][:64]):
        split = tree["splits"][0]
        threshold = split["threshold"]
        for label, value in [("below", math.nextafter(threshold, -math.inf)), ("at", threshold),
                             ("above", math.nextafter(threshold, math.inf))]:
            values = [10.0] * 11
            values[split["feature"]] = value
            vectors.append({"id": f"root-{index}-{label}", "values": values})
    return vectors


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--source", type=Path, required=True)
    parser.add_argument("--classifier", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    versions = {"lightgbm": lightgbm.__version__, "numpy": np.__version__, "scipy": scipy.__version__}
    assert versions == {"lightgbm": "4.6.0", "numpy": "2.2.6", "scipy": "1.15.3"}, versions
    source, classifier = args.source.read_bytes(), args.classifier.read_bytes()
    if digest(source) != SOURCE_SHA256 or digest(classifier) != CLASSIFIER_SHA256:
        raise ValueError("source or classifier differs from the reviewed pinned bytes")
    args.output.mkdir(parents=True, exist_ok=True)
    function = original_proxy(source)
    cases = proxy_cases()
    for case in cases:
        case["reference_score"] = measure_proxy(function, case)
    write(args.output / "proxy-controls.json", cases)
    model = lightgbm.Booster(model_file=str(args.classifier))
    dumped = model.dump_model(num_iteration=-1)
    assert dumped["num_class"] == 9 and dumped["num_tree_per_iteration"] == 9
    assert dumped["max_feature_idx"] == 10 and not dumped["average_output"]
    assert dumped["objective"] == "multiclass num_class:9"
    assert len(dumped["tree_info"]) == 1800
    spec = {"version": "unswell-llmdet-ensemble-v1", "feature_count": 11, "classes": CLASSES,
            "trees": [export_tree(tree, index) for index, tree in enumerate(dumped["tree_info"])]}
    write(args.output / "ensemble.json", spec)
    vectors = classifier_vectors(spec)
    matrix = np.array([row["values"] for row in vectors], dtype=np.float64)
    margins = model.predict(matrix, raw_score=True, num_iteration=-1, num_threads=1)
    responses = model.predict(matrix, raw_score=False, num_iteration=-1, num_threads=1)
    for index, vector in enumerate(vectors):
        vector["raw"] = margins[index].tolist()
        vector["responses"] = responses[index].tolist()
    write(args.output / "classifier-controls.json", vectors)
    write(args.output / "reference.json", {
        "version": "unswell-llmdet-component-reference-v1", "python": platform.python_version(),
        "platform": platform.platform(), "packages": versions, "source_sha256": digest(source),
        "classifier_sha256": digest(classifier), "script_sha256": digest(Path(__file__).read_bytes()),
        "scope": "Authored float64 probability rows and complete numeric classifier vectors; no text detection.",
        "proxy_controls": len(cases), "classifier_controls": len(vectors), "classes": CLASSES,
        "trees": 1800, "features": 11,
        "absolute_tolerance": 1e-12, "relative_tolerance": 1e-12,
        "artifacts": {name: digest((args.output/name).read_bytes()) for name in
                      ["proxy-controls.json", "classifier-controls.json", "ensemble.json"]},
    })


if __name__ == "__main__":
    main()
