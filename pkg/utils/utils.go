package utils

import (
	"bnb-48-ins-indexer/pkg/helper"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"regexp"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
)

var (
	bscChainID   = big.NewInt(56)
	londonSigner = types.NewLondonSigner(bscChainID)
	eIP155Signer = types.NewEIP155Signer(bscChainID)

	maxU256 = abi.MaxUint256
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
	if strings.HasPrefix(data, "-") {
		return nil, fmt.Errorf("%s invaild, can not support neg", data)
	}

	if data == "" {
		data = "0"
	}

	bigint, ok := new(big.Int).SetString(data, 10)
	if !ok {
		return nil, fmt.Errorf("%s invaild, can not parse to bigint", data)
	}

	return bigint, nil
}

func InputToBNB48Inscription(str string) (*helper.BNB48Inscription, error) {
	bytes, err := hexutil.Decode(str)
	if err != nil {
		return nil, err
	}

	utfStr := string(bytes)

	if len(utfStr) >= 6 && utfStr[:6] == "data:," {
		utfStr = utfStr[6:]

		obj := &helper.BNB48Inscription{}
		err := json.Unmarshal([]byte(utfStr), obj)
		if err != nil {
			return nil, err
		}

		if ok := verifyInscription(obj); !ok {
			return nil, nil
		}

		obj.To = strings.ToLower(obj.To)
		if len(obj.Miners) > 0 {
			obj.Miners = strings.Split(strings.ToLower(strings.Join(obj.Miners, ",")), ",")
		}

		return obj, nil
	} else {
		return nil, nil
	}
}

func verifyInscription(ins *helper.BNB48Inscription) bool {
	if len(ins.P) > 42 {
		return false
	}

	if len(ins.Tick) > 42 {
		return false
	}

	if len(ins.TickHash) > 66 {
		return false
	}

	if len(ins.To) > 42 {
		return false
	}

	miners := strings.Join(ins.Miners, ",")
	if len(miners) > 2048 {
		return false
	}

	if ins.Decimals != "" {
		decimals, err := StringToBigint(ins.Decimals)
		if err != nil {
			return false
		}

		if decimals.Uint64() > 18 {
			return false
		}
	}

	if ins.Max != "" {
		max, err := StringToBigint(ins.Max)
		if err != nil {
			return false
		}

		if max.Cmp(maxU256) > 0 || max.Uint64() < 1 {
			return false
		}
	}

	if ins.Lim != "" {
		lim, err := StringToBigint(ins.Lim)
		if err != nil {
			return false
		}

		if lim.Cmp(maxU256) > 0 || lim.Uint64() < 1 {
			return false
		}
	}

	if ins.Amt != "" {
		amt, err := StringToBigint(ins.Amt)
		if err != nil {
			return false
		}

		if amt.Cmp(maxU256) > 0 || amt.Uint64() < 1 {
			return false
		}
	}

	return true
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

func Address2Format(address string) []string {
	res := []string{}

	for _, v := range []string{
		strings.ToLower(address),
		strings.ToUpper(address),
		common.HexToAddress(address).Hex(),
	} {
		res = append(res, hex.EncodeToString([]byte(v)))
	}
	return res
}
