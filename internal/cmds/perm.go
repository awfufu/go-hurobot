package cmds

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/awfufu/go-hurobot/internal/config"
	"github.com/awfufu/go-hurobot/internal/db"
	"github.com/awfufu/qbot"
)

const permHelpMsg string = `Manage command permissions.
Usage: perm <subcommand> [args...]
Subcommands:
  set <cmd> <key> <value>
    Keys: user_allow (0-2/guest/admin/master), group_enable (0-1), whitelist_user (0-1), whitelist_group (0-1)
  special <cmd> <user|group> <add|rm|list> [targets...]
  user <target> <level>
    Levels: 0/guest, 1/admin, 2/master
Examples:
  /perm set draw user_allow 0
  /perm set draw group_enable 1
  /perm special draw user add @user
  /perm user @user admin`

type PermCommand struct {
	cmdBase
}

func NewPermCommand() *PermCommand {
	return &PermCommand{
		cmdBase: cmdBase{
			Name:       "perm",
			HelpMsg:    permHelpMsg,
			Permission: config.Master, // Only master can manage permissions
			NeedRawMsg: false,
			MinArgs:    2,
		},
	}
}

func (cmd *PermCommand) Self() *cmdBase {
	return &cmd.cmdBase
}

func (cmd *PermCommand) Exec(b *qbot.Sender, msg *qbot.Message) {
	if len(msg.Array) < 2 {
		b.SendGroupMsg(msg.GroupID, cmd.HelpMsg)
		return
	}

	getText := func(i int) string {
		if i < len(msg.Array) {
			if msg.Array[i].Type() == qbot.TextType {
				return msg.Array[i].Text()
			}
		}
		return ""
	}

	subCmd := getText(1)
	switch subCmd {
	case "set":
		cmd.handleSet(b, msg)
	case "special":
		cmd.handleSpecial(b, msg)
	case "user":
		cmd.handleUserRole(b, msg)
	default:
		b.SendGroupMsg(msg.GroupID, "Unknown subcommand: "+subCmd)
	}
}

func (cmd *PermCommand) handleSet(b *qbot.Sender, msg *qbot.Message) {
	getText := func(i int) string {
		if i < len(msg.Array) {
			if msg.Array[i].Type() == qbot.TextType {
				return msg.Array[i].Text()
			}
		}
		return ""
	}

	if len(msg.Array) < 5 {
		b.SendGroupMsg(msg.GroupID, "Usage: perm set <cmd> <key> <value>")
		return
	}

	cmdName := getText(2)
	key := getText(3)
	value := getText(4)

	perm := db.GetCommandPermission(cmdName)
	if perm == nil {
		// Initialize if not exists (should theoretically exist from startup, but safe fallback)
		perm = &db.DbPermissions{
			Command:           cmdName,
			UserAllow:         2,
			SpecialUsers:      "",
			IsWhitelistUsers:  0,
			SpecialGroups:     "",
			IsWhitelistGroups: 0,
		}
	}

	switch key {
	case "user_allow":
		valInt := 2
		switch value {
		case "0", "guest":
			valInt = 0
		case "1", "admin":
			valInt = 1
		case "2", "master":
			valInt = 2
		default:
			b.SendGroupMsg(msg.GroupID, "Invalid user_allow. Use 0/guest, 1/admin, 2/master")
			return
		}
		perm.UserAllow = valInt
	case "whitelist_user":
		if value == "1" || value == "enable" || value == "true" {
			perm.IsWhitelistUsers = 1
		} else {
			perm.IsWhitelistUsers = 0
		}
	case "whitelist_group":
		if value == "1" || value == "enable" || value == "true" {
			perm.IsWhitelistGroups = 1
		} else {
			perm.IsWhitelistGroups = 0
		}
	default:
		b.SendGroupMsg(msg.GroupID, "Unknown key: "+key)
		return
	}

	if err := db.SaveCommandPermission(perm); err != nil {
		b.SendGroupMsg(msg.GroupID, "Failed to save permission: "+err.Error())
	} else {
		b.SendGroupMsg(msg.GroupID, fmt.Sprintf("Updated %s %s to %s", cmdName, key, value))
	}
}

func (cmd *PermCommand) handleSpecial(b *qbot.Sender, msg *qbot.Message) {
	getText := func(i int) string {
		if i < len(msg.Array) {
			if msg.Array[i].Type() == qbot.TextType {
				return msg.Array[i].Text()
			}
		}
		return ""
	}

	// special <cmd> <user|group> <add|rm|list> [targets...]
	// 0       1     2            3             4
	if len(msg.Array) < 5 {
		b.SendGroupMsg(msg.GroupID, "Usage: perm special <cmd> <user|group> <add|rm|list> [targets...]")
		return
	}

	cmdName := getText(2)
	targetType := getText(3)
	action := getText(4)

	if targetType != "user" && targetType != "group" {
		b.SendGroupMsg(msg.GroupID, "Invalid target type. Must be user or group.")
		return
	}

	var currentList []uint64
	var rawStr string

	perm := db.GetCommandPermission(cmdName)
	if perm == nil {
		perm = &db.DbPermissions{
			Command:           cmdName,
			UserAllow:         2,
			SpecialUsers:      "",
			IsWhitelistUsers:  0,
			SpecialGroups:     "",
			IsWhitelistGroups: 0,
		}
	} else {
		if targetType == "user" {
			rawStr = perm.SpecialUsers
		} else {
			rawStr = perm.SpecialGroups
		}
	}

	currentList = db.ParseIDList(rawStr)

	switch action {
	case "add":
		targets := extractTargets(msg.Array, targetType)
		if len(targets) == 0 {
			b.SendGroupMsg(msg.GroupID, fmt.Sprintf("No %ss specified.", targetType))
			return
		}
		count := 0
		for _, t := range targets {
			if !slices.Contains(currentList, uint64(t)) {
				currentList = append(currentList, uint64(t))
				count++
			}
		}
		b.SendGroupMsg(msg.GroupID, fmt.Sprintf("Added %d %ss.", count, targetType))

	case "rm":
		targets := extractTargets(msg.Array, targetType)
		if len(targets) == 0 {
			b.SendGroupMsg(msg.GroupID, fmt.Sprintf("No %ss specified.", targetType))
			return
		}
		count := 0
		for _, t := range targets {
			idx := slices.Index(currentList, uint64(t))
			if idx != -1 {
				currentList = append(currentList[:idx], currentList[idx+1:]...)
				count++
			}
		}
		b.SendGroupMsg(msg.GroupID, fmt.Sprintf("Removed %d %ss.", count, targetType))

	case "list":
		if len(currentList) == 0 {
			b.SendGroupMsg(msg.GroupID, "List is empty.")
		} else {
			strs := make([]string, len(currentList))
			for i, v := range currentList {
				strs[i] = strconv.FormatUint(v, 10)
			}
			b.SendGroupMsg(msg.GroupID, strings.Join(strs, ", "))
		}
		return // No save needed for list

	default:
		b.SendGroupMsg(msg.GroupID, "Unknown action: "+action)
		return
	}

	// Update struct
	newStr := db.JoinIDList(currentList)

	if targetType == "user" {
		perm.SpecialUsers = newStr
	} else {
		perm.SpecialGroups = newStr
	}

	if err := db.SaveCommandPermission(perm); err != nil {
		b.SendGroupMsg(msg.GroupID, "Failed to save: "+err.Error())
	} else {
		b.SendGroupMsg(msg.GroupID, "Permission updated.")
	}
}

func extractTargets(args []qbot.MsgItem, targetType string) []uint64 {
	var targets []uint64
	for _, arg := range args {
		if targetType == "user" {
			if arg.Type() == qbot.AtType {
				targets = append(targets, uint64(arg.At()))
			} else if arg.Type() == qbot.TextType {
				// Try to parse raw ID from text if provided as argument
				parts := strings.Fields(arg.Text())
				for _, part := range parts {
					if id, err := strconv.ParseUint(part, 10, 64); err == nil {
						targets = append(targets, id)
					}
				}
			}
		} else {
			// Groups are usually just numbers in text
			if arg.Type() == qbot.TextType {
				parts := strings.Fields(arg.Text())
				for _, part := range parts {
					if id, err := strconv.ParseUint(part, 10, 64); err == nil {
						targets = append(targets, id)
					}
				}
			}
		}
	}
	return targets
}

func (cmd *PermCommand) handleUserRole(b *qbot.Sender, msg *qbot.Message) {
	getText := func(i int) string {
		if i < len(msg.Array) {
			if msg.Array[i].Type() == qbot.TextType {
				return msg.Array[i].Text()
			}
		}
		return ""
	}

	// user <target> <level>
	if len(msg.Array) < 4 {
		b.SendGroupMsg(msg.GroupID, "Usage: perm user <target> <level>\nLevels: 0/guest, 1/admin, 2/master")
		return
	}

	// Target should be an At item or Int ID
	targetID := qbot.InvalidUser
	if msg.Array[2].Type() == qbot.AtType {
		targetID = msg.Array[2].At()
	} else if msg.Array[2].Type() == qbot.TextType {
		if id, err := strconv.ParseUint(msg.Array[2].Text(), 10, 64); err == nil {
			targetID = qbot.UserID(id)
		}
	}

	if targetID == qbot.InvalidUser {
		b.SendGroupMsg(msg.GroupID, "Invalid user target.")
		return
	}

	levelStr := getText(3)
	var level int
	switch levelStr {
	case "0", "guest":
		level = 0
	case "1", "admin":
		level = 1
	case "2", "master":
		level = 2
	default:
		b.SendGroupMsg(msg.GroupID, "Invalid level. Usage: 0/guest, 1/admin, 2/master")
		return
	}

	if err := db.UpdateUserPerm(uint64(targetID), level); err != nil {
		b.SendGroupMsg(msg.GroupID, "Failed to update user role: "+err.Error())
	} else {
		b.SendGroupMsg(msg.GroupID, fmt.Sprintf("Updated user %d role to %d", targetID, level))
	}
}
