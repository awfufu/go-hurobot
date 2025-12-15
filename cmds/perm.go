package cmds

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/awfufu/go-hurobot/config"
	"github.com/awfufu/go-hurobot/db"
	"github.com/awfufu/qbot"
)

const permHelpMsg string = `Manage command permissions.
Usage: perm <subcommand> [args...]
Subcommands:
  set <cmd> user_default <guest|admin|master>
  set <cmd> group_default <enable|disable>
  allow <cmd> user <add|rm|list> [targets...]
  allow <cmd> group <add|rm|list> [targets...]
  reject <cmd> user <add|rm|list> [targets...]
  reject <cmd> group <add|rm|list> [targets...]
Examples:
  /perm set draw user_default guest
  /perm allow sh user add @user
  /perm reject llm group add 12345`

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
	case "allow":
		cmd.handleListModify(b, msg, true)
	case "reject":
		cmd.handleListModify(b, msg, false)
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
		b.SendGroupMsg(msg.GroupID, "Usage: perm set <cmd> <user_default|group_default> <value>")
		return
	}

	cmdName := getText(2)
	settingType := getText(3)
	value := getText(4)

	perm := db.GetCommandPermission(cmdName)
	if perm == nil {
		// Initialize if not exists
		perm = &db.DbPermissions{
			Command:      cmdName,
			UserDefault:  "master",
			GroupDefault: "disable",
			AllowUsers:   "[]",
			RejectUsers:  "[]",
			AllowGroups:  "[]",
			RejectGroups: "[]",
		}
	}

	switch settingType {
	case "user_default":
		if value != "guest" && value != "admin" && value != "master" {
			b.SendGroupMsg(msg.GroupID, "Invalid value for user_default. Must be guest, admin, or master.")
			return
		}
		perm.UserDefault = value
	case "group_default":
		if value != "enable" && value != "disable" {
			b.SendGroupMsg(msg.GroupID, "Invalid value for group_default. Must be enable or disable.")
			return
		}
		perm.GroupDefault = value
	default:
		b.SendGroupMsg(msg.GroupID, "Unknown setting type: "+settingType)
		return
	}

	if err := db.SaveCommandPermission(perm); err != nil {
		b.SendGroupMsg(msg.GroupID, "Failed to save permission: "+err.Error())
	} else {
		b.SendGroupMsg(msg.GroupID, fmt.Sprintf("Updated %s %s to %s", cmdName, settingType, value))
	}
}

func (cmd *PermCommand) handleListModify(b *qbot.Bot, msg *qbot.Message, isAllow bool) {
	getText := func(i int) string {
		if i < len(msg.Array) {
			if txt := msg.Array[i].GetTextItem(); txt != nil {
				return txt.Content
			}
		}
		return ""
	}

	// allow <cmd> <user|group> <add|rm|list> [targets...]
	// 0     1     2            3             4
	if len(msg.Array) < 4 {
		actionType := "allow"
		if !isAllow {
			actionType = "reject"
		}
		b.SendGroupMsg(msg.GroupID, fmt.Sprintf("Usage: perm %s <cmd> <user|group> <add|rm|list> [targets...]", actionType))
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
		// Initialize
		perm = &db.DbPermissions{
			Command:      cmdName,
			UserDefault:  "master",
			GroupDefault: "disable",
			AllowUsers:   "",
			RejectUsers:  "",
			AllowGroups:  "",
			RejectGroups: "",
		}
	} else {
		// Ensure we are working with correct struct fields
		if isAllow {
			if targetType == "user" {
				rawStr = perm.AllowUsers
			} else {
				rawStr = perm.AllowGroups
			}
		} else {
			if targetType == "user" {
				rawStr = perm.RejectUsers
			} else {
				rawStr = perm.RejectGroups
			}
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

	if isAllow {
		if targetType == "user" {
			perm.AllowUsers = newStr
		} else {
			perm.AllowGroups = newStr
		}
	} else {
		if targetType == "user" {
			perm.RejectUsers = newStr
		} else {
			perm.RejectGroups = newStr
		}
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
