package index

import (
	"bnb-48-ins-indexer/pkg/database"
	"bnb-48-ins-indexer/pkg/log"
	"bnb-48-ins-indexer/service"

	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "index",
		Short: "index",
		Run: func(cmd *cobra.Command, args []string) {
			setup()
		},
	}
}

func setup() {
	log.Init("index.log")
	database.NewMysql()

	bsc := service.NewBscScanService()

	if err := bsc.Scan(); err != nil {
		panic(err)
	}
}
