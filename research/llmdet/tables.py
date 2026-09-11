"""Convert one published LLMDet n-gram dictionary into a bounded binary table.

The dictionary is the ``npz`` file of one model from the upstream release.
Its object array holds six dictionaries: the continuations and probabilities
of unigram, bigram, and trigram contexts. The table keeps every context and
every retained value, with contexts sorted for binary search, tokens as
16-bit integers, and probabilities as the original 16-bit floats. Nothing is
sampled, rounded, or dropped, and the source digest is recorded.
"""

import argparse
import hashlib
import json
import struct
import sys

import numpy as np

VERSION = "unswell-llmdet-table-v1"
MAGIC = b"ULDT\x01"


def digest(path):
    h = hashlib.sha256()
    with open(path, "rb") as f:
        for chunk in iter(lambda: f.read(1 << 20), b""):
            h.update(chunk)
    return h.hexdigest()


def order_arrays(contexts, probabilities, n):
    keys = np.array(list(contexts.keys()), dtype=np.uint32).reshape(-1, n)
    order = np.lexsort([keys[:, i] for i in range(n - 1, -1, -1)])
    keys = keys[order]
    tokens, probs, counts = [], [], np.zeros(len(order), dtype=np.uint32)
    values = list(contexts.values())
    weights = list(probabilities.values())
    for position, index in enumerate(order):
        t, p = values[index], weights[index]
        if len(t) != len(p):
            raise SystemExit("continuations and probabilities differ in length")
        counts[position] = len(t)
        tokens.append(np.asarray(t, dtype=np.uint16))
        probs.append(np.asarray(p, dtype=np.float16))
    offsets = np.zeros(len(order) + 1, dtype=np.uint32)
    offsets[1:] = np.cumsum(counts)
    return keys, offsets, np.concatenate(tokens), np.concatenate(probs)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--model", required=True, help="Upstream model key, such as gpt2")
    parser.add_argument("--vocab-size", required=True, type=int,
                        help="The vocabulary size the reference passes for this model; it enters the residual mass only")
    parser.add_argument("--npz", required=True)
    parser.add_argument("--output", required=True)
    args = parser.parse_args()
    source = np.load(args.npz, allow_pickle=True)
    items = list(source[source.files[0]])
    if len(items) != 6 or not all(isinstance(item, dict) for item in items):
        raise SystemExit("dictionary requires six dictionaries")
    if not (0 < args.vocab_size <= 65536):
        raise SystemExit("this table keeps 16-bit tokens")
    orders = []
    payload = []
    for n, (contexts, probabilities) in enumerate([(items[0], items[1]), (items[2], items[3]), (items[4], items[5])], start=1):
        if contexts.keys() != probabilities.keys():
            raise SystemExit("continuation and probability contexts differ")
        keys, offsets, tokens, probs = order_arrays(contexts, probabilities, n)
        if keys.size and (keys.max() >= 65536 or tokens.max() >= 65536):
            raise SystemExit("token outside the 16-bit range of this table")
        orders.append({"n": n, "contexts": int(len(keys)), "values": int(len(tokens))})
        payload.extend([keys.astype("<u4").tobytes(), offsets.astype("<u4").tobytes(),
                        tokens.astype("<u2").tobytes(), probs.view(np.uint16).astype("<u2").tobytes()])
    header = {"version": VERSION, "model": args.model, "vocab_size": args.vocab_size,
              "source": {"file": args.npz.rsplit("/", 1)[-1], "sha256": digest(args.npz)}, "orders": orders}
    encoded = json.dumps(header, sort_keys=True).encode()
    with open(args.output, "wb") as out:
        out.write(MAGIC)
        out.write(struct.pack("<I", len(encoded)))
        out.write(encoded)
        for part in payload:
            out.write(part)
    print(json.dumps({"model": args.model, "orders": orders, "sha256": digest(args.output)}), file=sys.stdout)


if __name__ == "__main__":
    main()
