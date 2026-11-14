#!/usr/bin/env python3
"""Fit and evaluate the page index on signal counts that the product computed.

usage: index_fit.py <workspace> <signals jsonl> fit
       index_fit.py <workspace> <signals jsonl> evaluate <part> <weights json>

fit: logistic loss of class (pages before a de-slop commit against pages that people wrote before November 2022) plus
three times the logistic loss of the pair (before above after), on the development part; prints the weights, the
cross-validated AUC and pair result of the integer index, and candidate thresholds.
"""
import hashlib, json, math, random, sys

W, signals_path, mode = sys.argv[1].rstrip("/") + "/", sys.argv[2], sys.argv[3]
sys.path.insert(0, W + "assisted/scripts")
from english import share

NAMES = ["em-dash", "emoji-in-heading", "emoji-at-list-item", "arrow", "stock-word", "bold-label-in-list"]


def development(repository):
    return int(hashlib.sha256(repository.encode()).hexdigest(), 16) % 10 < 6


def english_ids():
    result = {}
    for name, path in (("deslop", "assisted/deslop-files.jsonl"), ("reserve", "assisted/deslop-reserve-files.jsonl"),
                       ("web", "assisted/manual-files.jsonl")):
        for line in open(W + path):
            doc = json.loads(line)
            if not doc.get("skip"):
                result[(name, doc["id"])] = share(doc["text"])[0] >= 0.10
    return result


def vector(row):
    rates = {s["name"]: s["per_thousand_words"] for s in row["ai_sounding"]["signals"]}
    return [math.log1p(rates[name]) for name in NAMES]


def load(part, sets=("deslop",)):
    english = english_ids()
    wanted = part == "development"
    rows = {}
    human = []
    for line in open(signals_path):
        row = json.loads(line)
        if row.get("failed") or not row.get("ai_sounding") or row["ai_sounding"]["status"] != "available":
            continue
        if part != "all" and development(row["repository"]) != wanted:
            continue
        if row["set"] == "human":
            human.append(vector(row))
        elif row["set"] in sets and english.get((row["set"], row["id"])):
            rows.setdefault((row["set"], row["id"]), {})[row["side"]] = vector(row)
    pairs = [(sides["before"], sides["after"]) for sides in rows.values() if len(sides) == 2]
    return [p[0] for p in pairs], [p[1] for p in pairs], human


def fit(before, after, human, pair_weight=3.0, rounds=600, rate=0.5, penalty=0.01):
    weights, bias = [0.0] * len(NAMES), 0.0
    count = len(before) + len(human)
    for _ in range(rounds):
        grad, gbias = [0.0] * len(NAMES), 0.0
        for rows, label in ((before, 1), (human, 0)):
            for x in rows:
                p = 1 / (1 + math.exp(-(bias + sum(w * v for w, v in zip(weights, x)))))
                for k, v in enumerate(x):
                    grad[k] += (p - label) * v / count
                gbias += (p - label) / count
        for x, y in zip(before, after):
            difference = [a - b for a, b in zip(x, y)]
            p = 1 / (1 + math.exp(-sum(w * v for w, v in zip(weights, difference))))
            for k, v in enumerate(difference):
                grad[k] += pair_weight * (p - 1) * v / len(before)
        weights = [w - rate * (g + penalty * w) for w, g in zip(weights, grad)]
        bias -= rate * gbias
    return weights, bias


def index(weights, bias, x):
    return round(100 / (1 + math.exp(-(bias + sum(w * v for w, v in zip(weights, x))))))


def auc(positive, negative):
    ranked = sorted([(v, 1) for v in positive] + [(v, 0) for v in negative])
    total, i = 0.0, 0
    while i < len(ranked):
        j = i
        while j < len(ranked) and ranked[j][0] == ranked[i][0]:
            j += 1
        total += (i + j + 1) / 2 * sum(1 for k in range(i, j) if ranked[k][1])
        i = j
    return (total - len(positive) * (len(positive) + 1) / 2) / (len(positive) * len(negative))


def report(label, positive, later, negative):
    lower = sum(a < b for b, a in zip(positive, later))
    higher = sum(a > b for b, a in zip(positive, later))
    print(f"{label}: {len(positive)} pairs, {len(negative)} pages that people wrote; AUC {auc(positive, negative):.3f}; "
          f"index lower after the commit in {lower}, higher in {higher}, same in {len(positive) - lower - higher} "
          f"({lower / max(1, lower + higher):.1%} of pairs that differ)")
    for threshold in (60, 65, 70, 75, 80, 85):
        print(f"   index at or above {threshold}: before {sum(p >= threshold for p in positive) / len(positive):5.1%}, "
              f"after {sum(p >= threshold for p in later) / len(later):5.1%}, people {sum(p >= threshold for p in negative) / max(1, len(negative)):5.1%}")


if mode == "fit":
    before, after, human = load("development")
    weights, bias = fit(before, after, human)
    print(json.dumps({"names": NAMES, "weights": [round(w, 2) for w in weights], "bias": round(bias, 2)}))
    rounded, constant = [round(w, 2) for w in weights], round(bias, 2)
    report("development, fitted on all of it", [index(rounded, constant, x) for x in before], [index(rounded, constant, x) for x in after],
           [index(rounded, constant, x) for x in human])
    rng = random.Random(5)
    order_b, order_h = list(range(len(before))), list(range(len(human)))
    rng.shuffle(order_b)
    rng.shuffle(order_h)
    positive, later, negative = [], [], []
    for fold in range(5):
        train_b = [i for k, i in enumerate(order_b) if k % 5 != fold]
        train_h = [i for k, i in enumerate(order_h) if k % 5 != fold]
        w, b = fit([before[i] for i in train_b], [after[i] for i in train_b], [human[i] for i in train_h], rounds=300)
        for k, i in enumerate(order_b):
            if k % 5 == fold:
                positive.append(index(w, b, before[i]))
                later.append(index(w, b, after[i]))
        negative += [index(w, b, human[i]) for k, i in enumerate(order_h) if k % 5 == fold]
    report("development, five-fold", positive, later, negative)
else:
    part, model = sys.argv[4], json.loads(open(sys.argv[5]).read())
    for sets in (("deslop",), ("reserve",), ("web",)):
        before, after, human = load(part, sets)
        if before:
            report(f"{part}, {sets[0]}", [index(model["weights"], model["bias"], x) for x in before],
                   [index(model["weights"], model["bias"], x) for x in after], [index(model["weights"], model["bias"], x) for x in human])
