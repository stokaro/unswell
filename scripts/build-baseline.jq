# Reduce a corpus frequencies run to one construction baseline artifact: the
# rates of one measure by role, and the identity of the analyzer that counted
# them. Rates are written to five decimal places, finer than any comparison
# the proposal makes and stable enough to diff.
def r5: (. * 100000 | round) / 100000;
{
  format: "unswell-construction-baseline-v1",
  measure: $measure,
  min_count: .min_count,
  human_corpus: .human_corpus,
  source: $source,
  nlp: .nlp,
  baselines: [
    .baselines[] | select(.measure == $measure) |
    {measure, role, sentences, words,
     terms: [.terms[] | {key, count, per_thousand_words: (.per_thousand_words | r5)}]}
  ]
}
