package cmds

import (
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

func (cmd *EssenceCommand) Exec(b *qbot.Sender, msg *qbot.Message) {
	if msg.ReplyID == 0 {
		b.SendGroupMsg(msg.GroupID, cmd.HelpMsg)
		return
	}

	if len(msg.Array) >= 2 {
		if msg.Array[1].Type() == qbot.TextType {
			switch msg.Array[1].Text() {
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
