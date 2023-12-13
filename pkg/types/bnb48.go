package types

type BNB48Inscription struct {
	P        string   `json:"p"`
	Op       string   `json:"op"`
	Tick     string   `json:"tick"`
	TickHash string   `json:"tick_hash"`
	Decimal  uint8    `json:"decimal"`
	Max      string   `json:"max"`
	Lim      string   `json:"lim"`
	Miners   []string `json:"miners"`
	Amt      string   `json:"amt"`
}
