package models

type SymbolData struct {
	Symbol    string
	BaseCoin  string
	QuoteCoin string
	LotSize   float64
}

type ArbitrageRelation struct {
	SymbolData
	FinalSymbol SymbolData
	IsBuyOrSell bool
}

type SymbolWithRelations struct {
	SymbolData
	ArbitrageRelations []ArbitrageRelation
}
