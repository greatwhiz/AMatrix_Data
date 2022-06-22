package binance_v2

import (
	"A-Matrix/src/models"
	"fmt"
	"github.com/tidwall/gjson"
	"log"
	"math"
)

func RunAnalysis(symbolWithRelations []models.SymbolWithRelations) {
	//var i int
	log.Println("Analysis started.")
	for {
		AnalyzeBatch(symbolWithRelations)
		//i += 1
		//println(i)
	}
}

func AnalyzeBatch(symbolWithRelations []models.SymbolWithRelations) {
	books := GetAPI("ticker/bookTicker", nil)
	booksArray := gjson.Parse(books).Array()
	symbolTickers := make(map[string]models.SymbolTicker)
	for _, book := range booksArray {
		askPrice := gjson.Get(book.Raw, "askPrice").Float()
		askQuantity := gjson.Get(book.Raw, "askQty").Float()
		bidPrice := gjson.Get(book.Raw, "bidPrice").Float()
		bidQuantity := gjson.Get(book.Raw, "bidQty").Float()

		if askPrice != 0 && askQuantity != 0 && bidPrice != 0 && bidQuantity != 0 {
			symbolTicker := models.SymbolTicker{
				SymbolBase: models.SymbolBase{
					Symbol:    gjson.Get(book.Raw, "symbol").Str,
					BaseCoin:  gjson.Get(book.Raw, "base").Str,
					QuoteCoin: gjson.Get(book.Raw, "quote").Str,
				},
				AskPrice:    askPrice,
				AskQuantity: askQuantity,
				BidPrice:    bidPrice,
				BidQuantity: bidQuantity,
			}
			symbolTickers[symbolTicker.Symbol] = symbolTicker
		}
	}
	for _, sr := range symbolWithRelations {
		for _, relation := range sr.ArbitrageRelations {
			go doAnalysis(sr, relation, symbolTickers)
		}
	}
}

func doAnalysis(symbolRelation models.SymbolWithRelations, relation models.ArbitrageRelation, symbolTickers map[string]models.SymbolTicker) {
	//get base ask price
	book := symbolTickers[symbolRelation.Symbol]
	mediumBook := symbolTickers[relation.Symbol]
	finalBook := symbolTickers[relation.FinalSymbol.Symbol]

	if book.Symbol == "" || mediumBook.Symbol == "" || finalBook.Symbol == "" {
		return
	}

	askPrice := book.AskPrice
	baseAskQty := book.AskQuantity

	var mediumPrice, mediumQty float64
	if relation.IsBuyOrSell {
		mediumPrice = mediumBook.AskPrice
		mediumQty = mediumBook.AskQuantity
	} else {
		mediumPrice = mediumBook.BidPrice
		mediumQty = mediumBook.BidQuantity
	}

	bidFinalPrice := finalBook.BidPrice
	bidFinalQty := finalBook.BidQuantity
	estimatedAmount := calculate(askPrice, mediumPrice, bidFinalPrice, commissionRate, relation.IsBuyOrSell)

	if estimatedAmount > arbitrageThreshold && estimatedAmount != math.Inf(0) {
		if relation.IsBuyOrSell {
			log.Println(fmt.Sprintf("%s(%.8f) %s(Buy %.8f) %s(%.8f): %.4f", symbolRelation.Symbol, askPrice, relation.Symbol, mediumPrice, relation.FinalSymbol.Symbol, bidFinalPrice, estimatedAmount))
		} else {
			log.Println(fmt.Sprintf("%s(%.8f) %s(Sell %.8f) %s(%.8f): %.4f", symbolRelation.Symbol, askPrice, relation.Symbol, mediumPrice, relation.FinalSymbol.Symbol, bidFinalPrice, estimatedAmount))
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
