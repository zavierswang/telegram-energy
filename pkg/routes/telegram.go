package routes

import (
	"github.com/mr-linch/go-tg"
	"github.com/mr-linch/go-tg/tgb"
	"regexp"
	"telegram-energy/pkg/controllers"
	"telegram-energy/pkg/middleware"
)

func Telegram(router *tgb.Router) {
	router.Use(middleware.SessionManager)
	router.Use(tgb.MiddlewareFunc(middleware.Hook))

	router.Message(controllers.Start, tgb.Command("start"), tgb.ChatType(tg.ChatTypePrivate))
	router.Message(controllers.Help, tgb.Any(tgb.Command("help"), tgb.TextEqual(controllers.Menu.Help)), tgb.ChatType(tg.ChatTypePrivate))
	router.Message(controllers.Rent, tgb.TextEqual(controllers.Menu.Rent), tgb.ChatType(tg.ChatTypePrivate))
	router.Message(controllers.Orders, tgb.TextEqual(controllers.Menu.Orders), tgb.ChatType(tg.ChatTypePrivate))
	router.Message(controllers.Personal, tgb.TextEqual(controllers.Menu.Personal), tgb.ChatType(tg.ChatTypePrivate))
	//router.Message(controllers.Receiver, tgb.Regexp(regexp.MustCompile(`T\w+`)), tgb.ChatType(tg.ChatTypePrivate))

	router.CallbackQuery(controllers.RentReset, middleware.IsSessionStep(middleware.SessionRentEnergy), tgb.TextEqual(controllers.Callback.RentReset), tgb.ChatType(tg.ChatTypePrivate))
	router.CallbackQuery(controllers.RentChoose, middleware.IsSessionStep(middleware.SessionRentEnergy), tgb.Regexp(regexp.MustCompile(`^rent\s+(?P<type>[a-z]+)\s+(?P<action>[a-z]+)?\s?(?P<number>(?:0|[1-9]\d*)(?:\.\d*)?)_?(?P<multiplier>\d+\.?\d+?)?`)), tgb.ChatType(tg.ChatTypePrivate))
	router.CallbackQuery(controllers.RentConfirm, middleware.IsSessionStep(middleware.SessionRentEnergy), tgb.TextEqual(controllers.Callback.RentConfirm), tgb.ChatType(tg.ChatTypePrivate))
	router.CallbackQuery(controllers.RentPayment, middleware.IsSessionStep(middleware.SessionRentEnergy), tgb.TextEqual(controllers.Callback.RentPayment), tgb.ChatType(tg.ChatTypePrivate))
	router.CallbackQuery(controllers.Cancel, middleware.IsSessionStep(middleware.SessionRentEnergy), tgb.TextEqual(controllers.Callback.Cancel), tgb.ChatType(tg.ChatTypePrivate))
	router.CallbackQuery(controllers.Close, tgb.TextEqual(controllers.Callback.Close), tgb.ChatType(tg.ChatTypePrivate))
	router.CallbackQuery(controllers.Empty, tgb.TextEqual(controllers.Callback.Empty), tgb.ChatType(tg.ChatTypePrivate))
}
