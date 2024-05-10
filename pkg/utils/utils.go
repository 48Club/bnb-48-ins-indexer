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
	opHasAmtMin1      = mapset.NewSet[string]()
	opHasAmtMin0      = mapset.NewSet[string]()
	opHasTo           = mapset.NewSet[string]()
)

const (
	FutureEnableBNForPR67 uint64 = 35_354_848 // more detail: https://github.com/48Club/bnb-48-ins-indexer/pull/67
)

func init() {
	bulkCannotContain.Append(config.GetConfig().App.BulkCannotContain...)
	opHasAmtMin1.Append("mint", "transfer", "burn")
	opHasAmtMin0.Append("approve", "transferFrom")
	opHasTo.Append("transfer", "transferFrom")
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
		return common.Big0, nil
	}

	bigint, ok := new(big.Int).SetString(data, 10)
	if !ok {
		return nil, fmt.Errorf("%s invaild, can not parse to bigint", data)
	}

	return bigint, nil
}

func InputToBNB48Inscription(i interface{}, bn ...uint64) ([]*helper.BNB48InscriptionVerified, error) {
	var utfStr string
	switch v := i.(type) {
	case string:
		b, e := hexutil.Decode(v)
		if e != nil {
			return nil, e
		}
		utfStr = string(b)
	case []byte:
		utfStr = string(v)
	default:
		return nil, fmt.Errorf("unsupported type: %T", i)
	}

	if len(bn) > 0 && bn[0] >= 34_778_248 /*支持 application/json 与 bulktTx 的区块高度*/ {
		return InputToBNB48Inscription2(utfStr, bn...)
	}

	if len(utfStr) >= 6 && utfStr[:6] == "data:," {
		utfStr = utfStr[6:]

		obj := &helper.BNB48Inscription{}
		if err := json.Unmarshal([]byte(utfStr), obj); err != nil {
			return nil, err
		}
		objV, ok := verifyInscription(obj)
		if !ok {
			return nil, nil
		}

		return []*helper.BNB48InscriptionVerified{objV}, nil
	} else {
		return nil, nil
	}
}

func InputToBNB48Inscription2(utfStr string, bn ...uint64) (inss []*helper.BNB48InscriptionVerified, err error) {

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

	_inss := []*helper.BNB48Inscription{}
	switch tmp.(type) {
	case map[string]interface{}:
		var i helper.BNB48Inscription
		if err = json.Unmarshal([]byte(s), &i); err == nil {
			_inss = append(_inss, &i)
		}
	case []interface{}:
		var i []*helper.BNB48Inscription
		if err = json.Unmarshal([]byte(s), &i); err == nil {
			_inss = append(_inss, i...)
		}
	}
	mustCheckOP := len(_inss) > 1

	for _, ins := range _inss {
		objV, ok := verifyInscription(ins, bn...)
		if !ok {
			return nil, nil
		}

		if mustCheckOP && bulkCannotContain.ContainsOne(ins.Op) {
			return nil, nil
		}
		inss = append(inss, objV)
	}

	return inss, err
}

func verifyInscription(_ins *helper.BNB48Inscription, bn ...uint64) (ins *helper.BNB48InscriptionVerified, b bool) {
	var target uint64 = 0
	if len(bn) > 0 {
		target = bn[0]
	}
	ins = &helper.BNB48InscriptionVerified{
		BNB48Inscription: _ins,
		ReservesV:        map[string]*big.Int{},
	}

	if len(ins.P) > 42 {
		return ins, false
	}

	if ins.Op != "deploy" && len(ins.TickHash) != 66 {
		return ins, false
	}

	var ok bool

	if opHasAmtMin0.ContainsOne(ins.Op) {
		if ins.Amt == "" /* mandatory check */ || !parseAmt(ins, false) {
			return ins, false
		}
	} else if opHasAmtMin1.ContainsOne(ins.Op) {
		if ins.Amt == "" /* mandatory check */ || !parseAmt(ins, true) {
			return ins, false
		}
	}

	if opHasTo.ContainsOne(ins.Op) {
		if ins.To == "" {
			// mandatory check
			return ins, false
		}
		if ins.To, ok = IsValidERCAddress(ins.To); !ok {
			return ins, false
		}
	}

	switch ins.Op {
	case "deploy":
		if ins.Lim == "" || ins.Max == "" || ins.Tick == "" {
			// mandatory check
			return ins, false
		}

		if len(ins.Tick) > 42 || !parseDecimals(ins) || !parseMax(ins) || !parseLim(ins) {
			return ins, false
		}

		if ins.MaxV.Cmp(ins.LimV) < 0 {
			return ins, false
		}

		if common.Big0.Rem(ins.MaxV, ins.LimV).Uint64() != 0 {
			return ins, false
		}
		for k, address := range ins.Miners {
			if ins.Miners[k], ok = IsValidERCAddress(address); !ok {
				return ins, false
			}
		}
		if target >= FutureEnableBNForPR67 {
			if ins.Commence != "" && !parseCommence(ins) {
				return ins, false
			}
			for address, amt := range ins.Reserves {
				add, ok := IsValidERCAddress(address)
				if !ok {
					return ins, false
				}
				amtV, err := StringToBigint(amt)
				if !checkBigBetween(amtV, err, true) {
					return ins, false
				}
				ins.ReservesV[add] = amtV
				ins.ReservesSum = new(big.Int).Add(ins.ReservesSum, amtV)
			}
			// check sum
			if ins.ReservesSum != nil {
				if ins.ReservesSum.Cmp(ins.MaxV) > 0 {
					return ins, false
				}
			} else {
				ins.ReservesSum = common.Big0 // default
			}
			for k, address := range ins.Minters {
				if ins.Minters[k], ok = IsValidERCAddress(address); !ok {
					return ins, false
				}
			}
		}

	case "recap":
		if ins.Max == "" /* mandatory check */ || !parseMax(ins) {
			return ins, false
		}
	case "approve":
		if ins.Spender, ok = IsValidERCAddress(ins.Spender); !ok {
			return ins, false
		}
	case "transferFrom":
		if ins.From, ok = IsValidERCAddress(ins.From); !ok {
			return ins, false
		}
	case "mint":
	case "transfer":
	case "burn":
	default:
		return ins, false
	}

	return ins, true
}

func IsValidERCAddress(address string) (string, bool) {
	if address == "" {
		return "", false
	}
	res := common.HexToAddress(address).Hex()
	if strings.EqualFold(res, address) {
		return strings.ToLower(address), true
	}
	return "", false
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

func parseDecimals(ins *helper.BNB48InscriptionVerified) bool {
	var err error

	ins.DecimalsV, err = StringToBigint(ins.Decimals)
	if err != nil || ins.DecimalsV.Uint64() > 18 {
		return false
	}
	return true
}

func parseMax(ins *helper.BNB48InscriptionVerified) bool {
	var err error

	ins.MaxV, err = StringToBigint(ins.Max)
	return checkBigBetween(ins.MaxV, err, true)
}

func parseLim(ins *helper.BNB48InscriptionVerified) bool {
	var err error

	ins.LimV, err = StringToBigint(ins.Lim)
	return checkBigBetween(ins.LimV, err, true)
}

func parseCommence(ins *helper.BNB48InscriptionVerified) bool {
	var err error

	ins.CommenceV, err = StringToBigint(ins.Commence)
	return checkBigBetween(ins.CommenceV, err, true)
}
func checkBigBetween(b *big.Int, err error, checkMin bool) bool {
	if err != nil {
		return false
	}
	if b.Cmp(maxU256) > 0 {
		return false
	}
	if checkMin && b.Uint64() < 1 {
		return false
	}
	return true
}

func parseAmt(ins *helper.BNB48InscriptionVerified, b bool) bool {
	var err error

	ins.AmtV, err = StringToBigint(ins.Amt)
	return checkBigBetween(ins.AmtV, err, b)
}

func Unpack(types []string, data []byte) ([]interface{}, error) {
	var (
		ts   []abi.Type
		args = abi.Arguments{}
	)

	for _, t := range types {
		_t, err := abi.NewType(t, t, nil)
		if err != nil {
			return nil, err
		}

		ts = append(ts, _t)
	}

	for _, t := range ts {
		args = append(args, abi.Argument{Type: t})
	}

	return args.Unpack(data)
}
