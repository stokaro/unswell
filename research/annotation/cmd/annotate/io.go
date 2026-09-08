package main

import "context"

type ioResult[T any] struct {
	value T
	err   error
}

// A terminal read may survive closing its descriptor on macOS. Let main return
// on cancellation; process exit ends any outstanding stdin/stdout operation.
// The buffered result permits completion after the waiting call has returned.
func awaitIO[T any](ctx context.Context, operation func() (T, error)) (T, error) {
	var zero T
	if err := ctx.Err(); err != nil {
		return zero, err
	}
	done := make(chan ioResult[T], 1)
	go func() {
		value, err := operation()
		done <- ioResult[T]{value: value, err: err}
	}()
	select {
	case <-ctx.Done():
		return zero, ctx.Err()
	case result := <-done:
		return result.value, result.err
	}
}
