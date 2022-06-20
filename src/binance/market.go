package binance

import (
	"A-Matrix/src/db"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"go.mongodb.org/mongo-driver/bson"
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

	for {
		_, content, err := c.ReadMessage()
		if err != nil {
			log.Println("websocket read:", err)
			// try to redial due to EOF error
			if websocket.IsCloseError(err, websocket.CloseAbnormalClosure) {
				c = dail()
				continue
			}
		}
		if content == nil {
			log.Println("read: nil")
			err = c.Close()
			if err != nil {
				log.Println("websocket close:", err)
			}
			c = dail()
			continue
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
			log.Println(ticker["s"], " find: ", err)
		}
		//println("Updated:", ticker["s"].(string))
		mongoDB.Close()
		if resultBSON != nil && resultBSON.Map()["arbitrage"] != nil && resultBSON.Map()["ticker"] != nil {
			Analyze(resultBSON)
		}
	}
}

func dail() *websocket.Conn {
	addr := websocketAddress

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
		log.Println("read:", err)
		err := c.Close() // we must close anyway
		if err == nil {  // we must not overwrite the actual error if it is happened, and we did all the best to cleanup anyway
			err = errors.Wrap(err, "close")
		}
		return dail()
	}
	println(string(content))
	return c
}
