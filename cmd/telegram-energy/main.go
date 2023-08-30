package main

import (
	"telegram-energy/pkg/bootstrap"
	"telegram-energy/pkg/core/cst"
)

func main() {
	// 初始配置文件
	bootstrap.LoadConfig(cst.AppName)
	//logger.Info("configs: %#v", global.App.Config)
	// 初始化DB
	bootstrap.ConnectDB()
	// 初始化私钥
	bootstrap.StoreKey()
	// Telegram Bot
	err := bootstrap.Telegram()
	if err != nil {
		panic(err)
	}
}
