package binance_v1

import "testing"

func TestSubscribeSymbols(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"Test of Subscription"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SubscribeSymbols()
		})
	}
}
