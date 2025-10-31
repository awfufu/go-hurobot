package cmds

import (
	"strconv"
	"strings"

	"github.com/google/shlex"

	"go-hurobot/config"
	"go-hurobot/qbot"
)

type srcMsg struct {
	MsgID   uint64
	UserID  uint64
	GroupID uint64
	Card    string
	Role    string
	Time    uint64
	Raw     string
}

const (
	atPrefix      string = "--at="
	replyPrefix   string = "--reply="
	facePrefix    string = "--face="
	imagePrefix   string = "--image="
	recordPrefix  string = "--record="
	filePrefix    string = "--file="
	forwardPrefix string = "--forward="
	jsonPrefix    string = "--json="
)

type command interface {
	Self() *cmdBase
	Exec(c *qbot.Client, args []string, src *srcMsg, begin int)
}

type cmdBase struct {
	Name        string            // Command name
	HelpMsg     string            // Help message
	Permission  config.Permission // Permission requirement
	AllowPrefix bool              // Whether to allow prefixes (like @)
	NeedRawMsg  bool              // Whether raw message is needed
	MaxArgs     int               // Maximum number of arguments
	MinArgs     int               // Minimum number of arguments
}

const commandPrefix = '/'

var cmdMap map[string]command

func init() {
	cmdMap = map[string]command{
		"callme":       NewCallmeCommand(),
		"config":       NewConfigCommand(),
		"crypto":       NewCryptoCommand(),
		"delete":       NewDeleteCommand(),
		"dice":         NewDiceCommand(),
		"draw":         NewDrawCommand(),
		"echo":         NewEchoCommand(),
		"essence":      NewEssenceCommand(),
		"fx":           NewErCommand(),
		"group":        NewGroupCommand(),
		"llm":          NewLlmCommand(),
		"mc":           NewMcCommand(),
		"memberinfo":   NewMemberinfoCommand(),
		"psql":         NewPsqlCommand(),
		"py":           NewPyCommand(),
		"rcon":         NewRconCommand(),
		"rps":          NewRpsCommand(),
		"sh":           NewShCommand(),
		"specialtitle": NewSpecialtitleCommand(),
		"testapi":      NewTestapiCommand(),
		"which":        NewWhichCommand(),
	}
}

func HandleCommand(c *qbot.Client, msg *qbot.Message) bool /* is command */ {
	cmdName, raw, skip := parseCmd(msg)
	if skip == -1 {
		return false
	}

	if cmdName == "" {
		return false
	}

	cmd, exists := cmdMap[cmdName]
	if !exists {
		return false
	}

	src := &srcMsg{
		MsgID:   msg.MsgID,
		UserID:  msg.UserID,
		GroupID: msg.GroupID,
		Card:    msg.Card,
		Role:    msg.Role,
		Time:    msg.Time,
		Raw:     raw,
	}
	cmdBase := cmd.Self()

	// check permission
	if !checkCmdPermission(cmdBase.Name, src.UserID) {
		c.SendMsg(src.GroupID, src.UserID, cmdBase.Name+": Permission denied")
		return true
	}

	// check if allow prefix
	if skip != 0 && !cmdBase.AllowPrefix {
		return true
	}

	// parse arguments
	var args []string
	if cmdBase.NeedRawMsg {
		args = []string{cmdName, raw}
	} else {
		args = splitArguments(msg)
	}
	if args == nil {
		return false
	}
	argCount := len(args) - skip
	if (cmdBase.MinArgs > 0 && argCount < cmdBase.MinArgs) || (cmdBase.MaxArgs > 0 && argCount > cmdBase.MaxArgs) {
		c.SendMsg(src.GroupID, src.UserID, cmdBase.HelpMsg)
		return true
	}

	// check if is help command
	if isHelpRequest(args, skip) {
		c.SendMsg(src.GroupID, src.UserID, cmdBase.HelpMsg)
		return true
	}

	// execute command
	cmd.Exec(c, args, src, skip)
	return true
}

// calculate the number of prefixes to skip
func parseCmd(msg *qbot.Message) (string, string, int) {
	if len(msg.Array) == 0 {
		return "", "", -1
	}

	for skip := 0; skip < len(msg.Array); skip++ {
		switch msg.Array[skip].Type {
		case qbot.Reply, qbot.At:
			continue
		case qbot.Text:
			content := msg.Array[skip].Content

			// Skip leading spaces
			offset := 0
			for offset < len(content) && content[offset] == ' ' {
				offset++
			}

			// Check if starts with command prefix
			if offset >= len(content) || content[offset] != commandPrefix {
				return "", "", -1
			}
			offset++ // skip the '/' character

			// Find the skip-th ']' character in msg.Raw and extract content after it
			rawStart := 0
			for i := 0; i < skip; i++ {
				idx := strings.Index(msg.Raw[rawStart:], "]")
				if idx == -1 {
					break
				}
				rawStart += idx + 1
			}

			// Skip to the command prefix in raw
			for rawStart < len(msg.Raw) && msg.Raw[rawStart] != commandPrefix {
				rawStart++
			}
			if rawStart >= len(msg.Raw) {
				return "", "", -1
			}
			rawStart++ // skip the '/' character

			raw := msg.Raw[rawStart:]

			// Find command name (up to first space)
			cmdIndex := strings.Index(raw, " ")
			if cmdIndex == -1 {
				return raw, "", skip
			}
			return raw[:cmdIndex], raw[cmdIndex+1:], skip
		default:
			return "", "", -1
		}
	}
	return "", "", -1
}

// isHelpRequest checks if it is a help request
func isHelpRequest(args []string, skip int) bool {
	if len(args) <= skip {
		return false
	}
	firstArg := args[skip]
	return firstArg == "-h" || firstArg == "-?" || firstArg == "--help"
}

func getCmdPermLevel(cmdName string) config.Permission {
	cmdCfg := config.Cfg.GetCmdConfig(cmdName)
	if cmdCfg == nil {
		return config.Master
	}
	return cmdCfg.GetPermissionLevel()
}

// checkCmdPermission checks if user has permission to execute specified command
// considering permission, allow_users, reject_users in command configuration
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

func splitArguments(msg *qbot.Message) []string {
	result := make([]string, 0, 20)

	firstText := true
	for _, item := range msg.Array {
		if item.Type == qbot.Text {
			content := item.Content
			if firstText {
				content = item.Content[1:]
				firstText = false
			}
			texts, err := shlex.Split(content)
			if err != nil {
				return nil
			}
			result = append(result, texts...)
		} else {
			result = append(result, msgItemToArg(item))
		}
	}
	return result
}

func msgItemToArg(item qbot.MsgItem) string {
	var prefix string
	switch item.Type {
	case qbot.At:
		prefix = atPrefix
	case qbot.Face:
		prefix = facePrefix
	case qbot.Image:
		prefix = imagePrefix
	case qbot.Record:
		prefix = recordPrefix
	case qbot.Reply:
		prefix = replyPrefix
	case qbot.File:
		prefix = filePrefix
	case qbot.Forward:
		prefix = forwardPrefix
	case qbot.Json:
		prefix = jsonPrefix
	default:
		return item.Content
	}
	return prefix + item.Content
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
