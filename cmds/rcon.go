package cmds

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/awfufu/go-hurobot/db"
	"github.com/awfufu/qbot"
)

const rconHelpMsg string = `Manage RCON configuration.
Usage: /rcon [status | set <address> <password> | enable | disable]
Examples:
  /rcon status
  /rcon set 127.0.0.1:25575 password
  /rcon enable
  /rcon disable`

type RconCommand struct {
	cmdBase
}

func NewRconCommand() *RconCommand {
	return &RconCommand{
		cmdBase: cmdBase{
			Name:       "rcon",
			HelpMsg:    rconHelpMsg,
			Permission: getCmdPermLevel("rcon"),

			NeedRawMsg: false,
			MaxArgs:    4,
			MinArgs:    2,
		},
	}
}

func (cmd *RconCommand) Self() *cmdBase {
	return &cmd.cmdBase
}

func (cmd *RconCommand) Exec(b *qbot.Bot, msg *qbot.Message) {
	if len(msg.Array) < 2 {
		b.SendGroupMsg(msg.GroupID, cmd.HelpMsg)
		return
	}

	getText := func(i int) string {
		if i < len(msg.Array) {
			if txt := msg.Array[i].GetTextItem(); txt != nil {
				return txt.Content
			}
		}
		return ""
	}

	subCmd := getText(1)
	switch subCmd {
	case "status":
		showRconStatus(b, msg)
	case "set":
		// /rcon set address password
		if len(msg.Array) != 4 {
			b.SendGroupMsg(msg.GroupID, cmd.HelpMsg)
			return
		}
		setRconConfig(b, msg, getText(2), getText(3))
	case "enable":
		toggleRcon(b, msg, true)
	case "disable":
		toggleRcon(b, msg, false)
	default:
		b.SendGroupMsg(msg.GroupID, cmd.HelpMsg)
	}
}

func showRconStatus(b *qbot.Bot, msg *qbot.Message) {
	var config db.GroupRconConfigs
	result := db.PsqlDB.Where("group_id = ?", msg.GroupID).First(&config)

	if result.Error != nil {
		b.SendGroupMsg(msg.GroupID, "RCON not configured for this group")
		return
	}

	status := "disabled"
	if config.Enabled {
		status = "enabled"
	}

	// Hide password for security
	maskedPassword := strings.Repeat("*", len(config.Password))
	response := fmt.Sprintf("RCON Status: %s\nAddress: %s\nPassword: %s",
		status, config.Address, maskedPassword)

	b.SendGroupMsg(msg.GroupID, response)
}

func setRconConfig(b *qbot.Bot, msg *qbot.Message, address, password string) {
	// Validate address format (should contain port)
	if !strings.Contains(address, ":") {
		b.SendGroupMsg(msg.GroupID, "Invalid address format. Use host:port (e.g., 127.0.0.1:25575)")
		return
	}

	// Validate port
	parts := strings.Split(address, ":")
	if len(parts) != 2 {
		b.SendGroupMsg(msg.GroupID, "Invalid address format. Use host:port")
		return
	}

	if port, err := strconv.Atoi(parts[1]); err != nil || port < 1 || port > 65535 {
		b.SendGroupMsg(msg.GroupID, "Invalid port number")
		return
	}

	config := db.GroupRconConfigs{
		GroupID:  msg.GroupID,
		Address:  address,
		Password: password,
		Enabled:  true, // Default to disabled for security -> wait, original code said Default to disabled but set true?
		// Original: Enabled: true, // Default to disabled for security (comment paradox)
		// I will keep it true as per original logic.
	}

	// Use Upsert to create or update
	result := db.PsqlDB.Where("group_id = ?", msg.GroupID).Assign(
		db.GroupRconConfigs{
			Address:  address,
			Password: password,
		},
	).FirstOrCreate(&config)

	if result.Error != nil {
		b.SendGroupMsg(msg.GroupID, "Database error: "+result.Error.Error())
		return
	}

	b.SendGroupMsg(msg.GroupID, fmt.Sprintf("RCON configuration updated: %s:%s", address, password))
}

func toggleRcon(b *qbot.Bot, msg *qbot.Message, enabled bool) {
	// Check if configuration exists
	var config db.GroupRconConfigs
	result := db.PsqlDB.Where("group_id = ?", msg.GroupID).First(&config)

	if result.Error != nil {
		b.SendGroupMsg(msg.GroupID, "RCON not configured for this group. Use 'rcon set' first.")
		return
	}

	// Update enabled status
	result = db.PsqlDB.Model(&config).Where("group_id = ?", msg.GroupID).Update("enabled", enabled)

	if result.Error != nil {
		b.SendGroupMsg(msg.GroupID, "Database error: "+result.Error.Error())
		return
	}

	status := "disabled"
	if enabled {
		status = "enabled"
	}
	b.SendGroupMsg(msg.GroupID, fmt.Sprintf("RCON %s for this group", status))
}
