package binance_v1

var (
	websocketAddress, restAPIHost                   string
	defaultCollection                               string
	fundamentalSymbol                               string
	AccountBalance                                  map[string]float64
	commissionRate                                  float64
	arbitrageThreshold, tradingAmountThreshold      float64
	blackList                                       map[string]bool
	apiKey                                          string
	secretKey                                       string
	leverage                                        float64
	tradeNumLimit, tradeCount, analyzingConcurrency int
)

func init() {
	websocketAddress = "stream.binance_v1.com:9443"
	defaultCollection = "symbols_binance"
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
