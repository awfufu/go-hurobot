package cmds

import (
	"github.com/awfufu/qbot"
)

const essenceHelpMsg string = `Manage essence messages.
Usage: [Reply to a message] /essence [add|rm]`

var essenceCommand *Command = &Command{
	Name:       "essence",
	HelpMsg:    essenceHelpMsg,
	Permission: getCmdPermLevel("essence"),
	NeedRawMsg: false,
	Exec:       execEssence,
}

func execEssence(b *qbot.Sender, msg *qbot.Message) {
	if msg.ReplyID == 0 {
		b.SendGroupMsg(msg.GroupID, essenceHelpMsg)
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
				b.SendGroupMsg(msg.GroupID, essenceHelpMsg)
			}
		} else {
			b.SendGroupMsg(msg.GroupID, essenceHelpMsg)
		}
	} else {
		b.SetGroupEssence(msg.ReplyID)
	}
}
