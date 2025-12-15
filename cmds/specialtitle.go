package cmds

import (
	"slices"
	"strconv"
	"strings"

	"github.com/awfufu/go-hurobot/config"
	"github.com/awfufu/qbot"
)

const specialtitleHelpMsg string = `Set special title for group members.
Usage: /specialtitle [@user] <title>
Example: /specialtitle @user qwq`

type SpecialtitleCommand struct {
	cmdBase
}

func NewSpecialtitleCommand() *SpecialtitleCommand {
	return &SpecialtitleCommand{
		cmdBase: cmdBase{
			Name:       "specialtitle",
			HelpMsg:    specialtitleHelpMsg,
			Permission: getCmdPermLevel("specialtitle"),

			NeedRawMsg: false,
			MaxArgs:    3,
			MinArgs:    2,
		},
	}
}

func (cmd *SpecialtitleCommand) Self() *cmdBase {
	return &cmd.cmdBase
}

func (cmd *SpecialtitleCommand) Exec(b *qbot.Bot, msg *qbot.Message) {
	if !slices.Contains(config.Cfg.Permissions.BotOwnerGroupIDs, msg.GroupID) {
		return
	}

	var targetUserID uint64
	var title string

	if len(msg.Array) == 3 {
		// Case 1: Command + 2 arguments
		// Check invalid inputs (Text + Text) or (At + At)
		item1 := msg.Array[1]
		item2 := msg.Array[2]

		if at1 := item1.GetAtItem(); at1 != nil {
			targetUserID = at1.TargetID
			if txt2 := item2.GetTextItem(); txt2 != nil {
				title = txt2.Content
			} else {
				b.SendGroupReplyMsg(msg.GroupID, msg.MsgID, "Arguments MUST be one @user and one Text title")
				return
			}
		} else if at2 := item2.GetAtItem(); at2 != nil {
			targetUserID = at2.TargetID
			if txt1 := item1.GetTextItem(); txt1 != nil {
				title = txt1.Content
			} else {
				b.SendGroupReplyMsg(msg.GroupID, msg.MsgID, "Arguments MUST be one @user and one Text title")
				return
			}
		} else {
			// No At item found, both are Text or other types
			b.SendGroupReplyMsg(msg.GroupID, msg.MsgID, "Please mention a user to set title for")
			return
		}

	} else if len(msg.Array) == 2 {
		// Case 2: Command + 1 argument
		// The argument must be Text
		item := msg.Array[1]
		if txt := item.GetTextItem(); txt != nil {
			// Check if the text looks like a mention (e.g. "@123456")
			if strings.HasPrefix(txt.Content, "@") && isNumeric(txt.Content[1:]) {
				b.SendGroupReplyMsg(msg.GroupID, msg.MsgID, "Arguments MUST be one @user and one Text title. If you are trying to mention a user, please ensure it is a valid mention.")
				return
			}

			title = txt.Content
			targetUserID = msg.UserID
		} else {
			b.SendGroupReplyMsg(msg.GroupID, msg.MsgID, "When setting title for self, please provide text only")
			return
		}
	} else {
		// Should be handled by MaxArgs/MinArgs, but safe fallback
		b.SendGroupReplyMsg(msg.GroupID, msg.MsgID, cmd.HelpMsg)
		return
	}

	if length := len([]byte(title)); length > 18 {
		b.SendGroupReplyMsg(msg.GroupID, msg.MsgID, "Title length not allowed to exceed 18 bytes, currently "+strconv.FormatInt(int64(length), 10)+" bytes")
		return
	}

	b.SetGroupSpecialTitle(msg.GroupID, targetUserID, decodeSpecialChars(title))
}
