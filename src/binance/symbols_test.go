package binance

import "testing"

func TestUpdateArbitrageRelation(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"Test of UpdateArbitrageRelation"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			UpdateArbitrageRelation()
		})
	}
}
