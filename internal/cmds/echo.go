package cmds

import (
	"github.com/awfufu/qbot"
)

const echoHelpMsg string = `Echoes messages to a target destination.
Usage: /echo <any>
Example: /echo helloworld`

var echoCommand *Command = &Command{
	Name:       "echo",
	HelpMsg:    echoHelpMsg,
	Permission: getCmdPermLevel("echo"),
	NeedRawMsg: false,
	MinArgs:    2,
	Exec: func(b *qbot.Sender, msg *qbot.Message) {
		b.SendGroupMsg(msg.GroupID, msg.Array[1:])
	},
}
