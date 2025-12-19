package cmds

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/awfufu/go-hurobot/internal/config"
	"github.com/awfufu/qbot"
)

const waHelpMsg string = `Calculates the result of a mathematical expression using WolframAlpha.
Usage: /wa <expression>
Example: /wa integrate x^2`

var waCommand *Command = &Command{
	Name:       "wa",
	HelpMsg:    waHelpMsg,
	Permission: getCmdPermLevel("wa"),
	NeedRawMsg: false,
	MinArgs:    2,
	Exec:       waExec,
}

func waExec(b *qbot.Sender, msg *qbot.Message) {
	appID := config.Cfg.ApiKeys.WolframAlphaAppID
	if appID == "" {
		b.SendGroupMsg(msg.GroupID, "WolframAlpha AppID not configured")
		return
	}

	exprString := ""
	for _, item := range msg.Array[1:] {
		if item.Type() == qbot.TextType {
			exprString += item.Text() + " "
		} else {
			b.SendGroupMsg(msg.GroupID, "invalid expression")
			return
		}
	}

	// Use WolframAlpha Simple API
	// https://products.wolframalpha.com/simple-api/documentation/
	apiURL := "https://api.wolframalpha.com/v1/simple"
	params := url.Values{}
	params.Set("appid", appID)
	params.Set("i", exprString)
	// params.Set("background", "black") // Optional: optimize for dark mode?
	// params.Set("foreground", "white")

	resp, err := http.Get(apiURL + "?" + params.Encode())
	if err != nil {
		b.SendGroupMsg(msg.GroupID, fmt.Sprintf("Request failed: %v", err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotImplemented {
		b.SendGroupMsg(msg.GroupID, "No short answer available")
		return
	} else if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		b.SendGroupMsg(msg.GroupID, fmt.Sprintf("Error: %s", string(body)))
		return
	}

	contentType := resp.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "image/") {
		// Create temporary file
		tmpFile, err := os.CreateTemp("", "wolfram_*.gif")
		if err != nil {
			b.SendGroupMsg(msg.GroupID, fmt.Sprintf("Failed to create temp file: %v", err))
			return
		}
		defer os.Remove(tmpFile.Name())
		defer tmpFile.Close()

		_, err = io.Copy(tmpFile, resp.Body)
		if err != nil {
			b.SendGroupMsg(msg.GroupID, fmt.Sprintf("Failed to write image: %v", err))
			return
		}

		// Send image
		b.SendGroupMsg(msg.GroupID, qbot.Image("file://"+tmpFile.Name()))
	} else {
		// Fallback for text response (though Simple API usually returns image)
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			b.SendGroupMsg(msg.GroupID, fmt.Sprintf("Failed to read response: %v", err))
			return
		}
		b.SendGroupMsg(msg.GroupID, string(body))
	}
}
