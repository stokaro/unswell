package e2e_test

import "testing"

func TestInstructionGrammarRevision(t *testing.T) {
	files := map[string]string{
		"draft.md": "The parser is capable of decoding a packet.\n\n" +
			"The utility includes a reader that has the ability to decode packets.",
		"revision.md": "The parser can decode a packet.\n\n" +
			"The utility includes a reader that can decode packets.",
		// Version 19 reports neither a conditional reader goal nor a reminder.
		"control.md": "The proxy allows you to run a local server while developing the UI.\n\n" +
			"If you want to read and write records, call the storage function.\n\n" +
			"Be sure to register the callback before starting the worker.\n\n" +
			"Be sure that the callback is registered before starting the worker.",
	}
	checkInstructionRevision(t, files, 2)
}
