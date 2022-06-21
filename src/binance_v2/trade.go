package binance_v2

import (
	"A-Matrix/src/models"
	"fmt"
	"github.com/tidwall/gjson"
	"log"
	"math"
	"strconv"
	"strings"
)

func OrderFull(orderRelation *models.OrderRelation) {
	// when hit num limit
	if tradeNumLimit != -1 && tradeCount >= tradeNumLimit {
		log.Println("Trade Num Limited reached.")
		return
	}
	startBalance := AccountBalance[fundamentalSymbol]
	tradingAmount, baseQty := calculateOrderQuantity(orderRelation)
	// skip when too small amount to trade
	if tradingAmount < tradingAmountThreshold {
		return
	}
	baseQty = getPreciseQuantity(baseQty, orderRelation.BaseLotSize)
	var executedQty, mediumQty, mediumExecutedQty float64
	executedQty = marketOrder(orderRelation.BaseSymbol, "BUY", baseQty)
	action := ""
	if orderRelation.MediumSellBuy {
		action = "BUY"
		mediumQty = executedQty * (1 - commissionRate) / orderRelation.MediumPrice
	} else {
		action = "SELL"
		mediumQty = executedQty * (1 - commissionRate) * orderRelation.MediumPrice
	}

	mediumQty = getPreciseQuantity(mediumQty, orderRelation.MediumLotSize)
	mediumExecutedQty = marketOrder(orderRelation.MediumRelation, action, mediumQty)
	finalQty := getPreciseQuantity(mediumExecutedQty*(1-commissionRate), orderRelation.FinalLotSize)
	finalExecutedQty := marketOrder(orderRelation.FinalSymbol, "SELL", finalQty)
	GetBalance(fundamentalSymbol)
	endBalance := AccountBalance[fundamentalSymbol]
	log.Println(fmt.Sprintf("Arbitrage completed: (%.4f) %.8f -> %.8f, %.2f", finalExecutedQty, startBalance, endBalance, endBalance/startBalance))
	tradeCount += 1 // Count one trade
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
	result := PostSignedAPI("order", order)
	executedQty := gjson.Get(result, "executedQty").Float()
	log.Print("Trade result: ", result)
	return executedQty
}

func calculateOrderQuantity(orderRelation *models.OrderRelation) (tradingAmount float64, baseQty float64) {
	baseBalance := AccountBalance[fundamentalSymbol]
	tradingAmount = baseBalance * leverage
	var mediumQty, finalQty float64
	isLessTrading := false // if trading amount is less than order book
	baseQty = tradingAmount / orderRelation.BaseAskPrice
	// compare first quantity to order book
	if baseQty > orderRelation.BaseAskQty {
		baseQty = orderRelation.BaseAskQty
		isLessTrading = true
	}
	mediumQty = baseQty * (1 - commissionRate) // add commission for the init trading
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
