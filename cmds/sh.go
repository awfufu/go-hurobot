package cmds

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/awfufu/go-hurobot/config"
	"github.com/awfufu/qbot"
)

const shHelpMsg string = `Execute shell commands.
Usage: /sh <command>
Example: /sh ls -la`

const user = "mc"
const masterUser = "awfufu"
const home = "/home/" + user
const masterHome = "/home/" + masterUser

var workingDir string = home
var masterWorkingDir string = masterHome

func truncateString(s string) string {
	s = encodeSpecialChars(s)
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}

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
			Name:       "sh",
			HelpMsg:    shHelpMsg,
			Permission: getCmdPermLevel("sh"),

			NeedRawMsg: true,
			MinArgs:    2,
		},
	}
}

func (cmd *ShCommand) Self() *cmdBase {
	return &cmd.cmdBase
}

func (cmd *ShCommand) Exec(b *qbot.Bot, msg *qbot.Message) {
	// For NeedRawMsg, msg.Array[1] contains the raw arguments string.
	// But let's verify if we need parsing for --reset
	// The original code checked args[1] == "--reset".
	// If raw message is used, msg.Array should still have split parts?
	// Step 77: if cmdBase.NeedRawMsg { args = ... cmdName, raw ... }
	// So Array[1] is the WHOLE raw string.
	// So we need to check if the raw string STARTS with --reset or IS --reset.

	if len(msg.Array) < 2 {
		return
	}

	rawArgs := ""
	if txt := msg.Array[1].GetTextItem(); txt != nil {
		rawArgs = txt.Content
	}

	isMaster := config.GetUserPermission(msg.UserID) == config.Master
	if strings.TrimSpace(rawArgs) == "--reset" {
		if isMaster {
			masterWorkingDir = masterHome
			b.SendGroupReplyMsg(msg.GroupID, msg.MsgID, "working dir reset to "+masterWorkingDir)
		} else {
			workingDir = home
			b.SendGroupReplyMsg(msg.GroupID, msg.MsgID, "working dir reset to "+workingDir)
		}
		return
	}

	rawcmd := decodeSpecialChars(rawArgs)

	var shellCmd *exec.Cmd = nil
	if isMaster {
		fullCmd := fmt.Sprintf("cd %s && %s; echo '\n'; pwd", masterWorkingDir, rawcmd)
		shellCmd = exec.Command("zsh", "-c", fullCmd)
	} else {
		fullCmd := fmt.Sprintf("cd %s && %s; echo '\n'; pwd", workingDir, rawcmd)
		shellCmd = exec.Command("ssh", user+"@127.0.0.1", "zsh", "-c", fullCmd)
	}

	done := make(chan error, 1)
	var output []byte

	go func() {
		var err error
		output, err = shellCmd.CombinedOutput()
		log.Printf("run command: %s, output: %s, error: %v",
			rawArgs, string(output), err)
		done <- err
	}()

	select {
	case err := <-done:
		outputStr := string(output)

		// 解析输出的最后一行来更新workingDir
		lines := strings.Split(strings.TrimSpace(outputStr), "\n")
		if len(lines) > 0 {
			lastLine := strings.TrimSpace(lines[len(lines)-1])
			if strings.HasPrefix(lastLine, "/") {
				if _, statErr := os.Stat(lastLine); statErr == nil {
					if isMaster {
						masterWorkingDir = lastLine
					} else {
						workingDir = lastLine
					}
					if len(lines) > 1 {
						outputStr = strings.Join(lines[:len(lines)-1], "\n")
					} else {
						outputStr = ""
					}
				}
			}
		}

		if err == nil {
			// success
			if outputStr != "" {
				b.SendGroupReplyMsg(msg.GroupID, msg.MsgID, truncateString(outputStr))
			} else {
				b.SendGroupReplyMsg(msg.GroupID, msg.MsgID, "ok")
			}
		} else {
			// failed
			b.SendGroupReplyMsg(msg.GroupID, msg.MsgID, fmt.Sprintf("%v\n%s", err, truncateString(outputStr)))
		}
	case <-time.After(300 * time.Second):
		shellCmd.Process.Kill()
		b.SendGroupReplyMsg(msg.GroupID, msg.MsgID, fmt.Sprintf("Timeout: %q", rawcmd))
	}
}
