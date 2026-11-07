# Lexical richness and entropy on complete technical documents

This **offline, exposed-data feasibility screen** measured two interpretable
features motivated by [Shan et al.](https://aclanthology.org/2026.emnlp-main.1725/)
and [El Attar et al.](https://aclanthology.org/2026.findings-aacl.10/).
It used the saved [long-prose-v3](../../../generation/studies/long-prose-v3/README.md)
inputs and outputs: 38 original technical documents from 20 repositories and
456 responses, split evenly between full generation and editing. No new model
call or downloaded data was used. The originals' historical authorship is
`unknown`; they are a paired source baseline, not verified human negatives.

The [v1 protocol](PROTOCOL.md), [script](run.py), and input archive hashes were
sealed in [freeze.json](freeze.json) before scoring. The run verified the two
archives and all 38 source hashes. All 456 requests produced feature scores;
there were no abstentions. The complete [v1 result](result.json) has SHA-256
`c7084e1fad0664f7b3a33dcb4e11afc011fb83b9238e09b8d67e6ba2ea7308e4`.

| Operation versus its own original | Pairs | Median lexical-richness delta | Median entropy delta | Positive repositories: richness / entropy |
| --- | ---: | ---: | ---: | ---: |
| Full generation, all lengths | 228 | +0.0323 | +0.0134 | 17/20 / 17/20 |
| AI editing, all lengths | 228 | -0.0007 | -0.0004 | 7/20 / 8/20 |

All three generator families had positive median deltas in both features.
The v1 positive-direction screen therefore passed its registered criterion.
This is **not detector accuracy**: no classifier, decision threshold, false-
positive rate, or authorship probability was computed.

## Length and formatting sensitivity

Generated responses had a median prose-token length ratio of **0.745** relative
to their source, and only 103/228 were inside the registered 0.75–1.33 band.
The median proportion of words retained after code/URL/HTML removal was 0.072
higher for generated responses, a content-format difference that can influence
the feature comparison. In the length-band subset, lexical richness was
positive in 15/18 represented repositories and entropy in 14/18. This subset
omits two repositories and does not replace the all-response result.

After seeing the length imbalance, a **separate**
[equal-window protocol](PROTOCOL-v2.md) and [script](equal_length.py) were sealed
in [freeze-v2.json](freeze-v2.json). It compares equal numbers of 50-token
windows, either from the beginning or evenly across each source and response.
The complete [v2 result](equal-length-result.json) has SHA-256
`e03374a48c64e6f16dcbff423d2baddf3c1d8e9d3e35d95abcee089eafdcb21b`.
Both methods retain all 456 responses with zero abstentions.

| Full generation, equal windows | Median lexical-richness delta | Median entropy delta | Positive repositories: richness / entropy |
| --- | ---: | ---: | ---: |
| Evenly spaced | +0.0300 | +0.0133 | 17/20 / 17/20 |
| Prefix | +0.0311 | +0.0129 | 19/20 / 18/20 |

All three family medians stayed positive for both features under both methods,
so the separately registered equal-window direction screen also passed. AI
editing remained near zero or slightly negative in both features. The v2 test
is a sensitivity analysis on the **same exposed texts**, not independent
confirmation. Equal window counts do not remove differences in omitted
sections, facts, headings, or code density.

## Interpretation and next evidence

These results justify testing lexical richness and word entropy as **origin
features** on a new, source-matched technical cohort with independently known
provenance, balanced document length and format, and repository- and generator-
held-out evaluation. They do not qualify a detector trained on these two
features. Mixed or edited text needs its own label and evaluation; on this
corpus the edit arm does not share the full-generation direction.

Neither feature identifies a specific editorial defect or changes issue
[#349](https://github.com/stokaro/unswell/issues/349), its frozen 123 events,
or its 80% recall and 85% accepted-finding gates. Do not combine an origin
score with the editorial gate on the strength of this screen.

Reproduce from the repository root, without model access:

```sh
python3 -B research/origin/runs/2026-11-07-richness-screen/run.py
python3 -B research/origin/runs/2026-11-07-richness-screen/equal_length.py
```
