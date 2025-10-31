package cmds

import (
	"go-hurobot/qbot"
	"strings"
)

const echoHelpMsg string = `Echoes messages to a target destination.
Usage: /echo [options] <content>
Options:
  -d, -r  Decode special characters
  -e      Encode special characters
Example: /echo "Hello, world!"`

type EchoCommand struct {
	cmdBase
}

func NewEchoCommand() *EchoCommand {
	return &EchoCommand{
		cmdBase: cmdBase{
			Name:        "echo",
			HelpMsg:     echoHelpMsg,
			Permission:  getCmdPermLevel("echo"),
			AllowPrefix: false,
			NeedRawMsg:  false,
			MinArgs:     2,
		},
	}
}

func (cmd *EchoCommand) Self() *cmdBase {
	return &cmd.cmdBase
}

func (cmd *EchoCommand) Exec(c *qbot.Client, args []string, src *srcMsg, _ int) {
	if len(args) >= 3 && args[1][0] == '-' {
		switch args[1] {
		case "-r":
			c.SendMsg(src.GroupID, src.UserID, encodeSpecialChars(src.Raw[3:]))
		case "-d":
			c.SendMsg(src.GroupID, src.UserID, decodeSpecialChars(args[2]))
		}
	} else {
		c.SendMsg(src.GroupID, src.UserID, strings.Join(args[1:], " "))
	}
}
