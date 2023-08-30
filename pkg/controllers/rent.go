package controllers

import (
	"context"
	"github.com/mr-linch/go-tg/tgb"
	"telegram-energy/pkg/core/global"
	"telegram-energy/pkg/middleware"
)

var Render RenderAlgo

type RenderAlgo interface {
	render(ctx context.Context, update *tgb.MessageUpdate) error
	reset(ctx context.Context, callback *tgb.CallbackQueryUpdate) error
	choose(ctx context.Context, callback *tgb.CallbackQueryUpdate) error
	confirm(ctx context.Context, callback *tgb.CallbackQueryUpdate) error
	payment(ctx context.Context, callback *tgb.CallbackQueryUpdate) error
}

type RentRender struct {
	algo  RenderAlgo
	ratio float64
}

func (r *RentRender) setRender(algo RenderAlgo) {
	r.algo = algo
	r.ratio = global.App.Config.Telegram.Ratio
}

func (r *RentRender) render(ctx context.Context, update *tgb.MessageUpdate) error {
	return r.algo.render(ctx, update)
}

func (r *RentRender) reset(ctx context.Context, callback *tgb.CallbackQueryUpdate) error {
	return r.algo.reset(ctx, callback)
}

func (r *RentRender) choose(ctx context.Context, callback *tgb.CallbackQueryUpdate) error {
	return r.algo.choose(ctx, callback)
}

func (r *RentRender) confirm(ctx context.Context, callback *tgb.CallbackQueryUpdate) error {
	return r.algo.confirm(ctx, callback)
}

func (r *RentRender) payment(ctx context.Context, callback *tgb.CallbackQueryUpdate) error {
	return r.algo.payment(ctx, callback)
}

func initRender() {
	var algo RenderAlgo
	algo = &GoodiesRender{}
	if global.App.Config.Telegram.EnableApi == "fee" {
		algo = &FeeRender{}
	}
	Render = &RentRender{
		algo:  algo,
		ratio: global.App.Config.Telegram.Ratio,
	}
}

func Rent(ctx context.Context, update *tgb.MessageUpdate) error {
	//userId := update.Message.From.ID.PeerID()
	//username := update.Message.From.Username.PeerID()
	//username = strings.ToLower(strings.ReplaceAll(username, "@", ""))
	//logger.Info("[%s %s] trigger action [rent] controller", userId, username)
	//sess := middleware.SessionManager.Get(ctx)
	//sess.Step = middleware.SessionRentEnergy
	//tmpl, err := template.ParseFiles(cst.RentTemplateFile)
	//if err != nil {
	//	logger.Error("[%s %s] template parse file %s, failed %v", userId, username, cst.RentTemplateFile, err)
	//	return err
	//}
	//buf := new(buffer.Buffer)
	//err = tmpl.Execute(buf, username)
	//if err != nil {
	//	logger.Error("[%s %s] template execute file %s, failed %v", userId, username, cst.RentTemplateFile, err)
	//	return err
	//}
	//inputFile, _ := tg.NewInputFileLocal(cst.GroupTitlePNG)
	//fileArg := tg.NewFileArgUpload(inputFile)
	//return update.AnswerPhoto(fileArg).Caption(buf.String()).
	//	ParseMode(tg.HTML).
	//	DoVoid(ctx)
	sess := middleware.SessionManager.Get(ctx)
	sess.Step = middleware.SessionRentEnergy
	initRender()
	return Render.render(ctx, update)
}
