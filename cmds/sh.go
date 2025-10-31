package cmds

import (
	"fmt"
	"go-hurobot/qbot"
	"log"
	"os/exec"
	"strings"
	"time"
)

const shHelpMsg string = `Execute shell commands.
Usage: /sh <command>
Example: /sh ls -la`

var workingDir string = "/tmp"

func truncateString(s string) string {
	s = encodeSpecialChars(s)
	const (
		maxLines    = 20
		maxChars    = 1024
		truncateMsg = "... (truncated)"
	)

	lineCount := strings.Count(s, "\n") + 1

	if lineCount >= 11 {
		index := 0
		for range 10 {
			index = strings.Index(s[index:], "\n") + 1 + index
			if index == 0 {
				return s
			}
		}
		return s[:index] + truncateMsg
	}

	if len(s) > maxChars {
		return s[:maxChars] + truncateMsg
	}

	return s
}

type ShCommand struct {
	cmdBase
}

func NewShCommand() *ShCommand {
	return &ShCommand{
		cmdBase: cmdBase{
			Name:        "sh",
			HelpMsg:     shHelpMsg,
			Permission:  getCmdPermLevel("sh"),
			AllowPrefix: false,
			NeedRawMsg:  true,
			MinArgs:     2,
		},
	}
}

func (cmd *ShCommand) Self() *cmdBase {
	return &cmd.cmdBase
}

func (cmd *ShCommand) Exec(c *qbot.Client, args []string, src *srcMsg, begin int) {
	if len(args) <= 1 {
		c.SendMsg(src.GroupID, src.UserID, qbot.CQReply(src.MsgID)+cmd.HelpMsg)
		return
	}

	rawcmd := decodeSpecialChars(src.Raw)

	if strings.HasPrefix(args[1], "cd") {
		absPath, err := exec.Command("bash", "-c",
			fmt.Sprintf("cd %s && %s && pwd", workingDir, rawcmd)).CombinedOutput()

		if err != nil {
			c.SendMsg(src.GroupID, src.UserID, qbot.CQReply(src.MsgID)+err.Error())
			return
		}

		workingDir = strings.TrimSpace(string(absPath))
		c.SendMsg(src.GroupID, src.UserID, qbot.CQReply(src.MsgID)+workingDir)
		return
	}

	shellCmd := exec.Command("bash", "-c", fmt.Sprintf("cd %s && %s", workingDir, rawcmd))

	done := make(chan error, 1)
	var output []byte

	go func() {
		var err error
		output, err = shellCmd.CombinedOutput()
		log.Printf("run command: %s, output: %s, error: %v",
			strings.Join(args[1:], " "), string(output), err)
		done <- err
	}()

	select {
	case err := <-done:
		if err == nil {
			// success
			c.SendMsg(src.GroupID, src.UserID, qbot.CQReply(src.MsgID)+truncateString(string(output)))
		} else {
			// failed
			c.SendMsg(src.GroupID, src.UserID, qbot.CQReply(src.MsgID)+fmt.Sprintf("%v\n%s", err, truncateString(string(output))))
		}
	case <-time.After(300 * time.Second):
		shellCmd.Process.Kill()
		c.SendMsg(src.GroupID, src.UserID, qbot.CQReply(src.MsgID)+fmt.Sprintf("Timeout: %q", rawcmd))
	}
}
