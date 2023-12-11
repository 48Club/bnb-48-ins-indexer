package utils

import "github.com/sony/sonyflake"

var sf *sonyflake.Sonyflake

func init() {
	var st = sonyflake.Settings{
		MachineID: func() (uint16, error) {
			return 100, nil
		},
	}

	if sf = sonyflake.NewSonyflake(st); sf == nil {
		panic("sonyflake not created")
	}
}

// GenSnowflakeID generates a next unique ID.
// After the Sonyflake time overflows, NextID returns an errs.
func GenSnowflakeID() (uint64, error) {
	return sf.NextID()
}
