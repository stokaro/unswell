module example.com/unswell-consumer

go 1.25.0

require (
	github.com/frankban/quicktest v1.14.6
	github.com/stokaro/unswell v0.1.0-alpha.1.0.20260908025147-fcb39738b4a0
	github.com/stokaro/unswell/goanalysis v0.0.0
	golang.org/x/tools v0.49.0
)

require (
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/jdkato/prose/v3 v3.2.1 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/rogpeppe/go-internal v1.9.0 // indirect
	github.com/santhosh-tekuri/jsonschema/v6 v6.0.3 // indirect
	github.com/stokaro/gotreesitter v0.52.1-0.20260913084044-276f5cc0ec6f // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/text v0.14.0 // indirect
	gopkg.in/neurosnap/sentences.v1 v1.0.7 // indirect
)

replace github.com/stokaro/unswell => ../..

replace github.com/stokaro/unswell/goanalysis => ../../goanalysis
