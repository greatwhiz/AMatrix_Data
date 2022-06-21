package binance_v1

import (
	"fmt"
	"github.com/tidwall/gjson"
	"log"
)

func GetBalance(symbol string) {
	balanceBody := GetSignedAPI("account", nil)
	AccountBalance[symbol] = gjson.Get(balanceBody, fmt.Sprintf("balances.#(asset==\"%s\").free", symbol)).Float()
	log.Print(fmt.Sprintf("Balance %s: %.4f", symbol, AccountBalance[symbol]))
}
