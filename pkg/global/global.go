package global

import (
	"bnb-48-ins-indexer/config"

	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	BNB48Prefix = "0x646174613a2c7b2270223a22626e622d3438222c226f70223a22"
)

var (
	BscClient *ethclient.Client
)

func init() {
	_bscRpc, err := ethclient.Dial(config.GetConfig().App.BscRpc)
	if err != nil {
		panic(err)
	}

	BscClient = _bscRpc
}
