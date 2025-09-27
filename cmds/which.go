package cmds

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go-hurobot/qbot"
)

type NbnhhshRequest struct {
	Text string `json:"text"`
}

type NbnhhshResponse []struct {
	Name      string   `json:"name"`
	Trans     []string `json:"trans,omitempty"`
	Inputting []string `json:"inputting,omitempty"`
}

func cmd_which(c *qbot.Client, msg *qbot.Message, args *ArgsList) {
	const helpMsg = `Usage: which <text>`
	if args.Size < 2 {
		c.SendMsg(msg, helpMsg)
		return
	}

	for i := 1; i < len(args.Types); i++ {
		if args.Types[i] != qbot.Text {
			c.SendMsg(msg, "Only plain text is allowed")
			return
		}
	}

	text := strings.Join(args.Contents[1:], " ")
	if text == "" {
		c.SendMsg(msg, helpMsg)
		return
	}

	if strings.Contains(text, ";") {
		c.SendMsg(msg, "Multiple queries are not allowed")
		return
	}

	reqData := NbnhhshRequest{
		Text: text,
	}

	jsonData, err := json.Marshal(reqData)
	if err != nil {
		c.SendMsg(msg, err.Error())
		return
	}

	req, err := http.NewRequest("POST", "https://lab.magiconch.com/api/nbnhhsh/guess", bytes.NewBuffer(jsonData))
	if err != nil {
		c.SendMsg(msg, err.Error())
		return
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		c.SendMsg(msg, err.Error())
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.SendMsg(msg, err.Error())
		return
	}

	if resp.StatusCode != 200 {
		c.SendMsg(msg, fmt.Sprintf("http error %d", resp.StatusCode))
		return
	}

	var nbnhhshResp NbnhhshResponse
	if err := json.Unmarshal(body, &nbnhhshResp); err != nil {
		c.SendMsg(msg, err.Error())
		return
	}

	if len(nbnhhshResp) == 0 {
		c.SendMsg(msg, "null")
		return
	}

	result := nbnhhshResp[0]

	if len(result.Trans) > 0 {
		c.SendMsg(msg, strings.Join(result.Trans, ", "))
		return
	}

	c.SendMsg(msg, "null")
}
