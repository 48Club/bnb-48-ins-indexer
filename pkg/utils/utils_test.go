package utils

import (
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
