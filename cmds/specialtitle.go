package cmds

import (
	"go-hurobot/config"
	"go-hurobot/qbot"
	"slices"
	"strconv"
	"strings"
)

const specialtitleHelpMsg string = `Set special title for group members.
Usage: /specialtitle [@user] <title>
Example: /specialtitle @user VIP`

type SpecialtitleCommand struct {
	cmdBase
}

func NewSpecialtitleCommand() *SpecialtitleCommand {
	return &SpecialtitleCommand{
		cmdBase: cmdBase{
			Name:        "specialtitle",
			HelpMsg:     specialtitleHelpMsg,
			Permission:  getCmdPermLevel("specialtitle"),
			AllowPrefix: false,
			NeedRawMsg:  false,
			MaxArgs:     3,
			MinArgs:     2,
		},
	}
}

func (cmd *SpecialtitleCommand) Self() *cmdBase {
	return &cmd.cmdBase
}

func (cmd *SpecialtitleCommand) Exec(c *qbot.Client, args []string, src *srcMsg, _ int) {
	if !slices.Contains(config.Cfg.Permissions.BotOwnerGroupIDs, src.GroupID) {
		return
	}

	// check if the second parameter is @
	var targetUserID uint64
	var titleIdx int
	if len(args) > 1 && strings.HasPrefix(args[1], "--at=") {
		targetUserID = str2uin64(strings.TrimPrefix(args[1], "--at="))
		titleIdx = 2
	} else {
		targetUserID = src.UserID
		titleIdx = 1
	}

	if titleIdx >= len(args) {
		c.SendMsg(src.GroupID, src.UserID, qbot.CQReply(src.MsgID)+cmd.HelpMsg)
		return
	}

	title := args[titleIdx]
	if length := len([]byte(title)); length > 18 {
		c.SendMsg(src.GroupID, src.UserID, qbot.CQReply(src.MsgID)+"Title length not allowed to exceed 18 bytes, currently "+strconv.FormatInt(int64(length), 10)+" bytes")
		return
	}

	c.SetGroupSpecialTitle(src.GroupID, targetUserID, decodeSpecialChars(title))
}
