package binance_tests

import (
	"A-Matrix/src/binance"
	"testing"
)

func TestUpdateArbitrageRelation(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"Test of UpdateArbitrageRelation"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			binance.UpdateArbitrageRelation()
		})
	}
}