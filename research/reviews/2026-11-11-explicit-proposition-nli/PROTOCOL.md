# #349 explicit proposition relation screen

Status: frozen before any inference. This is a bounded, local, non-generative development screen of one evidence-relation component, not a detector, admission policy, or #349 qualification. It uses the 19 available pages from the separately stopped clause-role v1 selection only as exposed development sources. The root Codex assistant prepared and reviewed the 16 source-bound pairs; no human annotator or reserved prospective page is involved.

## New hypothesis

The retired November 10 NLI screen paired an entire target-removed synthetic page with a target sentence. The model saw many irrelevant facts and missed source-function relations. Here, the premise contains only explicitly identified original-source evidence spans. The hypothesis is a manually normalized material proposition or a controlled negation of one original claim. This tests whether a non-generative relation model can distinguish already covered task/scope, missing material content, and explicit contradiction. It does **not** test whether those premises or normalized propositions can be found automatically, whether a redundant claim merits an editorial finding, or whether an edit is safe.

`packet.json` contains six entailments, six neutral pairs, and four contradictions. Every premise fragment is bound to a pinned GitHub commit, original UTF-8 byte span, and SHA-256 recorded in `../2026-11-11-clause-role-transfer/source-inventory.json`. Controlled negations are synthetic hypotheses and are identified as such. The source-only expected labels are frozen in the packet before inference; the runner feeds only premise and hypothesis to the model.

## Model and execution contract

- Model: [`MoritzLaurer/DeBERTa-v3-large-mnli-fever-anli-ling-wanli`](https://huggingface.co/MoritzLaurer/DeBERTa-v3-large-mnli-fever-anli-ling-wanli), revision `b3546ea6b0346eb6f8d5d68b13c7dc6d0376b3d7`, quantized ONNX CPU inference. The model card reports training on MultiNLI, FEVER-NLI, ANLI, LingNLI, and WANLI; these are external benchmark claims, not Unswell results.
- Assets: `config.json` SHA-256 `7f9d420b616691c5b575bab8839719aefbddc656d8e1316519414b097ce43e3d`; `tokenizer.json` SHA-256 `7aa118770f066a74530d161c7d0b994d0629cc0ff3a0df213f184192773f960a`; `onnx/model_quantized.onnx` SHA-256 `2eb57451c9d94a868c3468182177ce75f20ca80cfa41f851af1d7e7e4dc57228`, 642,587,601 bytes. Reject any mismatch.
- Container: `docker.io/library/python:3.12-slim@sha256:5e2bdb1b0d542d5039cffb8181c65930093949194a65c365dfb1b0fe13a4fa5b` on explicit `remote-dev-container`. Pin `numpy==2.5.3`, `onnxruntime==1.30.0`, and `tokenizers==0.23.2`. CPU only, two intra-op threads.
- Disable tokenizer JSON truncation explicitly. Reject a pair above 512 token IDs. Check ONNX input names, three finite logits, and config label order (`entailment`, `neutral`, `contradiction`). Use argmax once; no threshold, fit, retry, label repair, or alternate checkpoint after seeing results.
- One local inference process, at most 16 sequential pairs, 30-minute wall limit, stop at first error. No hosted model or paid inference call. The user authorized autonomous continuation; this fixed local run does not authorize a later model run. Remove the task container, image, model cache, and Python packages from the shared remote daemon after recording the result.

## Directional gate and limits

Advance this relation representation only if at least 14/16 labels are correct, all four contradictions are correct, at most one neutral pair is falsely called entailment, and all source bindings and model identities pass. Otherwise retire this exact model/representation without changing a cutoff on the exposed cases. Report errors by relation and source; a technically successful run with a failed gate is a negative research result.

A pass would permit a separate fresh page-disjoint end-to-end candidate experiment. It would not add a FULL event to #349 by itself. The actual combined detector must still satisfy the unchanged complete-cohort 80% FULL-event recall, 85% accepted-delivery, complete-page availability, and advice-safety requirements before reserved prospective confirmation. Preserve every earlier negative result and denominator.
