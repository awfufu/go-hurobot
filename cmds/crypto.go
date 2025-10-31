package cmds

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go-hurobot/config"
	"go-hurobot/qbot"
)

type OkxResponse struct {
	Code string          `json:"code"`
	Data []OkxInstrument `json:"data"`
	Msg  string          `json:"msg"`
}

type OkxInstrument struct {
	Alias            string `json:"alias"`
	AuctionEndTime   string `json:"auctionEndTime"`
	BaseCcy          string `json:"baseCcy"`
	Category         string `json:"category"`
	ContTdSwTime     string `json:"contTdSwTime"`
	CtMult           string `json:"ctMult"`
	CtType           string `json:"ctType"`
	CtVal            string `json:"ctVal"`
	CtValCcy         string `json:"ctValCcy"`
	ExpTime          string `json:"expTime"`
	FutureSettlement bool   `json:"futureSettlement"`
	InstFamily       string `json:"instFamily"`
	InstId           string `json:"instId"`
	InstType         string `json:"instType"`
	Lever            string `json:"lever"`
	ListTime         string `json:"listTime"`
	LotSz            string `json:"lotSz"`
	MaxIcebergSz     string `json:"maxIcebergSz"`
	MaxLmtAmt        string `json:"maxLmtAmt"`
	MaxLmtSz         string `json:"maxLmtSz"`
	MaxMktAmt        string `json:"maxMktAmt"`
	MaxMktSz         string `json:"maxMktSz"`
	MaxStopSz        string `json:"maxStopSz"`
	MaxTriggerSz     string `json:"maxTriggerSz"`
	MaxTwapSz        string `json:"maxTwapSz"`
	MinSz            string `json:"minSz"`
	OpenType         string `json:"openType"`
	OptType          string `json:"optType"`
	QuoteCcy         string `json:"quoteCcy"`
	RuleType         string `json:"ruleType"`
	SettleCcy        string `json:"settleCcy"`
	State            string `json:"state"`
	Stk              string `json:"stk"`
	TickSz           string `json:"tickSz"`
	Uly              string `json:"uly"`
}

type TickerResp struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
	Data []struct {
		InstId string `json:"instId"`
		Last   string `json:"last"`
	} `json:"data"`
}

const cryptoHelpMsg string = `Query cryptocurrency prices.
Usage:
  /crypto <coin>                 - Query coin price in USDT
  /crypto <from_coin> <to_coin>  - Query coin price in target currency
Examples:
  /crypto BTC
  /crypto BTC USD`

type CryptoCommand struct {
	cmdBase
}

func NewCryptoCommand() *CryptoCommand {
	return &CryptoCommand{
		cmdBase: cmdBase{
			Name:        "crypto",
			HelpMsg:     cryptoHelpMsg,
			Permission:  getCmdPermLevel("crypto"),
			AllowPrefix: false,
			NeedRawMsg:  false,
			MaxArgs:     3,
			MinArgs:     2,
		},
	}
}

func (cmd *CryptoCommand) Self() *cmdBase {
	return &cmd.cmdBase
}

func (cmd *CryptoCommand) Exec(c *qbot.Client, args []string, src *srcMsg, begin int) {
	if len(args) == 2 {
		coin := strings.ToUpper(args[1])
		handleSingleCrypto(c, src, coin)
	} else if len(args) == 3 {
		fromCoin := strings.ToUpper(args[1])
		toCurrency := strings.ToUpper(args[2])
		handleCryptoCurrencyPair(c, src, fromCoin, toCurrency)
	}
}

func handleSingleCrypto(c *qbot.Client, src *srcMsg, coin string) {
	log.Printf("Query single cryptocurrency: %s", coin)
	price, err := getCryptoPrice(coin, "USDT")
	if err != nil {
		log.Printf("Failed to query %s price: %v", coin, err)
		c.SendMsg(src.GroupID, src.UserID, fmt.Sprintf("Query failed: %s", err.Error()))
		return
	}
	c.SendMsg(src.GroupID, src.UserID, fmt.Sprintf("1 %s = %s USDT", coin, price))
}

func handleCryptoCurrencyPair(c *qbot.Client, src *srcMsg, fromCoin string, toCurrency string) {
	log.Printf("Query cryptocurrency pair: %s -> %s", fromCoin, toCurrency)

	usdPrice, err := getCryptoPrice(fromCoin, "USD")
	if err != nil {
		log.Printf("Failed to query %s USD price: %v", fromCoin, err)
		c.SendMsg(src.GroupID, src.UserID, fmt.Sprintf("Failed to query %s price: %s", fromCoin, err.Error()))
		return
	}

	usdPriceFloat, err := strconv.ParseFloat(usdPrice, 64)
	if err != nil {
		log.Printf("Price parsing failed: %v", err)
		c.SendMsg(src.GroupID, src.UserID, fmt.Sprintf("Price parsing failed: %s", err.Error()))
		return
	}

	if toCurrency == "USD" {
		c.SendMsg(src.GroupID, src.UserID, fmt.Sprintf("%s latest USD price: %.4f", fromCoin, usdPriceFloat))
		return
	}

	log.Printf("Need exchange rate conversion: USD -> %s", toCurrency)
	exchangeRate, err := getExchangeRate("USD", toCurrency)
	if err != nil {
		log.Printf("Failed to get exchange rate: %v", err)
		c.SendMsg(src.GroupID, src.UserID, fmt.Sprintf("Failed to get exchange rate: %s", err.Error()))
		return
	}

	finalPrice := usdPriceFloat * exchangeRate
	log.Printf("Conversion complete: %s USD price %.4f, exchange rate %.4f, final price %.4f %s", fromCoin, usdPriceFloat, exchangeRate, finalPrice, toCurrency)
	c.SendMsg(src.GroupID, src.UserID, fmt.Sprintf("1 %s=%.4f %s", fromCoin, finalPrice, toCurrency))
}

func getCryptoPrice(coin string, quoteCurrency string) (string, error) {
	instId := coin + "-" + quoteCurrency + "-SWAP"
	url := "https://bot-forward.lavacreeper.net/api/v5/market/ticker?instId=" + instId

	log.Printf("Request cryptocurrency price: %s", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("request creation failed: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Okx-Python-Client")
	req.Header.Set("X-API-Key", config.Cfg.ApiKeys.OkxMirrorAPIKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %v", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("Failed to close response body: %v", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	var ticker TickerResp
	if err := json.NewDecoder(resp.Body).Decode(&ticker); err != nil {
		return "", fmt.Errorf("parsing failed: %v", err)
	}

	if ticker.Code != "0" || len(ticker.Data) == 0 {
		return "", fmt.Errorf("API returned error: %s", ticker.Msg)
	}

	log.Printf("Got price: %s = %s", instId, ticker.Data[0].Last)
	return ticker.Data[0].Last, nil
}

func getExchangeRate(baseCode string, targetCode string) (float64, error) {
	if config.Cfg.ApiKeys.ExchangeRateAPIKey == "" {
		return 0, fmt.Errorf("exchange rate API key not configured")
	}

	url := fmt.Sprintf("https://v6.exchangerate-api.com/v6/%s/latest/%s", config.Cfg.ApiKeys.ExchangeRateAPIKey, baseCode)

	log.Printf("Request exchange rate: %s", url)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return 0, fmt.Errorf("exchange rate request failed: %v", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("Failed to close response body: %v", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("exchange rate API HTTP error: %d", resp.StatusCode)
	}

	var exchangeData ExchangeRateResponse
	if err := json.NewDecoder(resp.Body).Decode(&exchangeData); err != nil {
		return 0, fmt.Errorf("exchange rate data parsing failed: %v", err)
	}

	if exchangeData.Result != "success" {
		return 0, fmt.Errorf("exchange rate API returned error: %s", exchangeData.Result)
	}

	rate, exists := exchangeData.ConversionRates[targetCode]
	if !exists {
		return 0, fmt.Errorf("unsupported currency: %s", targetCode)
	}

	log.Printf("Got exchange rate: 1 %s = %f %s", baseCode, rate, targetCode)
	return rate, nil
}
