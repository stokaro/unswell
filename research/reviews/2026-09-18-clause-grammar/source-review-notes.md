# Source review notes

The original protocol and selection were frozen against draft PR #325. That
stack has now merged. The runtime baseline for this comparison is main commit
`ad06b689e0da23e6f9e31746553a3d041d02585a`. Recompute all baseline reports from
that revision; do not substitute reports produced by an earlier branch. This
amendment precedes runtime changes and confirmation diagnostics.

The historical long-text selection is libevent's CMakeLists.txt. Its metadata
classifies it as text; the extractor exposes code along with comments and help
strings. The review covers its natural-language content, retaining code as
context rather than labeling commands as bad prose. Keep the selected source;
this set has no ordinary historical long-prose page. The recorded 3,031-word
size includes code, so it does not establish a 3,031-word prose stratum.

The source-only overlap check found no complete normalized block of at least
12 whitespace tokens in the 142 prior source identities. This is not semantic
independence. During scratch preparation, c04's raw text path collided with its
rendered reading packet. Restore raw bytes from the frozen archive in a separate
raw directory and verify every source hash before annotating or checking overlap.
The archive and input freeze did not change.

The implementing Codex assistant reviewed all seven categories before seeing
confirmation diagnostics. These are maintainer-accepted development labels, not
independent human annotation or population estimates. Ambiguous improvements
remain uncertain; technical requirements and plausible novice guidance remain
controls. No origin label supplies an editorial label.
