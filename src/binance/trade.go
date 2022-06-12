package binance

import (
	"github.com/tidwall/gjson"
	"math"
	"strconv"
	"strings"
)

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

func OrderFull(orderRelation OrderRelation) {
	baseQty := calculateOrderQuantity(orderRelation)
	var executedQty, mediumExecutedQty float64
	executedQty = marketOrder(orderRelation.BaseSymbol, "BUY", baseQty)
	if orderRelation.MediumSellBuy {
		mediumExecutedQty = marketOrder(orderRelation.MediumRelation, "BUY", executedQty)
	} else {
		mediumExecutedQty = marketOrder(orderRelation.MediumRelation, "SELL", executedQty)
	}
	finalExecutedQty := marketOrder(orderRelation.MediumRelation, "SELL", mediumExecutedQty)
	println(finalExecutedQty)
}

func marketOrder(symbol string, side string, quantity float64) float64 {
	baseOrder := map[string]string{
		"symbol":           symbol,
		"side":             side,
		"type":             "MARKET",
		"quantity":         strconv.FormatFloat(quantity, 'f', -1, 64),
		"newOrderRespType": "RESULT",
	}
	result := PostSignedAPI("order/test", baseOrder)
	executedQty := gjson.Get(result, "executedQty").Float()
	return executedQty
}

func calculateOrderQuantity(orderRelation OrderRelation) (baseQty float64) {
	baseBalance := AccountBalance[balanceSymbol]
	tradingAmount := baseBalance * (1 - commissionRate) * leverage
	var mediumQty, finalQty float64
	baseQty = baseBalance * (1 - commissionRate) * leverage / orderRelation.BaseAskPrice
	baseQty = getPreciseQuantity(baseQty, orderRelation.BaseLotSize)
	if baseQty > orderRelation.BaseAskQty {
		baseQty = orderRelation.BaseAskQty
	}
	if orderRelation.MediumSellBuy {
		mediumQty = baseQty * (1 - commissionRate) / orderRelation.MediumPrice
		mediumQty = getPreciseQuantity(mediumQty, orderRelation.MediumLotSize)
	} else {
		mediumQty = baseQty * (1 - commissionRate) * orderRelation.MediumPrice
		mediumQty = getPreciseQuantity(mediumQty, orderRelation.MediumLotSize)
	}
	if mediumQty > orderRelation.MediumQty {
		mediumQty = orderRelation.MediumQty
	}
	finalQty = mediumQty * (1 - commissionRate) * orderRelation.FinalBid
	finalQty = getPreciseQuantity(finalQty, orderRelation.FinalLotSize)
	if finalQty > orderRelation.FinalQty {
		finalQty = orderRelation.FinalQty
	}
	finalTradedAmount := finalQty * orderRelation.FinalBid
	if finalTradedAmount < tradingAmount {
		baseQty = finalTradedAmount / orderRelation.BaseAskPrice
		baseQty = getPreciseQuantity(baseQty, orderRelation.BaseLotSize)
	}
	return baseQty
}

func getPreciseQuantity(quantity float64, lotSize float64) float64 {
	precision := float64(len(strings.Split(strings.TrimRight(strconv.FormatFloat(lotSize, 'f', 15, 64), "0"), ".")[1]))
	return math.Floor(quantity*math.Pow(10, precision)) / math.Pow(10, precision)
}
