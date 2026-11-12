# ADR 0052: Claims follow what was measured against people

Status: proposed on November 12, 2026. It takes effect when the maintainer
merges it. [ADR 0053](0053-trim-admitted-by-model-consensus.md) replaces
decision 5, narrows decision 6, and replaces the sentence about the README.

## Context

ADRs 0046 to 0051 qualify rules, defaults, and the review command with
assistant annotators and assistant reviewers. Every acceptance rate in them is
agreement between models. No person judged a finding, and
[ADR 0051](0051-review-command.md) says so about its own numbers.

The [human-reference record](../../research/reviews/2026-11-12-human-reference/README.md)
measures the same tool against what people did: pages of documentation that
maintainers reworded, model answers that experts edited, and model-written
fiction that professional writers edited and ranked. Its plan and decision
rules were registered on issue 525 before the results were read. Six results
matter here.

1. Three professional writers who edited the same paragraphs agree on which
   words to change with a Cohen's kappa of 0.11. One writer's edited sentences
   predict another's by 7.1 points over chance.
2. The same writers agree on which of two versions is better in four pairs
   out of five, and they prefer model-edited text to raw model text in 71.6%
   of judgments.
3. Three models tried as judges of writing agree with the writers' majority
   about half the time on "writer-edited or raw model text", and prefer the
   raw model text in about two thirds of the triplets. None met the
   registered floor.
4. On documentation that maintainers reworded, the default rules flag 4.0% of
   the reworded sentences and 4.2% of the others. With every rule enabled the
   rates are 18.8% and 23.0%.
5. A first blind comparison of paragraphs before and after revision had one
   reader. The maintainer later withdrew those answers as a measure, and
   ADR 0053 decides the question by a consensus of models.
6. `unswell review` flags reworded sentences more often than the others in all
   three genres. On the documentation test the margin is 18.8% against 11.6% with one reviewer model and 19.3% against 14.6% with the other.
   Confirmation by a second model shortens the list and does not move it closer to what people edited in any consistent way. The margin, in points, changes from +7.2 to +7.6 and from +4.7 to +5.8 on documentation, from +17.3 to +14.8 and from +14.1 to +13.2 on expert-edited answers, and from +7.4 to +3.3 and from -0.1 to +0.3 on fiction.

## Decision

1. **No claim that the tool finds what an editor would change.** Editors do
   not find what another editor would change. The documentation reports, for
   the rules and for the review, how their findings fall on sentences that
   people edited, with the agreement between editors next to it.

2. **The rules are a narrow check.** They stay: they are fast, reproducible,
   offline, and can gate a build on wording a project has decided to refuse.
   The README no longer presents them as a way to find wording that needs
   revision in general. Their measured relation to human edits is stated
   where they are introduced.

3. **Review findings are suggestions.** They are described as a model's
   suggestions that fall on edited sentences somewhat more often than on
   others. Confirmation by a second model stays available as a way to get a
   shorter list. It is no longer described as raising quality.

4. **No model judges writing quality for this project.** An acceptance rate
   from assistant reviewers is called agreement between models wherever it
   appears. It keeps its role in ADRs 0047 to 0050, which is to keep rule
   defaults consistent with a stated rubric. It is not evidence that a reader
   is helped.

5. **Replaced by ADR 0053.** This decision said that the tool reports findings
   and does not rewrite. It rested on answers that were withdrawn as a
   measure.

6. **A method that claims more needs a human reference.** A new rule family,
   prompt, or model-based method is admitted on a measurement against one of
   the human references of the record, on its test part, with the rule
   written before the result is read. ADR 0053 admits a revision option on a
   consensus of models; this decision still governs any claim that findings
   match what people edit.

## Consequences

The review documentation reports the human-referenced numbers before the
model-referenced ones. ADR 0053 keeps the limits there and out of the README.

The qualification record keeps its numbers and gains a sentence that they are
agreement between models and that the judge validation failed.

The work that the record points to, in order: a human reference of
documentation written by models and edited by people, which no public set
provides; the open-source linters that were registered as a comparison and not
run.
