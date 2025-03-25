package cmd

import (
	"context"
	"os"

	"github.com/pmalek/github-pm-groomer/internal/github/api"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "github-pm-groomer",
		Short: "A CLI to do common product management stuff on github",
	}
	ghClient api.Client
)

func Execute(ctx context.Context) error {
	ghClient = api.New(os.Getenv("GITHUB_API_TOKEN"))
	if err := ghClient.Ping(ctx); err != nil {
		return err
	}
	return rootCmd.ExecuteContext(ctx)
}
