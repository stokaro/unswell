#!/usr/bin/env python3
"""Embed every annotation item with a small local sentence encoder.

usage: embed.py <workspace> <docs file name> <output prefix>
Writes <prefix>.npy (float32 matrix) and <prefix>.keys.json (doc_id, item id per row).
"""
import json, sys, time
from pathlib import Path
import numpy as np
from fastembed import TextEmbedding

W = Path(sys.argv[1])
MODEL = "BAAI/bge-small-en-v1.5"
keys, texts = [], []
for line in open(W / sys.argv[2]):
    doc = json.loads(line)
    for item in doc["items"]:
        keys.append([doc["doc_id"], item["id"]])
        texts.append(" ".join(item["text"].split())[:2000])
started = time.time()
model = TextEmbedding(MODEL)
vectors = np.array(list(model.embed(texts, batch_size=128)), dtype=np.float32)
np.save(W / (sys.argv[3] + ".npy"), vectors)
(W / (sys.argv[3] + ".keys.json")).write_text(json.dumps({"model": MODEL, "keys": keys}))
print(f"{sys.argv[2]}: rows={len(keys)} dim={vectors.shape[1]} seconds={time.time() - started:.0f}")
