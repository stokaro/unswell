# Independent response instructions

Read the versioned rubric and the profile in your packet. Work independently until
you have submitted every assigned response. Contact the curator if the packet or
context is incomplete; do not consult a detector or another rater's answers.

Return one JSON array containing one object per answered unit. The example below
has placeholders and cannot be submitted as a valid response. Use your assigned
opaque actor ID and copy the packet's `sha256` exactly.

```json
[
  {
    "packet_sha256": "<packet.sha256>",
    "unit_id": "<unit.id>",
    "actor_id": "<assigned actor ID>",
    "label": "<acceptable, needs_revision, or uncertain>",
    "categories": [],
    "context": "<adequate or insufficient>",
    "rationale": "<your independent explanation>",
    "recorded_at": "<submission time in RFC 3339 with time zone>"
  }
]
```

For `needs_revision`, put at least one supported rubric category in `categories`.
For `acceptable`, keep that array empty. For `uncertain`, explain the missing
context or unresolved distinction; any selected categories are only candidates.
Insufficient context requires an uncertain label. Do not guess who wrote the text.

Omit an unanswered unit; do not turn nonresponse into `acceptable` or `uncertain`.
Record when you submit the response, not an invented time or the tutorial's time.
If you must correct a submitted response, notify the curator and preserve both
versions and the reason. Corrections require an auditable new round or version.

The curator stores each submitted file unchanged, checks its packet hash and actor
assignment, and appends its records to the round's `judgments`. Run `validate`
before measuring agreement or starting adjudication. Do not send the administrative
round or other responses to raters during the independent phase.
