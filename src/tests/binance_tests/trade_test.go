package binance_tests

import (
	"A-Matrix/src/binance_v1"
	"testing"
)

func TestOrderFull(t *testing.T) {
	type args struct {
		orderRelation binance_v1.OrderRelation
	}
	tests := []struct {
		name string
		args args
	}{
		{"Test for doing orders", args{binance_v1.OrderRelation{
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
		binance_v1.GetBalance("USDT")
		t.Run(tt.name, func(t *testing.T) {
			binance_v1.OrderFull(&tt.args.orderRelation)
		})
	}
}
