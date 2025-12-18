package cmds

import (
	"strconv"
	"strings"

	"github.com/awfufu/qbot"
)

const specialtitleHelpMsg string = `Set special title for group members.
Usage: /specialtitle [@user] <title>
Example: /specialtitle @user qwq`

var specialtitleCommand *Command = &Command{
	Name:       "specialtitle",
	HelpMsg:    specialtitleHelpMsg,
	Permission: getCmdPermLevel("specialtitle"),
	NeedRawMsg: false,
	MaxArgs:    3,
	MinArgs:    2,
	Exec:       execSpecialTitle,
}

func execSpecialTitle(b *qbot.Sender, msg *qbot.Message) {
	var targetUserID qbot.UserID
	var title string

	if len(msg.Array) == 3 {
		// Case 1: Command + 2 arguments
		// Check invalid inputs (Text + Text) or (At + At)

		if msg.Array[1].Type() == qbot.AtType {
			targetUserID = msg.Array[1].At()
			if msg.Array[2].Type() == qbot.TextType {
				title = msg.Array[2].Text()
			} else {
				b.SendGroupReplyMsg(msg.GroupID, msg.MsgID, "Arguments MUST be one @user and one Text title")
				return
			}
		} else if msg.Array[2].Type() == qbot.AtType {
			targetUserID = msg.Array[2].At()
			if msg.Array[1].Type() == qbot.TextType {
				title = msg.Array[1].Text()
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
		if msg.Array[1].Type() == qbot.TextType {
			// Check if the text looks like a mention (e.g. "@123456")
			if strings.HasPrefix(msg.Array[1].Text(), "@") && isNumeric(msg.Array[1].Text()[1:]) {
				b.SendGroupReplyMsg(msg.GroupID, msg.MsgID, "Arguments MUST be one @user and one Text title. If you are trying to mention a user, please ensure it is a valid mention.")
				return
			}

			title = msg.Array[1].Text()
			targetUserID = msg.UserID
		} else {
			b.SendGroupReplyMsg(msg.GroupID, msg.MsgID, "When setting title for self, please provide text only")
			return
		}
	} else {
		// Should be handled by MaxArgs/MinArgs, but safe fallback
		b.SendGroupReplyMsg(msg.GroupID, msg.MsgID, specialtitleHelpMsg)
		return
	}

	if length := len([]byte(title)); length > 18 {
		b.SendGroupReplyMsg(msg.GroupID, msg.MsgID, "Title length not allowed to exceed 18 bytes, currently "+strconv.FormatInt(int64(length), 10)+" bytes")
		return
	}

	b.SetGroupSpecialTitle(msg.GroupID, targetUserID, decodeSpecialChars(title))
}
