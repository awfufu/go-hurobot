package cmds

import (
	"strconv"
	"strings"

	"github.com/google/shlex"

	"go-hurobot/qbot"
)

var maxCommandLength int = 0

type ArgsList struct {
	Contents []string
	Types    []qbot.MsgType
	Size     int
}

type CmdHandler func(*qbot.Client, *qbot.Message, *ArgsList)

var cmdMap map[string]CmdHandler

func init() {
	cmdMap = map[string]CmdHandler{
		"echo":         cmd_echo,
		"specialtitle": cmd_specialtitle,
		"psql":         cmd_psql,
		"group":        cmd_group,
		"delete":       cmd_delete,
		"llm":          cmd_llm,
		"callme":       cmd_callme,
		"debug":        cmd_debug,
		"essence":      cmd_essence,
		"draw":         cmd_draw,
		"fx":           cmd_er,
		"crypto":       cmd_crypto,
		"event":        cmd_event,
		"sh":           cmd_sh,
		"rps":          cmd_rps,
		"dice":         cmd_dice,
		"memberinfo":   cmd_memberinfo,
		"info":         cmd_info,
		"rcon":         cmd_rcon,
		"mc":           cmd_mc,
	}

	for key := range cmdMap {
		if len(key) > maxCommandLength {
			maxCommandLength = len(key)
		}
	}
}

func HandleCommand(c *qbot.Client, msg *qbot.Message) bool {
	skip := 0
	if len(msg.Array) == 0 {
		return false
	}
	if msg.Array[0].Type == qbot.Reply {
		skip++
		if len(msg.Array) > 1 && msg.Array[1].Type == qbot.At {
			skip++
		}
	} else if msg.Array[0].Type == qbot.At {
		skip++
	}

	var raw string
	if skip != 0 {
		if p := findNthClosingBracket(msg.Raw, skip); p != len(msg.Raw) {
			raw = msg.Raw
			msg.Raw = msg.Raw[p:]
		} else {
			return false
		}
	}
	handler := findCommand(getCommandName(msg.Raw))
	if handler != nil {
		if args := splitArguments(msg, skip); args != nil {
			handler(c, msg, args)
		}
	}
	if raw != "" {
		msg.Raw = raw
	}
	return handler != nil
}

func splitArguments(msg *qbot.Message, skip int) *ArgsList {
	result := &ArgsList{
		Contents: make([]string, 0, 20),
		Types:    make([]qbot.MsgType, 0, 20),
		Size:     0,
	}

	if skip < 0 {
		skip = 0
	}

	if skip >= len(msg.Array) {
		return result
	}

	for _, item := range msg.Array[skip:] {
		if item.Type == qbot.Text {
			texts, err := shlex.Split(item.Content)
			if err != nil {
				return nil
			}
			result.Contents = append(result.Contents, texts...)
			result.Types = appendRepeatedValues(result.Types, qbot.Text, len(texts))
			result.Size += len(texts)
		} else {
			result.Contents = append(result.Contents, item.Content)
			result.Types = append(result.Types, item.Type)
			result.Size++
		}
	}
	return result
}

func findNthClosingBracket(s string, n int) int {
	count := 0
	for i, char := range s {
		if char == ']' {
			count++
			if count == n {
				i++
				for i < len(s) && s[i] == ' ' {
					i++
				}
				return i
			}
		}
	}
	return 0
}

func findCommand(cmd string) CmdHandler {
	if cmd == "" {
		return nil
	}
	return cmdMap[cmd]
}

func getCommandName(s string) string {
	sliced := false
	if len(s) > maxCommandLength+1 {
		s = s[:maxCommandLength+1]
		sliced = true
	}
	if i := strings.IndexAny(s, " \n"); i != -1 {
		return s[:i]
	}
	if sliced {
		return ""
	}
	return s
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

func appendRepeatedValues[T any](slice []T, value T, count int) []T {
	newSlice := make([]T, len(slice)+count)
	copy(newSlice, slice)
	for i := len(slice); i < len(newSlice); i++ {
		newSlice[i] = value
	}
	return newSlice
}
