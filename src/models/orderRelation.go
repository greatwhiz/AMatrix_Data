package models

type OrderRelation struct {
	BaseSymbol     string
	BaseAskPrice   float64
	BaseAskQty     float64
	BaseLotSize    float64
	MediumRelation string
	MediumPrice    float64
	MediumQty      float64
	MediumLotSize  float64
	MediumSellBuy  bool
	FinalSymbol    string
	FinalBid       float64
	FinalQty       float64
	FinalLotSize   float64
}
