package cmds

import (
	"slices"

	"github.com/awfufu/go-hurobot/config"
	"github.com/awfufu/qbot"
)

const essenceHelpMsg string = `Manage essence messages.
Usage: [Reply to a message] /essence [add|rm]`

type EssenceCommand struct {
	cmdBase
}

func NewEssenceCommand() *EssenceCommand {
	return &EssenceCommand{
		cmdBase: cmdBase{
			Name:       "essence",
			HelpMsg:    essenceHelpMsg,
			Permission: getCmdPermLevel("essence"),

			NeedRawMsg: false,
		},
	}
}

func (cmd *EssenceCommand) Self() *cmdBase {
	return &cmd.cmdBase
}

func (cmd *EssenceCommand) Exec(b *qbot.Bot, msg *qbot.Message) {
	if !slices.Contains(config.Cfg.Permissions.BotOwnerGroupIDs, msg.GroupID) {
		return
	}

	if msg.ReplyID == 0 {
		b.SendGroupMsg(msg.GroupID, cmd.HelpMsg)
		return
	}

	if len(msg.Array) >= 2 {
		if txt := msg.Array[1].GetTextItem(); txt != nil {
			switch txt.Content {
			case "rm":
				b.DeleteGroupEssence(msg.ReplyID)
			case "add":
				b.SetGroupEssence(msg.ReplyID)
			default:
				b.SendGroupMsg(msg.GroupID, cmd.HelpMsg)
			}
		} else {
			b.SendGroupMsg(msg.GroupID, cmd.HelpMsg)
		}
	} else {
		b.SetGroupEssence(msg.ReplyID)
	}
}
