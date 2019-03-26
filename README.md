# Currency

Currency library. Give only convert methods now.

## Work flow

Get fiat convert mods from http://currencies.apps.grandtrunk.net every 10 seconds.

Provide converter interface:

```go
type Converter interface {
	GetConvertModToUSD(fromSymbol string) (float64, error)
	AppendCustomFiat(key string, value float64)
	CustomFiat(symbol string) bool
	Fiat(symbol string) bool
}
```

You can add your custom fiats by `AppendCustomFiat`.
