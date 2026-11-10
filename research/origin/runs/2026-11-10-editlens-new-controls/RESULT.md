# New historical controls: low-alert pass, AI-edit detection gap

November 10, 2026. The source frame and score procedure were locally committed
as `706e5fb` **before** the first EditLens inference. The runner then scored
all 6,075 windows exactly once on the local CPU. It saved the calibration
threshold before test inference. No hosted model call, new AI text generation,
product detector, editorial rule, or frozen #349 label changed.

## Frozen control decision

| Measure | Result |
| --- | ---: |
| Calibration documents / windows | 404 / 3,127 |
| Strict `>` threshold, maximum calibration document mean | `0.47566521167755127` |
| Test-frame candidates | 403 |
| Eligible test documents / windows | 357 / 2,948 |
| Test coverage | **357/403 = 88.6%** |
| Historical test documents flagged | **1/357 = 0.28%** |
| Test repositories with a flag | 1/162 |
| Registered control gate, ≥70% coverage and ≤1% observed flags | **Passed** |

The one flagged page is `simonask/w`'s 2014
`docs/wayward/support/overview.md`, a C++ Support API index (score
`0.5312542683548398`, three windows). It remains in the denominator.
The 95% Wilson interval for the **repository-level any-flag fraction** is
0.11%–3.41%; it is descriptive and does not establish a human false-positive
rate. By project language, the flags were 0/86 C#, 1/183 C++, 0/74 Java, and
0/14 Python. By page length: 1/169 with one to four windows, 0/113 with five
to nine, and 0/75 with at least ten. These subgroup denominators are small,
especially Python. The pages predate August 2021 but individual authorship
and possible model training overlap remain unverified.

## Exposed positive replay: negative for the current configuration

The *unchanged* new threshold was applied without inference to saved
full-document scores from the earlier, already exposed 228 generated and 228
AI-polished technical responses. It flagged **94/228 generation (41.2%)** and
**4/228 polish (1.8%)**. Both fall below the planned 50% and 20% feasibility
floors; these exposed pages cannot qualify a new detector, but their failure
is enough to avoid spending a new generation budget on this scalar-mean
configuration.

The previously published academic threshold `0.27447814408224075` would
have flagged 32/357 (9.0%) of the new historical test controls. A *post hoc,
exploratory* threshold envelope on these observed controls and old AI
responses shows the tradeoff: allowing up to three control flags out of 357
permits at most 8/228 polished responses at the best strict threshold, while
reaching 46/228 polished responses would flag at least 38/357 controls.
This envelope was calculated after viewing the test scores. It is a diagnostic
of these two cohorts, **not** a registered gate or a reusable threshold.

**Decision:** retire this full-document EditLens mean and scalar threshold as
the AI-edit detector candidate. The control gate passed, but useful AI-edit
recall did not. The next candidate must change the information used for
AI-edit detection and be tested on a new source-disjoint control set; this
test set is now exposed development data. No product qualification or issue
[#349](https://github.com/stokaro/unswell/issues/349) completion is claimed.

## Verification and correction

`verify_score.py` independently recomputed every document mean, the maximum
calibration threshold, all flags, subgroup counts, and the gate from saved
window scores. It passed on the copied tracked artifacts. Replacing one saved
window score in a temporary copy made it reject the document mean. The
score-blind frame verifier independently checked all 860 candidate decisions
and all 852 available exact source hashes; its source-hash mutation test also
failed as expected.
The calibration file SHA-256 is
`cdafb6b8aee2a72a31dfb91a2771426dced1327532293b91ff7e7e1c1ed2cfa2`;
the result JSON SHA-256 is
`e000a5bd4e41a0e8447921f3a5bf54cb7615a2d13a33ba55e13d0ba69795f49a`.
The [exploratory replay](exposed_replay.json) pins its older and newer inputs.

The frozen [source-frame note](SOURCE_FRAME.md) says 323 represented
repositories split 160/163. Those were preliminary counts *before* the 35
score-blind source-review exclusions. The final selected frame contains
**319** represented repositories, 157 calibration and 162 test. The frozen
note is retained unchanged; this is its correction.
