package binance

var balanceSymbol string
var AccountBalance map[string]float64
var commissionRate float64
var threshold float64
var blackList map[string]bool
var apiKey string
var secretKey string
var leverage float64

func init() {
	balanceSymbol = "USDT"
	AccountBalance = map[string]float64{}
	commissionRate = 0.001
	threshold = 1.01
	blackList = map[string]bool{"HOTBNB": true, "PERLBNB": true, "SPELLBNB": true, "API3BNB": true, "TLMBNB": true}
	apiKey = "crhlAbsbBAC9j9WAwyic8WFhUWSLP0TJIScir1ny5HxtTehq2G19sKE0tCFqho2s"
	secretKey = "Ou7EtmQ5sfBMe2zV8Sm0sNuxihp5UyZVIYMWboRbpQ8FhTCwgqH0S6t5bV66Oc7Y"
	leverage = .95
}
