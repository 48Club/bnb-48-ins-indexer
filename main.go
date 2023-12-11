package main

import (
	"github.com/jwrookie/fans/cmd/fans_index"
	"github.com/spf13/cobra"
	"os"
)

func newCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fans",
		Short: "fans",
	}

	cmd.AddCommand(fans_index.NewCommand())
	return cmd
}

func main() {
	cmd := newCommand()
	if err := cmd.Execute(); err != nil {
		os.Exit(-1)
	}
}
