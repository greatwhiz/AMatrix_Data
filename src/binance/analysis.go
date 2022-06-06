package binance

import (
	"A-Matrix/src/db"
	"fmt"
	"github.com/tidwall/gjson"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var commissionRate float64 = 0.001
var threshold = 1.01

func Analyze(symbolBSON bson.D) {
	symbolBytes, _ := bson.MarshalExtJSON(symbolBSON, true, true)
	symbolJSON := string(symbolBytes)
	symbol := gjson.Get(symbolJSON, "symbol").Str
	baseSymbol := gjson.Get(symbolJSON, "base").Str
	askPrice := gjson.Get(symbolJSON, "ticker.ask").Float()
	for _, relation := range symbolBSON.Map()["arbitrage"].(bson.A) {
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
				continue
			} else {
				println("arbitrate medium: ")
				panic(err)
			}
		}

		if medium.Map()["ticker"] == nil {
			mongoDB.Close()
			continue
		}

		mediumBytes, _ := bson.MarshalExtJSON(medium, true, true)
		mediumJSON := string(mediumBytes)
		mediumRelation := gjson.Get(mediumJSON, "symbol").Str
		var mediumBuySell bool // true = Buy, false = Sell
		var mediumSymbol string
		var mediumPrice float64
		if baseSymbol == gjson.Get(mediumJSON, "base").Str {
			mediumPrice = gjson.Get(mediumJSON, "ticker.bid").Float() //get the bid to see how much I can sell
			mediumSymbol = gjson.Get(mediumJSON, "quote").Str
			mediumBuySell = false
		} else {
			mediumPrice = gjson.Get(mediumJSON, "ticker.ask").Float() // get the ask to see how much I should buy
			mediumSymbol = gjson.Get(mediumJSON, "base").Str
			mediumBuySell = true
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
				continue
			} else {
				println("arbit final: ")
				panic(err)
			}
		}
		mongoDB.Close()
		finalBytes, _ := bson.MarshalExtJSON(final, true, true)
		finalJSON := string(finalBytes)
		bidFinalPrice := gjson.Get(finalJSON, "ticker.bid").Float()
		finalSymbol := gjson.Get(finalJSON, "symbol").Str
		estimatedAmount := calculate(askPrice, mediumPrice, bidFinalPrice, commissionRate, mediumBuySell)
		if estimatedAmount > threshold {
			if mediumBuySell {
				println(fmt.Sprintf("%s(%.8f) %s(Buy %.8f) (%.8f): %.4f", baseSymbol, askPrice, mediumRelation, mediumPrice, bidFinalPrice, estimatedAmount))
			} else {
				println(fmt.Sprintf("%s(%.8f) %s(Sell %.8f) (%.8f): %.4f", baseSymbol, askPrice, mediumRelation, mediumPrice, bidFinalPrice, estimatedAmount))
			}
			Decide(symbol, mediumRelation, finalSymbol, mediumBuySell)
		}
	}
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
