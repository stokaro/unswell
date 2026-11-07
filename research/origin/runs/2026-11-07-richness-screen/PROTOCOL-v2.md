# Separate equal-window sensitivity screen

Date: November 7, 2026. The first frozen screen passed its directional
criterion, but generated responses had median prose-token length ratio 0.745
and only 103/228 generated pairs met the registered 0.75–1.33 length band.
This post-result sensitivity screen is separate from v1 and cannot change its
result or act as an unused confirmation sample. The same exposed archives,
source identities, request counts, and operation labels remain fixed.

For every source/response pair, remove code, inline code, URLs, and HTML tags
and tokenize exactly as in `run.py`. Form all 50-token, stride-25 windows.
Let `n` be the smaller number of windows in the two texts. Use exactly `n`
windows from each text in two fixed ways: (a) the first `n` windows; (b) `n`
windows evenly spaced from first to last, choosing index
`round(i*(m-1)/(n-1))` for `i=0..n-1` from a text with `m` windows; for
`n=1`, choose index zero. Compute the same lexical-richness and normalized
entropy means over these matched windows. Keep every request; if either side
has zero windows, record abstention and retain it in the 456-request total.

For each operation and matching method, report paired median deltas, counts
of positive/negative/zero repository-mean deltas across all 20 repositories,
and median deltas per model family. A robust directional result would require
at least 15/20 repositories positive in **both** features under **both**
matching methods, with positive family medians for both features under both
methods. Failure rejects a two-feature origin baseline on this corpus;
success only warrants a future classifier experiment with new sources.
Neither outcome changes #349 editorial recall or qualifies origin detection.
