package binance

import (
	"A-Matrix/src/db"
	"context"
	"fmt"
	"github.com/tidwall/gjson"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"io/ioutil"
	"log"
	"net/http"
)

func UpdateSymbols() {
	host := "https://api.binance.com/api/v3"

	url := fmt.Sprintf("%s/%s", host, "exchangeInfo")

	resp, err := http.Get(url)

	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Fatal(err)
	}

	mongoDB := db.GetMongoDB()
	defer mongoDB.Cancel()
	symbolCollection := mongoDB.GetCollection("symbols")

	symbols := gjson.Get(string(body), "symbols")
	var symbolBSONs []interface{}
	symbols.ForEach(func(key, symbol gjson.Result) bool {
		var result bson.M
		value := symbol.Map()
		symbolBSON := bson.D{{"symbol", value["symbol"].String()}, {"exchange", "binance"}, {"base", value["baseAsset"].String()}, {"quote", value["quoteAsset"].String()}}
		err := symbolCollection.FindOne(context.TODO(), bson.D{{"symbol", value["symbol"].String()}, {"exchange", "binance"}}).Decode(&result)
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
		results, err := symbolCollection.InsertMany(context.TODO(), symbolBSONs)
		// check for errors in the insertion
		if err != nil {
			panic(err)
		}
		// display the ids of the newly inserted objects
		fmt.Println(results.InsertedIDs)
	}
}

func UpdateArbitrageRelation() {
	mongoDB := db.GetMongoDB()
	defer mongoDB.Cancel()
	symbolCollection := mongoDB.GetCollection("symbols")

	filter := bson.M{
		"quote":    "USDT",
		"exchange": "binance",
	}

	cur, err := symbolCollection.Find(mongoDB.Ctx, filter)
	if err != nil {
		log.Fatal(err)
	}
	defer cur.Close(mongoDB.Ctx)
	for cur.Next(mongoDB.Ctx) {
		var result bson.D
		err := cur.Decode(&result)
		if err != nil {
			log.Fatal(err)
		}
		resultMap := result.Map()
		baseSymbol := resultMap["base"].(string)
		filterMedium := bson.M{
			"exchange": "binance",
			"$or": bson.A{
				bson.M{"quote": baseSymbol},
				bson.M{"baseSymbol": baseSymbol},
			},
		}
		curMedium, err := symbolCollection.Find(mongoDB.Ctx, filterMedium)
		var resultMediums []interface{}
		for curMedium.Next(mongoDB.Ctx) {
			var resultMedium bson.D
			err := curMedium.Decode(&resultMedium)
			if err != nil {
				log.Fatal(err)
			}
			resultMediumMap := resultMedium.Map()
			baseMediumSymbol := ""
			if resultMediumMap["quote"].(string) != "USDT" && resultMediumMap["base"].(string) != "USDT" {
				if baseSymbol == resultMediumMap["base"].(string) {
					baseMediumSymbol = resultMediumMap["quote"].(string)
				} else {
					baseMediumSymbol = resultMediumMap["base"].(string)
				}
			} else {
				continue
			}

			filterFinal := bson.M{
				"exchange": "binance",
				"quote":    "USDT",
				"base":     baseMediumSymbol,
			}

			var resultFinal bson.M
			// check for errors in the finding
			if err = symbolCollection.FindOne(mongoDB.Ctx, filterFinal).Decode(&resultFinal); err != nil {
				if err == mongo.ErrNoDocuments {
					continue
				}
			}

			resultMediums = append(resultMediums, resultMediumMap["symbol"])
		}

		_, err = symbolCollection.UpdateByID(mongoDB.Ctx, result[0].Value, bson.D{
			{"$set", bson.D{
				{"arbitrage", resultMediums},
			},
			},
		})
		if err != nil {
			log.Fatal(err)
		}
	}
	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}
}
