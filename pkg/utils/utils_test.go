package utils

import (
	"bnb-48-ins-indexer/pkg/helper"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInputToBNB48Inscription(t *testing.T) {
	// deploy
	input := "0x646174613a2c7b2270223a22626e622d3438222c226f70223a226465706c6f79222c227469636b223a2266616e73222c226d6178223a2233333838323330222c226c696d223a2231222c226d696e657273223a5b22307837326236316336303134333432643931343437306543376143323937356245333435373936633262225d7d"
	rs, err := InputToBNB48Inscription(input)
	assert.NoError(t, err)
	assert.Equal(t, rs.P, "bnb-48")
	assert.Equal(t, rs.Op, "deploy")
	assert.Equal(t, rs.Tick, "fans")
	assert.Equal(t, rs.Max, "3388230")
	assert.Equal(t, rs.Lim, "1")
	assert.Equal(t, rs.Miners, []string{"0x72b61c6014342d914470ec7ac2975be345796c2b"})
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
			assert.Equalf(t, tt.want, verifyInscription(tt.args.ins), "verifyInscription(%v)", tt.args.ins)
		})
	}
}
