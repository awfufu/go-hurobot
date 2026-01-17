package cmds

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/awfufu/qbot"
)

var forwardCommand = &Command{
	Name:       "forward",
	HelpMsg:    forwardHelpMsg,
	Permission: getCmdPermLevel("forward"),
	NeedRawMsg: false,
	MinArgs:    0,
	Exec:       forwardExec,
}

const forwardHelpMsg = `Construct a merged forward message.
Usage: /forward [options] [-u <user>] -m <message> ...
Options:
  -u, --user <user>      Set message sender (UserID or @mention). Default: yourself.
  -m, --message <msg>    Add message content
  -t, --title <title>    Set Title
  -p, --preview <text>   Set Preview
  -s, --summary <text>   Set Summary
Example:
  /forward -m Hello -m World
  /forward -m hyw? -u 114514 -m 1919810
  /forward -t 群聊的聊天记录 -p 你：你好 -s 查看n条转发消息 -m 你好`

func forwardExec(b *qbot.Sender, msg *qbot.Message) {
	// args start from index 1. Index 0 is command name.
	args := msg.Array[1:]

	if len(args) == 0 {
		b.SendGroupMsg(msg.GroupID, forwardHelpMsg)
		return
	}

	var forwardItems []qbot.ForwardBlockItem
	var currentUser uint64 = uint64(msg.UserID)
	currentName := msg.Name

	// Metadata
	title := ""
	preview := ""
	summary := ""

	i := 0
	for i < len(args) {
		arg := args[i]
		i++

		// Check for flags in TextItem
		if t, ok := arg.(qbot.TextItem); ok {
			s := t.String()
			switch s {
			case "-t", "--title":
				if i < len(args) {
					title = getTextArg(args[i])
					i++
				}
				continue
			case "-p", "--preview":
				if i < len(args) {
					preview = getTextArg(args[i])
					i++
				}
				continue
			case "-s", "--summary":
				if i < len(args) {
					summary = getTextArg(args[i])
					i++
				}
				continue
			case "-u", "--user":
				if i >= len(args) {
					b.SendGroupMsg(msg.GroupID, "Missing value for -u")
					return
				}
				// Next arg is user
				userArg := args[i]
				i++

				parsedUser, parsedName, err := parseUserArg(b, msg.GroupID, userArg)
				if err != nil {
					b.SendGroupMsg(msg.GroupID, fmt.Sprintf("Invalid user argument: %v", err))
					return
				}
				currentUser = parsedUser
				currentName = parsedName
				continue
			case "-m", "--message":
				// Collect message content until next flag
				var segments []qbot.Segment
				first := true

				// Loop until next flag or end
				for i < len(args) {
					// Check if next arg is a flag
					nextArg := args[i]
					if nextT, ok := nextArg.(qbot.TextItem); ok {
						nextS := nextT.String()
						if isFlag(nextS) {
							break
						}
					}

					// Add space if not first arg in this -m block
					if !first {
						segments = append(segments, qbot.Text(" "))
					}

					// Not a flag, append to segments
					seg := msgItemToSegment(nextArg)
					segments = append(segments, seg)
					first = false
					i++
				}

				if len(segments) == 0 {
					// Empty message? Skip or allow?
					continue
				}

				// Add to forward items
				forwardItems = append(forwardItems, qbot.ForwardBlockItem{
					Name:    currentName,
					UserID:  currentUser,
					Content: segments,
				})
				continue
			}
		}

	}

	if len(forwardItems) == 0 {
		b.SendGroupMsg(msg.GroupID, "Error: No messages provided. Use -m to add content.")
		return
	}

	// Construct Block
	block := qbot.ForwardBlock{
		Title:   title,
		Preview: preview,
		Summary: summary,
		Prompt:  "",
		Content: forwardItems,
	}

	_, _, err := b.SendGroupForward(msg.GroupID, block)
	if err != nil {
		b.SendGroupMsg(msg.GroupID, fmt.Sprintf("Failed to send forward message: %v", err))
	}
}

func parseUserArg(b *qbot.Sender, groupID qbot.GroupID, arg qbot.MsgItem) (uint64, string, error) {
	if at, ok := arg.(qbot.AtItem); ok {
		uid := uint64(at)
		return uid, getDisplayName(b, groupID, uid), nil
	}

	if t, ok := arg.(qbot.TextItem); ok {
		s := t.String()
		// Strip leading @ if present in text
		s = strings.TrimPrefix(s, "@")

		// Try parsing as uint64
		uid, err := strconv.ParseUint(s, 10, 64)
		if err == nil {
			return uid, getDisplayName(b, groupID, uid), nil
		}
		return 0, "", fmt.Errorf("invalid user ID: %s", s)
	}

	return 0, "", fmt.Errorf("unsupported argument type for user")
}

func getDisplayName(b *qbot.Sender, groupID qbot.GroupID, userID uint64) string {

	info, err := b.GetGroupMemberInfo(groupID, qbot.UserID(userID), false)
	if err != nil || info == nil {
		return strconv.FormatUint(userID, 10)
	}
	if info.Card != "" {
		return info.Card
	}
	if info.Nickname != "" {
		return info.Nickname
	}
	return strconv.FormatUint(userID, 10)
}

func msgItemToSegment(item qbot.MsgItem) qbot.Segment {
	switch v := item.(type) {
	case qbot.TextItem:
		return qbot.Text(string(v))
	case qbot.AtItem:
		return qbot.At(qbot.UserID(v))
	case qbot.FaceItem:
		return qbot.Face(qbot.FaceID(v))
	case *qbot.ImageItem:
		return qbot.Image(v.Url)
	default:
		return qbot.Text(fmt.Sprintf("%v", v))
	}
}

func isFlag(s string) bool {
	return s == "-u" ||
		s == "--user" ||
		s == "-m" ||
		s == "--message" ||
		s == "-t" ||
		s == "--title" ||
		s == "-p" ||
		s == "--preview" ||
		s == "-s" ||
		s == "--summary"
}

func getTextArg(item qbot.MsgItem) string {
	if t, ok := item.(qbot.TextItem); ok {
		return t.String()
	}
	return fmt.Sprintf("%v", item)
}
