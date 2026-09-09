package server_test

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// responseBarrier holds one response after initialization, then observes the
// actual transport write. Handler completion alone does not drain that write.
type responseBarrier struct {
	mcp.Transport
	armed    atomic.Bool
	started  chan struct{}
	release  chan struct{}
	finished chan error
	unblock  func()
}

func newResponseBarrier(transport mcp.Transport) *responseBarrier {
	barrier := &responseBarrier{
		Transport: transport,
		started:   make(chan struct{}),
		release:   make(chan struct{}),
		finished:  make(chan error, 1),
	}
	barrier.unblock = sync.OnceFunc(func() { close(barrier.release) })
	return barrier
}

func (b *responseBarrier) Connect(ctx context.Context) (mcp.Connection, error) {
	connection, err := b.Transport.Connect(ctx)
	if err != nil {
		return nil, err
	}
	return &responseConnection{Connection: connection, barrier: b}, nil
}

type responseConnection struct {
	mcp.Connection
	barrier *responseBarrier
}

func (c *responseConnection) Write(ctx context.Context, message jsonrpc.Message) error {
	if _, response := message.(*jsonrpc.Response); !response || !c.barrier.armed.CompareAndSwap(true, false) {
		return c.Connection.Write(ctx, message)
	}
	close(c.barrier.started)
	<-c.barrier.release
	err := c.Connection.Write(ctx, message)
	c.barrier.finished <- err
	return err
}
