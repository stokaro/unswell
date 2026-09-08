"""Compare a dependencyprobe observation with the pinned Python reference.

Research only; not required by Unswell builds, tests, or scans. Each call uses
exactly the same extracted prose and protected boundaries as the Go experiment.
"""

import argparse
import hashlib
import importlib.metadata
import json
import sys
from pathlib import Path

import spacy


def compare_input(pipeline, source, call, totals):
    reference = list(pipeline(call["text"]))
    go_tokens = []
    go_arcs = []
    for sentence in source["sentences"][call["first_sentence"]:call["end_sentence"]]:
        base = len(go_tokens)
        go_tokens.extend(sentence["tokens"])
        for arc in sentence["dependencies"]["arcs"]:
            go_arcs.append({"head": -1 if arc["head"] == -1 else base + arc["head"],
                            "relation": arc["relation"]})
    if [t.text for t in reference] != [t["text"] for t in go_tokens]:
        raise ValueError(f"tokenization mismatch in {source['id']}")
    row = {"source": source["id"], "text": call["text"], "tokens": []}
    for index, token in enumerate(reference):
        head = -1 if token.head.i == index and token.dep_ == "ROOT" else token.head.i
        arc = go_arcs[index]
        byte_start = call["offset"] + len(call["text"][:token.idx].encode("utf-8"))
        totals["tokens"] += 1
        totals["heads_equal"] += head == arc["head"]
        totals["labels_equal"] += token.dep_ == arc["relation"]
        totals["tags_equal"] += token.tag_ == go_tokens[index]["tag"]
        totals["offsets_equal"] += byte_start == go_tokens[index]["start"]
        row["tokens"].append(dict(text=token.text, rune_start=token.idx,
                                  head=head, relation=token.dep_, tag=token.tag_))
    return row


def compare(model, observation):
    if spacy.__version__ != "3.8.14":
        raise ValueError("reference requires spaCy 3.8.14")
    pipeline = spacy.load(model, disable=["ner", "lemmatizer"])
    rows = []
    totals = dict(tokens=0, heads_equal=0, labels_equal=0, tags_equal=0, offsets_equal=0)
    for source in observation["sources"]:
        for call in source["inputs"]:
            rows.append(compare_input(pipeline, source, call, totals))
    return {"schema": "unswell-spacy-reference-v1", "spacy": spacy.__version__,
            "environment": {d.metadata["Name"]: d.version for d in importlib.metadata.distributions()},
            "comparison": totals, "calls": rows}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--model", required=True)
    parser.add_argument("--observation", required=True)
    args = parser.parse_args()
    data = Path(args.observation).read_bytes()
    result = compare(args.model, json.loads(data))
    result["observation_sha256"] = hashlib.sha256(data).hexdigest()
    print(json.dumps(result, indent=2, sort_keys=True, ensure_ascii=False))
    comparison = result["comparison"]
    if any(value != comparison["tokens"] for value in comparison.values()):
        sys.exit(1)


if __name__ == "__main__":
    main()
