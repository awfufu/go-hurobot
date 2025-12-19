package cmds

import (
	"log"
	"slices"
	"strconv"
	"strings"

	"github.com/awfufu/go-hurobot/internal/config"
	"github.com/awfufu/go-hurobot/internal/db"
	"github.com/awfufu/qbot"
	"github.com/google/shlex"
)

type Command struct {
	Name       string            // Command name
	HelpMsg    string            // Help message
	Permission config.Permission // Permission requirement
	NeedRawMsg bool
	MaxArgs    int                                     // Maximum number of arguments
	MinArgs    int                                     // Minimum number of arguments
	Exec       func(b *qbot.Sender, msg *qbot.Message) // Execute function
}

const commandPrefix = '/'

var cmdMap map[string]*Command

func init() {
	cmdMap = map[string]*Command{
		"crypto":       cryptoCommand,
		"delete":       deleteCommand,
		"draw":         drawCommand,
		"echo":         echoCommand,
		"essence":      essenceCommand,
		"fx":           erCommand,
		"group":        groupCommand,
		"perm":         permCommand,
		"sh":           shCommand,
		"specialtitle": specialtitleCommand,
		"which":        whichCommand,
		"calc":         calcCommand,
	}
}

func InitCommandPermissions() {
	for name, cmd := range cmdMap {
		base := cmd

		perm := db.GetCommandPermission(name)
		if perm != nil {
			continue
		}

		// Map config.Permission to int
		// 0:guest, 1:admin, 2:master
		var userAllow int
		switch base.Permission {
		case config.Guest:
			userAllow = 0
		case config.Admin:
			userAllow = 1
		case config.Master:
			userAllow = 2
		default:
			userAllow = 2
		}

		newPerm := &db.DbPermissions{
			Command:          name,
			UserAllow:        userAllow,
			SpecialUsers:     "",
			IsWhitelistUsers: 0, // blacklist mode by default (empty blacklist = allow nobody? No, blacklist means if in list then block. If list empty, all allowed? Wait. Logic check needed.)
			// Wait, if IsWhitelist, only those in list are allowed.
			// If IsBlacklist (0), those in list are blocked.
			// Default blacklist empty means NO special blocks.
			// BUT GroupEnable is 0 (disable). So groups are disabled by default.
			// UserAllow depends on role.
			SpecialGroups:     "",
			IsWhitelistGroups: 0,
		}

		if err := db.SaveCommandPermission(newPerm); err != nil {
			log.Printf("Failed to init permission for %s: %v", name, err)
		} else {
			log.Printf("Initialized permission for command: %s (user_allow: %d)", name, userAllow)
		}
	}
}

func HandleCommand(b *qbot.Sender, msg *qbot.Message) {
	cmdName, argsItems, raw := parseCmd(msg)

	if cmdName == "" {
		return
	}

	cmd, exists := cmdMap[cmdName]
	if !exists {
		return
	}

	cmdBase := cmd

	// check permission
	if !checkCmdPermission(cmdBase.Name, msg.UserID, msg.GroupID) {
		b.SendGroupMsg(msg.GroupID, cmdBase.Name+": Permission denied")
		return
	}

	// parse arguments
	var args []qbot.MsgItem
	if cmdBase.NeedRawMsg {
		args = []qbot.MsgItem{qbot.TextItem(cmdName)}
		if len(raw) > 0 {
			args = append(args, qbot.TextItem(raw))
		}
	} else {
		// New logic: commands are strictly TextItem first.
		// args construction starts with command name, then parsed args.

		args = make([]qbot.MsgItem, 0, len(msg.Array)+len(argsItems))

		// Add command name
		args = append(args, qbot.TextItem(cmdName))

		// Add parsed args
		for _, item := range argsItems {
			if t, ok := item.(qbot.TextItem); ok {
				parts, err := shlex.Split(t.String())
				if err != nil {
					return
				}
				for _, part := range parts {
					args = append(args, qbot.TextItem(part))
				}
			} else {
				args = append(args, item)
			}
		}
	}
	if args == nil {
		return
	}
	// argCount logic simplified as there is no skip
	argCount := len(args)
	if (cmdBase.MinArgs > 0 && argCount < cmdBase.MinArgs) || (cmdBase.MaxArgs > 0 && argCount > cmdBase.MaxArgs) {
		b.SendGroupMsg(msg.GroupID, cmdBase.HelpMsg)
		return
	}

	// check if is help command
	if isHelpRequest(args) {
		b.SendGroupMsg(msg.GroupID, cmdBase.HelpMsg)
		return
	}

	// execute command
	newMsg := *msg
	newMsg.Array = args
	cmd.Exec(b, &newMsg)
}

// calculate the number of prefixes to skip
// calculate the number of prefixes to skip
func parseCmd(msg *qbot.Message) (string, []qbot.MsgItem, string) {
	if len(msg.Array) == 0 {
		return "", nil, ""
	}

	// Strictly check if the first item is TextType
	if msg.Array[0].Type() != qbot.TextType {
		return "", nil, ""
	}

	item, ok := msg.Array[0].(qbot.TextItem)
	if !ok {
		return "", nil, ""
	}
	content := string(item)

	// Skip leading spaces
	offset := 0
	for offset < len(content) && content[offset] == ' ' {
		offset++
	}

	// Check if starts with command prefix
	if offset >= len(content) || content[offset] != commandPrefix {
		return "", nil, ""
	}
	offset++ // skip the '/' character

	// In the raw message, since we are at the first item, we can skip leading spaces and finding '/' directly?
	// The original logic tried to find ']' for skips, but here skip is 0.
	// We need to find where the command starts in msg.Raw.
	// Assuming msg.Raw corresponds to the text content properly.
	// Simple approach: look for commandPrefix in msg.Raw, similar to before but without loop.

	rawStart := 0
	// Skip to the command prefix in raw
	for rawStart < len(msg.Raw) && msg.Raw[rawStart] != commandPrefix {
		rawStart++
	}
	if rawStart >= len(msg.Raw) {
		return "", nil, ""
	}
	rawStart++ // skip the '/' character

	raw := msg.Raw[rawStart:]

	// Find command name (up to first space)
	cmdIndex := strings.Index(raw, " ")
	var cmdName, rawArgs string
	if cmdIndex == -1 {
		cmdName = raw
		rawArgs = ""
	} else {
		cmdName = raw[:cmdIndex]
		rawArgs = raw[cmdIndex+1:]
	}

	// Construct args
	var args []qbot.MsgItem

	// Extract args from the current text item
	rest := content[offset:]
	idx := strings.Index(rest, " ")
	if idx != -1 {
		argContent := rest[idx+1:]
		if len(argContent) > 0 {
			args = append(args, qbot.TextItem(argContent))
		}
	}

	// Append remaining items
	if len(msg.Array) > 1 {
		args = append(args, msg.Array[1:]...)
	}

	return cmdName, args, rawArgs
}

// checks if it is a help request
func isHelpRequest(args []qbot.MsgItem) bool {
	if len(args) <= 0 {
		return false
	}
	firstArg := args[0]
	if t, ok := firstArg.(qbot.TextItem); ok {
		return t == "-h" || t == "-?" || t == "--help"
	}
	return false
}

func getCmdPermLevel(cmdName string) config.Permission {
	return config.Master
}

func checkCmdPermission(cmdName string, userID qbot.UserID, groupID qbot.GroupID) bool {
	// 1. Master Bypass
	if userID == config.Cfg.Permissions.MasterID {
		return true
	}

	// 2. Load Permissions from DB
	perm := db.GetCommandPermission(cmdName)

	var userAllow int = 2 // default master
	var specialUsers []uint64
	var isWhitelistUsers int = 0 // default blacklist
	var specialGroups []uint64
	var isWhitelistGroup int = 0 // default blacklist

	if perm != nil {
		userAllow = perm.UserAllow
		specialUsers = perm.ParseSpecialUsers()
		isWhitelistUsers = perm.IsWhitelistUsers
		specialGroups = perm.ParseSpecialGroups()
		isWhitelistGroup = perm.IsWhitelistGroups
	}

	// 3. User Special List Check
	if slices.Contains(specialUsers, uint64(userID)) {
		if isWhitelistUsers == 1 {
			// Whitelist mode: User in list -> Allow
			return true
		} else {
			// Blacklist mode: User in list -> Block
			return false
		}
	}

	// 4. User Role Check
	userRole := GetUserPermission(userID)
	var requiredPerm config.Permission
	// Map DB int (0,1,2) to config.Permission
	switch userAllow {
	case 0:
		requiredPerm = config.Guest
	case 1:
		requiredPerm = config.Admin
	case 2:
		requiredPerm = config.Master
	default:
		requiredPerm = config.Master
	}

	if userRole < requiredPerm {
		return false
	}

	// 5. Group Special List Check
	if slices.Contains(specialGroups, uint64(groupID)) {
		if isWhitelistGroup == 1 {
			// Whitelist mode: Group in list -> Allow
			return true
		} else {
			// Blacklist mode: Group in list -> Block
			return false
		}
	}
	return true
}

func GetUserPermission(userID qbot.UserID) config.Permission {
	if userID == config.Cfg.Permissions.MasterID {
		return config.Master
	}

	perm := db.GetUserPerm(uint64(userID))
	switch perm {
	case 0:
		return config.Guest
	case 1:
		return config.Admin
	case 2:
		return config.Master // Should not theoretically happen via DB if MasterID is separate, but allows granting Master via DB
	default:
		return config.Guest
	}
}

func decodeSpecialChars(raw string) string {
	replacer := strings.NewReplacer(
		"&#91;", "[",
		"&#93;", "]",
		"&amp;", "&",
		"&lt;", "<",
		"&gt;", ">",
	)
	return replacer.Replace(raw)
}

func isNumeric(s string) bool {
	_, err := strconv.ParseUint(s, 10, 64)
	return err == nil
}

func encodeSpecialChars(raw string) string {
	replacer := strings.NewReplacer(
		"[", "&#91;",
		"]", "&#93;",
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
	)
	return replacer.Replace(raw)
}

func str2int64(s string) int64 {
	value, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return value
}
