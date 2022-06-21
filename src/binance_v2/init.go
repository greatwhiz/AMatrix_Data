package binance_v1

var websocketAddress, restAPIHost string
var fundamentalSymbol string
var AccountBalance map[string]float64
var commissionRate float64
var arbitrageThreshold, tradingAmountThreshold float64
var blackList map[string]bool
var apiKey string
var secretKey string
var leverage float64
var tradeNumLimit, tradeCount, analyzingConcurrency int

func init() {
	websocketAddress = "stream.binance_v1.com:9443"
	restAPIHost = "https://api.binance.com/api/v3"
	fundamentalSymbol = "USDT"
	AccountBalance = map[string]float64{}
	commissionRate = 0.001
	arbitrageThreshold = 1.01
	tradingAmountThreshold = 100  // based on fundamentalSymbol
	blackList = map[string]bool{} //{"HOTBNB": true, "PERLBNB": true, "SPELLBNB": true, "API3BNB": true, "TLMBNB": true}
	apiKey = "crhlAbsbBAC9j9WAwyic8WFhUWSLP0TJIScir1ny5HxtTehq2G19sKE0tCFqho2s"
	secretKey = "Ou7EtmQ5sfBMe2zV8Sm0sNuxihp5UyZVIYMWboRbpQ8FhTCwgqH0S6t5bV66Oc7Y"
	leverage = .98
	tradeNumLimit = 3 // >0 normal limit, =0 off trading, =-1 unlimited
	analyzingConcurrency = 2
	GetBalance(fundamentalSymbol)
}