package main

import (
	"fmt"
	"go-hurobot/config"
	"go-hurobot/qbot"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// 初始化配置
	config.LoadConfigFile()
	qbot.InitDB()

	bot := qbot.NewClient()
	defer bot.Close()

	bot.HandleMessage(messageHandler)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	stopSignal := <-stop
	fmt.Println("shutting down:", stopSignal.String())
}
