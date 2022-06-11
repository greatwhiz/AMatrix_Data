package binance_tests

import (
	"A-Matrix/src/binance"
	"testing"
)

func TestGetBalance(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"Test of Account Balance"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			binance.GetBalance("USDT")
		})
	}
}
