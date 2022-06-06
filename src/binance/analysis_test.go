package binance

import "testing"

func TestDecide(t *testing.T) {
	type args struct {
		symbol         string
		mediumRelation string
		finalSymbol    string
		mediumBuySell  bool
	}
	tests := []struct {
		name string
		args args
	}{
		{"Test for Decide()",
			args{"BTCUSDT", "ETHBTC", "ETHUSDT", true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Decide(tt.args.symbol, tt.args.mediumRelation, tt.args.finalSymbol, tt.args.mediumBuySell)
		})
	}
}
