package cmds

import (
	"github.com/awfufu/qbot"
)

const echoHelpMsg string = `Echoes messages to a target destination.
Usage: /echo <any>
Example: /echo helloworld`

type EchoCommand struct {
	cmdBase
}

func NewEchoCommand() *EchoCommand {
	return &EchoCommand{
		cmdBase: cmdBase{
			Name:       "echo",
			HelpMsg:    echoHelpMsg,
			Permission: getCmdPermLevel("echo"),

			NeedRawMsg: false,
			MinArgs:    2,
		},
	}
}

func (cmd *EchoCommand) Self() *cmdBase {
	return &cmd.cmdBase
}

func (cmd *EchoCommand) Exec(b *qbot.Sender, msg *qbot.Message) {
	b.SendGroupMsg(msg.GroupID, msg.Array[1:])
}
