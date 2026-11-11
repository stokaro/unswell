#!/usr/bin/env python3
"""Word n-gram linear model: transfer from generated documents to real pages, and simpler feature sets.

usage: study3.py <workspace> <train docs,...> <test docs,...>
"""
import json, re, sys
from pathlib import Path
import numpy as np
from sklearn.feature_extraction.text import CountVectorizer, TfidfVectorizer
from sklearn.linear_model import LogisticRegression
from sklearn.metrics import average_precision_score, precision_recall_curve
from sklearn.model_selection import GroupKFold
from scipy.sparse import hstack

W = Path(sys.argv[1])
TOKEN = r"[A-Za-z]+(?:[-'][A-Za-z]+)*"


def load(names):
    rows = []
    for name in names.split(","):
        for line in open(W / name):
            doc = json.loads(line)
            paths = [W / "ann" / n / f"{doc['doc_id']}.json" for n in ("astra", "claude")]
            if not all(p.exists() for p in paths):
                continue
            astra, claude = [{f["item"]: f for f in json.loads(p.read_text())["flags"]} for p in paths]
            for item in doc["items"]:
                a, c = astra.get(item["id"]), claude.get(item["id"])
                pa = bool(a and a["label"] == "needs_revision")
                pc = bool(c and c["label"] == "needs_revision")
                if pa and pc:
                    label = 1
                elif not a and not c:
                    label = 0
                else:
                    continue
                rows.append((" ".join(item["text"].split()), label, doc["repository"]))
    return rows


def recall_at(y, score, floor=0.85):
    precision, recall, _ = precision_recall_curve(y, score)
    ok = [r for p, r in zip(precision[:-1], recall[:-1]) if p >= floor]
    return max(ok) if ok else 0.0


def report(name, y, score):
    print(f"{name:34s} positives={int(y.sum()):5d} of {len(y):6d} average precision={average_precision_score(y, score):.3f} "
          f"recall at 85%={recall_at(y, score):.3f} at 70%={recall_at(y, score, 0.7):.3f}", flush=True)


VARIANTS = {
    "word 1-3 tfidf + char 3-5": lambda: (TfidfVectorizer(ngram_range=(1, 3), min_df=2, sublinear_tf=True, token_pattern=TOKEN),
                                          TfidfVectorizer(analyzer="char_wb", ngram_range=(3, 5), min_df=3, sublinear_tf=True)),
    "word 1-3 binary": lambda: (CountVectorizer(ngram_range=(1, 3), min_df=2, binary=True, token_pattern=TOKEN),),
    "word 1-2 binary": lambda: (CountVectorizer(ngram_range=(1, 2), min_df=2, binary=True, token_pattern=TOKEN),),
}


def fit_predict(variant, train, test, C):
    vectorizers = VARIANTS[variant]()
    x_train = hstack([v.fit_transform([r[0] for r in train]) for v in vectorizers]).tocsr()
    x_test = hstack([v.transform([r[0] for r in test]) for v in vectorizers]).tocsr()
    model = LogisticRegression(C=C, class_weight="balanced", max_iter=3000)
    model.fit(x_train, np.array([r[1] for r in train]))
    return model.decision_function(x_test)


train, test = load(sys.argv[2]), load(sys.argv[3])
y_test = np.array([r[1] for r in test])
print(f"train items={len(train)} positives={sum(r[1] for r in train)}; test items={len(test)} positives={int(y_test.sum())}")
for variant, C in (("word 1-3 tfidf + char 3-5", 4.0), ("word 1-3 binary", 0.3), ("word 1-2 binary", 0.3)):
    report("transfer: " + variant, y_test, fit_predict(variant, train, test, C))
# cross-validation over repositories on the union: does adding real pages help?
union = train + test
y = np.array([r[1] for r in union])
groups = np.array([r[2] for r in union])
is_test = np.array([False] * len(train) + [True] * len(test))
scores = np.zeros(len(union))
for a, b in GroupKFold(n_splits=5).split(union, y, groups):
    scores[b] = fit_predict("word 1-3 binary", [union[i] for i in a], [union[i] for i in b], 0.3)
report("union cross-validation, all", y, scores)
report("union cross-validation, test part", y[is_test], scores[is_test])
