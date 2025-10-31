package cmds

import (
	"fmt"
	"go-hurobot/config"
	"go-hurobot/qbot"
	"log"
	"os/exec"
	"strings"
	"time"
)

const pyHelpMsg string = `Execute Python code.
Usage: /py <python_code>
Example: /py print("Hello, World!")`

type PyCommand struct {
	cmdBase
}

func NewPyCommand() *PyCommand {
	return &PyCommand{
		cmdBase: cmdBase{
			Name:        "py",
			HelpMsg:     pyHelpMsg,
			Permission:  getCmdPermLevel("py"),
			AllowPrefix: false,
			NeedRawMsg:  true,
			MinArgs:     2,
		},
	}
}

func (cmd *PyCommand) Self() *cmdBase {
	return &cmd.cmdBase
}

func (cmd *PyCommand) Exec(c *qbot.Client, args []string, src *srcMsg, begin int) {
	if len(args) <= 1 {
		c.SendMsg(src.GroupID, src.UserID, qbot.CQReply(src.MsgID)+cmd.HelpMsg)
		return
	}

	// get interpreter path
	interpreter := config.Cfg.Python.Interpreter
	if interpreter == "" {
		interpreter = "python3"
	}

	pythonCode := decodeSpecialChars(src.Raw)
	pythonCode = strings.TrimSpace(pythonCode)

	pythonCmd := exec.Command(interpreter, "-c", pythonCode)

	done := make(chan error, 1)
	var output []byte

	go func() {
		var err error
		output, err = pythonCmd.CombinedOutput()
		log.Printf("run python command: %s, output: %s, error: %v",
			pythonCode, string(output), err)
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
		pythonCmd.Process.Kill()
		c.SendMsg(src.GroupID, src.UserID, qbot.CQReply(src.MsgID)+fmt.Sprintf("Timeout: %q", pythonCode))
	}
}
