# WikiSQE vague-label transfer: negative screen for #349

The [protocol](PROTOCOL.md) was frozen before the download and first inference. This bounded offline screen trains a non-generative TF-IDF/logistic classifier on the authors' [WikiSQE `vague` split](https://huggingface.co/datasets/ando55/WikiSQE_experiment), then ranks original-byte-bound sentence rows on the complete exposed 36-page #349 cohort. It does not infer AI authorship or make a product change. The dataset revision, three Parquet SHA-256 values, source inputs, model parameters, versions, and output hashes are recorded in [result.json](result.json). The pinned [paper](https://ojs.aaai.org/index.php/AAAI/article/download/29717/31231) and [authors' repository](https://github.com/ken-ando/WikiSQE) describe the source task.

The model trained on 24,746 Wikipedia sentences. On the untouched, artificially balanced 1,000-sentence WikiSQE test split it reached 71.1% accuracy and 71.9% binary F1. This is only an in-domain sanity check. Wikipedia's untagged sentences are not confirmed clean editorial examples, and the paper explicitly does not evaluate realistic low-prevalence deployment precision.

The fixed top-three-per-page rule selected 108 Unswell candidates. Only **three previously non-FULL events** have all their frozen target spans overlapped by these candidates, below the protocol's necessary eight-event review gate. All three belong to `empty_framing`; **zero new `vague_claims` events** are covered. One already FULL `vague_claims` event is also overlapped. The cohort ceilings for *new* events are 2/57 whole page, 1/46 context, and 0/20 historical confirmation. These are span-overlap upper bounds; no candidate received editorial diagnosis credit or safe-edit review. The screen stopped before reviewing all 108 candidates, as predeclared.

The transfer failure is specific to the WikiSQE `vague` label, this sparse classifier, these source-bound fragments, and this top-three selection. It does not refute every small encoder or every external editorial dataset. The latest reviewed Unswell result remains 28/57, 16/46, and 4/20 FULL events, with whole-page technical accepted delivery 82/126. The 80% recall and 85% accepted-delivery goals remain unmet; #349 stays open. Do not tune this screen on the exposed pages and present the result as prospective evidence.

## Reproduce and verify

The offline verifier validates saved score integrity, original source hashes and byte spans, fixed ranking, all 123 frozen event joins, the latest seven reviewed FULL additions, and the early-stop decision using only Python's standard library:

```sh
python3 -B research/reviews/2026-11-09-wikisqe-vague-screen/verify.py
python3 -B -m unittest discover -s research/reviews/2026-11-09-wikisqe-vague-screen -p 'test_*.py'
```

To rerun training, fetch only `vague/{train,val,test}.parquet` from dataset revision `e6a3da39b44df3efa2c75706ccba1dc26dd17076` and verify the SHA-256 values in `result.json`. Use Python 3.12.15, scikit-learn 1.7.2, and pyarrow 21.0.0, then run:

```sh
python research/reviews/2026-11-09-wikisqe-vague-screen/run.py /path/to/pinned/vague
```

The saved score file and [top candidates](top-candidates.json) permit offline assessment without redistributing the external training data or model weights. The standard-library verifier cannot independently reproduce model inference; rerunning `run.py` with the pinned dataset does that. Its separate negative tests reject a changed selection and inflated event ceiling.
