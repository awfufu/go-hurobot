package main

import (
	"go-hurobot/cmds"
	"go-hurobot/config"
	"go-hurobot/llm"
	"go-hurobot/mc"
	"go-hurobot/qbot"
)

func messageHandler(c *qbot.Client, msg *qbot.Message) {
	if msg.UserID != config.BotID {
		isCommand := cmds.HandleCommand(c, msg)
		defer qbot.SaveDatabase(msg, isCommand)

		if isCommand {
			return
		} else {
			mc.ForwardMessageToMC(c, msg)
		}

		if llm.NeedLLMResponse(msg) {
			llm.LLMMsgHandle(c, msg)
			return
		}
		cmds.CheckUserEvents(c, msg)
	}
}
