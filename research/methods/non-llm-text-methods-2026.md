# Non-LLM text detection and revision methods: November 2026 review

Status: source review and linked exposed-data feature screens, dated November 7,
2026. The APE follow-ups produced negative source-local and span-example
screens; no #349 qualification is claimed.
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
| Matched web rewrites | [PAWN++ in the Wild, WAC 2026](https://aclanthology.org/2026.wac-1.25/) compares original human webpages with AI rewrites of those pages. In its evaluated setting, no detector exceeded 12% true-positive rate on frontier-model rewrites at 1% false-positive rate. PAWN++ itself uses two frozen language models for likelihood signals, so it is not a non-LLM runtime candidate. | Adopt the matched-rewrite test and low-false-positive reporting, not the detector or its resource requirements. |
| Edit-history labels | [AITDNA, 2026](https://arxiv.org/abs/2606.04906) defines alternative notions of AI text from human/AI edit histories. Its [project](https://github.com/UKPLab/arxiv2026-aitdna) exposes document, sentence, span, and other labels. | Use its labeling discipline to keep generated, AI-edited, mixed, and human-authored cases distinct. The project is evaluation infrastructure, not an editorial rule set. |
| Proposition overlap | [Ortame and Navigli, EMNLP 2026](https://aclanthology.org/2026.emnlp-main.2302/) provide [APE](https://github.com/SapienzaNLP/APE), a local discriminative proposition encoder. Its embeddings can compare claims across nearby sentences but do not decide whether either sentence needs editing. | A [pinned 36-page screen](../reviews/2026-11-07-ape-proposition-screen/README.md) tested adjacent-sentence overlap as an editorial candidate generator. The fixed top-three-per-page selection can overlap at most five newly missed FULL events, below its admission gate of eight. No editorial review or prospective claim followed. |
| Example-based span origin evidence | [ExaGPT, Findings of ACL 2026](https://aclanthology.org/2026.findings-acl.380/) compares local span embeddings with human and generated examples, with dynamic segmentation and a matched datastore. | A [frozen AITDNA screen](../origin/runs/2026-11-07-span-example-screen/README.md) tested a simpler author-disjoint, task-matched fixed-window margin with the local APE/BGE encoder. Its AUROC was 0.544 versus 0.697 for lexical richness on the same 352 rows, so the registered feasibility rule failed. This result does not test ExaGPT's full method or establish editorial value. |

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

The subsequent [dual-gate triage](../reviews/2026-11-07-dual-gate-triage/README.md)
shows that a recall-only candidate cannot plausibly qualify the whole-page
stream: the technical profile has 78 accepted of 122 delivered findings, while
22 more FULL events are needed. With one accepted new finding per newly FULL
event and no losses, at least 27 existing nonaccepted findings must also be
removed to reach the 85% acceptance gate. A metadiscourse-only operator cannot
fill any cohort's recall gap. The next screen must test selective suppression or
repair of noisy existing findings alongside multiple source-local editorial
mechanisms; it must still review the complete changed stream and preserve the
frozen labels.
