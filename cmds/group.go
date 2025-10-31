package cmds

import (
	"fmt"
	"go-hurobot/config"
	"go-hurobot/qbot"
	"slices"
	"strconv"
	"strings"
)

const groupHelpMsg string = `Manage group settings.
Usage: group [rename <name> | op [@users...] | deop [@users...] | banme <minutes> | ban @user <minutes>]
Examples:
  /group rename awa
  /group op @user1 @user2`

type GroupCommand struct {
	cmdBase
}

func NewGroupCommand() *GroupCommand {
	return &GroupCommand{
		cmdBase: cmdBase{
			Name:        "group",
			HelpMsg:     groupHelpMsg,
			Permission:  getCmdPermLevel("group"),
			AllowPrefix: false,
			NeedRawMsg:  false,
			MaxArgs:     4,
			MinArgs:     2,
		},
	}
}

func (cmd *GroupCommand) Self() *cmdBase {
	return &cmd.cmdBase
}

func (cmd *GroupCommand) Exec(c *qbot.Client, args []string, src *srcMsg, begin int) {
	if !slices.Contains(config.Cfg.Permissions.BotOwnerGroupIDs, src.GroupID) {
		return
	}
	if len(args) == 1 {
		c.SendMsg(src.GroupID, src.UserID, cmd.HelpMsg)
		return
	}
	switch args[1] {
	case "rename":
		newName := decodeSpecialChars(strings.Join(args[2:], " "))
		c.SendMsg(src.GroupID, src.UserID, fmt.Sprintf("rename: %q", newName))
		c.SetGroupName(src.GroupID, newName)
	case "op":
		setGroupAdmin(c, src, args, true)
	case "deop":
		setGroupAdmin(c, src, args, false)
	case "ban":
		time, err := strconv.Atoi(args[3])
		if err != nil || time < 1 || time > 24*60*30 {
			c.SendMsg(src.GroupID, src.UserID, "Invalid time duration")
			return
		}
		c.SetGroupBan(src.GroupID, str2uin64(strings.TrimPrefix(args[2], "--at=")), time*60)
	}
}

func setGroupAdmin(c *qbot.Client, src *srcMsg, args []string, isOp bool) {
	targetUserIDs, err := extractTargetUsers(args, 2, src.UserID)
	if err != nil {
		c.SendMsg(src.GroupID, src.UserID, "Invalid argument: "+err.Error())
		return
	}

	validUserIDs := make([]uint64, 0, len(targetUserIDs))
	userIDSet := make(map[uint64]bool)

	action := "op"
	if !isOp {
		action = "deop"
	}

	for _, userID := range targetUserIDs {
		if userID == config.Cfg.Permissions.BotID {
			c.SendMsg(src.GroupID, src.UserID, fmt.Sprintf("Cannot %s bot", action))
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
		c.SetGroupAdmin(src.GroupID, userID, isOp)
	}

	if len(validUserIDs) == 1 {
		c.SendMsg(src.GroupID, src.UserID, fmt.Sprintf("%s: %d", action, validUserIDs[0]))
	} else {
		userIDStrings := make([]string, len(validUserIDs))
		for i, id := range validUserIDs {
			userIDStrings[i] = strconv.FormatUint(id, 10)
		}
		c.SendMsg(src.GroupID, src.UserID, fmt.Sprintf("%s: %s", action, strings.Join(userIDStrings, ", ")))
	}
}

func extractTargetUsers(args []string, startIndex int, defaultUserID uint64) ([]uint64, error) {
	var targetUserIDs []uint64
	hasAtUsers := false

	for i := startIndex; i < len(args); i++ {
		if strings.HasPrefix(args[i], "--at=") {
			hasAtUsers = true
			targetUserIDs = append(targetUserIDs, str2uin64(strings.TrimPrefix(args[i], "--at=")))
		} else {
			return nil, fmt.Errorf("use @ to mention users")
		}
	}

	if !hasAtUsers {
		targetUserIDs = append(targetUserIDs, defaultUserID)
	}

	return targetUserIDs, nil
}
