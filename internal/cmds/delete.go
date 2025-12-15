package cmds

import (
	"log"

	"github.com/awfufu/qbot"
)

const deleteHelpMsg = `Delete a message by replying to it.
Usage: [Reply to a message] /delete`

type DeleteCommand struct {
	cmdBase
}

func NewDeleteCommand() *DeleteCommand {
	return &DeleteCommand{
		cmdBase: cmdBase{
			Name:       "delete",
			HelpMsg:    deleteHelpMsg,
			Permission: getCmdPermLevel("delete"),
			NeedRawMsg: false,
			MaxArgs:    1,
			MinArgs:    1,
		},
	}
}

func (cmd *DeleteCommand) Self() *cmdBase {
	return &cmd.cmdBase
}

func (cmd *DeleteCommand) Exec(b *qbot.Bot, msg *qbot.Message) {
	if msg.ReplyID != 0 {
		b.DeleteMsg(msg.ReplyID)
		log.Printf("delete message %d", msg.ReplyID)
	} else {
		b.SendGroupMsg(msg.GroupID, "Please reply to a message to delete it, and ensure the bot has permission to delete it")
	}
}
