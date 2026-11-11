# #349 explicit proposition relation screen: negative

The frozen v1 run completed once on `remote-dev-container` with the pinned quantized DeBERTa model and 16 source-bound pairs. The runner verified the model, tokenizer, config, packet, and code SHA-256 values before inference. Raw output is in `result.json` (SHA-256 `676465b44a849a1695dcc5ce681b2fbbe3ddaddf0ab78593dd3eb8c2ffdab4b4`); `score.json` is produced by the frozen scorer. The task container, downloaded model, packages, and task-created image were removed after recording the result.

| Frozen gate | Result |
| --- | ---: |
| Correct relation labels, at least 14/16 | 15/16 |
| Correct contradictions, all 4/4 | **3/4** |
| Neutral pairs falsely called entailment, at most 1 | 0/6 |
| Combined directional gate | **Fail** |

The sole mismatch is C03. The original ChatDoctor source says, “please do not apply it to real clinical scenarios”; the controlled hypothesis is “The ChatDoctor model is recommended for clinical use.” The frozen expected relation is contradiction, while the model returned neutral (logits in entailment/neutral/contradiction order: `-1.452`, `1.866`, `-0.515`). The full source also says clinical use is prohibited, but that sentence was not in this pair's premise; it cannot be added after seeing the result. The fixed representation/model is retired under the prespecified gate. No cutoff or label was changed.

This is a narrow component test with manually selected evidence spans and normalized hypotheses. It provides no automatic candidate discovery, editorial admission, edit safety, or FULL-event result. It does not change the frozen #349 denominator or current-main recall. The next useful experiment must establish an automatic, source-bound way to locate evidence and form material propositions on page-disjoint sources before relation classification can contribute to product recall. Do not promote this result as detector improvement.
