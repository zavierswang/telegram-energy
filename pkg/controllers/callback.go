package controllers

import (
	"context"
	"github.com/mr-linch/go-tg"
	"github.com/mr-linch/go-tg/tgb"
	"strings"
	"telegram-energy/pkg/core/cst"
	"telegram-energy/pkg/core/global"
	"telegram-energy/pkg/core/logger"
	"telegram-energy/pkg/middleware"
	"telegram-energy/pkg/models"
	"time"
)

var Callback = struct {
	Close        string
	Empty        string
	Cancel       string
	RetryChoose  string
	RentConfirm  string
	RentReset    string
	RentPayment  string
	SwitchCount  string
	SwitchEnergy string
}{
	Close:        "close",
	Empty:        "empty",
	Cancel:       "cancel",
	RetryChoose:  "retry_choose",
	RentConfirm:  "rent_confirm",
	RentReset:    "rent_reset",
	RentPayment:  "rent_payment",
	SwitchCount:  "switch_count",
	SwitchEnergy: "switch_energy",
}

func Empty(ctx context.Context, callback *tgb.CallbackQueryUpdate) error {
	return callback.AnswerText("点击下方按键，选择您需要的套餐", true).ShowAlert(true).DoVoid(ctx)
}

func Close(ctx context.Context, callback *tgb.CallbackQueryUpdate) error {
	chatId := callback.Message.Chat.ID
	userId := callback.Message.From.ID.PeerID()
	username := strings.Replace(callback.From.Username.PeerID(), "@", "", 1)
	messageId := callback.Message.ID
	logger.Info("[%s %s] trigger action [close] controller", userId, username)
	sess := middleware.SessionManager.Get(ctx)
	middleware.SessionManager.Reset(sess)
	return callback.Client.DeleteMessage(chatId, messageId).DoVoid(ctx)
}

func Cancel(ctx context.Context, callback *tgb.CallbackQueryUpdate) error {
	chatId := callback.Message.Chat.ID
	messageId := callback.Message.ID
	username := strings.Replace(callback.From.Username.PeerID(), "@", "", 1)
	userId := callback.CallbackQuery.From.ID.PeerID()
	logger.Info("[%s %s] trigger action [cancel] controller", userId, username)
	sess := middleware.SessionManager.Get(ctx)
	var order models.Order
	global.App.DB.Find(&order, "user_id = ? AND amount = ?", userId, sess.Amount)
	if order.Finished {
		return callback.Client.DeleteMessage(chatId, messageId).DoVoid(ctx)
	}
	order.Status = cst.OrderStatusCancel
	order.Finished = true
	order.Expired = true
	global.App.DB.Save(order)
	middleware.SessionManager.Reset(sess)
	return callback.Client.EditMessageText(chatId, messageId, "😔您取消了订单~").ParseMode(tg.HTML).DoVoid(ctx)
}

func RentReset(ctx context.Context, callback *tgb.CallbackQueryUpdate) error {
	return Render.reset(ctx, callback)
}

func RentChoose(ctx context.Context, callback *tgb.CallbackQueryUpdate) error {
	return Render.choose(ctx, callback)
}

func RentConfirm(ctx context.Context, callback *tgb.CallbackQueryUpdate) error {
	return Render.confirm(ctx, callback)
}

func RentPayment(ctx context.Context, callback *tgb.CallbackQueryUpdate) error {
	return Render.payment(ctx, callback)
}

func tickerTimeout(ctx context.Context, callback *tgb.CallbackQueryUpdate) {
	ticker := time.NewTicker(time.Minute * 10)
	defer ticker.Stop()

	chatId := callback.Message.Chat.ID
	messageId := callback.Message.ID
	username := strings.Replace(callback.From.Username.PeerID(), "@", "", 1)
	sess := middleware.SessionManager.Get(ctx)
	for {
		select {
		case <-ticker.C:
			var orders []models.Order
			global.App.DB.Find(&orders, "amount = ? AND to_address = ? AND finished = ? AND status = ?", sess.Amount, sess.RentAddress, false, false)
			if len(orders) == 0 {
				middleware.SessionManager.Reset(sess)
				return
			}
			middleware.SessionManager.Reset(sess)
			logger.Warn("[%s %s] rent energy order has been expired", chatId.PeerID(), username)
			global.App.DB.Delete(orders[0])
			_ = callback.Client.EditMessageText(chatId, messageId, "⏱️您的订单已过期~").ParseMode(tg.HTML).DoVoid(ctx)
			return
		}
	}
}
