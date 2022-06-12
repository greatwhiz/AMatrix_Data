package binance

import (
	"A-Matrix/src/db"
	"fmt"
	"github.com/tidwall/gjson"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"math"
)

func Analyze(symbolBSON bson.D) {
	symbolBytes, _ := bson.MarshalExtJSON(symbolBSON, true, true)
	symbolJSON := string(symbolBytes)
	symbol := gjson.Get(symbolJSON, "symbol").Str
	baseSymbol := gjson.Get(symbolJSON, "base").Str
	askPrice := gjson.Get(symbolJSON, "ticker.ask").Float()
	done := make(chan int, 1)
	for _, relation := range symbolBSON.Map()["arbitrage"].(bson.A) {
		if blackList[relation.(string)] {
			continue
		}
		done <- -1
		go doAnalysis(symbol, baseSymbol, symbolJSON, relation, askPrice, done)
	}
	close(done)
}

func doAnalysis(symbol string, baseSymbol string, symbolJSON string, relation interface{}, askPrice float64, done chan int) {
	mongoDB := db.GetMongoDB()
	symbolCollection := mongoDB.GetCollection("symbols")
	filter := bson.M{
		"exchange": "binance",
		"symbol":   relation,
	}

	var medium bson.D
	// check for errors in the finding
	if err := symbolCollection.FindOne(mongoDB.Ctx, filter).Decode(&medium); err != nil {
		mongoDB.Close()
		if err == mongo.ErrNoDocuments {
			<-done
			return
		} else {
			println("arbitrate medium: ")
			panic(err)
		}
	}

	if medium.Map()["ticker"] == nil {
		mongoDB.Close()
		<-done
		return
	}

	mediumBytes, _ := bson.MarshalExtJSON(medium, true, true)
	mediumJSON := string(mediumBytes)
	mediumRelation := gjson.Get(mediumJSON, "symbol").Str
	var mediumBuySell bool // true = Buy, false = Sell
	var mediumSymbol string
	var mediumPrice, mediumQty float64
	mediumBook := GetAPI("depth", map[string]string{"symbol": mediumRelation, "limit": "1"})
	if baseSymbol == gjson.Get(mediumJSON, "base").Str {
		mediumPrice = gjson.Get(mediumBook, "bids.0.0").Float()
		mediumQty = gjson.Get(mediumBook, "bids.0.1").Float()
		mediumSymbol = gjson.Get(mediumJSON, "quote").Str
		mediumBuySell = false // sell side
	} else {
		mediumPrice = gjson.Get(mediumBook, "asks.0.0").Float()
		mediumQty = gjson.Get(mediumBook, "asks.0.1").Float()
		mediumSymbol = gjson.Get(mediumJSON, "base").Str
		mediumBuySell = true // buy side
	}

	if mediumPrice == 0 {
		log.Println(fmt.Sprintf("%s unable get book: %s", mediumRelation, mediumBook))
		<-done
		return
	}

	var final bson.D
	filterFinal := bson.M{
		"exchange": "binance",
		"base":     mediumSymbol,
		"quote":    "USDT",
	}

	if err := symbolCollection.FindOne(mongoDB.Ctx, filterFinal).Decode(&final); err != nil {
		mongoDB.Close()
		if err == mongo.ErrNoDocuments {
			<-done
			return
		} else {
			println("arbit final: ")
			panic(err)
		}
	}
	mongoDB.Close()
	finalBytes, _ := bson.MarshalExtJSON(final, true, true)
	finalJSON := string(finalBytes)
	finalSymbol := gjson.Get(finalJSON, "symbol").Str
	finalBook := GetAPI("depth", map[string]string{"symbol": finalSymbol, "limit": "1"})
	bidFinalPrice := gjson.Get(finalBook, "bids.0.0").Float()
	estimatedAmount := calculate(askPrice, mediumPrice, bidFinalPrice, commissionRate, mediumBuySell)
	if estimatedAmount > arbitrageThreshold && estimatedAmount != math.Inf(0) {
		if mediumBuySell {
			log.Println(fmt.Sprintf("%s(%.8f) %s(Buy %.8f) (%.8f): %.4f", symbol, askPrice, mediumRelation, mediumPrice, bidFinalPrice, estimatedAmount))
		} else {
			log.Println(fmt.Sprintf("%s(%.8f) %s(Sell %.8f) (%.8f): %.4f", symbol, askPrice, mediumRelation, mediumPrice, bidFinalPrice, estimatedAmount))
		}
		baseAskQty := gjson.Get(symbolJSON, "ticker.ask_qty").Float()
		baseLotSize := gjson.Get(symbolJSON, "filters.#(filterType==\"LOT_SIZE\").minQty").Float()
		mediumLotSize := gjson.Get(mediumJSON, "filters.#(filterType==\"LOT_SIZE\").minQty").Float()
		bidFinalQty := gjson.Get(finalBook, "bids.0.1").Float()
		finalLotSize := gjson.Get(finalJSON, "filters.#(filterType==\"LOT_SIZE\").minQty").Float()
		orderRelation := OrderRelation{
			symbol,
			askPrice,
			baseAskQty,
			baseLotSize,
			mediumRelation,
			mediumPrice,
			mediumQty,
			mediumLotSize,
			mediumBuySell,
			finalSymbol,
			bidFinalPrice,
			bidFinalQty,
			finalLotSize,
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
