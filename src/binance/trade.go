package binance

func OrderFull(symbol string, baseSymbol string) {
	GetBalance(baseSymbol)
	println(symbol, AccountBalance[baseSymbol])
}
