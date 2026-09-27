package selfcheck

import (
	"errors"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/mcp/internal/server"
)

// failure retains the rejected response separately from successfully matched batches.
// Source identities describe the request even when no response was received.
type failure struct {
	Stage   string              `json:"stage"`
	Message string              `json:"message"`
	Batch   int                 `json:"batch,omitempty"`
	Probe   string              `json:"probe,omitempty"`
	Sources []sourceIdentity    `json:"sources,omitempty"`
	Actual  *server.CheckOutput `json:"actual,omitempty"`
	cause   error
}

type sourceIdentity struct {
	Name       string          `json:"name"`
	Format     document.Format `json:"format"`
	SourceHash string          `json:"source_hash"`
	Bytes      int             `json:"bytes"`
}

func (f *failure) Error() string { return f.Message }
func (f *failure) Unwrap() error { return f.cause }

func failureRecord(stage string, err error) *failure {
	if retained, ok := errors.AsType[*failure](err); ok {
		return retained
	}
	return &failure{Stage: stage, Message: err.Error(), cause: err}
}

func batchFailure(index int, documents []unswell.DocumentResult, checked server.CheckOutput, err error) error {
	record := failureRecord("repository", err)
	record.Batch = index
	for _, doc := range documents {
		record.Sources = append(record.Sources, sourceIdentity{
			Name: doc.Name, Format: doc.Format, SourceHash: doc.SourceHash, Bytes: doc.Bytes,
		})
	}
	record.Actual = failedOutput(checked)
	return record
}

func probeFailure(name string, checked server.CheckOutput, err error) error {
	record := failureRecord("probes", err)
	record.Probe = name
	record.Actual = failedOutput(checked)
	return record
}

func failedOutput(checked server.CheckOutput) *server.CheckOutput {
	if checked.Outcome == "" && checked.Result.Status == "" {
		return nil // No structured response arrived; do not invent one.
	}
	checked.Result = normalize(checked.Result)
	return &checked
}
