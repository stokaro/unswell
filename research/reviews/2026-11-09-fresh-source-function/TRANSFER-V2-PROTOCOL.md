# Page-disjoint transfer for the source-function candidate

This follow-up uses the **same frozen v1 diagnosis core** (SHA-256
`1291ffc4fec191f83d3811a4fa7b5698f1ccf2cd0aed244fa9c8a9bcc4777081`)
with a new document-structure edit planner. The planner withholds any deletion
from a numbered or bulleted line and includes one following ASCII space when
removing a sentence inside a paragraph. It does not assert semantic safety;
every proposed edit still requires source-context review. The previous nine
pages are exposed development data and will not be rescored with v2.

Before opening any new source content, freeze this protocol, the v2 code, and
the deterministic selection script. Search public GitHub code for the first
30 results of each query below, sorted by indexed time descending, and save
the complete JSON responses:

1. `"next generation" filename:README.md size:4000..20000`
2. `"powerful and flexible" filename:README.md size:4000..20000`
3. `"the following table" filename:README.md size:4000..20000`

Select the first four eligible distinct repositories per stratum, excluding
every repository used in the nine-page development packet and the earlier
source-role transfer as encoded in `select_transfer_v2.py`. Do not substitute
for missing defects, inconvenient controls, unavailable content, or mismatch.
Preserve search rank and response SHA-256. Pin the default-branch commit for
each selection, fetch the exact README, and require the Git blob SHA-1 to
match search. Preserve complete original UTF-8 bytes locally and tracked
source hashes and commit paths. Invalid sources remain in the denominator.

The root Codex assistant will review each complete pinned page **before**
running v2. Label all clear opportunities in the six #349 editorial purposes
with exact original-byte spans, concrete reader burden, content to retain,
and safe-edit status. Record at least two close `retain`, `optional`, or
`technical_uncertain` controls per valid page where available. The assistant
is not a human annotator. Seal labels and hashes before the candidate run.
The unchanged main CLI reports on all complete pages provide prepared source
units and a baseline; no second Markdown parser or model service call is used.

Run the v2 code once on **all** selected pages. Review every diagnosis, every
proposed edit, all frozen positives and controls, source binding, invalid-page
coverage, and all unchanged CLI findings. Report location, FULL events,
accepted/delivered, abstentions, protected-region violations, and unsafe or
unresolved edits separately. Require at least five clear positives across
three repositories for a meaningful within-cue recall denominator; otherwise
report an underpowered transfer. A directional success needs at least 80%
FULL recall, at least 85% accepted delivery, no `retain` control hit, and no
unsafe supplied edit on this packet. Passing still cannot qualify #349: the
unchanged 123-event development gates, complete combined stream, and reserved
prospective confirmation remain required. Failing retires this version;
there is no post-result tuning on the new pages.
