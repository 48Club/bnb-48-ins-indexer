package global

import (
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	BscRpcUrl = "http://23.154.136.150:8545"

	MintData = "0x646174613a2c7b2270223a22626e622d3438222c226f70223a226d696e74222c227469636b223a2266616e73222c22616d74223a2231227d"

	BNB48 = "0x72b61c6014342d914470ec7ac2975be345796c2b"
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
