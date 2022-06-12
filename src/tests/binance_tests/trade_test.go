package binance_tests

import (
	"A-Matrix/src/binance"
	"testing"
)

func TestOrderFull(t *testing.T) {
	type args struct {
		orderRelation binance.OrderRelation
	}
	tests := []struct {
		name string
		args args
	}{
		{"Test for doing orders", args{binance.OrderRelation{
			"BUSDUSDT",
			1.0010,
			100,
			1,
			"BTCBUSD",
			29276.69,
			100,
			1,
			true,
			"BTCUSDT",
			29306.32,
			100,
			1,
		}}},
	}
	for _, tt := range tests {
		binance.GetBalance("USDT")
		t.Run(tt.name, func(t *testing.T) {
			binance.OrderFull(tt.args.orderRelation)
		})
	}
}
