package controllers

import (
	"context"
	"github.com/mr-linch/go-tg"
	"go.uber.org/zap/buffer"
	"html/template"
	"strconv"
	"sync"
	"telegram-energy/pkg/core/cst"
	"telegram-energy/pkg/core/global"
	"telegram-energy/pkg/core/logger"
	"telegram-energy/pkg/models"
)

func Update(token string) {
	var users []models.User
	global.App.DB.Find(&users)
	tmpl, _ := template.ParseFiles(cst.UpdateTemplateFile)
	buf := new(buffer.Buffer)
	_ = tmpl.Execute(buf, global.App.Config.App.Support)
	var wg sync.WaitGroup
	for _, user := range users {
		wg.Add(1)
		go func(user models.User) {
			defer wg.Done()
			chatId, _ := strconv.ParseInt(user.UserID, 10, 64)
			err := global.App.Client.SendMessage(tg.ChatID(chatId), buf.String()).
				ParseMode(tg.HTML).
				DisableWebPagePreview(true).
				DoVoid(context.Background())
			if err != nil {
				logger.Error("[update] send message to user [%s %s] failed %v", user.UserID, user.Username, err)
			}
		}(user)
	}
	wg.Wait()
	//gid := tg.Username(global.App.Config.App.Group)
	//err := global.App.Client.SendMessage(gid, buf.String()).ParseMode(tg.HTML).DoVoid(context.Background())
	//if err != nil {
	//	logger.Error("[update] send message to group failed %v", err)
	//}
}
