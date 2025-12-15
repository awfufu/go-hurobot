package cmds

import (
	"slices"
	"strconv"
	"strings"

	"github.com/google/shlex"

	"github.com/awfufu/go-hurobot/config"
	"github.com/awfufu/go-hurobot/db"
	"github.com/awfufu/qbot"
)

type command interface {
	Self() *cmdBase
	Exec(b *qbot.Bot, msg *qbot.Message)
}

type cmdBase struct {
	Name       string            // Command name
	HelpMsg    string            // Help message
	Permission config.Permission // Permission requirement
	NeedRawMsg bool
	MaxArgs    int // Maximum number of arguments
	MinArgs    int // Minimum number of arguments
}

const commandPrefix = '/'

var cmdMap map[string]command

func init() {
	cmdMap = map[string]command{
		"crypto":       NewCryptoCommand(),
		"delete":       NewDeleteCommand(),
		"draw":         NewDrawCommand(),
		"echo":         NewEchoCommand(),
		"essence":      NewEssenceCommand(),
		"fx":           NewErCommand(),
		"group":        NewGroupCommand(),
		"perm":         NewPermCommand(),
		"psql":         NewPsqlCommand(),
		"sh":           NewShCommand(),
		"specialtitle": NewSpecialtitleCommand(),
		"which":        NewWhichCommand(),
	}
}

func HandleCommand(b *qbot.Bot, msg *qbot.Message) {
	cmdName, argsItems, raw := parseCmd(msg)

	if cmdName == "" {
		return
	}

	cmd, exists := cmdMap[cmdName]
	if !exists {
		return
	}

	cmdBase := cmd.Self()

	// check permission
	if !checkCmdPermission(cmdBase.Name, msg.UserID, msg.GroupID) {
		b.SendGroupMsg(msg.GroupID, cmdBase.Name+": Permission denied")
		return
	}

	// parse arguments
	var args []qbot.MsgItem
	if cmdBase.NeedRawMsg {
		args = []qbot.MsgItem{&qbot.TextItem{Content: cmdName}}
		if len(raw) > 0 {
			args = append(args, &qbot.TextItem{Content: raw})
		}
	} else {
		// New logic: commands are strictly TextItem first.
		// args construction starts with command name, then parsed args.

		args = make([]qbot.MsgItem, 0, len(msg.Array)+len(argsItems))

		// Add command name
		args = append(args, &qbot.TextItem{Content: cmdName})

		// Add parsed args
		for _, item := range argsItems {
			if t, ok := item.(*qbot.TextItem); ok {
				parts, err := shlex.Split(t.Content)
				if err != nil {
					return
				}
				for _, part := range parts {
					args = append(args, &qbot.TextItem{Content: part})
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

	item, ok := msg.Array[0].(*qbot.TextItem)
	if !ok {
		return "", nil, ""
	}
	content := item.Content

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
			args = append(args, &qbot.TextItem{Content: argContent})
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
	if t, ok := firstArg.(*qbot.TextItem); ok {
		return t.Content == "-h" || t.Content == "-?" || t.Content == "--help"
	}
	return false
}

func getCmdPermLevel(cmdName string) config.Permission {
	// Deprecated or can read from DB for default if permission struct changes
	// For now, logic handled in checkCmdPermission
	return config.Master
}

func checkCmdPermission(cmdName string, userID, groupID uint64) bool {
	// 1. Master Bypass
	if userID == config.Cfg.Permissions.MasterID {
		return true
	}

	// 2. Load Permissions from DB
	perm := db.GetCommandPermission(cmdName)

	var userDefault string = "master"
	var groupDefault string = "disable"
	var allowUsers []uint64
	var rejectUsers []uint64
	var allowGroups []uint64
	var rejectGroups []uint64

	if perm != nil {
		userDefault = perm.UserDefault
		groupDefault = perm.GroupDefault
		allowUsers = perm.ParseAllowUsers()
		rejectUsers = perm.ParseRejectUsers()
		allowGroups = perm.ParseAllowGroups()
		rejectGroups = perm.ParseRejectGroups()
	}

	// 3. User Allow/Reject Lists
	if slices.Contains(rejectUsers, userID) {
		return false
	}
	if slices.Contains(allowUsers, userID) {
		return true
	}

	// 4. User Role Check
	userRole := config.GetUserPermission(userID)
	var requiredPerm config.Permission
	switch userDefault {
	case "guest":
		requiredPerm = config.Guest
	case "admin":
		requiredPerm = config.Admin
	case "master":
		requiredPerm = config.Master
	default:
		requiredPerm = config.Master
	}
	if userRole < requiredPerm {
		return false
	}

	// 5. Group Allow/Reject Lists
	if slices.Contains(rejectGroups, groupID) {
		return false
	}
	if slices.Contains(allowGroups, groupID) {
		return true
	}

	// 6. Group Default Check
	if groupDefault == "disable" {
		return false
	}
	// enable
	return true
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

func str2uin64(s string) uint64 {
	value, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0
	}
	return value
}
