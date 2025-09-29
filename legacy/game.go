package legacy

import (
	"errors"
	"go-hurobot/qbot"
	"strings"

	"gorm.io/gorm"
)

func IsGameCommand(msg *qbot.Message) bool {
	msg0 := msg.Array[0].Content
	return msg.Array[0].Type == qbot.Text && strings.HasPrefix(msg0, "&")
}

func GameCommandHandle(c *qbot.Client, msg *qbot.Message) {
	cmd := msg.Array[0].Content[1:]

	var gameData qbot.LegacyGame
	result := qbot.PsqlDB.Where("user_id = ?", msg.UserID).First(&gameData)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			newGameData := qbot.LegacyGame{
				UserID: msg.UserID,
			}
			if err := qbot.PsqlDB.Create(&newGameData).Error; err != nil {
				return
			}
			gameData = newGameData
		} else {
			return
		}
	}
	_ = gameData
	_ = cmd
}
