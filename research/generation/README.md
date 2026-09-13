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

`corpus screen` then screens one partition of the plan with the runs that
enter selection; the [patterns README](../annotation/patterns/README.md)
describes it.

A later run passes the earlier task sets to `corpus tasks --exclude-tasks`,
so no task is drawn twice, and names its shards with
`corpus generations --shard-suffix`, so its shards sit beside the earlier
run's shards of the same repositories. `bash scripts/measure-corpus.sh
--resume` then measures the new shards only and rebuilds the tables.

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

Three harnesses have run. A Claude session agent gets the request text and
the harness line as its one message. A Codex CLI run under amendment 3 gets
the same message on standard input through `codex exec`. Its sandbox is
read-only, its working directory is empty, and its Codex home holds no
instruction file or skill. The CLI's header names the model. If a run
prints a command, the harness marks the response as an error. An endpoint
run sends the same message as the one user message of a chat request with
fixed decoding parameters. Its response row keeps the identifier, the token
count, and the duration the endpoint returned.

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
a Markdown file under the cohort's checkout directory, in a directory named
by the run. Request IDs repeat across runs on the same tasks, and one run
must not overwrite another's files. It carries the task's
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
- [2026-09-11-run2](runs/2026-09-11-run2/README.md): the second run of the
  same family on 140 new tasks, drawn without the pilot's tasks.
- [2026-09-12-haiku](runs/2026-09-12-haiku/README.md): a second model of
  the same family on the 200 tasks of the two earlier runs, so every task
  has a response from both models.
- [2026-09-12-codex-luna](runs/2026-09-12-codex-luna/README.md): the second family,
  OpenAI through the Codex CLI at its weakest model and lowest effort, on
  the same 200 tasks under amendment 3.
- [2026-09-12-qwen](runs/2026-09-12-qwen/README.md): the third family, Qwen
  through the organization's endpoint with thinking disabled, on the same
  200 tasks under amendment 3.
- [2026-09-12-qwen-27b](runs/2026-09-12-qwen-27b/README.md): the larger of
  the endpoint's two Qwen models under the same parameters, so the third
  family has two model strata like the first.
- [2026-09-12-round2-haiku](runs/2026-09-12-round2-haiku/README.md): 150 new
  tasks from the ten repositories amendment 5 pinned to `development`,
  answered by the second Claude model.
- [2026-09-12-round2-luna](runs/2026-09-12-round2-luna/README.md): the same
  150 tasks answered by the OpenAI model, so both selecting families cover
  the enlarged development partition.
- [2026-09-13-round3-haiku](runs/2026-09-13-round3-haiku/README.md): 150
  further tasks from the development repositories that carried no
  controlled response, which lifts the controlled arm to the cluster
  minimum of 20 components.
- [2026-09-13-round3-luna](runs/2026-09-13-round3-luna/README.md): the same
  150 tasks answered by the OpenAI model.
