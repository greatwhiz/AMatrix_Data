package binance_v1

type SymbolWithRelations struct {
	Symbol    string
	BaseCoin  string
	QuoteCoin string
	Relations []interface{}
}
