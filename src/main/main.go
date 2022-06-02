package main

import "A-Matrix/src/binance"

func main() {
	binance.UpdateSymbols()
	binance.UpdateArbitrageRelation()
	binance.SubscribeMarket()
}
