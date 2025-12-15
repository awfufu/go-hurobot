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
Examples:
  /perm set draw user_allow 0
  /perm set draw group_enable 1
  /perm special draw user add @user
  /perm special draw group list`

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

func (cmd *PermCommand) Exec(b *qbot.Bot, msg *qbot.Message) {
	if len(msg.Array) < 2 {
		b.SendGroupMsg(msg.GroupID, cmd.HelpMsg)
		return
	}

	getText := func(i int) string {
		if i < len(msg.Array) {
			if txt := msg.Array[i].GetTextItem(); txt != nil {
				return txt.Content
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
	default:
		b.SendGroupMsg(msg.GroupID, "Unknown subcommand: "+subCmd)
	}
}

func (cmd *PermCommand) handleSet(b *qbot.Bot, msg *qbot.Message) {
	getText := func(i int) string {
		if i < len(msg.Array) {
			if txt := msg.Array[i].GetTextItem(); txt != nil {
				return txt.Content
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
			GroupEnable:       0,
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
	case "group_enable":
		if value == "1" || value == "enable" || value == "true" {
			perm.GroupEnable = 1
		} else {
			perm.GroupEnable = 0
		}
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

func (cmd *PermCommand) handleSpecial(b *qbot.Bot, msg *qbot.Message) {
	getText := func(i int) string {
		if i < len(msg.Array) {
			if txt := msg.Array[i].GetTextItem(); txt != nil {
				return txt.Content
			}
		}
		return ""
	}

	// special <cmd> <user|group> <add|rm|list> [targets...]
	// 0       1     2            3             4
	if len(msg.Array) < 4 {
		b.SendGroupMsg(msg.GroupID, "Usage: perm special <cmd> <user|group> <add|rm|list> [targets...]")
		return
	}

	cmdName := getText(1)
	targetType := getText(2)
	action := getText(3)

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
			GroupEnable:       0,
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
			if !slices.Contains(currentList, t) {
				currentList = append(currentList, t)
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
			idx := slices.Index(currentList, t)
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
			if at := arg.GetAtItem(); at != nil {
				targets = append(targets, at.TargetID)
			} else if txt := arg.GetTextItem(); txt != nil {
				// Try to parse raw ID from text if provided as argument
				parts := strings.Fields(txt.Content)
				for _, part := range parts {
					if id, err := strconv.ParseUint(part, 10, 64); err == nil {
						targets = append(targets, id)
					}
				}
			}
		} else {
			// Groups are usually just numbers in text
			if txt := arg.GetTextItem(); txt != nil {
				parts := strings.Fields(txt.Content)
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
