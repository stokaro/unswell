// Command unswell checks English prose against a local editorial policy.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/stokaro/unswell/internal/cli"
)

func main() { os.Exit(run()) }

func run() int {
	root, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	return cli.Run(ctx, os.Args[1:], cli.Environment{Dir: root, In: os.Stdin, Out: os.Stdout, Err: os.Stderr})
}
