package currency

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/lillilli/logger"
)

const (
	baseURL                       = "https://api.openrates.io/latest?base=USD"
	additionalURL                 = "http://currencies.apps.grandtrunk.net/getlatest/"
	fiatsConvertModUpdateInterval = 1 * time.Hour
)

var (
	defaultSupportedFiats = []string{
		"KRW", "JPY", "TRY",
		"EUR", "GBP", "RUB",
		"CNY", "CHF", "AUD", "CAD", "BRL", "UAH", "IRR",
	}
	additionalFiats = []string{
		"UAH", "IRR",
	}
)

type converter struct {
	client *http.Client
	log    logger.Logger

	cache                    map[string]float64
	additionalSupportedFiats map[string]bool
	sync.RWMutex
}

// NewConverter - return new converter instance
func NewConverter() Converter {
	converter := &converter{
		client:                   &http.Client{Timeout: 10 * time.Second},
		log:                      logger.NewLogger("price converter"),
		cache:                    make(map[string]float64),
		additionalSupportedFiats: make(map[string]bool),
	}

	converter.updateConvertMods()
	go converter.startFetchConvertMods()

	return converter
}

func (c *converter) GetConvertModToUSD(fromSymbol string) (float64, error) {
	fromSymbol = strings.ToUpper(fromSymbol)

	if fromSymbol == PriceUSDType || fromSymbol == PriceZUSDType {
		return 1, nil
	}

	if !c.Fiat(fromSymbol) && !c.CustomFiat(fromSymbol) {
		return 0, fmt.Errorf("unexpected not fiat symbol %q", fromSymbol)
	}

	c.RLock()
	mod, ok := c.cache[fromSymbol]
	c.RUnlock()

	if !ok {
		return 0, fmt.Errorf("convert mod from %q not found in cache", fromSymbol)
	}

	return mod, nil
}

func (c *converter) startFetchConvertMods() {
	ticker := time.NewTicker(fiatsConvertModUpdateInterval)

	for {
		select {
		case <-ticker.C:
			c.updateConvertMods()
		}
	}
}

func (c *converter) updateConvertMods() {
	convertMods, err := c.getRemoteConvertMods()
	if err != nil {
		c.log.Errorf("Fetching convert mods failed: %v", err)
		return
	}

	for _, convertSymbol := range defaultSupportedFiats {
		mod, ok := convertMods[convertSymbol]
		if !ok {
			c.log.Warnf("Fiat %s not found in converter mods", convertSymbol)
			continue
		}

		if mod != 0 {
			mod = math.Pow(mod, -1)
		}

		c.Lock()
		c.cache[convertSymbol] = mod
		c.Unlock()
	}
}

func (c *converter) getRemoteConvertMods() (map[string]float64, error) {
	resp, err := c.client.Get(baseURL)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	convertMods := &usdConvertMods{}
	err = json.Unmarshal(b, convertMods)
	if err != nil {
		return nil, err
	}

	for _, mod := range additionalFiats {
		rate, err := c.getAdditionalRemoteConvertMods(mod)
		if err != nil {
			continue
		}

		if _, ok := convertMods.Rates[mod]; !ok {
			convertMods.Rates[mod] = rate
		}
	}

	return convertMods.Rates, nil
}

func (c *converter) AppendCustomFiat(key string, value float64) {
	key = strings.ToUpper(key)

	c.Lock()
	c.additionalSupportedFiats[key] = true
	c.cache[key] = value
	c.Unlock()
}

func (c *converter) Fiat(symbol string) bool {
	symbol = strings.ToUpper(symbol)

	if symbol == PriceUSDType || symbol == PriceZUSDType {
		return true
	}

	return indexOfString(symbol, defaultSupportedFiats) != -1
}

func (c *converter) CustomFiat(symbol string) bool {
	symbol = strings.ToUpper(symbol)
	_, ok := c.additionalSupportedFiats[symbol]

	return ok
}

func indexOfString(element string, data []string) int {
	for k, v := range data {
		if element == v {
			return k
		}
	}

	return -1
}

func (c *converter) getAdditionalRemoteConvertMods(mod string) (float64, error) {
	resp, err := c.client.Get(fmt.Sprintf("%s/%s/%s", additionalURL, "USD", mod))
	if err != nil {
		return 0, err
	}

	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	return strconv.ParseFloat(string(b), 64)
}
