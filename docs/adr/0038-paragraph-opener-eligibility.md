# Paragraph-opener eligibility

Status: accepted for implementation in #239.

## Problem

The diagnostic replay in #217 found three Ptah paragraphs starting with
`Measured on PostgreSQL`. Correctly splitting the first sentence after `18.6.`
made one paragraph ineligible because the rule checked only that sentence's
length. The full paragraph still exceeded the configured minimum.

## Decision

`repetition.paragraph-openers` version `2` applies `min_words` to the complete
paragraph's existing prose-word count. It selects the first `opener_words` words
from the first sentence only. It does not concatenate sentences to build a key
or skip to a later sentence when the first lacks enough words.

`repetition.sentence-openers` keeps its sentence minimum and version `1`.
Both rules retain paragraph-only scope and the provider's `Word` flags.
Protected code cannot increase the word count or supply opening words. The
existing treatment of eligible words beside a protected token is unchanged.

No sentences yield `no_sentences`. A paragraph below its word minimum yields
`insufficient_words`. An eligible-length paragraph whose first sentence lacks
enough opening words yields `no_eligible_tokens`. An accepted key below the
group occurrence threshold yields a measured zero. These observations come
from the same traversal that creates grouped diagnostics.

The sentence split, grouping threshold, evidence locations, activation formula,
weights, and default gates are unchanged. Repeated openings indicate a shared
form; they do not establish redundant meaning or justify removing technical
conditions, negation, identifiers, or different numeric limits.

## Identity and validation

Version `2` changes the rule catalog, finding and baseline identities, and each
activation source's `ruleset_hash`. Configuration bytes and their compiled hash
remain unchanged. The activation formula stays at version `1` because it still
measures the maximum raw activation in each block. Consumers of activation data
must check both feature and rule-catalog identities before reusing measurements.

Blackbox engine tests and root CLI goldens pair short and long opening
sentences, short paragraphs, protected prefixes, and missing first-sentence
keys. They preserve dotted-version boundaries, technical facts, original
locations, BOM/CRLF, JSON/SARIF, and applicability values. The pinned Ptah
reproducer is replayed separately; historical review records remain unchanged.
