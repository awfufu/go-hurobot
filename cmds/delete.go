package cmds

import (
	"go-hurobot/qbot"
	"log"
	"strconv"
	"strings"
)

const deleteHelpMsg = `Delete a message by replying to it.
Usage: [Reply to a message] /delete`

type DeleteCommand struct {
	cmdBase
}

func NewDeleteCommand() *DeleteCommand {
	return &DeleteCommand{
		cmdBase: cmdBase{
			Name:        "delete",
			HelpMsg:     deleteHelpMsg,
			Permission:  getCmdPermLevel("delete"),
			AllowPrefix: true, // Allow prefix
			NeedRawMsg:  false,
			MaxArgs:     1,
			MinArgs:     1,
		},
	}
}

func (cmd *DeleteCommand) Self() *cmdBase {
	return &cmd.cmdBase
}

func (cmd *DeleteCommand) Exec(c *qbot.Client, args []string, src *srcMsg, _ int) {
	// Check for --reply= parameter
	var replyMsgID uint64
	if after, ok := strings.CutPrefix(args[0], "--reply="); ok {
		if msgid, err := strconv.ParseUint(after, 10, 64); err == nil {
			replyMsgID = msgid
		}
	}

	if replyMsgID != 0 {
		c.DeleteMsg(replyMsgID)
		log.Printf("delete message %d", replyMsgID)
	} else {
		c.SendMsg(src.GroupID, src.UserID, "Please reply to a message to delete it, and ensure the bot has permission to delete it")
	}
}
