package binance_tests

import (
	"A-Matrix/src/binance"
	"testing"
)

func TestOrderFull(t *testing.T) {
	type args struct {
		symbol     string
		baseSymbol string
	}
	tests := []struct {
		name string
		args args
	}{
		{"Test for doing orders", args{"ETH", "USDT"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			binance.OrderFull(tt.args.symbol, tt.args.baseSymbol)
		})
	}
}
