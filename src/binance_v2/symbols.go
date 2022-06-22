package binance_v2

import (
	"A-Matrix/src/db"
	"context"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
)

func UpdateSymbols() {
	mongoDB := db.GetMongoDB()
	defer mongoDB.Close()
	// clean database
	symbolCollection := mongoDB.GetCollection(defaultCollection)
	if err := symbolCollection.Drop(mongoDB.Ctx); err != nil {
		log.Fatal(err)
	}
	symbolCollection = mongoDB.GetCollection(defaultCollection) // automatically created database
	content := GetAPI("exchangeInfo", nil)
	symbols := gjson.Get(content, "symbols")
	var symbolBSONs []interface{}
	symbols.ForEach(func(key, symbol gjson.Result) bool {
		var result bson.M
		value := symbol.Map()

		var filters bson.A
		err := bson.UnmarshalExtJSON([]byte(symbol.Map()["filters"].Raw), true, &filters)
		if err != nil {
			log.Println("get filter: ", err)
		}
		symbolBSON := bson.D{{"symbol", value["symbol"].String()}, {"exchange", "binance"}, {"base", value["baseAsset"].String()}, {"quote", value["quoteAsset"].String()}, {"filters", filters}}
		err = symbolCollection.FindOne(context.TODO(), bson.D{{"symbol", value["symbol"].String()}, {"exchange", "binance"}}).Decode(&result)
		if err != nil {
			// ErrNoDocuments means that the filter did not match any documents in
			//the collection.
			if err == mongo.ErrNoDocuments {
				symbolBSONs = append(symbolBSONs, symbolBSON)
			}
			//log.Fatal(err)
		}
		return true // keep iterating
	})

	if len(symbolBSONs) > 0 {
		_, err := symbolCollection.InsertMany(context.TODO(), symbolBSONs)
		// check for errors in the insertion
		if err != nil {
			panic(err)
		}
		// display the ids of the newly inserted objects
		log.Println("Data Synchronized.")
	}
}

func UpdateArbitrageRelation() {
	mongoDB := db.GetMongoDB()
	symbolCollection := mongoDB.GetCollection(defaultCollection)

	filter := bson.M{
		"quote":    fundamentalSymbol,
		"exchange": "binance",
	}

	cur, err := symbolCollection.Find(mongoDB.Ctx, filter)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err := cur.Close(mongoDB.Ctx) // we must close anyway
		if err == nil {               // we must not overwrite the actual error if it is happened, and we did all the best to cleanup anyway
			err = errors.Wrap(err, "close")
		}
	}()
	var res []interface{}
	err = cur.All(mongoDB.Ctx, &res)
	if err != nil {
		log.Fatal(err)
	}
	mongoDB.Close()

	for _, result := range res {
		resultMap := result.(bson.D).Map()
		baseSymbol := resultMap["base"].(string)
		filterMedium := bson.M{
			"exchange": "binance",
			"$or": bson.A{
				bson.M{"quote": baseSymbol},
				bson.M{"base": baseSymbol},
			},
		}
		mongoDB = db.GetMongoDB()
		symbolCollection = mongoDB.GetCollection(defaultCollection)
		curMedium, err := symbolCollection.Find(mongoDB.Ctx, filterMedium)
		if err != nil {
			log.Fatal(err)
		}
		var resMedium []interface{}
		err = curMedium.All(mongoDB.Ctx, &resMedium)
		if err != nil {
			log.Fatal(err)
		}
		mongoDB.Close()
		var resultMediums []interface{}
		for _, resultMedium := range resMedium {
			resultMediumMap := resultMedium.(bson.D).Map()
			baseMediumSymbol := ""
			var isBuyOrSell bool // true = Buy, false = Sell
			if resultMediumMap["quote"].(string) != fundamentalSymbol && resultMediumMap["base"].(string) != fundamentalSymbol {
				if baseSymbol == resultMediumMap["base"].(string) {
					baseMediumSymbol = resultMediumMap["quote"].(string)
					isBuyOrSell = false // sell base
				} else {
					baseMediumSymbol = resultMediumMap["base"].(string)
					isBuyOrSell = true //buy medium
				}
			} else {
				continue
			}

			filterFinal := bson.M{
				"exchange": "binance",
				"quote":    fundamentalSymbol,
				"base":     baseMediumSymbol,
			}

			var resultFinal bson.M
			// check for errors in the finding
			mongoDB = db.GetMongoDB()
			symbolCollection = mongoDB.GetCollection(defaultCollection)
			if err = symbolCollection.FindOne(mongoDB.Ctx, filterFinal).Decode(&resultFinal); err != nil {
				mongoDB.Close()
				if err == mongo.ErrNoDocuments {
					continue
				}
			}
			mongoDB.Close()
			resultMediums = append(resultMediums, bson.M{"symbol": resultMediumMap["symbol"], "base": resultMediumMap["base"], "quote": resultMediumMap["quote"], "buy_or_sell": isBuyOrSell, "filters": resultMediumMap["filters"], "final_filters": resultFinal["filters"]})
		}
		mongoDB = db.GetMongoDB()
		symbolCollection = mongoDB.GetCollection(defaultCollection)
		_, err = symbolCollection.UpdateByID(mongoDB.Ctx, resultMap["_id"], bson.D{
			{"$set", bson.D{
				{"arbitrage", resultMediums},
			},
			},
		})
		mongoDB.Close()

		if err != nil {
			log.Fatal(err)
		}
	}
	log.Println("Arbitrage Relationships Updated.")
}
