package controllers

import (
	"context"
	"github.com/mr-linch/go-tg"
	"github.com/mr-linch/go-tg/tgb"
	"go.uber.org/zap/buffer"
	"html/template"
	"strings"
	"telegram-energy/pkg/core/cst"
	"telegram-energy/pkg/core/logger"
)

func Start(ctx context.Context, update *tgb.MessageUpdate) error {
	bot := NewBot()
	err := update.Client.SetMyCommands(bot.Cmd).DoVoid(ctx)
	userId := update.Message.From.ID.PeerID()
	username := update.Message.From.Username.PeerID()
	username = strings.ToLower(strings.ReplaceAll(username, "@", ""))
	tmpl, err := template.ParseFiles(cst.StartTemplateFile)
	if err != nil {
		logger.Error("[%s %s] template parse file %s, failed %v", userId, username, cst.StartTemplateFile, err)
		return err
	}
	buf := new(buffer.Buffer)
	err = tmpl.Execute(buf, username)
	if err != nil {
		logger.Error("[%s %s] template execute file %s, failed %v", userId, username, cst.StartTemplateFile, err)
		return err
	}
	inputFile, _ := tg.NewInputFileLocal(cst.StartTitlePNG)
	fileArg := tg.NewFileArgUpload(inputFile)
	return update.AnswerPhoto(fileArg).Caption(buf.String()).
		ParseMode(tg.HTML).
		ReplyMarkup(bot.ReplayMarkup).
		DoVoid(ctx)
}
