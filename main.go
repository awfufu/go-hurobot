package main

import (
	"log"

	"github.com/awfufu/go-hurobot/cmds"
	"github.com/awfufu/go-hurobot/config"
	"github.com/awfufu/go-hurobot/db"
	"github.com/awfufu/qbot"
)

func main() {
	config.LoadConfigFile()
	db.InitDB()

	bot := qbot.NewBot(config.Cfg.ReverseHttpListen)
	bot.ConnectNapcat(config.Cfg.NapcatHttpServer)

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
