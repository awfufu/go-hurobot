package mc

import (
	"fmt"
	"strings"

	"go-hurobot/config"
	"go-hurobot/qbot"

	"github.com/gorcon/rcon"
)

// ForwardMessageToMC forwards a group message to Minecraft server if RCON is enabled
func ForwardMessageToMC(c *qbot.Client, msg *qbot.Message) {
	// Skip bot's own messages
	if msg.UserID == config.BotID {
		return
	}

	// Get RCON configuration for this group
	var rconConfig qbot.GroupRconConfigs
	result := qbot.PsqlDB.Where("group_id = ?", msg.GroupID).Limit(1).Find(&rconConfig)

	// Skip if RCON not configured or disabled
	if result.Error != nil {
		return
	}
	if result.RowsAffected == 0 {
		// Silently return if RCON not configured for this group
		return
	}

	if !rconConfig.Enabled {
		return
	}

	// Get user's nickname from database
	var user qbot.Users
	nickname := msg.Card // Default to group card name

	userResult := qbot.PsqlDB.Where("user_id = ?", msg.UserID).Limit(1).Find(&user)
	if userResult.Error == nil && userResult.RowsAffected > 0 && user.Nickname != "" {
		nickname = user.Nickname
	}

	// Process message array to create tellraw components
	tellrawComponents := buildTellrawComponents(nickname, msg.Array)

	// Create tellraw command with JSON array format
	tellrawCmd := fmt.Sprintf("tellraw @a %s", tellrawComponents)

	// Execute the command
	executeRconCommand(rconConfig.Address, rconConfig.Password, tellrawCmd)
}

// buildTellrawComponents creates JSON components for tellraw command
func buildTellrawComponents(senderNickname string, msgArray []qbot.MsgItem) string {
	// Start with the sender nickname in aqua color
	components := []string{
		`""`, // Empty string as the base
		fmt.Sprintf(`{"text":"<%s> ","color":"aqua"}`, escapeMinecraftText(senderNickname)),
	}

	// Process each message item
	for i, item := range msgArray {
		// Add space between different elements (except for the first one)
		if i > 0 {
			components = append(components, `{"text":" "}`)
		}

		switch item.Type {
		case qbot.Text:
			// Clean and add text content in white color
			cleanText := cleanMessageForMC(item.Content)
			if cleanText != "" {
				components = append(components,
					fmt.Sprintf(`{"text":"%s","color":"white"}`, escapeMinecraftText(cleanText)))
			}
		case qbot.At:
			// Handle @ mentions in gray with underline
			atNickname := getAtUserNickname(item.Content)
			components = append(components,
				fmt.Sprintf(`{"text":"@%s","underlined":true,"color":"gray"}`, escapeMinecraftText(atNickname)))
		case qbot.Image:
			components = append(components, `{"text":"[图片]","color":"gray"}`)
		case qbot.Record:
			components = append(components, `{"text":"[语音]","color":"gray"}`)
		case qbot.Face:
			// Get the actual face name
			faceName := qbot.GetQFaceNameByStrID(item.Content)
			if faceName != "" && faceName != item.Content {
				components = append(components,
					fmt.Sprintf(`{"text":"[%s]","color":"gray"}`, escapeMinecraftText(faceName)))
			} else {
				components = append(components, `{"text":"[表情]","color":"gray"}`)
			}
		case qbot.File:
			components = append(components, `{"text":"[文件]","color":"gray"}`)
		case qbot.Forward:
			components = append(components, `{"text":"[转发]","color":"gray"}`)
		case qbot.Reply:
			components = append(components, `{"text":"[回复]","color":"gray"}`)
		case qbot.Json:
			components = append(components, `{"text":"[JSON]","color":"gray"}`)
		default:
			components = append(components, `{"text":"[其他]","color":"gray"}`)
		}
	}

	// Join components into JSON array format
	return "[" + strings.Join(components, ",") + "]"
}

// getAtUserNickname gets the nickname for an @ mention
func getAtUserNickname(atContent string) string {
	// Extract user ID from @ content (usually just the user ID string)
	userID := str2uin64(atContent)
	if userID == 0 {
		return atContent // Return original if parsing fails
	}

	// Look up user in database
	var user qbot.Users
	result := qbot.PsqlDB.Where("user_id = ?", userID).Limit(1).Find(&user)

	// Use nick_name if available, otherwise try to get group member info
	if result.Error == nil && result.RowsAffected > 0 && user.Nickname != "" {
		return user.Nickname
	}

	// Fallback to user ID if no nickname found
	return fmt.Sprintf("用户%d", userID)
}

// cleanMessageForMC removes or replaces characters that might cause issues in Minecraft
func cleanMessageForMC(content string) string {
	// Remove or replace problematic characters
	content = strings.ReplaceAll(content, "\n", " ")
	content = strings.ReplaceAll(content, "\r", " ")
	content = strings.ReplaceAll(content, "\t", " ")

	// Remove multiple spaces
	for strings.Contains(content, "  ") {
		content = strings.ReplaceAll(content, "  ", " ")
	}

	return strings.TrimSpace(content)
}

// escapeMinecraftText escapes special characters for Minecraft JSON text
func escapeMinecraftText(text string) string {
	text = strings.ReplaceAll(text, "\\", "\\\\")
	text = strings.ReplaceAll(text, "\"", "\\\"")
	text = strings.ReplaceAll(text, "\n", "\\n")
	text = strings.ReplaceAll(text, "\r", "\\r")
	text = strings.ReplaceAll(text, "\t", "\\t")
	return text
}

// executeRconCommand executes a command on the Minecraft server via RCON
func executeRconCommand(address, password, command string) (string, error) {
	// Connect to RCON server
	conn, err := rcon.Dial(address, password)
	if err != nil {
		return "", fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	// Execute command
	response, err := conn.Execute(command)
	if err != nil {
		return "", fmt.Errorf("failed: %w", err)
	}

	return response, nil
}

// str2uin64 converts string to uint64
func str2uin64(s string) uint64 {
	// This function should match the implementation in cmds/cmds.go
	// We'll import it from there or duplicate the simple logic
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0
		}
	}

	var result uint64
	for _, c := range s {
		result = result*10 + uint64(c-'0')
	}
	return result
}
