package main

import (
	"log"

	"github.com/awfufu/go-hurobot/internal/cmds"
	"github.com/awfufu/go-hurobot/internal/config"
	"github.com/awfufu/go-hurobot/internal/db"
	"github.com/awfufu/qbot"
)

func main() {
	config.LoadConfigFile()
	db.InitDB()
	cmds.InitCommandPermissions()

	bot := qbot.NewBot(config.Cfg.HttpListen)
	bot.ConnectNapcat(config.Cfg.HttpRemote)

	bot.OnMessage(func(b *qbot.Bot, msg *qbot.Message) {
		if msg.ChatType != qbot.Group {
			return
		}
		defer db.SaveDatabase(msg)

		if msg.UserID != config.Cfg.Permissions.BotID {
			cmds.HandleCommand(b, msg)
		}
	})

	log.Fatal(bot.Run())
}
