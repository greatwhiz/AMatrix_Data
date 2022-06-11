package binance

import "strconv"

type OrderRelation struct {
	BaseSymbol     string
	AskPrice       float64
	MediumRelation string
	MediumPrice    float64
	MediumSellBuy  bool
	FinalSymbol    string
	FinalBid       float64
}

func OrderFull(orderRelation OrderRelation) {
	baseBalance := AccountBalance[balanceSymbol]
	baseQty := baseBalance * (1 - commissionRate) * leverage / orderRelation.AskPrice
	baseOrder := map[string]string{
		"symbol":           orderRelation.BaseSymbol,
		"side":             "BUY",
		"type":             "MARKET",
		"quantity":         strconv.FormatFloat(baseQty, 'f', 2, 64),
		"newOrderRespType": "RESULT",
	}
	result := PostSignedAPI("order/test", baseOrder)
	println(result)
}
