#!/usr/bin/env python3
"""Student models with sentence embeddings and document context.

usage: study2.py <workspace> [--curve]
Same labels and repository-grouped evaluation as study.py.
"""
import collections, json, math, re, sys
from pathlib import Path
import numpy as np
from scipy.sparse import csr_matrix, hstack
from sklearn.feature_extraction.text import TfidfVectorizer
from sklearn.linear_model import LogisticRegression
from sklearn.metrics import average_precision_score, precision_recall_curve
from sklearn.model_selection import GroupKFold
from sklearn.preprocessing import StandardScaler

W = Path(sys.argv[1])
CURVE = "--curve" in sys.argv
WORD = re.compile(r"[A-Za-z][A-Za-z'-]*")
HEADING = re.compile(rb"^#{1,6} +(.+)$", re.M)


def load(docs_name, emb_prefix):
    vectors = np.load(W / (emb_prefix + ".npy"))
    keys = json.loads((W / (emb_prefix + ".keys.json")).read_text())["keys"]
    row_of = {tuple(k): i for i, k in enumerate(keys)}
    rows, embed, context = [], [], []
    for line in open(W / docs_name):
        doc = json.loads(line)
        paths = [W / "ann" / name / f"{doc['doc_id']}.json" for name in ("astra", "claude")]
        if not all(p.exists() for p in paths):
            continue
        marks = [{f["item"]: f for f in json.loads(p.read_text())["flags"]} for p in paths]
        raw = doc["text"].encode()
        headings = [(m.start(), set(w.lower() for w in WORD.findall(m.group(1).decode("utf-8", "replace"))))
                    for m in HEADING.finditer(raw)]
        items = doc["items"]
        count = len(items)
        matrix = np.stack([vectors[row_of[(doc["doc_id"], it["id"])]] for it in items]) if items else np.zeros((0, vectors.shape[1]))
        sims = matrix @ matrix.T if count else np.zeros((0, 0))
        centroid = matrix.mean(axis=0) if count else None
        previous = []
        for index, item in enumerate(items):
            text = " ".join(item["text"].split())
            tokens = set(w.lower() for w in WORD.findall(text))
            before = [h for h in headings if h[0] < item["start"]]
            heading = before[-1] if before else (0, set())
            first_after = not any(heading[0] < other["start"] < item["start"] for other in items)
            last_before = not any(item["start"] < other["start"] and
                                  not any(item["start"] < h[0] < other["start"] for h in headings) for other in items)
            overlap = len(tokens & heading[1]) / max(1, len(heading[1]))
            repeat = max((len(tokens & p) / max(1, len(tokens | p)) for p in previous), default=0.0)
            previous.append(tokens)
            labels = []
            for mark in marks:
                flag = mark.get(item["id"])
                labels.append(1 if flag and flag["label"] == "needs_revision" else (None if flag else 0))
            earlier = sims[index, :index]
            later = sims[index, index + 1:]
            dense = [index / max(1, count - 1), float(item["kind"] == "fragment"), math.log1p(len(text.split())),
                     float(first_after), float(last_before), overlap, repeat, float(text.endswith(":")),
                     float(index == 0), float(index == count - 1),
                     float(earlier.max()) if earlier.size else 0.0, float(later.max()) if later.size else 0.0,
                     float(np.sort(earlier)[-3:].mean()) if earlier.size else 0.0,
                     float(matrix[index] @ centroid / (np.linalg.norm(centroid) + 1e-9)),
                     float(sims[index, index - 1]) if index else 0.0,
                     float(sims[index, index + 1]) if index + 1 < count else 0.0]
            zero = np.zeros(matrix.shape[1], dtype=np.float32)
            rows.append({"doc": doc["doc_id"], "repo": doc["repository"], "origin": doc["origin"],
                         "family": doc["family"] or doc["origin"], "id": item["id"], "text": text,
                         "a": labels[0], "c": labels[1]})
            embed.append(np.concatenate([matrix[index], matrix[index - 1] if index else zero,
                                         matrix[index + 1] if index + 1 < count else zero]))
            context.append(dense)
    return rows, np.array(embed, dtype=np.float32), np.array(context, dtype=np.float32)


def recall_at(y, score, floor=0.85):
    precision, recall, threshold = precision_recall_curve(y, score)
    ok = [(r, t) for p, r, t in zip(precision[:-1], recall[:-1], threshold) if p >= floor]
    return max(ok) if ok else (0.0, float("inf"))


def design(variant, train_index, test_index, rows, embed, context):
    parts_train, parts_test = [], []
    if "tfidf" in variant:
        word = TfidfVectorizer(ngram_range=(1, 3), min_df=2, sublinear_tf=True, max_features=300000)
        char = TfidfVectorizer(analyzer="char_wb", ngram_range=(3, 5), min_df=3, sublinear_tf=True, max_features=300000)
        train_text = [rows[i]["text"] for i in train_index]
        test_text = [rows[i]["text"] for i in test_index]
        parts_train += [word.fit_transform(train_text), char.fit_transform(train_text)]
        parts_test += [word.transform(test_text), char.transform(test_text)]
    dense = []
    if "emb" in variant:
        dense.append(embed[:, :384] if "self" in variant else embed)
    if "ctx" in variant:
        dense.append(context)
    if dense:
        block = np.hstack(dense)
        scaler = StandardScaler().fit(block[train_index])
        weight = 0.15 if "tfidf" in variant else 1.0
        parts_train.append(csr_matrix(scaler.transform(block[train_index]) * weight))
        parts_test.append(csr_matrix(scaler.transform(block[test_index]) * weight))
    return hstack(parts_train).tocsr(), hstack(parts_test).tocsr()


def crossval(variant, index, y, groups, rows, embed, context, fraction=1.0):
    scores = np.zeros(len(index))
    rng = np.random.default_rng(0)
    for train, test in GroupKFold(n_splits=5).split(index, y, groups):
        if fraction < 1.0:
            train = rng.choice(train, size=int(len(train) * fraction), replace=False)
        x_train, x_test = design(variant, index[train], index[test], rows, embed, context)
        model = LogisticRegression(C=0.5 if "tfidf" not in variant else 4.0, class_weight="balanced", max_iter=3000)
        model.fit(x_train, y[train])
        scores[test] = model.decision_function(x_test)
    return scores


rows, embed, context = load("docs.jsonl", "emb-v0")
index = np.array([i for i, r in enumerate(rows) if r["a"] is not None and r["c"] is not None and r["a"] == r["c"]])
y = np.array([rows[i]["a"] for i in index])
groups = np.array([rows[i]["repo"] for i in index])
print(f"labeled items={len(index)} positives={int(y.sum())} repositories={len(set(groups))}")
results = {}
for variant in ("tfidf", "emb-self", "emb", "emb+ctx", "tfidf+emb+ctx"):
    scores = crossval(variant, index, y, groups, rows, embed, context)
    results[variant] = scores
    print(f"{variant:14s} average precision={average_precision_score(y, scores):.3f} "
          f"recall at 85%={recall_at(y, scores)[0]:.3f} at 80%={recall_at(y, scores, 0.8)[0]:.3f} "
          f"at 70%={recall_at(y, scores, 0.7)[0]:.3f}", flush=True)
best = max(results, key=lambda v: average_precision_score(y, results[v]))
if CURVE:
    print(f"learning curve for {best}:")
    for fraction in (0.1, 0.25, 0.5, 1.0):
        scores = crossval(best, index, y, groups, rows, embed, context, fraction)
        print(f"   {fraction:4.2f} -> recall at 85%={recall_at(y, scores)[0]:.3f} average precision={average_precision_score(y, scores):.3f}", flush=True)
np.save(W / "oof-scores.npy", results[best])
json.dump({"variant": best, "index": index.tolist()}, open(W / "oof-index.json", "w"))
