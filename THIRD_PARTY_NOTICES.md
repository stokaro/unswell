# Third-party notices

The release executable includes the following modules on one or more supported
platforms. Original notices are reproduced in `licenses/` and packaged with every
binary archive. Tool dependencies are isolated in `tools/go.mod` and are not
included in the runtime. The per-platform CycloneDX SBOM records build selection.
The Go runtime and standard library's BSD notice is in [licenses/go_LICENSE](licenses/go_LICENSE).

| Module | Version | Upstream notices |
| --- | --- | --- |
| `github.com/inconshreveable/mousetrap` | `v1.1.0` | [LICENSE](licenses/github.com_inconshreveable_mousetrap_LICENSE) |
| `github.com/jdkato/prose/v3` | `v3.2.1` | [LICENSE](licenses/github.com_jdkato_prose_v3_LICENSE) |
| `github.com/santhosh-tekuri/jsonschema/v6` | `v6.0.3` | [LICENSE](licenses/github.com_santhosh-tekuri_jsonschema_v6_LICENSE) |
| `github.com/spf13/cobra` | `v1.10.2` | [LICENSE.txt](licenses/github.com_spf13_cobra_LICENSE.txt) |
| `github.com/spf13/pflag` | `v1.0.9` | [LICENSE](licenses/github.com_spf13_pflag_LICENSE) |
| `github.com/odvcencio/gotreesitter` | `v0.52.0` | [LICENSE](licenses/github.com_odvcencio_gotreesitter_LICENSE) |
| `go.yaml.in/yaml/v3` | `v3.0.5` | [LICENSE](licenses/go.yaml.in_yaml_v3_LICENSE), [NOTICE](licenses/go.yaml.in_yaml_v3_NOTICE) |
| `golang.org/x/text` | `v0.14.0` | [LICENSE](licenses/golang.org_x_text_LICENSE) |
| `gopkg.in/neurosnap/sentences.v1` | `v1.0.7` | [LICENSE.md](licenses/gopkg.in_neurosnap_sentences.v1_LICENSE.md) |

## MCP modules

The separate MCP executable additionally includes these modules. Its module file
and binary build metadata define the selected dependency graph; the core CLI does
not acquire the MCP SDK. Their notices are packaged with MCP distribution artifacts.

| Module | Version | Upstream notices |
| --- | --- | --- |
| `github.com/modelcontextprotocol/go-sdk` | `v1.7.0` | [LICENSE](licenses/github.com_modelcontextprotocol_go-sdk_LICENSE) |
| `github.com/google/jsonschema-go` | `v0.4.3` | [LICENSE](licenses/github.com_google_jsonschema-go_LICENSE) |
| `github.com/segmentio/encoding` | `v0.5.4` | [LICENSE](licenses/github.com_segmentio_encoding_LICENSE) |
| `github.com/segmentio/asm` | `v1.1.3` | [LICENSE](licenses/github.com_segmentio_asm_LICENSE) |
| `github.com/yosida95/uritemplate/v3` | `v3.0.2` | [LICENSE](licenses/github.com_yosida95_uritemplate_v3_LICENSE) |
| `golang.org/x/oauth2` | `v0.35.0` | [LICENSE](licenses/golang.org_x_oauth2_LICENSE) |
| `golang.org/x/sync` | `v0.20.0` | [LICENSE](licenses/golang.org_x_sync_LICENSE) |
| `golang.org/x/sys` | `v0.41.0` | [LICENSE](licenses/golang.org_x_sys_LICENSE) |
| `golang.org/x/time` | `v0.15.0` | [LICENSE](licenses/golang.org_x_time_LICENSE) |

## Model data

- POS: Prose `tag/aptagmodel/en.bin`, SHA-256
  `209282c71733883be5c08af24301cc8917079c9deb7721b89a8dd118d071a780`.
  Prose's pinned LICENSE explicitly distributes its embedded tagger models under
  MIT terms and credits Joseph Kato and Matthew Honnibal.
- Sentence boundaries: neurosnap `data/english.json`, SHA-256
  `2498632eaf8c3d0480c074e0331057ea9b20f1306523db06ff742689269a21a8`.
  The pinned neurosnap distribution carries Eric Bower's MIT license and describes
  its Punkt implementation as an NLTK port. This records upstream distribution
  terms; Unswell has not reconstructed the original model's training corpus.
- Shallow chunks and rule dictionaries: hand-authored Unswell code, MIT.

The NER model is not imported. The alpha includes no revision-probability model,
editorial training corpus, or downloadable runtime models.

## Test schema

`report/testdata/sarif-schema-2.1.0.json` is the official OASIS SARIF 2.1.0
errata 01 schema, retrieved from
https://docs.oasis-open.org/sarif/sarif/v2.1.0/errata01/os/schemas/sarif-schema-2.1.0.json.
Its embedded copyright and license notices are retained. It is test data and is
not embedded in the executable. The Unswell result schema is generated from
Unswell's public Go types and distributed under this project's MIT license.

## Test and research dependencies

Root e2e tests use `github.com/dlclark/regexp2` v1.11.0 (MIT) to exercise GitHub
Runner's ECMAScript matcher semantics. This test dependency is not linked into
the CLI or MCP server. Its original notices remain in the dependency source.

The isolated `research/dependencies` consumer module evaluates
`github.com/bioshock/gospacy/v3` at `v3.8.14-port.2` (MIT). Its selected dependency
graph and model component terms are documented in that module's README. It is not
linked into the Unswell CLI, MCP server, or published library module. No spaCy
model weights or Python environment are distributed in this repository.
The saved dependency traces are predictions on agent-authored structural probes,
not human grammatical judgments or editorial annotations.
