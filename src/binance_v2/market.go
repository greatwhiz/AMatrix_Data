package binance_v1

import (
	"A-Matrix/src/db"
	"go.mongodb.org/mongo-driver/bson"
	"log"
)

func SubscribeSymbols() {
	var symbols []SymbolWithRelations
	mongoDB := db.GetMongoDB()
	symbolCollection := mongoDB.GetCollection("symbols")
	filter := bson.M{
		"exchange": "binance_v1",
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
		symbolMap := symbolObject.Map()
		symbolRelation := SymbolWithRelations{}
		symbolRelation.Symbol = symbolMap["symbol"].(string)
		symbolRelation.BaseCoin = symbolMap["base"].(string)
		symbolRelation.QuoteCoin = symbolMap["quote"].(string)
		for _, symbolArbitrage := range symbolMap["arbitrage"].(bson.A) {
			symbolArbitrageMap := symbolArbitrage.(bson.D).Map()
			symbolRelation.Relations = append(symbolRelation.Relations, map[string]string{"base": symbolArbitrageMap["base"].(string), "quote": symbolArbitrageMap["quote"].(string), "symbol": symbolArbitrageMap["symbol"].(string)})
		}
		symbols = append(symbols, symbolRelation)
	}
}
