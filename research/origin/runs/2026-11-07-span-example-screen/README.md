# Author-disjoint span-example origin screen: negative result

This exposed-data screen tested a fixed-window, example-neighbor idea inspired
by [ExaGPT](https://aclanthology.org/2026.findings-acl.380/) on the public
[AITDNA](https://github.com/UKPLab/arxiv2026-aitdna) user-study documents.
Its [protocol](PROTOCOL.md), [runner](run.py), and [input freeze](freeze.json)
were committed before encoder inference. This is not an ExaGPT reproduction:
the encoder, segmentation, and retrieval decisions differ.

The offline APE/BGE encoder produced four 25-word span embeddings for each
eligible document. Each target span was compared with the five nearest
documents in each study class from the same task and other authors. The
AI-assisted minus human-only similarity margin was averaged over the four
spans. The unchanged lexical-richness feature was recomputed on the same
scored rows. No trained threshold or generated-model call was used.

| Measure | Span-neighbor margin | Lexical richness |
| --- | ---: | ---: |
| Overall directional AUROC, 352 comparable rows | 0.544 | 0.697 |
| Argumentative Writing, 121 rows | 0.415 | 0.703 |
| Creative Writing, 116 rows | 0.633 | 0.679 |
| Explanatory Writing, 115 rows | 0.615 | 0.711 |

All 362 ordered input rows were accounted for. Ten peer-review rows abstained
because that task had no human-only control. The encoder processed 1,408 spans
with no overlength rows. On the 78 author-and-task pairs with both conditions,
the margin rose in 44 and fell in 34. Against human-only rows, the post-result
Bot-majority and User-minority assisted splits had AUROC 0.564 and 0.444.

The [registered feasibility rule](PROTOCOL.md) required AUROC at least 0.70
overall, at least 0.60 in every comparable task, a gain of at least 0.05 over
richness, and positive direction in at least 60% of non-tied author-and-task
pairs. The fixed variant failed the overall, task, gain, and paired-direction
requirements. Retire this representation for the current origin track. The
contrast labels describe study conditions, not verified machine authorship of
every token. These results do not measure false-positive rates, generalization
to technical documentation, editorial quality, or [#349](https://github.com/stokaro/unswell/issues/349)
recall.

The [result](result.json) and [compressed score ledger](scores.json.gz) retain the denominator,
class/task information, pseudonymous author equivalence, neighbor row indices,
and numerical scores for independent verification. They contain no participant
text or neighbor excerpts. [The verifier](verify.py) recomputes neighbor
isolation, scores, AUROC, paired directions, coverage, and the frozen failure
decision; [negative tests](test_verify.py) demonstrate rejection of altered
records. The result pins the SHA-256 of the uncompressed JSON; the verifier
decompresses it before checking that hash. Run `make check-span-example-screen`
from the repository root. Replaying
encoder inference additionally needs the pinned local APE/BGE assets and
the source parquet listed in the protocol; the ordinary repository check does
not download or execute a model.

The earlier AITDNA lexical screen reported 0.686 richness AUROC on all 362
rows. The 0.697 value here is a recomputation on the 352 rows for which the
task-matched span method has a control. These numbers use different
denominators and should not be read as a richness improvement.
