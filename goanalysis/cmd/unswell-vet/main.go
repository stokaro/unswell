// Command unswell-vet runs the default Unswell policy through Go analysis drivers.
package main

import (
	"fmt"
	"os"

	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/stokaro/unswell/goanalysis"
)

func main() {
	analyzer, err := goanalysis.New(goanalysis.Options{})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	singlechecker.Main(analyzer)
}
