# Controlled generation

This directory holds the stage C records of the
[pattern protocol](../methods/llm-patterns-v1.md). Each run under `runs/`
keeps its sampled tasks, the exact requests, the raw responses, the
generation records, and the shard manifests of the `controlled` cohort.
Nothing here labels any text as good or bad prose. A generated document is
a measured cohort member with known provenance.

The pipeline runs offline in the research module and calls no model itself:

```sh
corpus tasks --plan artifacts/measurement/dataset-plan.json \
  --sources research/acquisition/sources-v1.json --work "$WORK" \
  --cohort historical --partitions training,development --roles comment \
  --count 60 --seed unswell-llm-patterns-v1 --candidates ... > tasks.json
corpus requests --tasks tasks.json --prompts research/methods/prompts \
  --run 2026-09-11-pilot > requests.json
# the harness sends every request text to one agent and saves responses.json
corpus generations --tasks tasks.json --requests requests.json \
  --responses responses.json --records records.json --work "$WORK" \
  --output artifacts/acquisition/controlled/shards \
  --historical artifacts/acquisition/historical/records
bash scripts/measure-corpus.sh
corpus paired --records records.json --tasks tasks.json \
  --classes research/methods/rule-classes-v1.json --findings ... > paired.json
```

## Tasks

`corpus tasks` draws documentation units of the historical cohort. The draw
is seeded and stratified by ecosystem. A unit is eligible when it is a
paragraph of an admitted role in an admitted partition, has at least twelve
words, and a declaration follows it in the source. A deterministic
extractor reads that declaration into a fact sheet: the signature, its
identifiers, its parameter list, and the numbers the original states. The
sheet copies no sentence. The task set records each stratum's eligible and
selected counts.

## Requests

`corpus requests` pairs each task with each operation and prompt condition.
A request text has three parts: the operation's opening line, the frozen
prompt body with the requested word count, and the material after a
separator. The material is the fact sheet for `generate` and the original
text for `polish`. The harness line that tells an agent to use no tools is
stored beside the text. The generator receives exactly that text, and the
harness saves each response with its status.

## Records

`corpus generations` joins the responses with the requests and tasks. It
writes the `unswell-generation-v1` record with these fields:

- the full input and the raw and trimmed output, each with a hash;
- the realized length against the request, with the protocol's 30%
  tolerance;
- the word 8-gram overlap with the original, which flags copying above 0.5;
- the share of fact-sheet identifiers and numbers the response carries;
- `unavailable` for any field the harness cannot supply.

Every complete response becomes a source of the `controlled` cohort. It is
a Markdown file under the cohort's checkout directory. It carries the task's
repository, role, rights, and notices, names its task as a provenance link,
and cites the record as its origin evidence. Refused, truncated, and failed
responses stay in the record as coverage and produce no source.

## Paired tables

`corpus paired` builds the E2 tables. For each rule and each arm it reports
the share of originals with a finding, the share of responses with one, and
the paired change between them. The interval is a joint cluster bootstrap
over provenance components. Each arm also reports its difference from the
H0 documents of the same role.

## Runs

- [2026-09-11-pilot](runs/2026-09-11-pilot/README.md): the development
  pilot under amendment 1, one family through session agents.
