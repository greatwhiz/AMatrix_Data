package binance

import (
	"A-Matrix/src/db"
	"context"
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

func SubscribeMarket() {
	c := dail()
	defer func() {
		err := c.Close() // we must close anyway
		if err == nil {  // we must not overwrite the actual error if it is happened, and we did all the best to cleanup anyway
			err = errors.Wrap(err, "close")
		}
	}()

	mongoDB := db.GetMongoDB()
	defer mongoDB.Cancel()
	symbolCollection := mongoDB.GetCollection("symbols")
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
		println("Updated:", ticker["s"].(string))
		if resultBSON != nil && resultBSON.Map()["arbitrage"] != nil && resultBSON.Map()["ticker"] != nil {
			analyze(symbolCollection, mongoDB.Ctx, resultBSON)
		}
	}
}

func analyze(symbolCollection *mongo.Collection, ctx context.Context, symbolBSON bson.D) {
	symbolBytes, _ := bson.MarshalExtJSON(symbolBSON, true, true)
	symbolJSON := string(symbolBytes)
	baseSymbol := gjson.Get(symbolJSON, "base").Str
	askPrice := gjson.Get(symbolJSON, "ticker.ask").Float()
	for _, relation := range symbolBSON.Map()["arbitrage"].(bson.A) {
		filter := bson.M{
			"exchange": "binance",
			"symbol":   relation,
		}

		var result bson.D
		// check for errors in the finding
		if err := symbolCollection.FindOne(ctx, filter).Decode(&result); err != nil {
			if err == mongo.ErrNoDocuments {
				continue
			}
		}
		resultBytes, _ := bson.MarshalExtJSON(result, true, true)
		resultJSON := string(resultBytes)
		var mediumSymbol string
		var mediumSellPrice, mediumBuyPrice float64
		if baseSymbol == gjson.Get(resultJSON, "base").Str {
			mediumSellPrice = gjson.Get(resultJSON, "ticker.bid").Float()
			mediumSymbol = gjson.Get(resultJSON, "quote").Str
		} else {
			mediumBuyPrice = gjson.Get(resultJSON, "ticker.ask").Float()
			mediumSymbol = gjson.Get(resultJSON, "base").Str
		}

		var resultFinal bson.D
		filterFinal := bson.M{
			"exchange": "binance",
			"base":     mediumSymbol,
			"quote":    "USDT",
		}

		if err := symbolCollection.FindOne(ctx, filterFinal).Decode(&resultFinal); err != nil {
			if err == mongo.ErrNoDocuments {
				continue
			} else {
				panic(err)
			}
		}
		resultFinalBytes, _ := bson.MarshalExtJSON(resultFinal, true, true)
		resultFinalJSON := string(resultFinalBytes)
		bidFinalPrice := gjson.Get(resultFinalJSON, "ticker.bid").Float()
		var estimatedAmount float64
		if mediumSellPrice != 0 {
			estimatedAmount = 10000 / askPrice * mediumSellPrice * bidFinalPrice
		} else {
			estimatedAmount = 10000 / askPrice / mediumBuyPrice * bidFinalPrice
		}
		println(fmt.Sprintf("%.2f", estimatedAmount/10000))
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
