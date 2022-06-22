package binance_v1_tests

import (
	"A-Matrix/src/binance_v1"
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
			binance_v1.GetBalance("USDT")
		})
	}
}
