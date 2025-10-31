package cmds

import (
	"fmt"
	"strconv"
	"strings"

	"go-hurobot/qbot"
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
			Name:        "rcon",
			HelpMsg:     rconHelpMsg,
			Permission:  getCmdPermLevel("rcon"),
			AllowPrefix: false,
			NeedRawMsg:  false,
			MaxArgs:     4,
			MinArgs:     2,
		},
	}
}

func (cmd *RconCommand) Self() *cmdBase {
	return &cmd.cmdBase
}

func (cmd *RconCommand) Exec(c *qbot.Client, args []string, src *srcMsg, begin int) {
	switch args[1] {
	case "status":
		showRconStatus(c, src)
	case "set":
		if len(args) != 4 {
			c.SendMsg(src.GroupID, src.UserID, cmd.HelpMsg)
			return
		}
		setRconConfig(c, src, args[2], args[3])
	case "enable":
		toggleRcon(c, src, true)
	case "disable":
		toggleRcon(c, src, false)
	default:
		c.SendMsg(src.GroupID, src.UserID, cmd.HelpMsg)
	}
}

func showRconStatus(c *qbot.Client, src *srcMsg) {
	var config qbot.GroupRconConfigs
	result := qbot.PsqlDB.Where("group_id = ?", src.GroupID).First(&config)

	if result.Error != nil {
		c.SendMsg(src.GroupID, src.UserID, "RCON not configured for this group")
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

	c.SendMsg(src.GroupID, src.UserID, response)
}

func setRconConfig(c *qbot.Client, src *srcMsg, address, password string) {
	// Validate address format (should contain port)
	if !strings.Contains(address, ":") {
		c.SendMsg(src.GroupID, src.UserID, "Invalid address format. Use host:port (e.g., 127.0.0.1:25575)")
		return
	}

	// Validate port
	parts := strings.Split(address, ":")
	if len(parts) != 2 {
		c.SendMsg(src.GroupID, src.UserID, "Invalid address format. Use host:port")
		return
	}

	if port, err := strconv.Atoi(parts[1]); err != nil || port < 1 || port > 65535 {
		c.SendMsg(src.GroupID, src.UserID, "Invalid port number")
		return
	}

	config := qbot.GroupRconConfigs{
		GroupID:  src.GroupID,
		Address:  address,
		Password: password,
		Enabled:  true, // Default to disabled for security
	}

	// Use Upsert to create or update
	result := qbot.PsqlDB.Where("group_id = ?", src.GroupID).Assign(
		qbot.GroupRconConfigs{
			Address:  address,
			Password: password,
		},
	).FirstOrCreate(&config)

	if result.Error != nil {
		c.SendMsg(src.GroupID, src.UserID, "Database error: "+result.Error.Error())
		return
	}

	c.SendMsg(src.GroupID, src.UserID, fmt.Sprintf("RCON configuration updated: %s:%s", address, password))
}

func toggleRcon(c *qbot.Client, src *srcMsg, enabled bool) {
	// Check if configuration exists
	var config qbot.GroupRconConfigs
	result := qbot.PsqlDB.Where("group_id = ?", src.GroupID).First(&config)

	if result.Error != nil {
		c.SendMsg(src.GroupID, src.UserID, "RCON not configured for this group. Use 'rcon set' first.")
		return
	}

	// Update enabled status
	result = qbot.PsqlDB.Model(&config).Where("group_id = ?", src.GroupID).Update("enabled", enabled)

	if result.Error != nil {
		c.SendMsg(src.GroupID, src.UserID, "Database error: "+result.Error.Error())
		return
	}

	status := "disabled"
	if enabled {
		status = "enabled"
	}
	c.SendMsg(src.GroupID, src.UserID, fmt.Sprintf("RCON %s for this group", status))
}
