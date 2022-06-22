package models

type SymbolBase struct {
	Symbol    string
	BaseCoin  string
	QuoteCoin string
}

type SymbolTicker struct {
	SymbolBase
	AskPrice    float64
	AskQuantity float64
	BidPrice    float64
	BidQuantity float64
}

type SymbolData struct {
	SymbolBase
	LotSize float64
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
