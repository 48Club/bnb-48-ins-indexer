package fans_index

import (
	"github.com/jwrookie/fans/pkg/database"
	"github.com/jwrookie/fans/pkg/log"
	"github.com/jwrookie/fans/servers"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "fans_index",
		Short: "fans_index",
		Run: func(cmd *cobra.Command, args []string) {
			setup()
		},
	}
}

func setup() {
	log.Init("fans_index.log")
	database.NewMysql()

	bsc := servers.NewBscScanService()

	if err := bsc.Scan(); err != nil {
		panic(err)
	}
}
