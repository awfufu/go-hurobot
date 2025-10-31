package cmds

import (
	"fmt"
	"go-hurobot/config"
	"go-hurobot/qbot"
	"strconv"
	"strings"
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
			Name:        "config",
			HelpMsg:     configHelpMsg,
			Permission:  config.Admin,
			AllowPrefix: false,
			NeedRawMsg:  false,
			MinArgs:     2,
		},
	}
}

func (cmd *ConfigCommand) Self() *cmdBase {
	return &cmd.cmdBase
}

func (cmd *ConfigCommand) Exec(c *qbot.Client, args []string, src *srcMsg, begin int) {
	subCmd := args[1]
	switch subCmd {
	case "admin":
		cmd.handleAdmin(c, args, src)
	case "allow":
		cmd.handleAllow(c, args, src)
	case "reject":
		cmd.handleReject(c, args, src)
	case "reload":
		cmd.handleReload(c, src)
	default:
		c.SendMsg(src.GroupID, src.UserID, "Unknown subcommand: "+subCmd)
	}
}

func (cmd *ConfigCommand) handleAdmin(c *qbot.Client, args []string, src *srcMsg) {
	if len(args) < 3 {
		c.SendMsg(src.GroupID, src.UserID, "Usage: config admin [add|rm|list] [@user...]")
		return
	}

	action := args[2]
	switch action {
	case "add":
		if len(args) < 4 {
			c.SendMsg(src.GroupID, src.UserID, "Usage: config admin add @user...")
			return
		}
		userIDs := extractUserIDs(args[3:])
		if len(userIDs) == 0 {
			c.SendMsg(src.GroupID, src.UserID, "Please mention at least one user with @")
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
		c.SendMsg(src.GroupID, src.UserID, strings.Join(results, "\n"))

	case "rm":
		if len(args) < 4 {
			c.SendMsg(src.GroupID, src.UserID, "Usage: config admin rm @user...")
			return
		}
		userIDs := extractUserIDs(args[3:])
		if len(userIDs) == 0 {
			c.SendMsg(src.GroupID, src.UserID, "Please mention at least one user with @")
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
		c.SendMsg(src.GroupID, src.UserID, strings.Join(results, "\n"))

	case "list":
		admins := config.GetAdmins()
		if len(admins) == 0 {
			c.SendMsg(src.GroupID, src.UserID, "Admin list is empty")
		} else {
			adminStrs := make([]string, len(admins))
			for i, u := range admins {
				adminStrs[i] = strconv.FormatUint(u, 10)
			}
			c.SendMsg(src.GroupID, src.UserID, fmt.Sprintf("Admins: %s", strings.Join(adminStrs, ", ")))
		}

	default:
		c.SendMsg(src.GroupID, src.UserID, "Unknown action: "+action)
	}
}

func extractUserIDs(args []string) []uint64 {
	userIDs := make([]uint64, 0, len(args))
	seen := make(map[uint64]bool)

	for _, arg := range args {
		if !strings.HasPrefix(arg, atPrefix) {
			continue
		}
		userID := str2uin64(strings.TrimPrefix(arg, atPrefix))
		if userID != 0 && !seen[userID] {
			seen[userID] = true
			userIDs = append(userIDs, userID)
		}
	}
	return userIDs
}

func (cmd *ConfigCommand) handleAllow(c *qbot.Client, args []string, src *srcMsg) {
	if len(args) < 3 {
		c.SendMsg(src.GroupID, src.UserID, "Usage: config allow <cmd> [add|rm|list] [@user...]")
		return
	}

	cmdName := args[2]
	if len(args) < 4 {
		c.SendMsg(src.GroupID, src.UserID, "Usage: config allow <cmd> [add|rm|list] [@user...]")
		return
	}

	action := args[3]
	switch action {
	case "add":
		if len(args) < 5 {
			c.SendMsg(src.GroupID, src.UserID, "Usage: config allow <cmd> add @user...")
			return
		}
		userIDs := extractUserIDs(args[4:])
		if len(userIDs) == 0 {
			c.SendMsg(src.GroupID, src.UserID, "Please mention at least one user with @")
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
		c.SendMsg(src.GroupID, src.UserID, strings.Join(results, "\n"))

	case "rm":
		if len(args) < 5 {
			c.SendMsg(src.GroupID, src.UserID, "Usage: config allow <cmd> rm @user...")
			return
		}
		userIDs := extractUserIDs(args[4:])
		if len(userIDs) == 0 {
			c.SendMsg(src.GroupID, src.UserID, "Please mention at least one user with @")
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
		c.SendMsg(src.GroupID, src.UserID, strings.Join(results, "\n"))

	case "list":
		users, err := config.GetAllowUsers(cmdName)
		if err != nil {
			c.SendMsg(src.GroupID, src.UserID, "Failed to get allow list: "+err.Error())
			return
		}
		if len(users) == 0 {
			c.SendMsg(src.GroupID, src.UserID, fmt.Sprintf("Allow list for %s is empty", cmdName))
		} else {
			userStrs := make([]string, len(users))
			for i, u := range users {
				userStrs[i] = strconv.FormatUint(u, 10)
			}
			c.SendMsg(src.GroupID, src.UserID, fmt.Sprintf("Allow list for %s: %s", cmdName, strings.Join(userStrs, ", ")))
		}
	default:
		c.SendMsg(src.GroupID, src.UserID, "Unknown action: "+action)
	}
}

func (cmd *ConfigCommand) handleReject(c *qbot.Client, args []string, src *srcMsg) {
	if len(args) < 3 {
		c.SendMsg(src.GroupID, src.UserID, "Usage: config reject <cmd> [add|rm|list] [@user...]")
		return
	}

	cmdName := args[2]
	if len(args) < 4 {
		c.SendMsg(src.GroupID, src.UserID, "Usage: config reject <cmd> [add|rm|list] [@user...]")
		return
	}

	action := args[3]
	switch action {
	case "add":
		if len(args) < 5 {
			c.SendMsg(src.GroupID, src.UserID, "Usage: config reject <cmd> add @user...")
			return
		}
		userIDs := extractUserIDs(args[4:])
		if len(userIDs) == 0 {
			c.SendMsg(src.GroupID, src.UserID, "Please mention at least one user with @")
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
		c.SendMsg(src.GroupID, src.UserID, strings.Join(results, "\n"))

	case "rm":
		if len(args) < 5 {
			c.SendMsg(src.GroupID, src.UserID, "Usage: config reject <cmd> rm @user...")
			return
		}
		userIDs := extractUserIDs(args[4:])
		if len(userIDs) == 0 {
			c.SendMsg(src.GroupID, src.UserID, "Please mention at least one user with @")
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
		c.SendMsg(src.GroupID, src.UserID, strings.Join(results, "\n"))

	case "list":
		users, err := config.GetRejectUsers(cmdName)
		if err != nil {
			c.SendMsg(src.GroupID, src.UserID, "Failed to get reject list: "+err.Error())
			return
		}
		if len(users) == 0 {
			c.SendMsg(src.GroupID, src.UserID, fmt.Sprintf("Reject list for %s is empty", cmdName))
		} else {
			userStrs := make([]string, len(users))
			for i, u := range users {
				userStrs[i] = strconv.FormatUint(u, 10)
			}
			c.SendMsg(src.GroupID, src.UserID, fmt.Sprintf("Reject list for %s: %s", cmdName, strings.Join(userStrs, ", ")))
		}
	default:
		c.SendMsg(src.GroupID, src.UserID, "Unknown action: "+action)
	}
}

func (cmd *ConfigCommand) handleReload(c *qbot.Client, src *srcMsg) {
	err := config.ReloadConfig()
	if err != nil {
		c.SendMsg(src.GroupID, src.UserID, "Failed to reload config: "+err.Error())
	} else {
		c.SendMsg(src.GroupID, src.UserID, "Configuration reloaded successfully")
	}
}
