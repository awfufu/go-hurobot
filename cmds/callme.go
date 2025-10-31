package cmds

import (
	"go-hurobot/qbot"
	"strings"
)

var callmeHelpMsg string = `Set or query your nickname.
Usage:
 /callme
 /callme <nickname>
 /callme <@user>
 /callme <@user> <nickname>`

type CallmeCommand struct {
	cmdBase
}

func NewCallmeCommand() *CallmeCommand {
	return &CallmeCommand{
		cmdBase: cmdBase{
			Name:        "callme",
			HelpMsg:     callmeHelpMsg,
			Permission:  getCmdPermLevel("callme"),
			AllowPrefix: false,
			NeedRawMsg:  false,
			MaxArgs:     3, // callme <nickname> <@id>
			MinArgs:     1, // callme
		},
	}
}

func (cmd *CallmeCommand) Self() *cmdBase {
	return &cmd.cmdBase
}

func (cmd *CallmeCommand) Exec(c *qbot.Client, args []string, src *srcMsg, _ int) {
	var targetID uint64
	var nickname string
	var isQuery bool = true

	switch len(args) {
	case 1: // callme
		targetID = src.UserID
	case 2: // callme <nickname> / callme <@id>
		if strings.HasPrefix(args[1], "--at=") {
			// callme <@id>
			targetID = str2uin64(strings.TrimPrefix(args[1], "--at="))
		} else {
			// callme <nickname>
			targetID = src.UserID
			nickname = args[1]
			isQuery = false
		}
	case 3: // callme <nickname> <@id>
		isQuery = false
		if userid, ok := strings.CutPrefix(args[2], "--at="); ok {
			targetID = str2uin64(userid)
			nickname = args[1]
		} else {
			return
		}
	default:
		c.SendMsg(src.GroupID, src.UserID, callmeHelpMsg)
		return
	}

	if isQuery {
		var user qbot.Users
		result := qbot.PsqlDB.Where("user_id = ?", targetID).First(&user)
		if result.Error != nil || user.Nickname == "" {
			c.SendMsg(src.GroupID, src.UserID, "")
			return
		}
		c.SendMsg(src.GroupID, src.UserID, user.Nickname)
	} else {
		user := qbot.Users{
			UserID:   targetID,
			Name:     nickname,
			Nickname: nickname,
		}

		result := qbot.PsqlDB.Where("user_id = ?", targetID).Assign(
			qbot.Users{Nickname: nickname},
		).FirstOrCreate(&user)

		if result.Error != nil {
			c.SendMsg(src.GroupID, src.UserID, "failed")
			return
		}
		if targetID == src.UserID {
			c.SendMsg(src.GroupID, src.UserID, "Update nickname: "+nickname)
		} else {
			c.SendMsg(src.GroupID, src.UserID, "Update nickname for ["+qbot.CQAt(targetID)+"]: "+nickname)
		}
	}
}
