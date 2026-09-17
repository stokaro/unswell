package unswell_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestNarratedInstruction(t *testing.T) {
	for _, text := range []string{
		"Now that we have a client to work with we need to pull an image.",
		"Now that we have a connection, we need to send the request.",
		"Now that we have a task in the created state we need to make sure that we wait on the task to exit.",
		"We need to make sure that we wait on the process to exit.",
		"We will need to ensure that we configure the queue before starting the worker.",
		"Now that we have a record, we have to validate its checksum before storing it.",
	} {
		t.Run(text, func(t *testing.T) {
			c := qt.New(t)
			r := singleRuleResult(t, "filler.instruction-scaffolding", text, "", "")
			c.Assert(r.Findings, qt.HasLen, 1)
			c.Assert(r.Findings[0].Primary.Snippet, qt.Equals, strings.TrimSuffix(text, "."))
		})
	}
}

func TestNarratedInstructionControls(t *testing.T) {
	for _, text := range []string{
		"We need to pull an image.",
		"We have to wait on the task before submitting the request.",
		"Now the worker needs to pull an image.",
		"Now that we have a client, the worker needs to pull an image.",
		"We need to make sure that the worker exits.",
		"We need to ensure that the queue is empty.",
		"We need to ensure that we never expose a secret.",
		"Now that we have permission, we need to set the queue size.",
		"Now that we have a client, we might need to configure the queue.",
		"Now that we have a client, we need to avoid losing the connection.",
		"The manual says we need to make sure that we wait on the task.",
		"We need to `make sure that we wait` on the task.",
		"We need to make sure that `we` wait on the task.",
		"Do we need to make sure that we wait on the task?",
	} {
		t.Run(text, func(t *testing.T) {
			qt.New(t).Assert(singleRuleResult(t, "filler.instruction-scaffolding", text, "", "").Findings, qt.HasLen, 0)
		})
	}
}

func TestCapabilityNeedsInstructionContext(t *testing.T) {
	for _, text := range []string{
		"This allows you to run a normal server to handle API requests, while iterating separately on the UI.",
		"The interface allows users to configure the service.",
		"The library enables you to parse an archive.",
		"The console allows the reader to configure the queue.",
	} {
		qt.New(t).Assert(singleRuleResult(t, "filler.instruction-scaffolding", text, "", "").Findings, qt.HasLen, 0)
	}
	for _, gap := range []string{" ", "\n\n"} {
		r := singleRuleResult(t, "filler.instruction-scaffolding", "The interface allows users to configure the queue."+gap+
			"This is done by using the queue option.", "", "")
		qt.New(t).Assert(r.Findings, qt.HasLen, 1)
		qt.New(t).Assert(r.Findings[0].Related, qt.HasLen, 1)
	}
}

func TestOperationSupportClauses(t *testing.T) {
	for _, text := range []string{
		"Query method is used to determine if there exist a possible cache link between the input and a vertex.",
		"Load method is used to load a specific record into a result reference.",
		"The `Read` method is used to fetch a stored record.",
		"The lookup function is used to check whether the index contains a record.",
		"It can also be used for associating messages with the vertex that can be helpful for tracing purposes.",
		"The client may also be used for sending requests to a local service.",
		"The adapter can be used to query the index: the error is reported separately.",
	} {
		t.Run(text, func(t *testing.T) {
			qt.New(t).Assert(singleRuleResult(t, "filler.instruction-scaffolding", text, "", "").Findings, qt.HasLen, 1)
		})
	}
	for _, text := range []string{
		"The adapter is used to query the index.",
		"The method is used to query the index.",
		"The cache can be used for storage.",
		"The read function is not used to fetch a record.",
		"The old method is used only to read the cache.",
		"The manual says: the read function is used to fetch a record.",
		"The manual says \"The read function is used to fetch a record\".",
		"The method `is used to query` returns a result.",
	} {
		qt.New(t).Assert(singleRuleResult(t, "filler.instruction-scaffolding", text, "", "").Findings, qt.HasLen, 0)
	}
}

func TestNestedEnabledActors(t *testing.T) {
	for _, text := range []string{
		"The library is designed to allow other applications to use it.",
		"The adapter has the ability to enable workers to process batches.",
		"The map is intended to allow multiple arguments to be passed to templates.",
	} {
		qt.New(t).Assert(singleRuleResult(t, "filler.instruction-scaffolding", text, "", "").Findings, qt.HasLen, 1)
	}
	for _, text := range []string{
		"The library allows other applications to use it.",
		"The adapter allows workers to process batches.",
		"The console is designed to allow administrators to grant access.",
		"The map is intended to be passed to templates.",
	} {
		qt.New(t).Assert(singleRuleResult(t, "filler.instruction-scaffolding", text, "", "").Findings, qt.HasLen, 0)
	}
}
