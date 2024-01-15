package utils

import (
	"bnb-48-ins-indexer/config"
	"bnb-48-ins-indexer/pkg/helper"
	"bnb-48-ins-indexer/pkg/log"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
)

var (
	bscChainID        = big.NewInt(56)
	londonSigner      = types.NewLondonSigner(bscChainID)
	eIP155Signer      = types.NewEIP155Signer(bscChainID)
	dataRe, _         = regexp.Compile("data:([^\"]*),(.*)")
	maxU256           = abi.MaxUint256
	bulkCannotContain = mapset.NewSet[string]()
)

func init() {
	bulkCannotContain.Append(config.GetConfig().App.BulkCannotContain...)
}

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

func Error(err, ignoreErr error, tx, mgs string) error {
	if errors.Is(err, ignoreErr) {
		log.Sugar.Debugf("tx: %s, error: %s", tx, mgs)
		return nil
	}

	return err
}

func MustStringToBigint(data string) *big.Int {
	res, err := StringToBigint(data)
	if err != nil {
		panic(err)
	}

	return res
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

func InputToBNB48Inscription(str string, bn ...uint64) ([]*helper.BNB48Inscription, error) {

	bytes, err := hexutil.Decode(str)
	if err != nil {
		return nil, err
	}

	utfStr := string(bytes)

	if len(bn) > 0 && bn[0] >= 34_778_248 /*支持 application/json 与 bulktTx 的区块高度*/ {
		return InputToBNB48Inscription2(utfStr)
	}

	if len(utfStr) >= 6 && utfStr[:6] == "data:," {
		utfStr = utfStr[6:]

		obj := &helper.BNB48Inscription{}
		if err := json.Unmarshal([]byte(utfStr), obj); err != nil {
			return nil, err
		}

		if ok := verifyInscription(obj); !ok {
			return nil, nil
		}

		obj.To = strings.ToLower(obj.To)
		// obj.From = strings.ToLower(obj.From)
		// obj.Spender = strings.ToLower(obj.Spender)
		if len(obj.Miners) > 0 {
			obj.Miners = strings.Split(strings.ToLower(strings.Join(obj.Miners, ",")), ",")
		}

		return []*helper.BNB48Inscription{obj}, nil
	} else {
		return nil, nil
	}
}

func InputToBNB48Inscription2(utfStr string) (inss []*helper.BNB48Inscription, err error) {

	rs := dataRe.FindStringSubmatch(utfStr)
	if len(rs) != 3 {
		return nil, nil
	}
	if rs[1] != "" && rs[1] != "application/json" {
		return nil, nil
	}

	s := rs[2]
	var tmp interface{}
	if err := json.Unmarshal([]byte(s), &tmp); err != nil {
		return nil, err
	}

	switch tmp.(type) {
	case map[string]interface{}:
		var i helper.BNB48Inscription
		if err = json.Unmarshal([]byte(s), &i); err == nil {
			inss = append(inss, &i)
		}
	case []interface{}:
		var i []*helper.BNB48Inscription
		if err = json.Unmarshal([]byte(s), &i); err == nil {
			inss = append(inss, i...)
		}
	}
	mustCheckOP := len(inss) > 1

	for k, ins := range inss {
		if ok := verifyInscription(ins); !ok {
			return nil, nil
		}
		inss[k].To = strings.ToLower(ins.To)
		inss[k].Spender = strings.ToLower(ins.Spender)
		inss[k].From = strings.ToLower(ins.From)
		if len(ins.Miners) > 0 {
			inss[k].Miners = strings.Split(strings.ToLower(strings.Join(ins.Miners, ",")), ",")
		}

		if mustCheckOP && bulkCannotContain.ContainsOne(ins.Op) {
			return nil, nil
		}
	}

	return inss, err
}

func verifyInscription(ins *helper.BNB48Inscription) bool {
	for _, address := range []string{ins.To, ins.From, ins.Spender} {
		if address != "" && !IsValidERCAddress(address) {
			return false
		}
	}

	if len(ins.P) > 42 {
		return false
	}

	if len(ins.Tick) > 42 {
		return false
	}

	if len(ins.TickHash) > 66 {
		return false
	}

	for _, addresss := range [][]string{ins.Miners, ins.Minters} {
		for _, address := range addresss {
			if !IsValidERCAddress(address) {
				return false
			}
		}
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

	for _, v := range []string{ins.Max, ins.Lim, ins.Amt, ins.Commence} {
		if v != "" {
			if bv, err := StringToBigint(v); err != nil {
				return false
			} else if bv.Cmp(maxU256) > 0 || bv.Uint64() < 1 {
				return false
			}

		}
	}

	for address, amt := range ins.Reserves {
		if !IsValidERCAddress(address) {
			return false
		}

		if bv, err := StringToBigint(amt); err != nil {
			return false
		} else if bv.Cmp(maxU256) > 0 || bv.Uint64() < 1 {
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
