#!/usr/bin/env python3
"""Agreement between the two annotators and a first local student model.

usage: study.py <workspace> [--curve]

Labels: an item is positive when both annotators mark needs_revision, negative
when neither lists it, and masked otherwise. Evaluation holds out whole
repositories. The reported number is recall at 85% precision on held-out items.
"""
import collections, json, math, re, sys
from pathlib import Path
import numpy as np
from scipy.sparse import csr_matrix, hstack
from sklearn.feature_extraction.text import TfidfVectorizer
from sklearn.linear_model import LogisticRegression
from sklearn.metrics import average_precision_score, precision_recall_curve
from sklearn.model_selection import GroupKFold

W = Path(sys.argv[1])
CURVE = "--curve" in sys.argv
WORD = re.compile(r"[A-Za-z][A-Za-z'-]*")
HEADING = re.compile(rb"^#{1,6} +(.+)$", re.M)


def load():
    rows = []
    for line in open(W / "docs.jsonl"):
        doc = json.loads(line)
        paths = [W / "ann" / name / f"{doc['doc_id']}.json" for name in ("astra", "claude")]
        if not all(p.exists() for p in paths):
            continue
        answers = [json.loads(p.read_text()) for p in paths]
        marks = [{f["item"]: f for f in a["flags"]} for a in answers]
        raw = doc["text"].encode()
        headings = [(m.start(), set(w.lower() for w in WORD.findall(m.group(1).decode("utf-8", "replace"))))
                    for m in HEADING.finditer(raw)]
        previous = []
        count = len(doc["items"])
        for index, item in enumerate(doc["items"]):
            text = " ".join(item["text"].split())
            tokens = set(w.lower() for w in WORD.findall(text))
            before = [h for h in headings if h[0] < item["start"]]
            heading = before[-1] if before else (0, set())
            first_after = not any(heading[0] < other["start"] < item["start"] for other in doc["items"])
            overlap = len(tokens & heading[1]) / max(1, len(heading[1]))
            repeat = max((len(tokens & p) / max(1, len(tokens | p)) for p in previous), default=0.0)
            previous.append(tokens)
            labels = []
            for mark in marks:
                flag = mark.get(item["id"])
                labels.append(1 if flag and flag["label"] == "needs_revision" else (None if flag else 0))
            rows.append({"doc": doc["doc_id"], "repo": doc["repository"], "origin": doc["origin"],
                         "family": doc["family"] or "original", "operation": doc["operation"],
                         "id": item["id"], "text": text, "a": labels[0], "c": labels[1],
                         "cats": [m[item["id"]]["categories"] for m in marks if item["id"] in m],
                         "dense": [index / max(1, count - 1), float(item["kind"] == "fragment"),
                                   math.log1p(len(text.split())), float(first_after), overlap, repeat,
                                   float(text.endswith(":")), float(index == 0), float(index == count - 1)]})
    return rows


def kappa(pairs):
    n = len(pairs)
    agree = sum(a == c for a, c in pairs) / n
    pa = sum(a for a, _ in pairs) / n
    pc = sum(c for _, c in pairs) / n
    expected = pa * pc + (1 - pa) * (1 - pc)
    return agree, (agree - expected) / (1 - expected) if expected < 1 else float("nan")


def recall_at(y, score, floor=0.85):
    precision, recall, threshold = precision_recall_curve(y, score)
    ok = [(r, t) for p, r, t in zip(precision[:-1], recall[:-1], threshold) if p >= floor]
    return max(ok) if ok else (0.0, float("inf"))


def features(train, test):
    word = TfidfVectorizer(ngram_range=(1, 3), min_df=2, sublinear_tf=True, lowercase=True, max_features=300000)
    char = TfidfVectorizer(analyzer="char_wb", ngram_range=(3, 5), min_df=3, sublinear_tf=True, max_features=300000)
    xs = []
    for rows, fit in ((train, True), (test, False)):
        texts = [r["text"] for r in rows]
        dense = csr_matrix(np.array([r["dense"] for r in rows]))
        xs.append(hstack([word.fit_transform(texts) if fit else word.transform(texts),
                          char.fit_transform(texts) if fit else char.transform(texts), dense]).tocsr())
    return xs


def crossval(rows, fraction=1.0, seed=0):
    y = np.array([r["y"] for r in rows])
    groups = np.array([r["repo"] for r in rows])
    scores = np.zeros(len(rows))
    rng = np.random.default_rng(seed)
    for train_index, test_index in GroupKFold(n_splits=5).split(rows, y, groups):
        if fraction < 1.0:
            train_index = rng.choice(train_index, size=int(len(train_index) * fraction), replace=False)
        train = [rows[i] for i in train_index]
        test = [rows[i] for i in test_index]
        x_train, x_test = features(train, test)
        model = LogisticRegression(C=4.0, class_weight="balanced", max_iter=2000)
        model.fit(x_train, y[train_index])
        scores[test_index] = model.decision_function(x_test)
    return y, scores


rows = load()
docs = len({r["doc"] for r in rows})
both = [r for r in rows if r["a"] is not None and r["c"] is not None]
agree, k = kappa([(r["a"], r["c"]) for r in both])
pos = [r for r in both if r["a"] == 1 and r["c"] == 1]
neg = [r for r in both if r["a"] == 0 and r["c"] == 0]
print(f"documents with both annotations={docs} items={len(rows)} masked as uncertain={len(rows) - len(both)}")
print(f"astra flags={sum(r['a'] == 1 for r in both)} claude flags={sum(r['c'] == 1 for r in both)} "
      f"both={len(pos)} either={sum(r['a'] == 1 or r['c'] == 1 for r in both)}")
print(f"raw agreement={agree:.4f} Cohen kappa={k:.3f}")
by = collections.defaultdict(lambda: [0, 0])
for r in both:
    by[r["family"]][1] += 1
    by[r["family"]][0] += r["a"] == 1 and r["c"] == 1
for family, (p, n) in sorted(by.items()):
    print(f"   {family:9s} consensus positives {p:5d} of {n:6d} items ({100 * p / n:.2f}%)")
cats = collections.Counter(c for r in pos for group in r["cats"] for c in group)
print("categories on consensus positives:", dict(cats.most_common()))

labeled = []
for r in pos + neg:
    r = dict(r)
    r["y"] = int(r["a"] == 1 and r["c"] == 1)
    labeled.append(r)
if len(pos) < 50:
    print("too few consensus positives for a model yet")
    sys.exit(0)
y, scores = crossval(labeled)
recall, threshold = recall_at(y, scores)
print(f"\nstudent, held-out repositories: items={len(y)} positives={int(y.sum())} "
      f"average precision={average_precision_score(y, scores):.3f} recall at 85% precision={recall:.3f}")
for floor in (0.7, 0.8, 0.9):
    print(f"   recall at {int(floor * 100)}% precision={recall_at(y, scores, floor)[0]:.3f}")
order = np.argsort(-scores)
print("top held-out predictions:")
for i in order[:12]:
    print(f"   {'POS' if y[i] else 'neg'} {scores[i]:+.2f} {labeled[i]['text'][:110]}")
if CURVE:
    print("learning curve (fraction of training items -> recall at 85% precision, average precision):")
    for fraction in (0.1, 0.25, 0.5, 1.0):
        yy, ss = crossval(labeled, fraction)
        print(f"   {fraction:4.2f} -> {recall_at(yy, ss)[0]:.3f}  {average_precision_score(yy, ss):.3f}")
