package binance_tests

import (
	"A-Matrix/src/binance_v1"
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
			binance_v1.UpdateArbitrageRelation()
		})
	}
}
