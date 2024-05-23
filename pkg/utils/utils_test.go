package utils

import (
	"bnb-48-ins-indexer/pkg/helper"
	"context"
	"fmt"
	"github.com/status-im/keycard-go/hexutils"
	"math/big"
	"strings"
	"testing"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/assert"
)

func TestStringJoin(t *testing.T) {
	assert.Equal(t, strings.Join([]string{}, ","), "")
}

func TestMapSet(t *testing.T) {
	m := mapset.NewSet[string]()
	m.Append("deploy", "recap", "mint")
	assert.Equal(t, m.ToSlice(), []string{"deploy", "recap", "mint"})
	assert.Equal(t, m.ContainsOne("deploy"), true)
	m.Clear()
	m.Append([]string{}...)
	assert.Equal(t, m.Cardinality(), 0)
}

func TestUpdateinscriptions(t *testing.T) {

	type inscription struct {
		Id       uint64
		Miners   mapset.Set[string]
		Max      *big.Int
		Lim      *big.Int
		Minted   *big.Int
		Tick     string
		TickHash string
		DeployBy string
	}
	a := struct {
		inscriptions map[string]*inscription // tick-hash : inscription

	}{
		inscriptions: map[string]*inscription{
			"test": {
				Max: big.NewInt(100),
			},
		},
	}
	b, ok := a.inscriptions["test"]
	assert.Equal(t, ok, true)
	b.Max = big.NewInt(200)
	t.Log(a.inscriptions["test"].Max)
}

func TestInputToBNB48Inscription(t *testing.T) {
	// deploy
	input := "0x646174613a2c7b2270223a22626e622d3438222c226f70223a226465706c6f79222c227469636b223a2266616e73222c226d6178223a2233333838323330222c226c696d223a2231222c226d696e657273223a5b22307837326236316336303134333432643931343437306543376143323937356245333435373936633262225d7d"
	for _, v := range []uint64{0, 48_484_848} {
		inss, err := InputToBNB48Inscription(input, v)
		assert.NoError(t, err)
		assert.Equal(t, len(inss), 1)
		rs := inss[0]
		assert.Equal(t, rs.P, "bnb-48")
		assert.Equal(t, rs.Op, "deploy")
		assert.Equal(t, rs.Tick, "fans")
		assert.Equal(t, rs.Max, "3388230")
		assert.Equal(t, rs.Lim, "1")
		assert.Equal(t, rs.Miners, []string{"0x72b61c6014342d914470ec7ac2975be345796c2b"})
	}
}

func TestDecodeData(t *testing.T) {
	ec, err := ethclient.Dial("https://1gwei.48.club")
	assert.NoError(t, err)
	tx, _, err := ec.TransactionByHash(context.Background(), common.HexToHash("0x8eb3c6f9159188a6f34b6db83886680b83f95a87c9c43a6ba18bcff7a601ad34"))
	assert.NoError(t, err)

	inss, err := InputToBNB48Inscription(tx.Data(), 36_244_355)
	assert.NoError(t, err)
	assert.Equal(t, len(inss), 1)
}

func Test_verifyInscription(t *testing.T) {
	getTestMiners := func(number int) []string {
		var res []string
		for i := 0; i < number; i++ {
			res = append(res, "0x72b61c6014342d914470eC7aC2975bE345796c2b")
		}

		return res
	}

	type args struct {
		ins *helper.BNB48Inscription
	}

	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "P invalid",
			args: args{
				ins: &helper.BNB48Inscription{
					P: "bnb48-48bnb48-48bnb48-48bnb48-48bnb48-48bnb48-48bnb48-48bnb48-48bnb48-48bnb48-48",
				},
			},
			want: false,
		},

		{
			name: "Tick invalid",
			args: args{
				ins: &helper.BNB48Inscription{
					P:    "bnb-48",
					Tick: "fans-fans-fans-fans-fans-fans-fans-fans-fans-fans-fans-fans-fans-fans-fans-fans",
				},
			},
			want: false,
		},

		{
			name: "TickHash invalid",
			args: args{
				ins: &helper.BNB48Inscription{
					P:        "bnb-48",
					Tick:     "fans",
					TickHash: "0xd893ca77b3122cb6c480da7f8a12cb82e19542076f5895f21446258dc473a7c2+++",
				},
			},
			want: false,
		},

		{
			name: "To invalid",
			args: args{
				ins: &helper.BNB48Inscription{
					P:        "bnb-48",
					Tick:     "fans",
					TickHash: "0xd893ca77b3122cb6c480da7f8a12cb82e19542076f5895f21446258dc473a7c2",
					To:       "0x72b61c6014342d914470eC7aC2975bE345796c2b+",
				},
			},
			want: false,
		},

		{
			name: "Miners invalid",
			args: args{
				ins: &helper.BNB48Inscription{
					P:        "bnb-48",
					Tick:     "fans",
					TickHash: "0xd893ca77b3122cb6c480da7f8a12cb82e19542076f5895f21446258dc473a7c2",
					To:       "0x72b61c6014342d914470eC7aC2975bE345796c2b",
					Miners:   getTestMiners(100),
				},
			},
			want: false,
		},

		{
			name: "Decimals invalid",
			args: args{
				ins: &helper.BNB48Inscription{
					P:        "bnb-48",
					Tick:     "fans",
					TickHash: "0xd893ca77b3122cb6c480da7f8a12cb82e19542076f5895f21446258dc473a7c2",
					To:       "0x72b61c6014342d914470eC7aC2975bE345796c2b",
					Miners:   getTestMiners(10),
					Decimals: "19",
				},
			},
			want: false,
		},

		{
			name: "Max invalid",
			args: args{
				ins: &helper.BNB48Inscription{
					P:        "bnb-48",
					Tick:     "fans",
					TickHash: "0xd893ca77b3122cb6c480da7f8a12cb82e19542076f5895f21446258dc473a7c2",
					To:       "0x72b61c6014342d914470eC7aC2975bE345796c2b",
					Miners:   getTestMiners(10),
					Decimals: "8",
					// u256 max : 115792089237316195423570985008687907853269984665640564039457584007913129639935
					Max: "115792089237316195423570985008687907853269984665640564039457584007913129639936",
				},
			},
			want: false,
		},

		{
			name: "Lim invalid",
			args: args{
				ins: &helper.BNB48Inscription{
					P:        "bnb-48",
					Tick:     "fans",
					TickHash: "0xd893ca77b3122cb6c480da7f8a12cb82e19542076f5895f21446258dc473a7c2",
					To:       "0x72b61c6014342d914470eC7aC2975bE345796c2b",
					Miners:   getTestMiners(10),
					Decimals: "8",
					// u256 max : 115792089237316195423570985008687907853269984665640564039457584007913129639935
					Max: "115792089237316195423570985008687907853269984665640564039457584007913129639935",
					Lim: "115792089237316195423570985008687907853269984665640564039457584007913129639936",
				},
			},
			want: false,
		},

		{
			name: "Amt invalid",
			args: args{
				ins: &helper.BNB48Inscription{
					P:        "bnb-48",
					Tick:     "fans",
					TickHash: "0xd893ca77b3122cb6c480da7f8a12cb82e19542076f5895f21446258dc473a7c2",
					To:       "0x72b61c6014342d914470eC7aC2975bE345796c2b",
					Miners:   getTestMiners(10),
					Decimals: "8",
					// u256 max : 115792089237316195423570985008687907853269984665640564039457584007913129639935
					Max: "115792089237316195423570985008687907853269984665640564039457584007913129639935",
					Lim: "115792089237316195423570985008687907853269984665640564039457584007913129639935",
					Amt: "115792089237316195423570985008687907853269984665640564039457584007913129639936",
				},
			},
			want: false,
		},

		{
			name: "all right",
			args: args{
				ins: &helper.BNB48Inscription{
					P:        "bnb-48",
					Tick:     "fans",
					TickHash: "0xd893ca77b3122cb6c480da7f8a12cb82e19542076f5895f21446258dc473a7c2",
					To:       "0x72b61c6014342d914470eC7aC2975bE345796c2b",
					Miners:   getTestMiners(10),
					Decimals: "8",
					// u256 max : 115792089237316195423570985008687907853269984665640564039457584007913129639935
					Max: "115792089237316195423570985008687907853269984",
					Lim: "115792089237316195423570985008687907853269984",
					Amt: "115792089237316195423570985008687907853269984",
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ojbV, ok := verifyInscription(tt.args.ins)
			assert.True(t, ok)
			assert.Equalf(t, tt.want, ojbV.BNB48Inscription, "verifyInscription(%v)", tt.args.ins)
		})
	}
}

func Test111(t *testing.T) {
	b := hexutils.HexToBytes("5ef7f1d90000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000017d646174613a6170706c69636174696f6e2f6a736f6e2c5b7b2270223a22626e622d3438222c226f70223a227472616e73666572222c227469636b2d68617368223a22307864383933636137376233313232636236633438306461376638613132636238326531393534323037366635383935663231343436323538646334373361376332222c22746f223a22307833613732623138663934333833356465323662393735663038376332376666613562613565353063222c22616d74223a223130303030303030227d2c7b2270223a22626e622d3438222c226f70223a227472616e73666572222c227469636b2d68617368223a22307864383933636137376233313232636236633438306461376638613132636238326531393534323037366635383935663231343436323538646334373361376332222c22746f223a22307831313361636666306537646263353530343338636535316362653663306235636435366562383861222c22616d74223a223130303030303030227d5d000000")

	r, err := Unpack([]string{"string"}, b[4:])
	assert.NoError(t, err)

	var txData string

	switch r[0].(type) {
	case string:
		txData = r[0].(string)
	}

	fmt.Println(txData)
	datas, err := InputToBNB48Inscription([]byte(txData), 36_244_355)
	assert.NoError(t, err)

	fmt.Println(datas)
}
