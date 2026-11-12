#!/usr/bin/env python3
"""Agreement between LAMP writers at the word level, as a check of the sentence-level ceiling.

usage: ceiling_words.py <human directory>
Added after the sentence-level ceiling was read; it is not a registered measure.
"""
import collections, itertools, re, sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))
from common import byte_span, overlaps, read_jsonl

docs = [d for d in read_jsonl(Path(sys.argv[1]) / "docs.jsonl") if d["split"] == "ceiling"]
by_text = collections.defaultdict(dict)
for doc in docs:
    by_text[doc["text"]][doc["editor"]] = [e["span"] for e in doc["reference"]]
pooled = collections.Counter()
shares = collections.defaultdict(list)
all_three = collections.Counter()
for text, writers in by_text.items():
    words = [byte_span(text, m.start(), m.end()) for m in re.finditer(r"\S+", text)]
    marks = {name: [overlaps(word, spans) for word in words] for name, spans in writers.items()}
    for name, row in marks.items():
        shares[name].append(sum(row) / len(row))
    for first, second in itertools.permutations(sorted(writers), 2):
        for a, b in zip(marks[first], marks[second]):
            pooled["words"] += 1
            pooled["b"] += b
            pooled["a_and_b"] += a and b
            pooled["a_not_b"] += a and not b
    if len(marks) == 3:
        for votes in zip(*marks.values()):
            all_three[sum(votes)] += 1
tpr = pooled["a_and_b"] / pooled["b"]
fpr = pooled["a_not_b"] / (pooled["words"] - pooled["b"])
po = (pooled["a_and_b"] + (pooled["words"] - pooled["b"] - pooled["a_not_b"])) / pooled["words"]
pa = (pooled["a_and_b"] + pooled["a_not_b"]) / pooled["words"]
pb = pooled["b"] / pooled["words"]
pe = pa * pb + (1 - pa) * (1 - pb)
print(f"paragraphs {len(by_text)}; word share edited per writer:",
      {name: round(sum(v) / len(v), 3) for name, v in sorted(shares.items())})
print(f"word level, pooled ordered pairs: TPR {tpr:.1%} FPR {fpr:.1%} J {100 * (tpr - fpr):+.1f}; Cohen kappa {(po - pe) / (1 - pe):.3f}")
total = sum(all_three.values())
print("words marked by 0/1/2/3 of three writers:", {k: f"{all_three[k] / total:.1%}" for k in sorted(all_three)})
