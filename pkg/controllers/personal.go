package controllers

import (
	"context"
	"errors"
	"github.com/mr-linch/go-tg"
	"github.com/mr-linch/go-tg/tgb"
	"go.uber.org/zap/buffer"
	"html/template"
	"os"
	"strings"
	"telegram-energy/pkg/core/cst"
	"telegram-energy/pkg/core/global"
	"telegram-energy/pkg/core/logger"
	"telegram-energy/pkg/models"
	"telegram-energy/pkg/services"
	"telegram-energy/pkg/utils"
)

type UserOrderTmpl struct {
	Username  string
	FirstName string
	Orders    []models.Order
}

func Personal(ctx context.Context, update *tgb.MessageUpdate) error {
	userId := update.From.ID.PeerID()
	username := update.From.Username.PeerID()
	username = strings.ReplaceAll(username, "@", "")
	logger.Info("[%s %s] trigger action [personal] controller", userId, username)
	var orders []models.Order
	global.App.DB.Find(&orders, "user_id = ? AND finished = ? AND expired = ?", userId, true, false).Order("-id").Limit(20)

	pf, err := os.ReadFile(cst.PersonalTemplateFile)
	if err != nil {
		logger.Error("[%s %s] read template file %s, failed %v", userId, username, cst.PersonalTemplateFile, err)
		return err
	}
	tmpl, err := template.New("personal").Funcs(template.FuncMap{"format": utils.DateTime}).Parse(string(pf))
	if err != nil {
		logger.Error("[%s %s] template parse file %s, failed %v", userId, username, cst.PersonalTemplateFile, err)
		return err
	}
	buf := new(buffer.Buffer)

	link := update.From.Username.Link()
	info, err := services.GetUserInfo(link)
	if err != nil {
		logger.Error("[%s %s] get telegram userinfo failed %v", userId, username, err)
		return err
	}
	if !info.Exist && info.FirstName == "" {
		logger.Error("[%s %s] not found telegram user", userId, username)
		return errors.New("not found telegram user")
	}

	tpl := UserOrderTmpl{
		Username:  username,
		FirstName: info.FirstName,
		Orders:    orders,
	}
	err = tmpl.Execute(buf, tpl)
	if err != nil {
		logger.Error("[%s %s] template execute file %s, failed %v", userId, username, cst.PersonalTemplateFile, err)
		return err
	}
	fileArg := tg.NewFileArgURL(info.LogUrl)
	if info.Exist {
		return update.AnswerPhoto(fileArg).Caption(buf.String()).ParseMode(tg.HTML).DoVoid(ctx)
	}
	return update.Answer(buf.String()).ParseMode(tg.HTML).DoVoid(ctx)
}
