package global

import (
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	BscRpcUrl = "https://rpc.ankr.com/bsc/ace69abcacd8190d01f7c729847a286095340c412af9e4fff3c20547235d0364"

	BNB48Prefix = "0x646174613a2c7b2270223a22626e622d3438222c226f70223a22"
)

var (
	BscClient *ethclient.Client
)

func init() {
	_bscRpc, err := ethclient.Dial(BscRpcUrl)
	if err != nil {
		panic(err)
	}

	BscClient = _bscRpc
}
