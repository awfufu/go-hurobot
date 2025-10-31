package cmds

import (
	"go-hurobot/qbot"
)

const diceHelpMsg = "Roll a dice.\nUsage: /dice"

type DiceCommand struct {
	cmdBase
}

func NewDiceCommand() *DiceCommand {
	return &DiceCommand{
		cmdBase: cmdBase{
			Name:        "dice",
			HelpMsg:     diceHelpMsg,
			Permission:  getCmdPermLevel("dice"),
			AllowPrefix: false,
			NeedRawMsg:  false,
			MaxArgs:     1,
			MinArgs:     1,
		},
	}
}

func (cmd *DiceCommand) Self() *cmdBase {
	return &cmd.cmdBase
}

func (cmd *DiceCommand) Exec(c *qbot.Client, args []string, src *srcMsg, _ int) {
	c.SendMsg(src.GroupID, src.UserID, qbot.CQDice())
}
