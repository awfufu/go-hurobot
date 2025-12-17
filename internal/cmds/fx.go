package cmds

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/awfufu/go-hurobot/internal/config"
	"github.com/awfufu/qbot"
)

type FxRateResponse struct {
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
			Name:       "fx",
			HelpMsg:    erHelpMsg,
			Permission: getCmdPermLevel("fx"),

			NeedRawMsg: false,
			MaxArgs:    3,
			MinArgs:    3,
		},
	}
}

func (cmd *ErCommand) Self() *cmdBase {
	return &cmd.cmdBase
}

func (cmd *ErCommand) Exec(b *qbot.Sender, msg *qbot.Message) {
	if config.Cfg.ApiKeys.ExchangeRateAPIKey == "" {
		return
	}

	if len(msg.Array) < 3 {
		b.SendGroupMsg(msg.GroupID, cmd.HelpMsg)
		return
	}

	getText := func(i int) string {
		if i < len(msg.Array) {
			if msg.Array[i].Type() == qbot.TextType {
				return msg.Array[i].Text()
			}
		}
		return ""
	}

	fromCurrency := strings.ToUpper(getText(1))
	toCurrency := strings.ToUpper(getText(2))

	url := fmt.Sprintf("https://v6.exchangerate-api.com/v6/%s/latest/%s", config.Cfg.ApiKeys.ExchangeRateAPIKey, fromCurrency)

	log.Println(url)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		b.SendGroupMsg(msg.GroupID, fmt.Sprintf("%v", err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b.SendGroupMsg(msg.GroupID, fmt.Sprintf("%d", resp.StatusCode))
		return
	}

	var exchangeData FxRateResponse
	if err := json.NewDecoder(resp.Body).Decode(&exchangeData); err != nil {
		b.SendGroupMsg(msg.GroupID, fmt.Sprintf("%v", err))
		return
	}

	if exchangeData.Result != "success" {
		b.SendGroupMsg(msg.GroupID, fmt.Sprintf("%v", exchangeData.Result))
		return
	}

	toRate, exists := exchangeData.ConversionRates[toCurrency]
	if !exists {
		b.SendGroupMsg(msg.GroupID, fmt.Sprintf("Unsupported %s", toCurrency))
		return
	}

	fromRate, exists := exchangeData.ConversionRates[fromCurrency]
	if !exists {
		b.SendGroupMsg(msg.GroupID, fmt.Sprintf("Unsupported %s", fromCurrency))
		return
	}

	rate1to2 := toRate / fromRate
	rate2to1 := fromRate / toRate

	result := fmt.Sprintf("1 %s = %.4f %s\n1 %s = %.4f %s",
		fromCurrency, rate1to2, toCurrency,
		toCurrency, rate2to1, fromCurrency)

	b.SendGroupMsg(msg.GroupID, result)
}
