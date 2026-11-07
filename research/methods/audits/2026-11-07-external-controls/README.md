# External-method control pass on current Unswell

Status: exposed, constructed-control baseline for [#434](https://github.com/stokaro/unswell/issues/434),
November 7, 2026. This is not an upstream benchmark reproduction, an estimate of
population precision, or a #349 cohort result. The source audit in #434 supplied
method ideas and counterexamples; no upstream code or data was imported.

The baseline is the unmodified `bae57aa96afd22e5f64735800f85a3be81d8ad46`
CLI with the repository configuration. [The 18 inputs](controls.json) include
six legitimate word repetitions, one accidental duplicate, additive-adverb and
technical-vocabulary controls, two meaning-sensitive sentences, and three
editorial constructions. The [freeze](freeze.json) sealed the inputs and runner
before execution. The first runner stopped on its fourth case: it read
`finding.span` although the public JSON report uses `finding.primary.span`.
No score or judgment was made from that incomplete pass. The
[v2 amendment](freeze-v2.json) changes only report extraction and preserves the
same cases, binary, and current-main baseline.

The v2 replay completed all 18 cases. [Complete raw reports](reports-v2.json.gz)
and the [compact result](result-v2.json) are retained. All reports were complete,
with zero operational errors or abstentions. The result SHA-256 is
`bafac0f55431c3201e2a1a4533243ab6ec7bf404834b42c88cfc6ce1b9b11177`;
the deterministic compressed raw-report SHA-256 is
`0bd97158adfe50f6951df92f9354294d4253e3ca523ae318f2b6f3da17f97e7c`.
The binary SHA-256 is
`3de87af84bab750e0c0768f3df833910b12a4123e2869ba43f8034398bc2d2ac`.
The independent [verifier](verify.py) checked all 18 report/source identities,
the original byte spans, two delivered findings, one shared configuration,
zero skipped rules, and the raw and compact output hashes. Its
[negative test](test_verify.py) rejects a changed quote.

| Control family | Observed current behavior | Disposition |
| --- | --- | --- |
| Correct homographs, hyphen boundary, quoted example, protected code | Six legitimate cases, zero findings | Existing adjacent-word guards cover these probes; no new rule justified. |
| Accidental `requires requires` | One `repetition.adjacent-word` finding on the original byte span | Already implemented. |
| Additive adverbs and technical vocabulary | Six legitimate cases, zero findings | Useful negative controls. No synonym or blacklist rule justified. |
| Condition, negation, version, identifier | Two legitimate cases, zero findings | Useful advice-safety controls; absence of a finding alone does not prove rewrite safety. |
| Deprecated-features table lead | Zero findings | Potential missing source-local heading/table scaffolding. The existing `repetition.heading-echo` requires a longer heading and close lexical-set overlap; this short lead is outside that contract. The prior user review judged the redundant table introduction removable, but this constructed case is not a frozen #349 event. |
| Rephrased token assertion | Zero findings | Potential missing short paraphrastic restatement. `repetition.repeated-claim` handles narrow exact and closed paraphrase forms; this inversion is outside those forms. Necessity and safe revision still need source-context review. |
| `It is possible... This is done by...` | One `filler.instruction-scaffolding` finding | Already implemented; a second generic instruction rule would duplicate it. |

The two potential misses are distinct mechanisms, but neither is admitted as a
new operator from this tiny set. The next bounded step is to map them to the
unchanged #349 event ledger and source roles, check legitimate near-neighbors,
then freeze at most two exact operators only if an incremental gain is plausible.
No generated-text origin feature, upstream watermark score, or model call can
substitute for that source-local editorial check. The 123 exposed reference
events and 80% FULL-event / 85% accepted-delivery gates remain unchanged.

## Gap disposition after the control pass

The frozen #349 labels contain 57 `empty_framing`, 26 `wordiness`, 21
`vague_claims`, 11 `unjustified_intensifiers`, seven `needless_repetition`, and
one `needless_complexity` events. These are reference categories, not outcomes
from the 18 controls. Neither constructed potential miss is itself one of the
123 target quotes. The table lead is a possible member of the broad framing
family; the inverted token assertion is a possible member of the narrow
repetition family. Neither relationship proves that a new rule would recover
any specific event or deliver an accepted finding.

| #434 source or method | Current gap decision |
| --- | --- |
| Harper and LanguageTool repetition exceptions | Useful controls. The current adjacent-word rule passed the bounded positive and negative probes, so no missing exception was found. The pinned Harper method is recorded in [registry revision 5](../../registry-v1.json). |
| Vale recognition context and narrow target | Research-only design reference for any later multi-location operator. Current source spans and protected regions remain the contract. |
| textlint source-code-fixer | Deferred: no automatic rewrite is proposed; its UTF-16 offsets cannot replace original UTF-8 byte spans. |
| retext, proselint, write-good | Phrase and readability baselines only. Technical-vocabulary controls provide counterexamples to blanket synonym rules. |
| slopscore | Research-only weak-signal and suppression design; no weights, sigmoid, or author score imported. No current-engine score comparison was run. |
| TextDescriptives, LexicalRichness, GEN, ling-feats, ELFEN | Measurement references. The completed richness/entropy screen is origin-only; these statistics do not identify a mandatory edit in #349. No added feature or classifier is selected. |
| DisSim, Simple PPDB, ASSET, EASSE | Research-only revision and evaluation references. A transformed sentence still needs proposition and technical-preservation review. |
| SynthID, KGW, TextSeal, SWEET, other watermark projects | Deferred known-profile origin channel, separately from editorial diagnostics and gates. |

This is a **stop decision for operator implementation from the present control
set**, not a stop for #349. The next #434 pass needs actual complete-page
context and an event-level gap map before proposing a frozen operator. Repeated
threshold tuning of the retired #349 selector or a document-level origin score
would not supply that evidence.

Watermark verification is **deferred as a separate known-profile origin
capability**. The methods in #434 need a compatible generator scheme, key,
tokenizer, and detection parameters. They cannot verify arbitrary unsourced
technical prose, and a missing profile must never become a human-authorship
verdict. No runtime watermark code is justified by this editorial control pass.

Reproduce from the repository root with an exact-main CLI binary:

```sh
python3 -B research/methods/audits/2026-11-07-external-controls/run-v2.py \
  --binary /path/to/unswell
```

The script verifies its frozen control and code hashes, then stops on malformed
or incomplete output. It records the binary hash so a replay with another build
remains distinguishable.

```sh
python3 -B research/methods/audits/2026-11-07-external-controls/verify.py
python3 -B -m unittest research/methods/audits/2026-11-07-external-controls/test_verify.py
```
