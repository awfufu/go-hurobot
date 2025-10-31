package cmds

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"go-hurobot/qbot"
)

const testapiHelpMsg string = `Test API functionality.
Usage: /testapi <action> [arg1="value1" arg2=value2 ...]
Example: /testapi get_login_info`

type TestapiCommand struct {
	cmdBase
}

func NewTestapiCommand() *TestapiCommand {
	return &TestapiCommand{
		cmdBase: cmdBase{
			Name:        "testapi",
			HelpMsg:     testapiHelpMsg,
			Permission:  getCmdPermLevel("testapi"),
			AllowPrefix: false,
			NeedRawMsg:  false,
		},
	}
}

func (cmd *TestapiCommand) Self() *cmdBase {
	return &cmd.cmdBase
}

func (cmd *TestapiCommand) Exec(c *qbot.Client, args []string, src *srcMsg, _ int) {
	if len(args) < 2 {
		c.SendMsg(src.GroupID, src.UserID, cmd.HelpMsg)
		return
	}

	action := args[1]

	// parse parameters
	params := make(map[string]any)

	for i := 2; i < len(args); i++ {
		arg := args[i]

		// find equal sign separator
		if idx := strings.Index(arg, "="); idx != -1 {
			key := arg[:idx]
			value := arg[idx+1:]

			// handle $group variable replacement
			if value == "$group" {
				params[key] = src.GroupID
				continue
			}

			// check if it is a string type (surrounded by quotes)
			if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
				// string type, remove quotes
				params[key] = value[1 : len(value)-1]
			} else {
				// try to parse as numeric type
				if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
					params[key] = intVal
				} else if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
					params[key] = floatVal
				} else if boolVal, err := strconv.ParseBool(value); err == nil {
					params[key] = boolVal
				} else {
					// if all parsing fails, treat as string
					params[key] = value
				}
			}
		}
	}

	// call test API
	resp, err := c.SendTestAPIRequest(action, params)

	var result string
	if err != nil {
		result = fmt.Sprintf("Error: %v", err)
	} else if resp == "" {
		result = "null"
	} else {
		// directly use the returned JSON string, and format it
		var jsonMap map[string]interface{}
		if err := json.Unmarshal([]byte(resp), &jsonMap); err == nil {
			if jsonBytes, err := json.MarshalIndent(jsonMap, "", "  "); err == nil {
				result = string(jsonBytes)
			} else {
				result = resp // if formatting fails, return the original string
			}
		} else {
			result = resp // if parsing fails, return the original string
		}
	}

	c.SendMsg(src.GroupID, src.UserID, encodeSpecialChars(result))
}
