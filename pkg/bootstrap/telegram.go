package bootstrap

import (
	"context"
	"github.com/mr-linch/go-tg"
	"github.com/mr-linch/go-tg/tgb"
	"os"
	"os/signal"
	"syscall"
	"telegram-energy/pkg/controllers"
	"telegram-energy/pkg/core/cst"
	"telegram-energy/pkg/core/global"
	"telegram-energy/pkg/core/logger"
	"telegram-energy/pkg/routes"
)

func Telegram() error {
	opts := []tg.ClientOption{tg.WithClientServerURL(cst.TelegramApi)}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill, syscall.SIGTERM)
	defer cancel()
	token := global.App.Config.Telegram.Token
	if global.App.Config.App.Env == "release" || global.App.Config.App.Env == "prod" {
		token = cst.TelegramToken
	}
	global.App.Client = tg.New(token, opts...)
	me, err := global.App.Client.Me(ctx)
	if err != nil {
		logger.Error("authorized failed %v", err)
		return err
	}
	logger.Info("authorized %s successfully.", me.Username.Link())
	//telegram认证成功，启动cron任务
	StartCron()
	controllers.Update(token)
	r := tgb.NewRouter()
	routes.Telegram(r)
	return tgb.NewPoller(r, global.App.Client).Run(ctx)
}
