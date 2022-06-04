package binance

import (
	"A-Matrix/src/db"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"
)

var commissionRate float64 = 0.001

func SubscribeMarket() {
	c := dail()
	defer func() {
		err := c.Close() // we must close anyway
		if err == nil {  // we must not overwrite the actual error if it is happened, and we did all the best to cleanup anyway
			err = errors.Wrap(err, "close")
		}
	}()

	for {
		_, content, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			// try to redial due to EOF error
			if websocket.IsCloseError(err, websocket.CloseAbnormalClosure) {
				c = dail()
				continue
			}
		}
		ticker := gjson.Parse(string(content)).Value().(map[string]interface{})
		var resultBSON bson.D
		mongoDB := db.GetMongoDB()
		symbolCollection := mongoDB.GetCollection("symbols")
		err = symbolCollection.FindOneAndUpdate(mongoDB.Ctx, bson.M{"symbol": ticker["s"], "exchange": "binance"}, bson.D{
			{"$set", bson.D{
				{"ticker", bson.D{
					{"bid", ticker["b"]},
					{"bid_qty", ticker["B"]},
					{"ask", ticker["a"]},
					{"ask_qty", ticker["A"]},
					{"last_updated", time.Now()},
				}},
			},
			},
		}).Decode(&resultBSON)
		if err != nil {
			log.Println("find:", err)
		}
		//println("Updated:", ticker["s"].(string))
		mongoDB.Close()
		if resultBSON != nil && resultBSON.Map()["arbitrage"] != nil && resultBSON.Map()["ticker"] != nil {
			analyze(resultBSON)
		}
	}
}

func analyze(symbolBSON bson.D) {
	symbolBytes, _ := bson.MarshalExtJSON(symbolBSON, true, true)
	symbolJSON := string(symbolBytes)
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
				println("arbit medium: ")
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
		var mediumSymbol string
		var mediumSellPrice, mediumBuyPrice float64
		if baseSymbol == gjson.Get(mediumJSON, "base").Str {
			mediumSellPrice = gjson.Get(mediumJSON, "ticker.bid").Float()
			mediumSymbol = gjson.Get(mediumJSON, "quote").Str
		} else {
			mediumBuyPrice = gjson.Get(mediumJSON, "ticker.ask").Float()
			mediumSymbol = gjson.Get(mediumJSON, "base").Str
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
		var estimatedAmount float64
		if mediumSellPrice != 0 {
			estimatedAmount = 10000 / askPrice * (1 - commissionRate) * mediumSellPrice * (1 - commissionRate) * bidFinalPrice * (1 - commissionRate)
			println(fmt.Sprintf("%s(%.8f) %s(Sell %.8f) (%.8f): %.4f", baseSymbol, askPrice, mediumRelation, mediumSellPrice, bidFinalPrice, estimatedAmount/10000))
		} else if mediumBuyPrice != 0 {
			estimatedAmount = 10000 / askPrice * (1 - commissionRate) / mediumBuyPrice * (1 - commissionRate) * bidFinalPrice * (1 - commissionRate)
			println(fmt.Sprintf("%s(%.8f) %s(Buy %.8f) (%.8f): %.4f", baseSymbol, askPrice, mediumRelation, mediumBuyPrice, bidFinalPrice, estimatedAmount/10000))
		}
	}
}

func dail() *websocket.Conn {
	addr := "stream.binance.com:9443"

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "wss", Host: addr, Path: "/ws/bookTicker"}
	log.Printf("connecting to %s", u.String())
	dialer := websocket.DefaultDialer
	dialer.HandshakeTimeout = 100 * time.Second //increase handshake timeout due to EOF
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	message := `{"id": 1, "method":"SUBSCRIBE", "params":["!bookTicker"]}` //["spot/depth:ETH-USDT"]}
	err = c.WriteMessage(websocket.TextMessage, []byte(message))
	if err != nil {
		log.Fatal("write:", err)
	}
	_, content, err := c.ReadMessage()
	if err != nil {
		log.Fatal("read:", err)
	}
	println(string(content))
	return c
}
