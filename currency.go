package currency

const (
	// PriceUSDType - usd price type
	PriceUSDType = "USD"

	// PriceZUSDType - zusd price type
	PriceZUSDType = "ZUSD"

	// PriceUSDTType - usdt price type
	PriceUSDTType = "USDT"
)

// Converter - fiat price converter interface
type Converter interface {
	GetConvertModToUSD(fromSymbol string) (float64, error)
	AppendCustomFiat(key string, value float64)
	CustomFiat(symbol string) bool
	Fiat(symbol string) bool
}
