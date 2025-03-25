package main

import (
	"context"
	"fmt"
	"os"

	"github.com/pmalek/github-pm-groomer/cmd"
)

func main() {
	ctx := context.Background()
	if err := cmd.Execute(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
