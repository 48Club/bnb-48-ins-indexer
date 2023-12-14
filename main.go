package main

import (
	"github.com/jwrookie/fans/cmd/api"
	"github.com/jwrookie/fans/cmd/index"
	"github.com/spf13/cobra"
	"os"
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
