package binance_tests

import (
	"A-Matrix/src/binance"
	"testing"
)

func TestSubscribeMarket(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"Test SubscribeMarket"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			binance.SubscribeMarket()
		})
	}
}