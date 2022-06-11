package binance

import (
	"strconv"
	"strings"
)

type OrderRelation struct {
	BaseSymbol     string
	AskPrice       float64
	MediumRelation string
	MediumPrice    float64
	MediumSellBuy  bool
	FinalSymbol    string
	FinalBid       float64
	LotSize        float64
}

func OrderFull(orderRelation OrderRelation) {
	baseBalance := AccountBalance[balanceSymbol]
	baseQty := baseBalance * (1 - commissionRate) * leverage / orderRelation.AskPrice
	precision := len(strings.Split(strings.TrimRight(strconv.FormatFloat(orderRelation.LotSize, 'f', 10, 64), "0"), ".")[1])
	baseOrder := map[string]string{
		"symbol":           orderRelation.BaseSymbol,
		"side":             "BUY",
		"type":             "MARKET",
		"quantity":         strconv.FormatFloat(baseQty, 'f', precision, 64),
		"newOrderRespType": "RESULT",
	}
	result := PostSignedAPI("order/test", baseOrder)
	println(result)
}
