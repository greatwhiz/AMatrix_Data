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

var commissionRate = 0.001
var threshold = 1.01

func Analyze(symbolBSON bson.D) {
	symbolBytes, _ := bson.MarshalExtJSON(symbolBSON, true, true)
	symbolJSON := string(symbolBytes)
	symbol := gjson.Get(symbolJSON, "symbol").Str
	baseSymbol := gjson.Get(symbolJSON, "base").Str
	askPrice := gjson.Get(symbolJSON, "ticker.ask").Float()
	done := make(chan int, 2)
	for _, relation := range symbolBSON.Map()["arbitrage"].(bson.A) {
		done <- -1
		go doAnalysis(symbol, baseSymbol, relation, askPrice, done)
	}
	close(done)
}

func doAnalysis(symbol string, baseSymbol string, relation interface{}, askPrice float64, done chan int) {
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
	var mediumPrice float64
	mediumBook := GetAPI("depth", map[string]string{"symbol": mediumRelation, "limit": "1"})
	if baseSymbol == gjson.Get(mediumJSON, "base").Str {
		mediumPrice = gjson.Get(mediumBook, "bids.0.0").Float()
		mediumSymbol = gjson.Get(mediumJSON, "quote").Str
		mediumBuySell = false
	} else {
		mediumPrice = gjson.Get(mediumBook, "asks.0.0").Float()
		mediumSymbol = gjson.Get(mediumJSON, "base").Str
		mediumBuySell = true
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
	if estimatedAmount > threshold && estimatedAmount != math.Inf(0) {
		if mediumBuySell {
			log.Println(fmt.Sprintf("%s(%.8f) %s(Buy %.8f) (%.8f): %.4f", symbol, askPrice, mediumRelation, mediumPrice, bidFinalPrice, estimatedAmount))
		} else {
			log.Println(fmt.Sprintf("%s(%.8f) %s(Sell %.8f) (%.8f): %.4f", symbol, askPrice, mediumRelation, mediumPrice, bidFinalPrice, estimatedAmount))
		}
	}
	//if estimatedAmount > threshold {
	//	Decide(symbol, mediumRelation, finalSymbol, mediumBuySell)
	//}
	<-done
}

func Decide(symbol string, mediumRelation string, finalSymbol string, mediumBuySell bool) {
	startBook := GetAPI("depth", map[string]string{"symbol": symbol, "limit": "5"})
	startAsk := gjson.Get(startBook, "asks.0.0").Float()
	// startAskQty := gjson.Get(startBook, "asks.0.1").Float()
	mediumBook := GetAPI("depth", map[string]string{"symbol": mediumRelation, "limit": "1"})
	var mediumPrice float64 //, mediumQty float64
	if mediumBuySell {
		mediumPrice = gjson.Get(mediumBook, "asks.0.0").Float()
		//mediumQty = gjson.Get(mediumBook, "asks.0.1").Float()
	} else {
		mediumPrice = gjson.Get(mediumBook, "bids.0.0").Float()
		//mediumQty = gjson.Get(mediumBook, "bids.0.1").Float()
	}
	finalBook := GetAPI("depth", map[string]string{"symbol": finalSymbol, "limit": "1"})
	finalBid := gjson.Get(finalBook, "bids.0.0").Float()
	//finalAskQty := gjson.Get(finalBook, "bids.0.1").Float()
	estimatedAmount := calculate(startAsk, mediumPrice, finalBid, commissionRate, mediumBuySell)
	println(fmt.Sprintf("%s(%.8f) %s(%.8f) (%.8f): %.4f", symbol, startAsk, mediumRelation, mediumPrice, finalBid, estimatedAmount))
}

func calculate(askPrice float64, mediumPrice float64, bidFinalPrice float64, commissionRate float64, mediumBuySell bool) (estimatedAmount float64) {
	if mediumBuySell {
		estimatedAmount = 10000 / askPrice * (1 - commissionRate) / mediumPrice * (1 - commissionRate) * bidFinalPrice * (1 - commissionRate) / 10000
	} else {
		estimatedAmount = 10000 / askPrice * (1 - commissionRate) * mediumPrice * (1 - commissionRate) * bidFinalPrice * (1 - commissionRate) / 10000
	}
	return estimatedAmount
}
