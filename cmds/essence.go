package cmds

import (
	"go-hurobot/config"
	"go-hurobot/qbot"
	"slices"
	"strconv"
	"strings"
)

const essenceHelpMsg string = `Manage essence messages.
Usage: [Reply to a message] /essence [add|rm]`

type EssenceCommand struct {
	cmdBase
}

func NewEssenceCommand() *EssenceCommand {
	return &EssenceCommand{
		cmdBase: cmdBase{
			Name:        "essence",
			HelpMsg:     essenceHelpMsg,
			Permission:  getCmdPermLevel("essence"),
			AllowPrefix: false,
			NeedRawMsg:  false,
		},
	}
}

func (cmd *EssenceCommand) Self() *cmdBase {
	return &cmd.cmdBase
}

func (cmd *EssenceCommand) Exec(c *qbot.Client, args []string, src *srcMsg, _ int) {
	if !slices.Contains(config.Cfg.Permissions.BotOwnerGroupIDs, src.GroupID) {
		return
	}

	// 查找 --reply= 参数
	var msgID uint64
	for _, arg := range args {
		if strings.HasPrefix(arg, "--reply=") {
			if id, err := strconv.ParseUint(strings.TrimPrefix(arg, "--reply="), 10, 64); err == nil {
				msgID = id
				break
			}
		}
	}

	if msgID == 0 {
		c.SendMsg(src.GroupID, src.UserID, cmd.HelpMsg)
		return
	}

	if len(args) == 2 {
		switch args[1] {
		case "rm":
			c.DeleteGroupEssence(msgID)
		case "add":
			c.SetGroupEssence(msgID)
		default:
			c.SendMsg(src.GroupID, src.UserID, cmd.HelpMsg)
		}
	} else if len(args) == 1 {
		c.SetGroupEssence(msgID)
	} else {
		c.SendMsg(src.GroupID, src.UserID, cmd.HelpMsg)
	}
}
