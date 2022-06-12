package binance

import (
	"github.com/tidwall/gjson"
	"log"
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

func OrderFull(orderRelation *OrderRelation) {
	tradingAmount, baseQty := calculateOrderQuantity(orderRelation)
	// skip when too small amount to trade
	if tradingAmount < tradingAmountThreshold {
		return
	}
	baseQty = getPreciseQuantity(baseQty, orderRelation.BaseLotSize)
	var executedQty, mediumExecutedQty float64
	executedQty = marketOrder(orderRelation.BaseSymbol, "BUY", baseQty)
	mediumQty := getPreciseQuantity(executedQty, orderRelation.MediumLotSize)
	if orderRelation.MediumSellBuy {
		mediumExecutedQty = marketOrder(orderRelation.MediumRelation, "BUY", mediumQty)
	} else {
		mediumExecutedQty = marketOrder(orderRelation.MediumRelation, "SELL", mediumQty)
	}
	finalQty := getPreciseQuantity(mediumExecutedQty, orderRelation.FinalLotSize)
	finalExecutedQty := marketOrder(orderRelation.MediumRelation, "SELL", finalQty)
	println("Arbitrage completed: %.2f -> %.2f, %.2f", tradingAmount, finalExecutedQty, tradingAmount/finalExecutedQty)
}

func marketOrder(symbol string, side string, quantity float64) float64 {
	order := map[string]string{
		"symbol":           symbol,
		"side":             side,
		"type":             "MARKET",
		"quantity":         strconv.FormatFloat(quantity, 'f', -1, 64),
		"newOrderRespType": "RESULT",
	}
	log.Print("Trade request: ", order)
	result := PostSignedAPI("order/test", order)
	executedQty := gjson.Get(result, "executedQty").Float()
	log.Print("Trade result: ", result)
	return executedQty
}

func calculateOrderQuantity(orderRelation *OrderRelation) (tradingAmount float64, baseQty float64) {
	baseBalance := AccountBalance[fundamentalSymbol]
	tradingAmount = baseBalance * (1 - commissionRate) * leverage
	var mediumQty, finalQty float64
	isLessTrading := false // if trading amount is less than order book
	baseQty = tradingAmount / orderRelation.BaseAskPrice
	// compare first quantity to order book
	if baseQty > orderRelation.BaseAskQty {
		baseQty = orderRelation.BaseAskQty
		isLessTrading = true
	}
	if orderRelation.MediumSellBuy {
		mediumQty = baseQty * (1 - commissionRate) / orderRelation.MediumPrice
	} else {
		mediumQty = baseQty * (1 - commissionRate) * orderRelation.MediumPrice
	}
	// compare medium quantity to order book
	if mediumQty > orderRelation.MediumQty {
		mediumQty = orderRelation.MediumQty
		isLessTrading = true
	}
	finalQty = mediumQty
	// compare final quantity to order book
	if finalQty > orderRelation.FinalQty {
		finalQty = orderRelation.FinalQty
		isLessTrading = true
	}

	finalTradingAmount := finalQty * (1 - commissionRate) * orderRelation.FinalBid
	finalTradingAmount = finalTradingAmount / math.Pow(1-commissionRate, 3)

	if isLessTrading {
		baseQty = finalTradingAmount * (1 - commissionRate) / orderRelation.BaseAskPrice
		tradingAmount = finalTradingAmount
	}
	return tradingAmount, baseQty
}

func getPreciseQuantity(quantity float64, lotSize float64) float64 {
	precision := float64(len(strings.Split(strings.TrimRight(strconv.FormatFloat(lotSize, 'f', 15, 64), "0"), ".")[1]))
	return math.Floor(quantity*math.Pow(10, precision)) / math.Pow(10, precision)
}
