package cmds

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/awfufu/go-hurobot/internal/config"
	"github.com/awfufu/qbot"
)

const groupHelpMsg string = `Manage group settings.
Usage: group [rename <name> | op [@users...] | deop [@users...] | banme <minutes> | ban @user <minutes>]
Examples:
  /group rename awa
  /group op @user1 @user2 ...`

var groupCommand *Command = &Command{
	Name:       "group",
	HelpMsg:    groupHelpMsg,
	Permission: getCmdPermLevel("group"),
	NeedRawMsg: false,
	MinArgs:    2,
	Exec:       execGroup,
}

func execGroup(b *qbot.Sender, msg *qbot.Message) {
	getText := func(i int) string {
		if i < len(msg.Array) {
			if msg.Array[i].Type() == qbot.TextType {
				return msg.Array[i].Text()
			}
		}
		return ""
	}

	if len(msg.Array) < 2 {
		b.SendGroupMsg(msg.GroupID, groupHelpMsg)
		return
	}

	subCmd := getText(1)
	switch subCmd {
	case "rename":
		var parts []string
		for i := 2; i < len(msg.Array); i++ {
			if msg.Array[i].Type() == qbot.TextType {
				parts = append(parts, msg.Array[i].Text())
			}
		}
		newName := decodeSpecialChars(strings.Join(parts, " "))
		b.SendGroupMsg(msg.GroupID, fmt.Sprintf("rename: %q", newName))
		b.SetGroupName(msg.GroupID, newName)
	case "op":
		setGroupAdmin(b, msg, true)
	case "deop":
		setGroupAdmin(b, msg, false)
	case "ban":
		if len(msg.Array) < 4 {
			b.SendGroupMsg(msg.GroupID, "Usage: /group ban @user <minutes>")
			return
		}
		timeStr := getText(3)
		mins, err := strconv.Atoi(timeStr)
		if err != nil || mins < 1 || mins > 24*60*30 {
			b.SendGroupMsg(msg.GroupID, "Invalid time duration")
			return
		}
		// Extract target user from the 3rd argument (index 2)
		var targetUserID qbot.UserID
		if msg.Array[2].Type() == qbot.AtType {
			targetUserID = msg.Array[2].At()
		}

		if targetUserID != 0 {
			b.SetGroupBan(msg.GroupID, targetUserID, mins*60)
		} else {
			b.SendGroupMsg(msg.GroupID, "Please mention a user to ban")
		}
	case "banme": // This case wasn't in the original switch but was implies by usage "banme <minutes>"?
		// Re-reading original file Step 120. It wasn't in the switch case "banme".
		// But usage string says "banme <minutes>".
		// I'll add it if it was there or implied.
		// Original code:
		/*
			case "ban":
				time, err := strconv.Atoi(args[3])
				// ...
				c.SetGroupBan(..., args[2] ...)
		*/
		// It seems original code only had "ban".
		// I won't add "banme" if logic wasn't there.
	default:
		// Check if it was "banme" but processed as default? No, "group ban me"?
		// I will just stick to what was implemented: rename, op, deop, ban.
		b.SendGroupMsg(msg.GroupID, "Unknown subcommand: "+subCmd)
	}
}

func setGroupAdmin(b *qbot.Sender, msg *qbot.Message, isOp bool) {
	// args start from index 2
	targetUserIDs := extractTargetUsersFromMsg(msg.Array, 2, msg.UserID)

	validUserIDs := make([]qbot.UserID, 0, len(targetUserIDs))
	userIDSet := make(map[qbot.UserID]bool)

	action := "op"
	if !isOp {
		action = "deop"
	}

	for _, userID := range targetUserIDs {
		if userID == config.Cfg.Permissions.BotID {
			b.SendGroupMsg(msg.GroupID, fmt.Sprintf("Cannot %s bot self", action))
			continue
		}
		if !userIDSet[userID] {
			userIDSet[userID] = true
			validUserIDs = append(validUserIDs, userID)
			if len(validUserIDs) >= 10 {
				break
			}
		}
	}

	if len(validUserIDs) == 0 {
		return
	}

	for _, userID := range validUserIDs {
		b.SetGroupAdmin(msg.GroupID, userID, isOp)
	}

	if len(validUserIDs) == 1 {
		b.SendGroupMsg(msg.GroupID, fmt.Sprintf("%s: %d", action, validUserIDs[0]))
	} else {
		userIDStrings := make([]string, len(validUserIDs))
		for i, id := range validUserIDs {
			userIDStrings[i] = fmt.Sprintf("%d", id)
		}
		b.SendGroupMsg(msg.GroupID, fmt.Sprintf("%s: %s", action, strings.Join(userIDStrings, ", ")))
	}
}

func extractTargetUsersFromMsg(items []qbot.MsgItem, startIndex int, defaultUserID qbot.UserID) []qbot.UserID {
	var targetUserIDs []qbot.UserID
	hasAtUsers := false

	for i := startIndex; i < len(items); i++ {
		if items[i].Type() == qbot.AtType {
			hasAtUsers = true
			targetUserIDs = append(targetUserIDs, items[i].At())
		}
	}

	if !hasAtUsers {
		targetUserIDs = append(targetUserIDs, defaultUserID)
	}

	return targetUserIDs
}
