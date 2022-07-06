package binance_v2

const (
	restAPIHost            = "https://api.binance.com/api/v3"
	defaultCollection      = "symbols_binance"
	fundamentalSymbol      = "USDT"
	commissionRate         = 0.001
	arbitrageThreshold     = 1.01
	tradingAmountThreshold = 100 // based on fundamentalSymbol
	apiKey                 = "crhlAbsbBAC9j9WAwyic8WFhUWSLP0TJIScir1ny5HxtTehq2G19sKE0tCFqho2s"
	secretKey              = "Ou7EtmQ5sfBMe2zV8Sm0sNuxihp5UyZVIYMWboRbpQ8FhTCwgqH0S6t5bV66Oc7Y"
	leverage               = .95
	tradeNumLimit          = 2 // >0 normal limit, =0 off trading, =-1 unlimited
)

var AccountBalance map[string]float64
var tradeCount int
var tradingChan chan int

func init() {
	AccountBalance = map[string]float64{}
	tradingChan = make(chan int)
	GetBalance(fundamentalSymbol)
}
