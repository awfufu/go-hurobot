package cmds

import (
	"go-hurobot/qbot"
)

const rpsHelpMsg string = `Play rock-paper-scissors.
Usage: /rps`

type RpsCommand struct {
	cmdBase
}

func NewRpsCommand() *RpsCommand {
	return &RpsCommand{
		cmdBase: cmdBase{
			Name:        "rps",
			HelpMsg:     rpsHelpMsg,
			Permission:  getCmdPermLevel("rps"),
			AllowPrefix: false,
			NeedRawMsg:  false,
			MaxArgs:     1,
			MinArgs:     1,
		},
	}
}

func (cmd *RpsCommand) Self() *cmdBase {
	return &cmd.cmdBase
}

func (cmd *RpsCommand) Exec(c *qbot.Client, args []string, src *srcMsg, _ int) {
	c.SendMsg(src.GroupID, src.UserID, qbot.CQRps())
}
