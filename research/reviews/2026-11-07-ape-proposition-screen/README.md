# APE proposition screen: negative result for #349

The [protocol](PROTOCOL.md) was frozen before inference against Unswell main
`6e724618005623eddbd325653c0e04e7c38b317d`. The run used the published
[APE implementation](https://github.com/SapienzaNLP/APE) and
[`ape-bge-small` checkpoint](https://huggingface.co/sapienzanlp/ape-bge-small),
with the exact commits, revisions, asset hashes, and offline settings in
[`assets.json`](assets.json) and [`run.json`](run.json). No model assets are
redistributed here. APE is a proposition similarity model, not an editorial
classifier or an AI-origin detector.

The complete 36-page frozen source set yielded 4,139 eligible sentences and
2,192 directional source-local pair scores. The fixed top-three-per-page rule
selected 108 candidates. The [independent assessment](assessment.json) verified
source hashes, original-byte spans, pair boundaries, scoring maxima, ranking,
and all 123 unchanged target quotes. Only seven selected candidates overlap
any frozen event. They cover every target span of just five events that were
not already FULL in the current-main editorial baseline: two whole-page, two
contextual, and one historical confirmation event. This is **only a maximum
possible span overlap**, not editorial diagnosis credit. The predeclared
admission gate requires eight genuinely new FULL events, so the configuration
cannot pass even if all five diagnoses prove valid.

The predeclared review order called for inspecting positive and legitimate
control pair scores and reviewing all 108 candidates in complete page context.
We stopped before that review because a necessary admission gate had already
failed. No candidate was accepted or rejected editorially, and no finding was
delivered to users. Acceptance precision, gained or lost FULL credit, supplied
advice safety, and prospective performance remain unmeasured. This explicit
early stop is a deviation from the planned full review, not a completed #349
qualification. The frozen pages were exposed; no reserved prospective pages
were opened. The current-main #349 result remains 24/57, 13/46, and 4/20
FULL events, with accepted delivered findings of 78/122, 18/24, and 6/9.
Even adding every new span overlap as if it earned credit would reach only
26/57, 15/46, and 5/20 FULL events, versus the required 46, 37, and 16.

The selected high-score examples include factual release and parser behavior
statements. APE's similarity score alone does not establish that the writing
needs a change. This result disfavors this specific adjacent-sentence,
top-three-per-page configuration for #349. It does not test other APE uses or
other local methods. Do not tune the rank limit or score cutoff on these
already-exposed pages and call the result prospective.

## Reproduce the assessment

From the repository root, with Python 3, run the committed positive and
negative verifier checks with `make check-ape-screen`. For a standalone JSON
assessment:

```sh
PYTHONDONTWRITEBYTECODE=1 python3 research/reviews/2026-11-07-ape-proposition-screen/assess.py \
  research/reviews/2026-11-07-ape-proposition-screen \
  /tmp/unswell-ape-assessment.json
```

The [compressed scores](scores.json.gz), [selected candidates](top-candidates.json),
and [source hash manifest](source-hashes.json) are committed for the offline
check. To repeat inference, use [`prepare_inputs.py`](prepare_inputs.py), the
[`export` command](export/main.go.txt), and [`score.py`](score.py) with the pinned
upstream code and model assets from the protocol. The exporter is archived as
`main.go.txt` so it does not become a public Go package subject to Unswell's
API ledger; copy it to a temporary `main.go` before running it with `go run`.
The offline assessment needs no model weights or network access.
