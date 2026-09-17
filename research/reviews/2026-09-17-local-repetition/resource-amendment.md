# Resource accounting amendment

The first replay used implementation commit
`299cf61f02216993f761db6a9d70f68f7b84b9d3`. The repeated-claim rule exhausted
the default candidate budget on the already exposed 93,341-byte Ptah schema
commands page. The implementation charged a source-map traversal before checking
whether a protected operand was eligible. An invalid code operand returned
before that traversal ran, but still consumed its budget.

Move that charge immediately before the traversal. Keep the same candidate
limit, matcher, source checks and policy. Add a public-engine regression with
unsupported command operands followed by two eligible assertions. This change
corrects work accounting; it does not tune matching from confirmation output.

Confirmation outputs had been opened before this correction. Preserve the
initial reports and compare every finding, assessment and gate with the corrected
replay. Report any difference explicitly. Do not amend the frozen labels, selected
sources or protocol. Any later matcher tuning requires new confirmation pages.
