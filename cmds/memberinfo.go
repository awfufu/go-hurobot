package cmds

import (
	"fmt"
	"go-hurobot/qbot"
	"strings"
	"time"
)

const memberinfoHelpMsg string = `Query group member information.
Usage: /memberinfo [@user]
Example: /memberinfo @user`

type MemberinfoCommand struct {
	cmdBase
}

func NewMemberinfoCommand() *MemberinfoCommand {
	return &MemberinfoCommand{
		cmdBase: cmdBase{
			Name:        "memberinfo",
			HelpMsg:     memberinfoHelpMsg,
			Permission:  getCmdPermLevel("memberinfo"),
			AllowPrefix: false,
			NeedRawMsg:  false,
			MaxArgs:     2,
			MinArgs:     2,
		},
	}
}

func (cmd *MemberinfoCommand) Self() *cmdBase {
	return &cmd.cmdBase
}

func (cmd *MemberinfoCommand) Exec(c *qbot.Client, args []string, src *srcMsg, begin int) {
	// Only available in group chats
	if src.GroupID == 0 {
		return
	}

	var targetUserID uint64

	if len(args) >= 2 && strings.HasPrefix(args[1], "--at=") {
		targetUserID = str2uin64(strings.TrimPrefix(args[1], "--at="))
	} else {
		targetUserID = src.UserID
	}

	if targetUserID == 0 {
		c.SendMsg(src.GroupID, src.UserID, qbot.CQReply(src.MsgID)+"Invalid user ID")
		return
	}

	// Get group member information
	memberInfo, err := c.GetGroupMemberInfo(src.GroupID, targetUserID, false)
	if err != nil {
		c.SendMsg(src.GroupID, src.UserID, qbot.CQReply(src.MsgID)+fmt.Sprintf("Failed to get member info: %v", err))
		return
	}

	response := fmt.Sprintf(
		"QQ: %d\n"+
			"Nickname: %s\n"+
			"Card: %s\n"+
			"Gender: %s\n"+
			"Role: %s\n"+
			"Level: Lv %s",
		memberInfo.UserID,
		memberInfo.Nickname,
		getCardOrNickname(memberInfo.Card, memberInfo.Nickname),
		getSexString(memberInfo.Sex),
		getRoleString(memberInfo.Role),
		memberInfo.Level)

	if memberInfo.Age > 0 {
		response += fmt.Sprintf("\nAge: %d", memberInfo.Age)
	}

	if memberInfo.Area != "" {
		response += fmt.Sprintf("\nArea: %s", memberInfo.Area)
	}

	if memberInfo.Title != "" {
		response += fmt.Sprintf("\nTitle: %s", memberInfo.Title)
	}

	if memberInfo.ShutUpTimestamp > 0 {
		shutUpTime := time.Unix(memberInfo.ShutUpTimestamp, 0)
		if shutUpTime.After(time.Now()) {
			response += fmt.Sprintf("\nMuted until: %s", shutUpTime.Format("2006-01-02 15:04:05"))
		}
	}

	if memberInfo.JoinTime > 0 {
		joinTime := time.Unix(int64(memberInfo.JoinTime), 0)
		response += fmt.Sprintf("\nJoined: %s", joinTime.Format("2006-01-02 15:04:05"))
	}

	c.SendMsg(src.GroupID, src.UserID, qbot.CQReply(src.MsgID)+response)
}

func getCardOrNickname(card, nickname string) string {
	if card != "" {
		return card
	}
	return nickname
}

func getSexString(sex string) string {
	switch sex {
	case "male":
		return "♂"
	case "female":
		return "♀"
	default:
		return "?"
	}
}

func getRoleString(role string) string {
	switch role {
	case "owner":
		return "Owner"
	case "admin":
		return "Admin"
	case "member":
		return "Member"
	default:
		return role
	}
}
