# Operational-error handling after sealed labels

The transfer labels were sealed in commit `51b2907` before any v2 candidate
output. The unchanged main CLI then returned an incomplete report for selected
page 3 (`open-source-flash/open-source-flash`): its Markdown grammar returned an
incomplete or invalid syntax tree. The source bytes and Git blob were valid.
This is an operational failure, not an absence of editorial findings. Keep
page 3 in the 12-page denominator and mark its three retained controls as
unassessed by the candidate. Do not substitute another page.

The frozen v2 driver assumed every valid UTF-8 source would have a complete
CLI report. Before running or viewing any v2 candidate output, amend only the
driver's report handling: an incomplete report yields zero diagnoses and a
recorded operational error. The diagnosis core and edit planner remain
unchanged. This amendment follows observation of the baseline error, so the
result is exploratory, not a clean preregistered transfer pass. Any candidate
metrics on the other 11 pages cannot pass the registered 12-page gate because
operational errors cannot pass a gate. The next version needs a new page-
disjoint packet after the parser failure is addressed.
