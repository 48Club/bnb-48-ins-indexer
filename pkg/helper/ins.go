package helper

import "math/big"

type BNB48Inscription struct {
	P        string            `json:"p"`
	Op       string            `json:"op"`
	Tick     string            `json:"tick"`
	TickHash string            `json:"tick-hash"`
	To       string            `json:"to"`
	Decimals string            `json:"decimals"`
	Max      string            `json:"max"`
	Lim      string            `json:"lim"`
	Miners   []string          `json:"miners"`
	Amt      string            `json:"amt"`
	Spender  string            `json:"spender"`
	From     string            `json:"from"`
	Minters  []string          `json:"minters"`
	Reserves map[string]string `json:"reserves"`
	Commence string            `json:"commence"`
}

type BNB48InscriptionVerified struct {
	*BNB48Inscription
	DecimalsV   *big.Int
	MaxV        *big.Int
	LimV        *big.Int
	AmtV        *big.Int
	ReservesV   map[string]*big.Int
	ReservesSum *big.Int
	CommenceV   *big.Int
}
