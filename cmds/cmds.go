package cmds

import (
	"strconv"
	"strings"

	"github.com/google/shlex"

	"github.com/awfufu/go-hurobot/config"
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
	if !checkCmdPermission(cmdBase.Name, msg.UserID) {
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
	cmdCfg := config.Cfg.GetCmdConfig(cmdName)
	if cmdCfg == nil {
		return config.Master
	}
	return cmdCfg.GetPermissionLevel()
}

func checkCmdPermission(cmdName string, userID uint64) bool {
	cmdCfg := config.Cfg.GetCmdConfig(cmdName)

	// If not configured, use default permission check
	if cmdCfg == nil {
		return true
	}

	userPerm := config.GetUserPermission(userID)
	requiredPerm := cmdCfg.GetPermissionLevel()

	// Master users are not restricted by reject_users
	if userPerm == config.Master {
		return true
	}

	// Check reject_users (effective when command permission < master)
	if requiredPerm < config.Master && cmdCfg.IsInRejectList(userID) {
		return false
	}

	// Check allow_users (effective when command permission > guest)
	if requiredPerm > config.Guest {
		// If there is an allow_users list, only allow users in the list
		if len(cmdCfg.AllowUsers) > 0 {
			return cmdCfg.IsInAllowList(userID)
		}
	}

	// Check basic permission
	return userPerm >= requiredPerm
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
