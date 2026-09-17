package unswell_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestInstructionActionGrammar(t *testing.T) {
	for _, text := range []string{
		"If you want to read and write records, call the storage function.",
		"If you need to inspect the records, you can query the index.",
		"If you want to synchronize the repositories, run the synchronization command.",
		"If we intend to decode a packet, we need to read its header.",
		"Be sure to register the callback before starting the worker.",
		"The parser is capable of decoding a packet.",
		"The utility includes a reader that has the ability to decode packets.",
		"The utility includes a reader which can be used to decode packets.",
		"The `Read` method, which is used to fetch records, returns a stream.",
		"The toolkit provides a function that is designed to allow readers to decode packets.",
		"To configure the queue, you need to follow these steps.",
		"In order to configure the queue, you have to perform the following steps.",
		"To configure the queue, follow the following steps.",
		"The client allows you to specify the header to be used on the connection.",
	} {
		t.Run(text, func(t *testing.T) {
			c := qt.New(t)
			r := singleRuleResult(t, "filler.instruction-scaffolding", text, "", "")
			c.Assert(r.Findings, qt.HasLen, 1)
			c.Assert(r.Findings[0].Primary.Snippet, qt.Equals, strings.TrimSuffix(text, "."))
			c.Assert(r.Findings[0].RuleVersion, qt.Equals, "9")
		})
	}
}

func TestInstructionActionGrammarControls(t *testing.T) {
	for _, text := range []string{
		"If you believe the records are current, the worker can resume.",
		"If the worker needs to inspect the records, it queries the index.",
		"If you want to read the records, the worker updates the index.",
		"If you want to inspect records, you may need to ask an administrator.",
		"If you need permission to query the index, request approval.",
		"If you want to avoid a failure, do not restart the service.",
		"If you want to read and the worker writes records, check the result.",
		"Be sure that the callback is registered before starting the worker.",
		"Be sure to never delete the checkpoint.",
		"The operator is sure to read the records.",
		"The parser can decode a packet.",
		"The parser is capable of storage.",
		"The parser is not capable of decoding a packet.",
		"The utility includes a reader that can decode packets.",
		"The utility includes a reader that allows users to decode packets.",
		"The utility includes a reader that grants access to records.",
		"The utility includes a reader that has the ability to leak credentials.",
		"The log says: the utility includes a reader that has the ability to decode packets.",
		"The utility includes a reader `that has the ability to decode` packets.",
		"The utility includes a reader that has the ability to `decode` packets.",
		"To configure the queue, the scheduler performs the following steps.",
		"To configure the queue, follow the procedure only after approval.",
		"To configure the queue, set the size.",
		"The client allows you to specify a header.",
		"The client allows you to specify the header to be stored on the connection.",
		"The client allows you to run the server while developing the UI.",
		"The client allows you to bind the same variable in multiple arguments.",
	} {
		t.Run(text, func(t *testing.T) {
			qt.New(t).Assert(singleRuleResult(t, "filler.instruction-scaffolding", text, "", "").Findings, qt.HasLen, 0)
		})
	}
}

func TestInstructionRelatedObject(t *testing.T) {
	first := "The client allows the user to set the transfer speed conditions that must be met to let the transfer keep going."
	second := "By using the switch `-y` and `-Y` you can make curl abort transfers " +
		"if the transfer speed is below the specified lowest limit for a specified time."
	for _, gap := range []string{" ", "\n\n"} {
		r := singleRuleResult(t, "filler.instruction-scaffolding", first+gap+second, "", "")
		c := qt.New(t)
		c.Assert(r.Findings, qt.HasLen, 1)
		c.Assert(r.Findings[0].Primary.Snippet, qt.Equals, strings.TrimSuffix(first, "."))
		c.Assert(r.Findings[0].Related, qt.HasLen, 1)
		c.Assert(r.Findings[0].Related[0].Snippet, qt.Equals, strings.TrimSuffix(second, "."))
	}
	for _, gap := range []string{"\n\n## Another topic\n\n", "\n\n```text\nelsewhere\n```\n\n"} {
		qt.New(t).Assert(singleRuleResult(t, "filler.instruction-scaffolding", first+gap+second, "", "").Findings, qt.HasLen, 0)
	}
	text := first + " By using the switch you can configure the cache capacity."
	qt.New(t).Assert(singleRuleResult(t, "filler.instruction-scaffolding", text, "", "").Findings, qt.HasLen, 0)
}
