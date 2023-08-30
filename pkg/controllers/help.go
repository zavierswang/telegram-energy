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

func Help(ctx context.Context, update *tgb.MessageUpdate) error {
	userId := update.Message.From.ID.PeerID()
	username := update.Message.From.Username.PeerID()
	username = strings.ToLower(strings.ReplaceAll(username, "@", ""))
	logger.Info("[%s %s] trigger action [help] controller", userId, username)
	tmpl, err := template.ParseFiles(cst.HelpTemplateFile)
	if err != nil {
		logger.Error("[%s %s] template parse file %s, failed %v", userId, username, cst.HelpTemplateFile, err)
		return err
	}
	buf := new(buffer.Buffer)
	err = tmpl.Execute(buf, username)
	if err != nil {
		logger.Error("[%s %s] template execute file %s, failed %v", userId, username, cst.HelpTemplateFile, err)
		return err
	}
	inputFile, err := tg.NewInputFileLocal(cst.StartTitlePNG)
	if err != nil {
		logger.Error("[%s %s] load png %s, failed %v", userId, username, cst.StartTitlePNG, err)
		return err
	}
	fileArg := tg.NewFileArgUpload(inputFile)
	layout := tg.NewButtonLayout[tg.InlineKeyboardButton](2).Row()
	layout.Insert(
		tg.NewInlineKeyboardButtonURL("更多定制需求", "https://t.me/tg_llama"),
	)
	inlineKeyboard := tg.NewInlineKeyboardMarkup(layout.Keyboard()...)
	return update.AnswerPhoto(fileArg).Caption(buf.String()).
		ParseMode(tg.HTML).
		ReplyMarkup(inlineKeyboard).
		DoVoid(ctx)
}
