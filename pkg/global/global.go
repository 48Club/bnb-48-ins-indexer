package global

import (
	"bnb-48-ins-indexer/config"

	"github.com/ethereum/go-ethereum/ethclient"
)

var BscClient *ethclient.Client

func init() {
	_bscRpc, err := ethclient.Dial(config.GetConfig().App.BscRpc)
	if err != nil {
		panic(err)
	}

	BscClient = _bscRpc
}
