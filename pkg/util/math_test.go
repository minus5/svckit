package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRound(t *testing.T) {
	assert.Equal(t, 123.45, Round(123.45123, 2))
	assert.Equal(t, 123.46, Round(123.455, 2))

	// Slucajevi koji ne daju ocekivane rezultate zbog kojih
	assert.NotEqual(t, 17.32, Round(17.315, 2))
	assert.NotEqual(t, 17.37, Round(17.365, 2))
	assert.NotEqual(t, 17.39, Round(17.385, 2))
}

func TestRoundAwayFromZero(t *testing.T) {
	tst := []float64{
		17.305, 17.315, 17.325, 17.335, 17.345, 17.355, 17.365, 17.375, 17.385, 17.395,
		-17.305, -17.315, -17.325, -17.335, -17.345, -17.355, -17.365, -17.375, -17.385, -17.395,
	}

	exp := []float64{
		17.31, 17.32, 17.33, 17.34, 17.35, 17.36, 17.37, 17.38, 17.39, 17.40,
		-17.31, -17.32, -17.33, -17.34, -17.35, -17.36, -17.37, -17.38, -17.39, -17.40,
	}

	for i := 0; i < len(tst); i++ {
		assert.Equal(t, exp[i], RoundAwayFromZero(tst[i], 2))
	}

	var a, b float64
	a = 23.61
	b = 1.24
	assert.Equal(t, 24.85, RoundAwayFromZero(a+b, 2))

	a = 20.42
	b = 1.07
	assert.Equal(t, 21.49, RoundAwayFromZero(a+b, 2))

	a = 2.64
	b = 0.14
	assert.Equal(t, 2.78, RoundAwayFromZero(a+b, 2))
}
