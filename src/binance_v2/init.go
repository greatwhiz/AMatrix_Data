package binance_v2

var (
	restAPIHost                                string
	defaultCollection                          string
	fundamentalSymbol                          string
	AccountBalance                             map[string]float64
	commissionRate                             float64
	arbitrageThreshold, tradingAmountThreshold float64
	apiKey                                     string
	secretKey                                  string
	leverage                                   float64
	tradeNumLimit, tradeCount                  int
)

func init() {
	restAPIHost = "https://api.binance.com/api/v3"
	defaultCollection = "symbols_binance"
	fundamentalSymbol = "USDT"
	AccountBalance = map[string]float64{}
	commissionRate = 0.001
	arbitrageThreshold = 1.01
	tradingAmountThreshold = 100 // based on fundamentalSymbol
	apiKey = "crhlAbsbBAC9j9WAwyic8WFhUWSLP0TJIScir1ny5HxtTehq2G19sKE0tCFqho2s"
	secretKey = "Ou7EtmQ5sfBMe2zV8Sm0sNuxihp5UyZVIYMWboRbpQ8FhTCwgqH0S6t5bV66Oc7Y"
	leverage = .98
	tradeNumLimit = 0 // >0 normal limit, =0 off trading, =-1 unlimited
	GetBalance(fundamentalSymbol)
}
