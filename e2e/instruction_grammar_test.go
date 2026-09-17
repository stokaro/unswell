package e2e_test

import "testing"

func TestInstructionGrammarRevision(t *testing.T) {
	files := map[string]string{
		"draft.md": "If you want to read and write records, call the storage function.\n\n" +
			"The utility includes a reader that has the ability to decode packets.\n\n" +
			"Be sure to register the callback before starting the worker.",
		"revision.md": "To read and write records, call the storage function.\n\n" +
			"The utility includes a reader that can decode packets.\n\nRegister the callback before starting the worker.",
		"control.md": "The proxy allows you to run a local server while developing the UI.\n\n" +
			"Be sure that the callback is registered before starting the worker.",
	}
	checkInstructionRevision(t, files, 3)
}
