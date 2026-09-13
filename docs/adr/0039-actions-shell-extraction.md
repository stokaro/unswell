# Shell prose in GitHub Actions workflows

Status: accepted for implementation in #238.

## Problem

The #217 replay found `syntax.parenthetical-load` on shell substitutions and
subshells inside a YAML `run` scalar. The YAML decoder correctly preserved the
scalar, but the whole program became one prose block. A rule-level exception
would also hide real comments and messages in that block.

## Decision

Add `extraction.github_actions` with modes `shell` (default) and `strings`.
Library extraction also treats the zero value as `shell`. The selector applies
only to direct `.yml` and `.yaml` children of `.github/workflows`, including
absolute source names, and only to scalar strings at `jobs.*.steps[*].run`.
It does not select action inputs named `run`, composite actions, or arbitrary
YAML files. It does not validate the complete Actions workflow schema.

Use the existing YAML tree and scalar decoder to find the value. Determine the
shell from the step, job defaults, or workflow defaults, in that order. Without
an explicit shell, a declared container selects sh. Known literal hosted-runner
labels select Bash on Linux/macOS and PowerShell on Windows. Dynamic runner
values and unrecognized labels remain unknown; they are not evaluated locally.
Recognized label shapes are `ubuntu-latest` or `ubuntu-NN.NN` with optional
`-arm`; `macos-latest` or `macos-N` with optional `-large`, `-xlarge`, or `-intel`;
and `windows-latest` or `windows-NNNN` with optional `-arm`. This classification
does not assert that a particular runner image exists or is available.

The accepted shell names are `bash`, `sh`, `zsh`, `fish`, `pwsh`, and `powershell`.
A custom template must start with one of those names, contain one standalone
`{0}`, and otherwise contain simple literal option tokens. Executable paths, wrappers,
quoted command templates, cmd, and Python remain unsupported by this selector.
Existing grammar limitations still apply, including Zsh's shared Bash grammar.

Decode folding, chomping, quotes, and YAML escapes before calling the existing
shell extractor. Compose each resulting source map back to the original YAML
bytes. Comments and string literals retain independent blocks; substitutions,
operators, and other program syntax cannot contribute prose words. Bare command
words retain the shell extractor's existing treatment as syntax. Strings used by
`echo`, `printf`, and PowerShell output commands remain eligible. No program runs.

The outer YAML string context and its explicit exceptions apply first. The
nested shell format then applies its own context set and matching exceptions,
using the original workflow path. Returned block kinds remain `string`, so the
outer filter does not reinterpret a shell comment as a YAML comment. When
structural context is requested, labels include `embedded:<format>:<kind>` plus
the inner and outer grammar owners. Shell comment suppressions map back to the
same YAML source and target the selected prose blocks.

## Coverage and failures

Unknown shells produce `actions-shell-unknown`. Unresolved `${{ ... }}` in a
script produces `actions-expression` for the entire scalar: substitution could
change the program's syntax, including its comments and quoting. These explicit
exclusions do not assert that embedded analysis completed. Other step names,
ordinary values, and independent scripts remain checked. The usual empty-scan
gate applies; a file may pass if it still has eligible prose outside an excluded
program. Inspect exclusions when complete script coverage is required.

Known scripts that fail grammar validation, decoded source validation, or resource
limits cause an operational error. They never fall back to YAML prose. Explicit
YAML exceptions and disabled string contexts skip nested decoding and parsing;
the outer YAML grammar still must be valid. The `strings` mode is an explicit
policy choice to use ordinary scalar analysis for all run values.

Record omitted program ranges as `actions-shell-syntax`; retain the existing
reasons for nested context exclusions, interpolations, directives, and embedded
program literals. Exclusions must not overlap selected prose through a blanket
whole-script span. Total block limits, cancellation, and original source bytes
remain under the existing extraction contract.

## Identity and validation

Changing this mode changes the config and extraction policy hashes. Prepared
model inputs also bind the mode. Check these hashes before reusing measurements
or model bindings. Keep historical records intact. Prepared units also bind the
grammar context, block boundaries, target text, and source segments. No NLP
provider, rule version, scoring formula, or model target changes here.

Blackbox and root CLI tests reproduce the reported shell warning. They check
retained prose, shell choice, quoted and folded values, escapes, and original byte
ranges. Cases include Unicode, BOM/CRLF, mapped suppressions, explicit exceptions,
unknown shells, and invalid programs.
The CLI verifies JSON/SARIF locations and unchanged inputs. Policy and prepared
identity tests verify that opting into scalar analysis changes the bindings.

References: [shell selection and custom templates](https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-syntax),
[default precedence](https://docs.github.com/en/actions/how-tos/write-workflows/choose-what-workflows-do/set-default-values-for-jobs),
and [the container default shell](https://docs.github.com/en/actions/how-tos/write-workflows/choose-where-workflows-run/run-jobs-in-a-container).
