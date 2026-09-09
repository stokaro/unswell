package main

// White-box tests: Inject reader and writer failures and cancellation into the command runner;
// the executable interface cannot supply these in-process I/O implementations.

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"
	"time"

	qt "github.com/frankban/quicktest"
)

type brokenIO struct{}

func (brokenIO) Read([]byte) (int, error)  { return 0, io.ErrUnexpectedEOF }
func (brokenIO) Write([]byte) (int, error) { return 0, io.ErrClosedPipe }

func TestCommandFailures(t *testing.T) {
	c := qt.New(t)
	data, err := os.ReadFile("../../testdata/tutorial.json")
	c.Assert(err, qt.IsNil)
	c.Assert(run(t.Context(), nil, brokenIO{}, io.Discard), qt.ErrorMatches, "usage:.*")
	c.Assert(run(t.Context(), []string{"unknown"}, brokenIO{}, io.Discard), qt.ErrorMatches, "unknown annotation command.*")
	c.Assert(run(t.Context(), []string{"validate"}, brokenIO{}, io.Discard), qt.ErrorIs, io.ErrUnexpectedEOF)
	cases := []struct{ command string }{{"validate"}, {"packet"}, {"agreement"}, {"decisions"}}
	for _, tc := range cases {
		t.Run(tc.command, func(t *testing.T) {
			c := qt.New(t)
			c.Assert(run(t.Context(), []string{tc.command}, bytes.NewReader(data), brokenIO{}), qt.ErrorIs, io.ErrClosedPipe)
		})
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	c.Assert(run(ctx, []string{"validate"}, brokenIO{}, io.Discard), qt.ErrorIs, context.Canceled)
}

type blockedIO struct {
	started  chan struct{}
	release  chan struct{}
	finished chan struct{}
}

func (b blockedIO) Read([]byte) (int, error) {
	close(b.started)
	<-b.release
	close(b.finished)
	return 0, io.EOF
}

func (b blockedIO) Write([]byte) (int, error) {
	close(b.started)
	<-b.release
	close(b.finished)
	return 0, io.ErrClosedPipe
}

func (b blockedIO) close() {
	close(b.release)
	select {
	case <-b.started:
		<-b.finished
	default:
	}
}

func TestCancellationDuringIO(t *testing.T) {
	cases := []struct{ name string }{{"reader"}, {"writer"}}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := qt.New(t)
			data, err := os.ReadFile("../../testdata/tutorial.json")
			c.Assert(err, qt.IsNil)
			blocked := blockedIO{started: make(chan struct{}), release: make(chan struct{}), finished: make(chan struct{})}
			defer blocked.close()
			var input io.Reader = bytes.NewReader(data)
			output := io.Discard
			if tc.name == "reader" {
				input = blocked
			} else {
				output = blocked
			}
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			done := make(chan error, 1)
			go func() { done <- run(ctx, []string{"validate"}, input, output) }()
			select {
			case <-blocked.started:
			case <-time.After(5 * time.Second):
				t.Fatal("command did not enter I/O")
			}
			cancel()
			select {
			case err := <-done:
				c.Assert(err, qt.ErrorIs, context.Canceled)
			case <-time.After(5 * time.Second):
				t.Fatal("command remained blocked after cancellation")
			}
		})
	}
}
