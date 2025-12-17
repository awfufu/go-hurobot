package main

import (
	"github.com/awfufu/go-hurobot/internal/cmds"
	"github.com/awfufu/go-hurobot/internal/config"
	"github.com/awfufu/go-hurobot/internal/db"
	"github.com/awfufu/qbot"
)

func main() {
	config.LoadConfigFile()
	db.InitDB()
	cmds.InitCommandPermissions()

	receiver := qbot.HttpServer(config.Cfg.HttpListen)
	sender := qbot.HttpClient(config.Cfg.HttpRemote)

	for {
		select {
		case msg := <-receiver.OnMessage():
			go func() {
				if msg.ChatType != qbot.Group {
					return
				}
				defer db.SaveDatabase(msg)

				if msg.UserID != config.Cfg.Permissions.BotID {
					cmds.HandleCommand(sender, msg)
				}
			}()
		}
	}
}
