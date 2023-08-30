package controllers

import (
	"context"
	"github.com/mr-linch/go-tg/tgb"
	"strings"
	"telegram-energy/pkg/core/logger"
	"telegram-energy/pkg/middleware"
	"telegram-energy/pkg/services/grid"
	"time"
)

type RentTmpl struct {
	RentAddress  string
	RentDuration string
	RentEnergy   int64
	MinEnergy    int64
	CastTRX      string
	CurrentTime  time.Time
}

type RentPaymentTmpl struct {
	Address   string
	Amount    string
	ExpiredAt time.Time
}

func Receiver(ctx context.Context, update *tgb.MessageUpdate) error {
	userId := update.Message.From.ID.PeerID()
	username := update.Message.From.Username.PeerID()
	username = strings.ToLower(strings.ReplaceAll(username, "@", ""))
	logger.Info("[%s %s] trigger action [receiver] controller", userId, username)
	sess := middleware.SessionManager.Get(ctx)
	resp, err := grid.ValidateAddress(update.Text)
	if err != nil || !resp.Result {
		logger.Error("[%s %s] input address err %v", err)
		return update.Answer("无效地址, 请输入TRC20地址").DoVoid(ctx)
	}
	sess.RentAddress = update.Text
	initRender()
	return Render.render(ctx, update)
}
