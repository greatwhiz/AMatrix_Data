package binance

import (
	"fmt"
	"github.com/tidwall/gjson"
)

func GetBalance(symbol string) {
	balanceBody := GetSignedAPI("account", nil)
	AccountBalance[symbol] = gjson.Get(balanceBody, fmt.Sprintf("balances.#(asset==\"%s\").free", symbol)).Float()
}
