# #349 current-engine event reachability at the 80% target

This frozen offline audit checked all 123 exposed editorial events against
source-mapped units exported by the current technical engine. It is an
availability test, not a new detector or a quality estimate. No model call,
label change, or runtime change occurred.

The exporter from main `52b6612` processed all 36 pinned pages and produced
7,625 eligible units. Every frozen event has at least one alphanumeric target
byte inside an eligible source segment. Thus none of these events is completely
inaccessible through the current unit contract. Full target-byte availability
clears the 80% FULL count in each cohort, but by only one event in the
whole-page cohort. This does **not** establish that a source-local rule can
recognize the editorial defect or meet the actual FULL gate.

| Cohort | Available events | Every target overlaps eligible prose | Every target's content fully eligible | Required FULL recall |
| --- | ---: | ---: | ---: | ---: |
| Whole page | 57/57 | 57/57 | 47/57 | 46/57 |
| Contextual | 46/46 | 46/46 | 44/46 | 37/46 |
| Historical confirmation | 20/20 | 20/20 | 19/20 | 16/20 |

Nine of the ten whole-page events without full target-byte coverage intersect
explicit protected spans. The tenth includes Markdown link syntax: its prose
is accessible, while URL bytes are not prose units. The corresponding counts
are two protected target events in the contextual cohort and one in historical
confirmation. These partial spans matter for exact quoting and proposed edits;
they are not evidence that the editorial burden is unreachable. No protected
byte was silently made editable.

The next #349 candidate should therefore address semantic omissions across
multiple editorial relations and select precise eligible target segments while
retaining technical facts, links, code, and table context. The separate
accepted-delivery and supplied-edit safety gaps remain. The 80% FULL / 85%
accepted complete-stream gates and reserved prospective confirmation are
unchanged; #349 remains open.

`PROTOCOL.md`, `audit.py`, the archived source identity, current-main tree,
and exporter binary were hashed in `freeze.json` before execution. The public
`result.json` contains aggregate counts only. Per-event details and full
source bytes remain local. `python3 -B verify_result.py` and `python3 -B -m
unittest test_verify_result.py` check the public artifact and rejection paths.

To reproduce with the private archive and pinned exporter:

```sh
(cd research/annotation && go build -o /tmp/unswell-reviewtext ./cmd/reviewtext)
python3 audit.py --archive-root /path/to/archived/development \
  --exporter /tmp/unswell-reviewtext \
  --output /tmp/event-reachability-replay.json \
  --private-detail /tmp/event-reachability-detail.json
cmp result.json /tmp/event-reachability-replay.json
```

Run `go build` from the `research/annotation` module at the frozen main tree.
The binary SHA-256 must match `freeze.json`; the script refuses a different
binary or archive. The replay output paths must not already exist.
