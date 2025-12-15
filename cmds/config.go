package cmds

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/awfufu/go-hurobot/config"
	"github.com/awfufu/qbot"
)

const configHelpMsg string = `Manage bot configuration.
Usage: config <subcommand> [args...]
Subcommands:
  admin add @user...    Add user(s) as admin
  admin rm @user...     Remove user(s) from admin
  admin list            List all admins
  allow <cmd> add @user...  Add user(s) to command allow list
  allow <cmd> rm @user...   Remove user(s) from command allow list
  allow <cmd> list          List users in command allow list
  reject <cmd> add @user... Add user(s) to command reject list
  reject <cmd> rm @user...  Remove user(s) from command reject list
  reject <cmd> list         List users in command reject list
  reload                    Reload configuration from file
Examples:
  /config admin add @user1 @user2
  /config admin list
  /config allow sh add @user
  /config reject llm list
  /config reload`

type ConfigCommand struct {
	cmdBase
}

func NewConfigCommand() *ConfigCommand {
	return &ConfigCommand{
		cmdBase: cmdBase{
			Name:       "config",
			HelpMsg:    configHelpMsg,
			Permission: config.Admin,
			NeedRawMsg: false,
			MinArgs:    2,
		},
	}
}

func (cmd *ConfigCommand) Self() *cmdBase {
	return &cmd.cmdBase
}

func (cmd *ConfigCommand) Exec(b *qbot.Bot, msg *qbot.Message) {
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
	case "admin":
		cmd.handleAdmin(b, msg)
	case "allow":
		cmd.handleAllow(b, msg)
	case "reject":
		cmd.handleReject(b, msg)
	case "reload":
		cmd.handleReload(b, msg)
	default:
		b.SendGroupMsg(msg.GroupID, "Unknown subcommand: "+subCmd)
	}
}

func (cmd *ConfigCommand) handleAdmin(b *qbot.Bot, msg *qbot.Message) {
	getText := func(i int) string {
		if i < len(msg.Array) {
			if txt := msg.Array[i].GetTextItem(); txt != nil {
				return txt.Content
			}
		}
		return ""
	}

	if len(msg.Array) < 3 {
		b.SendGroupMsg(msg.GroupID, "Usage: config admin [add|rm|list] [@user...]")
		return
	}

	action := getText(2)
	switch action {
	case "add":
		if len(msg.Array) < 4 {
			b.SendGroupMsg(msg.GroupID, "Usage: config admin add @user...")
			return
		}
		userIDs := extractUserIDs(msg.Array)
		if len(userIDs) == 0 {
			b.SendGroupMsg(msg.GroupID, "Please mention at least one user with @")
			return
		}

		results := make([]string, 0, len(userIDs))
		for _, userID := range userIDs {
			err := config.AddAdmin(userID)
			if err != nil {
				results = append(results, fmt.Sprintf("❌ %d: %s", userID, err.Error()))
			} else {
				results = append(results, fmt.Sprintf("✅ %d: set as admin", userID))
			}
		}
		b.SendGroupMsg(msg.GroupID, strings.Join(results, "\n"))

	case "rm":
		if len(msg.Array) < 4 {
			b.SendGroupMsg(msg.GroupID, "Usage: config admin rm @user...")
			return
		}
		userIDs := extractUserIDs(msg.Array)
		if len(userIDs) == 0 {
			b.SendGroupMsg(msg.GroupID, "Please mention at least one user with @")
			return
		}

		results := make([]string, 0, len(userIDs))
		for _, userID := range userIDs {
			err := config.RemoveAdmin(userID)
			if err != nil {
				results = append(results, fmt.Sprintf("❌ %d: %s", userID, err.Error()))
			} else {
				results = append(results, fmt.Sprintf("✅ %d: removed from admin", userID))
			}
		}
		b.SendGroupMsg(msg.GroupID, strings.Join(results, "\n"))

	case "list":
		admins := config.GetAdmins()
		if len(admins) == 0 {
			b.SendGroupMsg(msg.GroupID, "Admin list is empty")
		} else {
			adminStrs := make([]string, len(admins))
			for i, u := range admins {
				adminStrs[i] = strconv.FormatUint(u, 10)
			}
			b.SendGroupMsg(msg.GroupID, fmt.Sprintf("Admins: %s", strings.Join(adminStrs, ", ")))
		}

	default:
		b.SendGroupMsg(msg.GroupID, "Unknown action: "+action)
	}
}

func extractUserIDs(args []qbot.MsgItem) []uint64 {
	userIDs := make([]uint64, 0, len(args))
	seen := make(map[uint64]bool)

	for _, arg := range args {
		if atItem := arg.GetAtItem(); atItem != nil {
			userID := atItem.TargetID
			if userID != 0 && !seen[userID] {
				seen[userID] = true
				userIDs = append(userIDs, userID)
			}
		}
	}
	return userIDs
}

func (cmd *ConfigCommand) handleAllow(b *qbot.Bot, msg *qbot.Message) {
	getText := func(i int) string {
		if i < len(msg.Array) {
			if txt := msg.Array[i].GetTextItem(); txt != nil {
				return txt.Content
			}
		}
		return ""
	}

	if len(msg.Array) < 3 {
		b.SendGroupMsg(msg.GroupID, "Usage: config allow <cmd> [add|rm|list] [@user...]")
		return
	}

	cmdName := getText(2)
	if len(msg.Array) < 4 {
		b.SendGroupMsg(msg.GroupID, "Usage: config allow <cmd> [add|rm|list] [@user...]")
		return
	}

	action := getText(3)
	switch action {
	case "add":
		if len(msg.Array) < 5 {
			b.SendGroupMsg(msg.GroupID, "Usage: config allow <cmd> add @user...")
			return
		}
		userIDs := extractUserIDs(msg.Array)
		if len(userIDs) == 0 {
			b.SendGroupMsg(msg.GroupID, "Please mention at least one user with @")
			return
		}

		results := make([]string, 0, len(userIDs))
		for _, userID := range userIDs {
			err := config.AddAllowUser(cmdName, userID)
			if err != nil {
				results = append(results, fmt.Sprintf("❌ %d: %s", userID, err.Error()))
			} else {
				results = append(results, fmt.Sprintf("✅ %d: added to %s allow list", userID, cmdName))
			}
		}
		b.SendGroupMsg(msg.GroupID, strings.Join(results, "\n"))

	case "rm":
		if len(msg.Array) < 5 {
			b.SendGroupMsg(msg.GroupID, "Usage: config allow <cmd> rm @user...")
			return
		}
		userIDs := extractUserIDs(msg.Array)
		if len(userIDs) == 0 {
			b.SendGroupMsg(msg.GroupID, "Please mention at least one user with @")
			return
		}

		results := make([]string, 0, len(userIDs))
		for _, userID := range userIDs {
			err := config.RemoveAllowUser(cmdName, userID)
			if err != nil {
				results = append(results, fmt.Sprintf("❌ %d: %s", userID, err.Error()))
			} else {
				results = append(results, fmt.Sprintf("✅ %d: removed from %s allow list", userID, cmdName))
			}
		}
		b.SendGroupMsg(msg.GroupID, strings.Join(results, "\n"))

	case "list":
		users, err := config.GetAllowUsers(cmdName)
		if err != nil {
			b.SendGroupMsg(msg.GroupID, "Failed to get allow list: "+err.Error())
			return
		}
		if len(users) == 0 {
			b.SendGroupMsg(msg.GroupID, fmt.Sprintf("Allow list for %s is empty", cmdName))
		} else {
			userStrs := make([]string, len(users))
			for i, u := range users {
				userStrs[i] = strconv.FormatUint(u, 10)
			}
			b.SendGroupMsg(msg.GroupID, fmt.Sprintf("Allow list for %s: %s", cmdName, strings.Join(userStrs, ", ")))
		}
	default:
		b.SendGroupMsg(msg.GroupID, "Unknown action: "+action)
	}
}

func (cmd *ConfigCommand) handleReject(b *qbot.Bot, msg *qbot.Message) {
	getText := func(i int) string {
		if i < len(msg.Array) {
			if txt := msg.Array[i].GetTextItem(); txt != nil {
				return txt.Content
			}
		}
		return ""
	}

	if len(msg.Array) < 3 {
		b.SendGroupMsg(msg.GroupID, "Usage: config reject <cmd> [add|rm|list] [@user...]")
		return
	}

	cmdName := getText(2)
	if len(msg.Array) < 4 {
		b.SendGroupMsg(msg.GroupID, "Usage: config reject <cmd> [add|rm|list] [@user...]")
		return
	}

	action := getText(3)
	switch action {
	case "add":
		if len(msg.Array) < 5 {
			b.SendGroupMsg(msg.GroupID, "Usage: config reject <cmd> add @user...")
			return
		}
		userIDs := extractUserIDs(msg.Array)
		if len(userIDs) == 0 {
			b.SendGroupMsg(msg.GroupID, "Please mention at least one user with @")
			return
		}

		results := make([]string, 0, len(userIDs))
		for _, userID := range userIDs {
			err := config.AddRejectUser(cmdName, userID)
			if err != nil {
				results = append(results, fmt.Sprintf("❌ %d: %s", userID, err.Error()))
			} else {
				results = append(results, fmt.Sprintf("✅ %d: added to %s reject list", userID, cmdName))
			}
		}
		b.SendGroupMsg(msg.GroupID, strings.Join(results, "\n"))

	case "rm":
		if len(msg.Array) < 5 {
			b.SendGroupMsg(msg.GroupID, "Usage: config reject <cmd> rm @user...")
			return
		}
		userIDs := extractUserIDs(msg.Array)
		if len(userIDs) == 0 {
			b.SendGroupMsg(msg.GroupID, "Please mention at least one user with @")
			return
		}

		results := make([]string, 0, len(userIDs))
		for _, userID := range userIDs {
			err := config.RemoveRejectUser(cmdName, userID)
			if err != nil {
				results = append(results, fmt.Sprintf("❌ %d: %s", userID, err.Error()))
			} else {
				results = append(results, fmt.Sprintf("✅ %d: removed from %s reject list", userID, cmdName))
			}
		}
		b.SendGroupMsg(msg.GroupID, strings.Join(results, "\n"))

	case "list":
		users, err := config.GetRejectUsers(cmdName)
		if err != nil {
			b.SendGroupMsg(msg.GroupID, "Failed to get reject list: "+err.Error())
			return
		}
		if len(users) == 0 {
			b.SendGroupMsg(msg.GroupID, fmt.Sprintf("Reject list for %s is empty", cmdName))
		} else {
			userStrs := make([]string, len(users))
			for i, u := range users {
				userStrs[i] = strconv.FormatUint(u, 10)
			}
			b.SendGroupMsg(msg.GroupID, fmt.Sprintf("Reject list for %s: %s", cmdName, strings.Join(userStrs, ", ")))
		}
	default:
		b.SendGroupMsg(msg.GroupID, "Unknown action: "+action)
	}
}

func (cmd *ConfigCommand) handleReload(b *qbot.Bot, msg *qbot.Message) {
	err := config.ReloadConfig()
	if err != nil {
		b.SendGroupMsg(msg.GroupID, "Failed to reload config: "+err.Error())
	} else {
		b.SendGroupMsg(msg.GroupID, "Configuration reloaded successfully")
	}
}
