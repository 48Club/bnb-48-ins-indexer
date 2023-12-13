package utils

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	bnb48types "github.com/jwrookie/fans/pkg/types"
	"math/big"
	"regexp"
	"strings"
)

var (
	bscChainID   = big.NewInt(56)
	londonSigner = types.NewLondonSigner(bscChainID)
	eIP155Signer = types.NewEIP155Signer(bscChainID)
)

func GetTxFrom(tx *types.Transaction) common.Address {
	var from common.Address
	switch tx.Type() {
	case types.LegacyTxType:
		from, _ = types.Sender(eIP155Signer, tx)
	case types.DynamicFeeTxType:
		from, _ = types.Sender(londonSigner, tx)
	}

	return from
}

func StringToBigint(data string) (*big.Int, error) {
	bigint, ok := new(big.Int).SetString(data, 10)
	if !ok {
		return nil, errors.New(fmt.Sprintf("%s can not parse to bigint", data))
	}

	return bigint, nil
}

func InputToBNB48Inscription(str string) (*bnb48types.BNB48Inscription, error) {
	if str[:2] == "0x" {
		str = str[2:]
	}

	bytes, err := hex.DecodeString(str)
	if err != nil {
		return nil, err
	}

	utfStr := strings.ToLower(string(bytes))
	if utfStr[:6] == "data:," {
		utfStr = utfStr[6:]

		obj := &bnb48types.BNB48Inscription{}
		err := json.Unmarshal([]byte(utfStr), obj)
		if err != nil {
			return nil, err
		}
		return obj, nil
	} else {
		return nil, fmt.Errorf("invalid str")
	}
}

func IsValidERCAddress(address interface{}) bool {
	re := regexp.MustCompile("^0x[0-9a-fA-F]{40}$")
	switch v := address.(type) {
	case string:
		return re.MatchString(v)
	case common.Address:
		return re.MatchString(v.Hex())
	default:
		return false
	}
}
