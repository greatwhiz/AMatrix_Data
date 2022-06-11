package binance

import (
	"fmt"
	"github.com/tidwall/gjson"
)

var AccountBalance = make(map[string]float64)

func GetBalance(symbol string) {
	balanceBody := GetSignedAPI("account", nil)
	AccountBalance[symbol] = gjson.Get(balanceBody, fmt.Sprintf("balances.#(asset==\"%s\").free", symbol)).Float()
}
