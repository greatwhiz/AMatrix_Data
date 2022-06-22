package binance_v2_tests

import (
	"A-Matrix/src/binance_v2"
	"testing"
)

func TestSubscribeSymbols(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"Test of Subscription"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			binance_v2.SubscribeSymbols()
		})
	}
}
