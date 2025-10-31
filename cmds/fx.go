package cmds

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"go-hurobot/config"
	"go-hurobot/qbot"
)

type ExchangeRateResponse struct {
	Result          string             `json:"result"`
	BaseCode        string             `json:"base_code"`
	ConversionRates map[string]float64 `json:"conversion_rates"`
}

const erHelpMsg string = `Query foreign exchange rates.
Usage: fx <from_currency> <to_currency>
Example: fx CNY HKD`

type ErCommand struct {
	cmdBase
}

func NewErCommand() *ErCommand {
	return &ErCommand{
		cmdBase: cmdBase{
			Name:        "fx",
			HelpMsg:     erHelpMsg,
			Permission:  getCmdPermLevel("fx"),
			AllowPrefix: false,
			NeedRawMsg:  false,
			MaxArgs:     3,
			MinArgs:     3,
		},
	}
}

func (cmd *ErCommand) Self() *cmdBase {
	return &cmd.cmdBase
}

func (cmd *ErCommand) Exec(c *qbot.Client, args []string, src *srcMsg, begin int) {
	if config.Cfg.ApiKeys.ExchangeRateAPIKey == "" {
		return
	}

	fromCurrency := strings.ToUpper(args[1])
	toCurrency := strings.ToUpper(args[2])

	url := fmt.Sprintf("https://v6.exchangerate-api.com/v6/%s/latest/%s", config.Cfg.ApiKeys.ExchangeRateAPIKey, fromCurrency)

	log.Println(url)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		c.SendMsg(src.GroupID, src.UserID, fmt.Sprintf("%v", err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.SendMsg(src.GroupID, src.UserID, fmt.Sprintf("%d", resp.StatusCode))
		return
	}

	var exchangeData ExchangeRateResponse
	if err := json.NewDecoder(resp.Body).Decode(&exchangeData); err != nil {
		c.SendMsg(src.GroupID, src.UserID, fmt.Sprintf("%v", err))
		return
	}

	if exchangeData.Result != "success" {
		c.SendMsg(src.GroupID, src.UserID, fmt.Sprintf("%v", exchangeData.Result))
		return
	}

	toRate, exists := exchangeData.ConversionRates[toCurrency]
	if !exists {
		c.SendMsg(src.GroupID, src.UserID, fmt.Sprintf("Unsupported %s", toCurrency))
		return
	}

	fromRate, exists := exchangeData.ConversionRates[fromCurrency]
	if !exists {
		c.SendMsg(src.GroupID, src.UserID, fmt.Sprintf("Unsupported %s", fromCurrency))
		return
	}

	rate1to2 := toRate / fromRate
	rate2to1 := fromRate / toRate

	result := fmt.Sprintf("1 %s = %.4f %s\n1 %s = %.4f %s",
		fromCurrency, rate1to2, toCurrency,
		toCurrency, rate2to1, fromCurrency)

	c.SendMsg(src.GroupID, src.UserID, result)
}
