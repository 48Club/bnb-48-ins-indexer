package main

import (
	"bnb-48-ins-indexer/cmd/api"
	"bnb-48-ins-indexer/cmd/index"
	"os"

	"github.com/spf13/cobra"
)

func newCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bnb48",
		Short: "bnb48",
	}

	cmd.AddCommand(index.NewCommand())
	cmd.AddCommand(api.NewCommand())
	return cmd
}

func main() {
	cmd := newCommand()
	if err := cmd.Execute(); err != nil {
		os.Exit(-1)
	}
}
