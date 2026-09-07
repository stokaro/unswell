// Command repocheck validates the repository's source and CI contracts.
package main

import (
	"fmt"
	"os"

	"github.com/stokaro/unswell/internal/repopolicy"
)

func main() {
	if err := repopolicy.Check(os.DirFS(".")); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
