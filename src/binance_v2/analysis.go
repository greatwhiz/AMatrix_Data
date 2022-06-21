package binance_v2

import (
	"A-Matrix/src/models"
	"fmt"
	"github.com/tidwall/gjson"
	"log"
	"math"
)

func RunAnalysis(symbolWithRelations []models.SymbolWithRelations) {
	for _, relations := range symbolWithRelations {
		Analyze(relations)
	}
}

func Analyze(symbolRelation models.SymbolWithRelations) {
	done := make(chan int, analyzingConcurrency)
	for _, relation := range symbolRelation.ArbitrageRelations {
		if blackList[relation.Symbol] {
			continue
		}
		done <- -1
		go doAnalysis(symbolRelation, relation, done)
	}
	close(done)
}

func doAnalysis(symbolRelation models.SymbolWithRelations, relation models.ArbitrageRelation, done chan int) {
	//get base ask price
	book := GetAPI("depth", map[string]string{"symbol": symbolRelation.Symbol, "limit": "1"})
	askPrice := gjson.Get(book, "asks.0.0").Float()
	baseAskQty := gjson.Get(book, "asks.0.1").Float()
	var mediumPrice, mediumQty float64
	mediumBook := GetAPI("depth", map[string]string{"symbol": relation.Symbol, "limit": "1"})
	if relation.IsBuyOrSell {
		mediumPrice = gjson.Get(mediumBook, "asks.0.0").Float()
		mediumQty = gjson.Get(mediumBook, "asks.0.1").Float()
	} else {
		mediumPrice = gjson.Get(mediumBook, "bids.0.0").Float()
		mediumQty = gjson.Get(mediumBook, "bids.0.1").Float()
	}

	if mediumPrice == 0 {
		log.Println(fmt.Sprintf("%s unable get book: %s", relation.Symbol, mediumBook))
		<-done
		return
	}
	finalBook := GetAPI("depth", map[string]string{"symbol": relation.FinalSymbol.Symbol, "limit": "1"})
	bidFinalPrice := gjson.Get(finalBook, "bids.0.0").Float()
	bidFinalQty := gjson.Get(finalBook, "bids.0.1").Float()
	estimatedAmount := calculate(askPrice, mediumPrice, bidFinalPrice, commissionRate, relation.IsBuyOrSell)
	if estimatedAmount > arbitrageThreshold && estimatedAmount != math.Inf(0) {
		if relation.IsBuyOrSell {
			log.Println(fmt.Sprintf("%s(%.8f) %s(Buy %.8f) (%.8f): %.4f", symbolRelation.Symbol, askPrice, relation.Symbol, mediumPrice, bidFinalPrice, estimatedAmount))
		} else {
			log.Println(fmt.Sprintf("%s(%.8f) %s(Sell %.8f) (%.8f): %.4f", symbolRelation.Symbol, askPrice, relation.Symbol, mediumPrice, bidFinalPrice, estimatedAmount))
		}
		orderRelation := models.OrderRelation{
			BaseSymbol:     symbolRelation.Symbol,
			BaseAskPrice:   askPrice,
			BaseAskQty:     baseAskQty,
			BaseLotSize:    symbolRelation.LotSize,
			MediumRelation: relation.Symbol,
			MediumPrice:    mediumPrice,
			MediumQty:      mediumQty,
			MediumLotSize:  relation.LotSize,
			MediumSellBuy:  relation.IsBuyOrSell,
			FinalSymbol:    relation.FinalSymbol.Symbol,
			FinalBid:       bidFinalPrice,
			FinalQty:       bidFinalQty,
			FinalLotSize:   relation.FinalSymbol.LotSize,
		}
		OrderFull(&orderRelation)
		//log.Println(orderRelation)
	}
	<-done
}

func calculate(askPrice float64, mediumPrice float64, bidFinalPrice float64, commissionRate float64, mediumBuySell bool) (estimatedAmount float64) {
	// buy sell
	if mediumBuySell {
		estimatedAmount = 10000 / askPrice * (1 - commissionRate) / mediumPrice * (1 - commissionRate) * bidFinalPrice * (1 - commissionRate) / 10000
	} else { //sell side
		estimatedAmount = 10000 / askPrice * (1 - commissionRate) * mediumPrice * (1 - commissionRate) * bidFinalPrice * (1 - commissionRate) / 10000
	}
	return estimatedAmount
}
