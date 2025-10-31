package main

import (
	"go-hurobot/cmds"
	"go-hurobot/config"
	"go-hurobot/legacy"
	"go-hurobot/llm"
	"go-hurobot/mc"
	"go-hurobot/qbot"
)

func messageHandler(c *qbot.Client, msg *qbot.Message) {
	if msg.UserID != config.Cfg.Permissions.BotID {
		isCommand := cmds.HandleCommand(c, msg)
		defer qbot.SaveDatabase(msg, isCommand)

		mc.ForwardMessageToMC(c, msg)

		if isCommand {
			return
		}

		if llm.NeedLLMResponse(msg) {
			llm.LLMMsgHandle(c, msg)
			return
		}
		if legacy.IsGameCommand(msg) {
			legacy.GameCommandHandle(c, msg)
			return
		}
	}
}
