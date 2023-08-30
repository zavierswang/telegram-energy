package controllers

import (
	"context"
	"fmt"
	"github.com/alexeyco/simpletable"
	"github.com/mr-linch/go-tg"
	"github.com/mr-linch/go-tg/tgb"
	"go.uber.org/zap/buffer"
	"html/template"
	"strings"
	"telegram-energy/pkg/core/cst"
	"telegram-energy/pkg/core/global"
	"telegram-energy/pkg/core/logger"
	"telegram-energy/pkg/models"
)

func Orders(ctx context.Context, update *tgb.MessageUpdate) error {
	username := strings.ReplaceAll(update.From.Username.PeerID(), "@", "")
	userId := update.From.ID.PeerID()
	logger.Info("[%s %s] trigger action [orders] controller", userId, username)
	var orders []models.Order
	global.App.DB.Find(&orders, "user_id = ? AND finished = ? AND expired = ?", userId, true, false).Order("-id").Limit(20)
	table := simpletable.New()
	table.Header = &simpletable.Header{Cells: []*simpletable.Cell{
		{Align: simpletable.AlignCenter, Text: "金额"},
		{Align: simpletable.AlignCenter, Text: "地址"},
		{Align: simpletable.AlignCenter, Text: "时间"},
	}}
	for _, item := range orders {
		a := []rune(item.ToAddress)
		b := fmt.Sprintf("%s...%s", string(a[0:4]), string(a[30:]))
		row := []*simpletable.Cell{
			{Align: simpletable.AlignCenter, Text: fmt.Sprintf("%.3f", item.Payments)},
			{Align: simpletable.AlignCenter, Text: b},
			{Align: simpletable.AlignCenter, Text: item.CreatedAt.Format(cst.TimeFormatter)},
		}
		table.Body.Cells = append(table.Body.Cells, row)
	}
	if len(orders) == 0 {
		table.Body.Cells = append(table.Body.Cells, []*simpletable.Cell{
			{},
			{},
			{Align: simpletable.AlignCenter, Text: "暂无记录~"},
		})
	}
	table.SetStyle(simpletable.StyleCompactLite)

	tmpl, err := template.ParseFiles(cst.OrdersTemplateFile)
	if err != nil {
		logger.Error("[%s %s] template parse file %s, failed %v", userId, username, cst.OrdersTemplateFile, err)
		return err
	}

	//link := update.From.Username.Link()
	//info, err := services.GetUserInfo(link)
	//if err != nil {
	//	logger.Error("[%s %s] get telegram userinfo failed %v", userId, username, err)
	//	return err
	//}
	//if !info.Exist && info.FirstName == "" {
	//	logger.Error("[%s %s] not found telegram user", userId, username)
	//	return errors.New("not found telegram user")
	//}
	//fileArg := tg.NewFileArgURL(info.LogUrl)
	inputFile, _ := tg.NewInputFileLocal(cst.HelpTitlePNG)
	fileArg := tg.NewFileArgUpload(inputFile)
	buf := new(buffer.Buffer)
	err = tmpl.Execute(buf, nil)
	if err != nil {
		logger.Error("[%s %s] template execute file %s, failed %v", userId, username, cst.OrdersTemplateFile, err)
		return err
	}
	fmt.Println(table.String())
	tpl := fmt.Sprintf("%s\n\n<pre>%s</pre>", buf.String(), table.String())
	return update.AnswerPhoto(fileArg).Caption(tpl).ParseMode(tg.HTML).DoVoid(ctx)
}
