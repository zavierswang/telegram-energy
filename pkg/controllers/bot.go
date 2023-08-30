package controllers

import "github.com/mr-linch/go-tg"

var Menu = struct {
	Start    string
	Rent     string
	Help     string
	Voucher  string
	Orders   string
	Personal string
}{
	Start:    "🥳 开始",
	Help:     "🌟帮助中心",
	Rent:     "🔋能量租赁",
	Voucher:  "🏦充值余额",
	Orders:   "📋消费订单",
	Personal: "👤个人中心",
}

type Bot struct {
	ReplayMarkup *tg.ReplyKeyboardMarkup
	Cmd          []tg.BotCommand
}

func NewBot() *Bot {
	layout := tg.NewReplyKeyboardMarkup(
		tg.NewButtonRow(
			tg.NewKeyboardButton(Menu.Rent),
			tg.NewKeyboardButton(Menu.Orders),
			tg.NewKeyboardButton(Menu.Help),
		),
		//tg.NewButtonRow(
		//	tg.NewKeyboardButton(Menu.Help),
		//	tg.NewKeyboardButton(Menu.Personal),
		//),
	)
	layout.ResizeKeyboard = true

	botCmd := []tg.BotCommand{
		{Command: "start", Description: Menu.Start},
		{Command: "help", Description: Menu.Help},
	}
	return &Bot{
		ReplayMarkup: layout,
		Cmd:          botCmd,
	}
}
