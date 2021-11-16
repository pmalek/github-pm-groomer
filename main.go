package main

import (
	"fmt"
	"github.com/lahabana/github-pm-groomer/cmd"
	"os"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
