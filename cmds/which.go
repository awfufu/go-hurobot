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

const whichHelpMsg string = `Query abbreviation meanings.
Usage: /which <text>
Example: /which yyds`

type WhichCommand struct {
	cmdBase
}

func NewWhichCommand() *WhichCommand {
	return &WhichCommand{
		cmdBase: cmdBase{
			Name:        "which",
			HelpMsg:     whichHelpMsg,
			Permission:  getCmdPermLevel("which"),
			AllowPrefix: false,
			NeedRawMsg:  false,
			MaxArgs:     2,
			MinArgs:     2,
		},
	}
}

func (cmd *WhichCommand) Self() *cmdBase {
	return &cmd.cmdBase
}

func (cmd *WhichCommand) Exec(c *qbot.Client, args []string, src *srcMsg, begin int) {
	// Check for non-text type parameters
	for i := 1; i < len(args); i++ {
		if strings.HasPrefix(args[i], "--") {
			c.SendMsg(src.GroupID, src.UserID, "Only plain text is allowed")
			return
		}
	}

	text := strings.Join(args[1:], " ")
	if text == "" {
		c.SendMsg(src.GroupID, src.UserID, cmd.HelpMsg)
		return
	}

	if strings.Contains(text, ";") {
		c.SendMsg(src.GroupID, src.UserID, "Multiple queries are not allowed")
		return
	}

	reqData := NbnhhshRequest{
		Text: text,
	}

	jsonData, err := json.Marshal(reqData)
	if err != nil {
		c.SendMsg(src.GroupID, src.UserID, err.Error())
		return
	}

	req, err := http.NewRequest("POST", "https://lab.magiconch.com/api/nbnhhsh/guess", bytes.NewBuffer(jsonData))
	if err != nil {
		c.SendMsg(src.GroupID, src.UserID, err.Error())
		return
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		c.SendMsg(src.GroupID, src.UserID, err.Error())
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.SendMsg(src.GroupID, src.UserID, err.Error())
		return
	}

	if resp.StatusCode != 200 {
		c.SendMsg(src.GroupID, src.UserID, fmt.Sprintf("http error %d", resp.StatusCode))
		return
	}

	var nbnhhshResp NbnhhshResponse
	if err := json.Unmarshal(body, &nbnhhshResp); err != nil {
		c.SendMsg(src.GroupID, src.UserID, err.Error())
		return
	}

	if len(nbnhhshResp) == 0 {
		c.SendMsg(src.GroupID, src.UserID, "null")
		return
	}

	result := nbnhhshResp[0]

	if len(result.Trans) > 0 {
		c.SendMsg(src.GroupID, src.UserID, strings.Join(result.Trans, ", "))
		return
	}

	c.SendMsg(src.GroupID, src.UserID, "null")
}
