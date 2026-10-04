The following object uses `unswell-source-table-v2` with kind `public`.
It contains the entire unchanged source and all fallible machine hypotheses.
It contains no review grades or reference events.

Each `units` row is `[unit_id, start_byte, end_byte, words]`. The first three
entries are the original unit index. `words` is the original eligible anchor
index for that unit, in order. Each word is
`[one_based_word_ordinal, absolute_start_byte, text]`.
Its existing ordinal is explicit, so you do not need to count array positions.
Its end byte is its start plus the ASCII text length. For example, ordinal 2
of `u0000` is anchor `u0000:w0002`.
Use these existing ordinals as `start_word` and `end_word` in the unchanged
response schema. Words inside protected source regions are absent from the
anchor index; the complete original source still includes those regions.
No source, unit, hypothesis, or eligible word was filtered.
