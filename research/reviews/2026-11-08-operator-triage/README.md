# #349 operator triage: an exploratory near miss

This offline audit asks whether coarse, source-visible proposal categories and
operators could improve the rejected [separated-admission
bank](../2026-10-07-separated-admission/README.md). It uses the same 36 exposed
development pages, 123 original events, 2,847 previously admitted proposals,
their archived assistant grades, and the five unresolved supplied edits. It
does not run a model, change the runtime, or touch reserved prospective pages.

The inputs are five SHA-256-pinned local artifacts from the archived six-stream
bank. The private source and review files are not republished. `audit.py`
requires all five exact hashes before replaying the calculation; `result.json`
contains only aggregate counts and the learned category/operator exclusions.
The public `verify_result.py` checks identities, denominators, arithmetic,
fold accounting, and negative outcome claims. It cannot independently verify
the private editorial grades.

## Measurements

| Whole-page variant | FULL events | Useful actual criticisms | Unresolved supplied edits |
| --- | ---: | ---: | ---: |
| Archived six-stream admission | 48/57 | 1,627/2,101 (77.4%) | 5 |
| Exact span + category + operator deduplication | 48/57 | 896/1,157 (77.4%) | 5 |
| Four exclusions selected with all exposed grades | 46/57 | 1,310/1,533 (85.5%) | 5 |
| Leave-one-repository-out learned exclusions | **45/57** | **1,443/1,700 (84.9%)** | **5** |

The duplicate grouping roughly halves the number of delivered criticisms
without improving their useful fraction. The four-exclusion row is a
retrospective, grade-selected construction. Its apparent numerical pass is not
held-out evidence and it still has unresolved edits. For the transfer row, each
of the 13 repository folds chooses at most eight category/operator exclusions
using only the other repositories' grades, keeping training recall at or above
80% and the smaller cohorts' useful fraction at or above 85%. Predictions are
then applied to the excluded repository without updating the policy. Across
the held-out predictions, whole-page recall is 45/57 (78.9%) and usefulness is
1,443/1,700 (84.9%): **both miss the unchanged 80% / 85% gates**. The context
cohort yields 38/46 FULL and 389/413 useful; the previously used confirmation
cohort yields 17/20 FULL and 190/198 useful. Those are still exposed
development measurements, not new prospective confirmation.

The five unresolved actual suggestions remain in every measured variant. A
diagnostic-only policy would remove supplied-edit risk by construction, but it
was not executed or editorially reviewed here and would not repair the two
held-out numeric failures. The cached six-stream proposal bank also does not
qualify standalone generation.

This near miss supports testing a source-specific semantic discriminator on
new pages. Further threshold or category sweeps over these exposed grades
would not establish transfer. #349 remains open; the original event labels,
denominators, prospective boundary, and product defaults are unchanged.

For an authorized holder of the archived private inputs, reproduce the
aggregate with:

```sh
python3 -B audit.py <development-root> result.json
```

The development root must contain `recovered-admission-bank-v1`,
`full-separated-judgment-source-necessity-astra-medium-v1-evaluation`, and
`joint-component-sol-high-evaluation` with the pinned contents. Public
accounting and negative controls run with:

```sh
python3 -B verify_result.py
python3 -B -m unittest discover -s . -p 'test_verify_result.py'
```
