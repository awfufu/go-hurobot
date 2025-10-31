package cmds

import (
	"database/sql"
	"fmt"
	"go-hurobot/qbot"
	"strings"
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
			Name:        "psql",
			HelpMsg:     psqlHelpMsg,
			Permission:  getCmdPermLevel("psql"),
			AllowPrefix: false,
			NeedRawMsg:  true, // uses raw message
			MinArgs:     2,
		},
	}
}

func (cmd *PsqlCommand) Self() *cmdBase {
	return &cmd.cmdBase
}

func (cmd *PsqlCommand) Exec(c *qbot.Client, args []string, src *srcMsg, _ int) {
	query := args[len(args)-1]
	rows, err := qbot.PsqlDB.Raw(decodeSpecialChars(query)).Rows()
	if err != nil {
		c.SendMsg(src.GroupID, src.UserID, err.Error())
		return
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		c.SendMsg(src.GroupID, src.UserID, err.Error())
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
			c.SendMsg(src.GroupID, src.UserID, err.Error())
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
		c.SendMsg(src.GroupID, src.UserID, err.Error())
	} else if result == "" {
		c.SendMsg(src.GroupID, src.UserID, "[]")
	} else {
		c.SendMsg(src.GroupID, src.UserID, encodeSpecialChars(result))
	}
}
