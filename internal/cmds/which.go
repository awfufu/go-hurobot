package cmds

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/awfufu/qbot"
)

type NbnhhshRequest struct {
	Text string `json:"text"`
}

type NbnhhshResponse []struct {
	Name      string   `json:"name"`
	Trans     []string `json:"trans,omitempty"`
	Inputting []string `json:"inputting,omitempty"`
}

const whichHelpMsg string = `Query abbreviation meanings.
Usage: /which <text>
Example: /which yyds`

type WhichCommand struct {
	cmdBase
}

func NewWhichCommand() *WhichCommand {
	return &WhichCommand{
		cmdBase: cmdBase{
			Name:       "which",
			HelpMsg:    whichHelpMsg,
			Permission: getCmdPermLevel("which"),

			NeedRawMsg: false,
			MaxArgs:    2,
			MinArgs:    2,
		},
	}
}

func (cmd *WhichCommand) Self() *cmdBase {
	return &cmd.cmdBase
}

func (cmd *WhichCommand) Exec(b *qbot.Sender, msg *qbot.Message) {
	// Check for non-text type parameters
	for i := 1; i < len(msg.Array); i++ {
		str := ""
		if msg.Array[i].Type() == qbot.TextType {
			str = msg.Array[i].Text()
		}
		if strings.HasPrefix(str, "--") {
			b.SendGroupMsg(msg.GroupID, "Only plain text is allowed")
			return
		}
	}

	var parts []string
	for i := 1; i < len(msg.Array); i++ {
		if msg.Array[i].Type() == qbot.TextType {
			parts = append(parts, msg.Array[i].Text())
		}
	}
	text := strings.Join(parts, " ")

	if text == "" {
		b.SendGroupMsg(msg.GroupID, cmd.HelpMsg)
		return
	}

	if strings.Contains(text, ";") {
		b.SendGroupMsg(msg.GroupID, "Multiple queries are not allowed")
		return
	}

	reqData := NbnhhshRequest{
		Text: text,
	}

	jsonData, err := json.Marshal(reqData)
	if err != nil {
		b.SendGroupMsg(msg.GroupID, err.Error())
		return
	}

	req, err := http.NewRequest("POST", "https://lab.magiconch.com/api/nbnhhsh/guess", bytes.NewBuffer(jsonData))
	if err != nil {
		b.SendGroupMsg(msg.GroupID, err.Error())
		return
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		b.SendGroupMsg(msg.GroupID, err.Error())
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		b.SendGroupMsg(msg.GroupID, err.Error())
		return
	}

	if resp.StatusCode != 200 {
		b.SendGroupMsg(msg.GroupID, fmt.Sprintf("http error %d", resp.StatusCode))
		return
	}

	var nbnhhshResp NbnhhshResponse
	if err := json.Unmarshal(body, &nbnhhshResp); err != nil {
		b.SendGroupMsg(msg.GroupID, err.Error())
		return
	}

	if len(nbnhhshResp) == 0 {
		b.SendGroupMsg(msg.GroupID, "null")
		return
	}

	result := nbnhhshResp[0]

	if len(result.Trans) > 0 {
		b.SendGroupMsg(msg.GroupID, strings.Join(result.Trans, ", "))
		return
	}

	b.SendGroupMsg(msg.GroupID, "null")
}
