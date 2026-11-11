# Non-LLM text detection and revision methods: November 2026 review

Status: source review and linked exposed-data feature screens, updated November
11, 2026. The APE follow-ups produced negative source-local and span-example
screens. EditLens passed paired directional screens on exposed complete
technical documents. Its fixed academic threshold flagged 20 of 462 historical
technical controls. A later, newly frozen historical-control threshold passed
its low-alert gate but detected only 4 of 228 exposed AI-polished responses.
A pinned MELD follow-up detected 147/228 complete generations but only 31/228
AI-polished responses at 1/351 observed historical-control alerts, missing its
registered polish floor. Neither is a qualified origin detector, and no #349
editorial qualification is claimed.

This review extracts methodological ideas; it does not select code or data to
import. "Non-LLM" below means that detection can run without calling a
generative language model or a remote service. A local parser or a compact
discriminative classifier may still be needed. Proxy language model detectors
are outside this runtime scope; their benchmarks can still test robustness.

## Two different targets

An **origin** signal estimates whether a document resembles generated text. An
**editorial** finding identifies a specific passage that needs a justified
revision. These are separate outcomes under [ADR 0015](../../docs/adr/0015-research-methodology.md)
and [ADR 0036](../../docs/adr/0036-llm-pattern-evidence.md). A document-level
authorship classifier cannot by itself localize an editable problem or establish
that a suggested revision preserves meaning. Issue [#349](https://github.com/stokaro/unswell/issues/349)
requires full-event recall, finding acceptance, source localization, and edit
safety on frozen editorial references.

## Reviewed studies and methods

| Method | Published evidence | Candidate use and limit |
| --- | --- | --- |
| Interpretable linguistic features plus linear SVM | [El Attar et al., Findings of AACL 2026](https://aclanthology.org/2026.findings-aacl.10/) evaluate 284 features, 30 generators, 10 domains, and eight generalization testbeds. Their aggregate unseen-domain/unseen-model test has macro F1 0.811 and AUROC 0.883, while leave-one-domain/model-pair-out macro F1 is 0.602. Lexical richness is the most consistent feature group; many other features depend on context. [Code and experiment layout](https://github.com/YassirELATTAR/ling-feats-mgt). | Reproduce the *evaluation design*: hold out both documentation source and generator, ablate feature groups, and report per-domain failure. A small offline classifier is an origin baseline, not an editorial detector. The paper uses Python linguistic tooling; its reported numbers are not performance claims for Unswell. |
| Generation versus editing stylometry | [Shan et al., EMNLP 2026](https://aclanthology.org/2026.emnlp-main.1725/) find that entropy and lexical diversity distinguish generated from human text across eight generators and five domains, while AI editing has a different, weaker signature. | Measure generated and AI-edited documents separately. Do not infer the need for an edit from an origin score. |
| Sentence transition structure | [Park and Kim, Findings of EMNLP 2026](https://aclanthology.org/2026.findings-emnlp.1988/) study variance and repeated similarity bursts between sentences, particularly near paragraph boundaries. Their detector uses a graph neural network; the authors state the relational formulation is the contribution. | Test simple source-local transition features and repeated discourse frames offline. These are new hypotheses; the published graph model's accuracy does not transfer to a rule implementation. |
| Span-based editorial taxonomy | [Measuring AI "Slop" in Text, October 2026 revision](https://arxiv.org/html/2509.19163v3) annotates information utility, quality, and style, including repetition, templatedness, coherence, tone, and verbosity. Binary slop judgments have weak agreement even when some span categories agree better. | Map one frozen #349 event to a concrete mechanism and source span. Preserve category-level results instead of collapsing them into a binary slop label. |
| Source-conditioned revision safety | [Centonze, WASP 2026](https://aclanthology.org/2026.wasp-1.15/) proposes checking propositions, modality, hedging, evidential relations, attribution, and argumentative role when assessing rewrites. Its 40-claim demonstration is exploratory. | Add explicit checks for changed numbers, names, negation, modality, and attribution. This is a review rubric, not a validated automatic checker. |
| Benchmark artifact checks | [Voznyuk and Grabovoy, DocInsights 2026](https://aclanthology.org/2026.docinsights-1.13/) find that PDF extraction and line-wrapping artifacts explain some near-perfect detector results; rewrapping alone changes decisions. [Gao et al., ACL 2026](https://aclanthology.org/2026.acl-long.1998/) find feature-effect inversions under style imitation. | Normalize transport artifacts only in a separately declared counterfactual; audit source format, genre, author style, and model shifts before accepting a feature. |
| Matched web rewrites | [PAWN++ in the Wild, WAC 2026](https://aclanthology.org/2026.wac-1.25/) compares original human webpages with AI rewrites of those pages. In its evaluated setting, no detector exceeded 12% true-positive rate on frontier-model rewrites at 1% false-positive rate. PAWN++ itself uses two frozen language models for likelihood signals, so it is not a non-LLM runtime candidate. Its [WATTLE repository](https://github.com/OlegHSE/PAWN-in-the-Wild/tree/main/data/wattle) currently publishes documentation and a SHA-256 manifest but explicitly omits the benchmark CSV data. | Adopt the matched-rewrite test and low-false-positive reporting, not the detector or its resource requirements. The absent CSV files prevent an independent WATTLE replay at present. |
| Four-regime source matching | [Perrone and Romano, ARB 2026](https://arxiv.org/abs/2607.29539) compare human originals, direct generation, AI rewrites of human text, and rewrites of AI text at the same 1% false-positive operating point. Their preprint's data-availability section has a placeholder, but the [public ARB dataset](https://huggingface.co/datasets/giper45/ARB-Dataset) is available separately. The [pinned Unswell transfer screen](../origin/runs/2026-11-08-arb-transfer/README.md) found H2L richness AUROC 0.632 and 4.5% TPR at a threshold calibrated on separate human sources for 1% FPR; its registered feasibility rule failed. Direct generation had AUROC 0.459 in the frozen high-score direction. | Keep generated, rewritten-human, and rewritten-AI labels separate. Source matching and a fixed low-FPR threshold exposed the failure of a lexical feature that looked promising on academic-paper rewrites. The public data enable replay, but neither this result nor the preprint qualifies an editorial finding or a general origin detector. |
| Paired academic prose | The [AIWritingEffect public dataset](https://huggingface.co/datasets/Jerry999/AIWritingEffect) links pre-ChatGPT conference-paper originals to rewrites from seven generators and four prompts, split by paper. A [pinned offline screen](../origin/runs/2026-11-08-aiwritingeffect-screen/README.md) found richness AUROC 0.686 on 2,131 eligible test paragraph pairs, but only 30.4% pair coverage and 6.0% TPR at a validation-human 1% FPR threshold; its registered feasibility rule failed. A later, exposed document-level sensitivity gave AUROC 0.902 on 143 variants from 89 papers, without a reliable 1% document FPR estimate. | The positive document contrast motivates a new independent, document-level low-FPR test, not deployment. Academic PDF extraction and rewrite prompts remain potential confounds, and neither screen supplies an editorial finding for #349. |
| Edit-magnitude classifier | [EditLens (ICLR 2026)](https://arxiv.org/html/2510.03154) uses source/edited pairs to supervise a local discriminative model that scores the edited text alone; the [authors' code](https://github.com/pangramlabs/EditLens) and checkpoint are published. Our [frozen ARB transfer](../origin/runs/2026-11-09-editlens-transfer/RESULT.md) used a SHA-pinned FP32 conversion and obtained H2L AUROC 0.977, 214/299 TPR at 3/300 human false positives; free generation was 298/299 TPR. On [saved technical-document prefixes](../origin/runs/2026-11-09-editlens-technical-prefix/RESULT.md), generation and polish increased the score in mean across all 20 repositories. A [pre-2022 academic-paragraph transfer](../origin/runs/2026-11-10-editlens-academic-transfer/RESULT.md) retained 290 matched pairs from 58 papers: AUROC 0.835 and 262/290 pairwise wins, but only 21/290 rewrites detected at 2/290 human false positives, failing its frozen low-FPR recall gate. A [five-window paper screen](../origin/runs/2026-11-10-editlens-five-window/RESULT.md) then detected 26/58 rewrites at 0/100 separate human-only and 1/58 matched-original false positives; paper-level AUROC was 0.919 on the exposed academic test papers. The [complete-window technical screen](../origin/runs/2026-11-10-editlens-technical-fullwindows/README.md) scored all 6,484 disjoint cleaned-prose windows: 221/228 generation and 186/228 polish responses scored strictly above their own source, with six polish ties. Its registered directional gates passed. The subsequent [historical-control transfer](../origin/runs/2026-11-10-editlens-historical-controls/README.md) applied the fixed academic threshold to 462 pre-August 2021 technical documents from 153 repositories and flagged 20/462 (4.33%), failing its zero-flag gate. | This is an origin ranking signal, not an editorial diagnosis or qualified low-FPR detector. ARB includes XSum and WritingPrompts, also used to train EditLens; exact training-source overlap is unverified. The technical pairs are exposed, their source authorship is unknown, and relative length or formatting may contribute to the score difference. Complete cleaned-prose window coverage is now demonstrated, but it does not establish a field false-positive rate. The historical-control flag rate is not a verified human false-positive rate; the academic threshold did not transfer to technical documentation. Any revised threshold or length policy needs a new predeclared source-disjoint control set with checked human provenance and an independent generator holdout. The 1.42 GB model also needs a runtime and packaging decision. |
| Family-balanced local encoder | [MELD, September 2026 revision](https://arxiv.org/html/2605.06903v2) trains a shared Ettin-1B encoder on family-specific AI-versus-human tasks, including matched rewrites, and pools token evidence. The authors report 99.7% TPR at 1% FPR on their MELD-eval pool of paired human texts and four unseen generators. They explicitly warn that the threshold is calibrated on each evaluation pool's human scores and may not transfer across domains, registers, or lengths. Their [checkpoint](https://huggingface.co/anon-review-meld-2026/meld) and scoring code are available; the original checkpoint requires access agreement, and a [public mirror](https://huggingface.co/odeveloper/meld) identifies itself as an unmodified copy. | The [pinned technical-document screen](../origin/runs/2026-11-11-meld-feasibility/RESULT.md) passed the observed control-alert and generation floors but failed the AI-polish floor, with severe generator-family variation. Retire this checkpoint/configuration for that screen. The published 1% operating point is not a technical-document guarantee; the historical controls lack verified human-authorship labels. The 4.1 GB FP32 weights create a substantial runtime and packaging cost. Token evidence is an origin cue, not proof that a span needs editorial revision. |
| Word-level AI-edit occurrence and intensity | [MixDetect, September 2026](https://arxiv.org/html/2609.32625v1) aligns human sources with AI edits using exact anchors and order-preserving semantic matches, then trains a DeBERTa-v3-base encoder to predict edit occurrence and conditional edit intensity from the edited text alone. On the EditLens test split it reports word-level AUROC 0.885 against alignment-derived labels; its text-level ternary macro-F1 is 0.916. The paper uses overlapping windows for texts beyond 512 tokens. | Separate *where* a change happened from *how much* changed. The word labels are constructed by the authors' alignment, not independent editorial judgments. The paper does not establish low-FPR transfer to technical documentation, and no author code or checkpoint link was identified in the paper on November 11, 2026. Reproducing this architecture would require new training and an independent span audit. |
| Progressive edit-operation benchmark | [OpAI-Bench, June 2026](https://arxiv.org/html/2606.06481v1) and its [repository](https://github.com/VILA-Lab/OpAI-Bench) release nine-version trajectories across polish, paraphrase, style rewrite, compression, and expansion with sentence and span provenance. The published main split has 15,722 human source documents across essays, news, government reports, and scientific abstracts, with a held-out generator. Intermediate mixed-authorship versions can be harder to detect than either endpoint. | Use operation-by-coverage curves and source-group/generator holdouts to stress-test an origin candidate. None of its four domains is technical documentation; its synthetic revision labels do not judge whether a passage is poor prose or whether a suggested edit preserves meaning. |
| Source-matched mixed authorship | The [Heterogeneous AI Spans public dataset](https://huggingface.co/datasets/open-text-detector/heterogeneous-ai-spans) pairs 4,400 synthetic mixed documents with unchanged human controls and exact AI replacement spans. Its source-group splits and layout-neutral view support a [pinned model-free transfer screen](../origin/runs/2026-11-08-heterogeneous-spans-transfer/README.md). Richness on 566 eligible, oracle-located test span pairs gave AUROC 0.652 but only 5.65% TPR at a train-human 1% FPR threshold; complete documents gave AUROC 0.608. Both registered feasibility rules failed. | Exact span provenance makes the corpus useful for testing localized origin features, while its human sources remain provenance candidates and its generated replacements may change facts. The score used the known spans and did not locate them; a deployed detector needs independent source localization and a substantially stronger low-FPR result. |
| Edit-history labels | [AITDNA, 2026](https://arxiv.org/abs/2606.04906) defines alternative notions of AI text from human/AI edit histories. Its [project](https://github.com/UKPLab/arxiv2026-aitdna) exposes document, sentence, span, and other labels. | Use its labeling discipline to keep generated, AI-edited, mixed, and human-authored cases distinct. The project is evaluation infrastructure, not an editorial rule set. |
| Proposition overlap | [Ortame and Navigli, EMNLP 2026](https://aclanthology.org/2026.emnlp-main.2302/) provide [APE](https://github.com/SapienzaNLP/APE), a local discriminative proposition encoder. Its embeddings can compare claims across nearby sentences but do not decide whether either sentence needs editing. | A [pinned 36-page screen](../reviews/2026-11-07-ape-proposition-screen/README.md) tested adjacent-sentence overlap as an editorial candidate generator. The fixed top-three-per-page selection can overlap at most five newly missed FULL events, below its admission gate of eight. No editorial review or prospective claim followed. |
| Example-based span origin evidence | [ExaGPT, Findings of ACL 2026](https://aclanthology.org/2026.findings-acl.380/) compares local span embeddings with human and generated examples, with dynamic segmentation and a matched datastore. | A [frozen AITDNA screen](../origin/runs/2026-11-07-span-example-screen/README.md) tested a simpler author-disjoint, task-matched fixed-window margin with the local APE/BGE encoder. Its AUROC was 0.544 versus 0.697 for lexical richness on the same 352 rows, so the registered feasibility rule failed. This result does not test ExaGPT's full method or establish editorial value. |

## Origin-screen outcome

The [new historical-control run](../origin/runs/2026-11-10-editlens-new-controls/RESULT.md)
passed its registered observed-alert gate at 1/357 controls but detected only
4/228 exposed AI-polished technical responses at its fixed threshold. Its
full-document EditLens mean is retired. The 357 controls and 228 response pairs
are now development data; their authorship is not independently verified.

The [frozen MELD transfer](../origin/runs/2026-11-11-meld-feasibility/RESULT.md)
used a SHA-pinned local checkpoint, 402 eligible historical calibration
documents, 351 eligible historical test documents, and 456 complete saved
responses. Its fixed threshold flagged 1/351 test controls and detected
147/228 generations, but only 31/228 AI-polished responses. The registered
20% polish floor failed, so this checkpoint/configuration is retired for the
declared technical-document screen. Generation detection varied sharply by
saved model family: 67/76 Haiku, 8/76 Luna, and 72/76 Qwen. A verifier
amendment restored one overwritten input-family metadata field from the sealed
manifest without changing any score or threshold. The historical controls
have no independent human-authorship labels, and one test document hit the
model's input cap.

No source-disjoint confirmation or product origin claim follows from this
failed feasibility screen. A new origin method would need newly frozen inputs,
checked human provenance, and held-out generators before a qualification
attempt. Keep #349's editorial-event recall, finding acceptance, and revision
safety as separate gates; a document origin score cannot satisfy them.

## GitHub project audit

* [bheijden/slop](https://github.com/bheijden/slop) is a deterministic prose
  linter with source-local findings. Its [18-pair matched audit](https://github.com/bheijden/slop/blob/main/research/audit.md)
  controls topic and key points, uses human sources published before 2022, and
  reports that two default phrase sets fired more often on human writing
  (1.96 versus 1.01 hits per 1,000 words). In technical documentation the same
  sets fired 2.9 versus 0.3. This is a small project audit, not a general
  population estimate. Its [time-window experiment](https://github.com/bheijden/slop/blob/main/research/k-and-window.md)
  shows that a contrast corpus contaminated with newer machine writing can
  erase the apparent signature. The useful method is matched controls, per-rule
  error accounting, and a fixed reference period.
* [salesforce/creativity_eval](https://github.com/salesforce/creativity_eval/tree/main/Writing_Alignment)
  publishes LAMP: 1,057 generated draft and professionally edited text pairs
  with categorized span edits. It is a useful example of revision-pair labels;
  its literary fiction domain does not qualify technical-document findings.
* [YassirELATTAR/ling-feats-mgt](https://github.com/YassirELATTAR/ling-feats-mgt)
  publishes the feature extraction, cross-domain/model testbeds, and ablations
  behind the AACL study. [UKPLab/arxiv2026-aitdna](https://github.com/UKPLab/arxiv2026-aitdna)
  publishes alternative authorship-label construction and evaluation.
* [google-deepmind/synthid-text](https://github.com/google-deepmind/synthid-text)
  and [facebookresearch/textseal](https://github.com/facebookresearch/textseal)
  provide watermark methods. Verification is a separate, opt-in origin channel
  that needs a participating generation scheme and its detection configuration
  or key. [EntroBench](https://aclanthology.org/2026.findings-acl.2089/)
  and its [benchmark project](https://github.com/py-qin/EntroBench) test watermark
  recovery across entropy levels, text tasks, and user edits. They supply a
  robustness protocol, not a general detector for an arbitrary existing page.

## Proposed offline screen for #349

An [Unswell feature screen](../origin/runs/2026-11-07-richness-screen/README.md)
has now checked lexical richness and entropy on the existing 38-document,
456-response technical corpus. Full generations show the paper's positive
direction even with equal numbers of text windows; AI edits do not show that
direction. This is origin-feature feasibility on exposed data and provides no
new #349 event or accepted finding. The earlier [repository-disjoint candidate
selector](https://github.com/stokaro/unswell/issues/349#issuecomment-6651505353)
and [source-only category retrieval](https://github.com/stokaro/unswell/issues/349#issuecomment-6652866782)
have already failed their registered screens. Their fixed representations
must not be retuned as a substitute for finding new editorial operators.

The [AITDNA user-study screen](../origin/runs/2026-11-07-aitdna-screen/README.md)
then tested the same two features on 362 documents with human-only versus
AI-assisted *study-setting* labels. Its registered feasibility screen passed:
overall directional AUROC was 0.686 for richness and 0.665 for entropy.
Post-result splits show a much weaker signal in AI-assisted documents whose
document notion is `User`. This is useful origin-feature evidence with a clear
boundary, not a qualified detector or editorial event.

1. Keep the existing 123 exposed events, candidate denominators, and reserved
   confirmation pages unchanged. Record new hypotheses and rule versions before
   scoring. Do not use the confirmation pages for feature selection.
2. Build matched controls by documentation role, topic, length, and source
   format. Distinguish human originals, full generations, and AI edits. Hold out
   entire repositories and generator families; add a formatting counterfactual
   that changes line wrapping without changing words.
3. Start with source-local hypotheses: repeated clause and paragraph frames,
   generic introductions without information, redundant restatements, and
   inconsistent transitions. Treat lexical richness and entropy as separate
   document-level origin baselines. Test each feature group's added value and
   failure cases independently. Do not add a trigger solely because it appears
   more often in a generated corpus.
4. For every proposed editorial finding, retain original UTF-8 byte spans,
   protected regions, and exact source evidence. Review candidate edits for
   numbers, versions, identifiers, negation, modality, and attribution.
5. Report full-event recall, accepted-findings fraction, invalid/abstaining
   page coverage, and suggestion safety separately, by mechanism and source
   role. Compare with the unchanged #349 thresholds of 80% full-event recall
   and 85% accepted findings on the required complete cohorts. A feature with
   useful origin discrimination but no localized editorial gain remains in the
   origin research track.

This plan is a hypothesis for an offline development screen, not evidence that
the #349 targets are met. A new frozen protocol must define implementation,
source identities, thresholds, and failure handling before an experiment.

The later [#434 source audit and external-control pass](audits/2026-11-07-external-controls/README.md)
checked the unchanged current engine on 18 constructed cases. It found two
potential editorial gaps but admitted neither as a new operator from the small
controls alone. The original 123-event #349 accounting remains untouched.

The earlier [dual-gate triage](../reviews/2026-11-07-dual-gate-triage/README.md)
showed why recall-only additions cannot qualify the whole-page stream. Its
snapshot had 78 accepted of 122 delivered findings and needed 22 more FULL
events; even one accepted finding per new FULL event would require removing at
least 27 nonaccepted findings to reach 85% acceptance. The later
[category frontier](../reviews/2026-11-08-dual-gate-feasibility/category-frontier.json)
put the exposed whole-page baseline at 28/57 FULL, below the required 46/57;
perfect recovery of the largest single missing category would still reach only
44/57. The next screen must test selective suppression or repair alongside
multiple source-local editorial mechanisms, review the complete changed
stream, and preserve the frozen labels.

The separately frozen [five-feature origin ensemble](../origin/runs/2026-11-08-interpretable-ensemble/README.md)
tested an interpretable, non-LLM proxy for the El Attar et al. feature
approach. It improved ARB human-to-LLM rewrite AUROC from 0.616 to 0.736 and
low-FPR recall from 1.24% to 7.84% on the same source-disjoint test, but missed
its registered 10% recall floor. Its unchanged threshold falsely flagged
120/400 academic human-only papers, so it failed transfer. This is a measured
genre shift, not a qualified origin classifier or an editorial gain for #349.
