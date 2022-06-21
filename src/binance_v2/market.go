package binance_v2

import (
	"A-Matrix/src/db"
	"A-Matrix/src/models"
	"github.com/tidwall/gjson"
	"go.mongodb.org/mongo-driver/bson"
	"log"
)

func SubscribeSymbols() {
	var symbols []models.SymbolWithRelations
	mongoDB := db.GetMongoDB()
	defer mongoDB.Close()
	symbolCollection := mongoDB.GetCollection("symbols")
	filter := bson.M{
		"exchange": "binance",
		"arbitrage": bson.M{
			"$exists": true,
			"$ne":     nil,
		},
	}
	cur, _ := symbolCollection.Find(mongoDB.Ctx, filter)

	for cur.Next(mongoDB.Ctx) {
		var symbolObject bson.D
		err := cur.Decode(&symbolObject)
		if err != nil {
			log.Fatal(err)
		}
		symbolBytes, _ := bson.MarshalExtJSON(symbolObject, true, true)
		symbolJSON := string(symbolBytes)
		symbolRelation := models.SymbolWithRelations{}
		symbolRelation.Symbol = gjson.Get(symbolJSON, "symbol").Str
		symbolRelation.BaseCoin = gjson.Get(symbolJSON, "base").Str
		symbolRelation.QuoteCoin = gjson.Get(symbolJSON, "quote").Str
		symbolRelation.LotSize = gjson.Get(symbolJSON, "filters.#(filterType==\"LOT_SIZE\").minQty").Float()
		for _, symbolArbitrage := range symbolObject.Map()["arbitrage"].(bson.A) {
			symbolArbitrageBytes, _ := bson.MarshalExtJSON(symbolArbitrage, true, true)
			symbolArbitrageJSON := string(symbolArbitrageBytes)
			isBuyOrSell := gjson.Get(symbolArbitrageJSON, "buy_or_sell").Bool()
			var finalSymbolBaseCoin string
			if isBuyOrSell {
				finalSymbolBaseCoin = gjson.Get(symbolArbitrageJSON, "base").Str
			} else {
				finalSymbolBaseCoin = gjson.Get(symbolArbitrageJSON, "quote").Str
			}
			symbolRelation.ArbitrageRelations = append(symbolRelation.ArbitrageRelations, models.ArbitrageRelation{
				SymbolData: models.SymbolData{
					BaseCoin:  gjson.Get(symbolArbitrageJSON, "base").Str,
					QuoteCoin: gjson.Get(symbolArbitrageJSON, "quote").Str,
					Symbol:    gjson.Get(symbolArbitrageJSON, "symbol").Str,
					LotSize:   gjson.Get(symbolArbitrageJSON, "filters.#(filterType==\"LOT_SIZE\").minQty").Float(),
				},
				FinalSymbol: models.SymbolData{
					BaseCoin:  finalSymbolBaseCoin,
					QuoteCoin: fundamentalSymbol,
					Symbol:    finalSymbolBaseCoin + fundamentalSymbol,
					LotSize:   gjson.Get(symbolArbitrageJSON, "final_filters.#(filterType==\"LOT_SIZE\").minQty").Float(),
				},
				IsBuyOrSell: isBuyOrSell,
			})
		}
		symbols = append(symbols, symbolRelation)
	}

	RunAnalysis(symbols)
}
