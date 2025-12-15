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

func (cmd *EchoCommand) Exec(b *qbot.Bot, msg *qbot.Message) {
	result := make([]qbot.Segment, 0, 8)
	for _, e := range msg.Array {
		switch v := e.(type) {
		case *qbot.TextItem:
			result = append(result, qbot.Text(v.Content))
		case *qbot.AtItem:
			result = append(result, qbot.At(v.TargetID))
		case *qbot.FaceItem:
			result = append(result, qbot.Face(v.ID))
		case *qbot.ImageItem:
			result = append(result, qbot.Image(v.URL))
		}
	}
	b.SendGroupMsg(msg.GroupID, result)
}
