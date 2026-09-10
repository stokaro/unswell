# SARIF output and consumers

Unswell writes SARIF 2.1.0 from the same result as every other report. This page
records how that output is checked against the specification and against a real
consumer, and what those checks do and do not establish.

## Schema validation

`report.ValidateSARIF` checks one report against the published SARIF 2.1.0
schema. The schema is embedded, and the loader refuses every external resource,
so validation needs no network and no downloads. The pinned copy comes from
`https://json.schemastore.org/sarif-2.1.0.json`, the location the reports
themselves declare, and its digest is recorded next to the embedded file.

Tests validate clean reports, reports with findings, and reports over Unicode
headings, list items and tables. They also confirm the check rejects a broken
document: invalid JSON, a version the specification does not define, a missing
`runs` property, and a column kind outside the enumeration. A valid document is
well formed. It is not evidence that a particular consumer accepts it.

## A real consumer

`scripts/verify-sarif-consumer.sh` imports a report into Sarif.Multitool, the
reference implementation from the specification authors. It runs in a pinned
container through an explicit Docker context, so the check needs no local .NET
installation:

```sh
make build
bash scripts/verify-sarif-consumer.sh --context remote-dev-container
```

The script writes the consumer's own SARIF validation log and a summary of its
results by level. It is not part of `make check`, because it needs a container
runtime and a network to fetch the consumer.

[The recorded check](sarif/consumer-check.json) imported a 14,134,807 byte
report of this repository, produced by the build at the recorded commit.
Sarif.Multitool 5.7.0 completed its analysis successfully and reported no
validation results at any level. That is one report from one build accepted by
one consumer. Other consumers apply their own rules, and GitHub code scanning in
particular enforces limits this check does not test.

## What the reports carry

Findings become results with their rule identity, level, message, and original
byte-mapped locations. The run properties carry the gate decision, assessments,
manifest, suppressions, and optional measurements, so a consumer that keeps
`properties` can recover everything the JSON report holds. Source text appears
only when a run explicitly includes it.
