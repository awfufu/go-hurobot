package cmds

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/awfufu/go-hurobot/db"
	"github.com/awfufu/qbot"
)

const psqlHelpMsg = `Execute PostgreSQL queries.
Usage: /psql <query>
Example: /psql SELECT * FROM users LIMIT 10`

type PsqlCommand struct {
	cmdBase
}

func NewPsqlCommand() *PsqlCommand {
	return &PsqlCommand{
		cmdBase: cmdBase{
			Name:       "psql",
			HelpMsg:    psqlHelpMsg,
			Permission: getCmdPermLevel("psql"),

			NeedRawMsg: true, // uses raw message
			MinArgs:    2,
		},
	}
}

func (cmd *PsqlCommand) Self() *cmdBase {
	return &cmd.cmdBase
}

func (cmd *PsqlCommand) Exec(b *qbot.Bot, msg *qbot.Message) {
	// For raw message commands, we typically want the arguments after the command name.
	// Since handleCommand splits args for text items, msg.Array contains parsed parts.
	// But psql wants the raw query string probably to avoiding parsing issues?
	// New logic in handleCommand: if NeedRawMsg, args includes raw string as second element?
	// Let's check handleCommand refactoring in Step 77.
	/*
		if cmdBase.NeedRawMsg {
			args = []qbot.MsgItem{&qbot.TextItem{Content: cmdName}}
			if len(raw) > 0 {
				args = append(args, &qbot.TextItem{Content: raw})
			}
		}
	*/
	// Yes, if NeedRawMsg is true, msg.Array[1] will be the raw content after command.

	if len(msg.Array) < 2 {
		b.SendGroupMsg(msg.GroupID, cmd.HelpMsg)
		return
	}

	query := ""
	if txt := msg.Array[1].GetTextItem(); txt != nil {
		query = txt.Content
	}

	if query == "" {
		b.SendGroupMsg(msg.GroupID, cmd.HelpMsg)
		return
	}

	rows, err := db.PsqlDB.Raw(decodeSpecialChars(query)).Rows()
	if err != nil {
		b.SendGroupMsg(msg.GroupID, err.Error())
		return
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		b.SendGroupMsg(msg.GroupID, err.Error())
		return
	}

	result := ""
	count := 1
	for rows.Next() {
		if count == 10 {
			result += "\n\n** more... **"
			break
		}

		values := make([]any, len(columns))
		for i := range values {
			values[i] = new(sql.RawBytes)
		}

		if err := rows.Scan(values...); err != nil {
			b.SendGroupMsg(msg.GroupID, err.Error())
			return
		}

		var rowStrings []string
		for i, col := range values {
			rowStrings = append(rowStrings, fmt.Sprintf("%s: %s", columns[i], string(*col.(*sql.RawBytes))))
		}

		if result != "" {
			result += "\n\n"
		}
		result += fmt.Sprintf("** %d **\n", count)
		result += strings.Join(rowStrings, "\n")
		count++
	}
	if err = rows.Err(); err != nil {
		b.SendGroupMsg(msg.GroupID, err.Error())
	} else if result == "" {
		b.SendGroupMsg(msg.GroupID, "[]")
	} else {
		b.SendGroupMsg(msg.GroupID, encodeSpecialChars(result))
	}
}
