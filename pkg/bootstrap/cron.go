package bootstrap

import (
	"context"
	"fmt"
	"github.com/mr-linch/go-tg"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap/buffer"
	"html/template"
	"math"
	"os"
	"strconv"
	"telegram-energy/pkg/core/cst"
	"telegram-energy/pkg/core/global"
	"telegram-energy/pkg/core/logger"
	"telegram-energy/pkg/models"
	"telegram-energy/pkg/services"
	"telegram-energy/pkg/services/grid"
	"telegram-energy/pkg/services/tron"
	"telegram-energy/pkg/utils"
	"time"
)

func StartCron() {
	global.App.Cron = cron.New(cron.WithSeconds())
	go func() {
		listenTRX := &ListenTRX{
			ctx: context.Background(),
		}
		_, err := global.App.Cron.AddJob("*/30 * * * * *", listenTRX)
		if err != nil {
			logger.Error("cron job listenUSDT failed %v", err)
			panic(err)
		}
		global.App.Cron.Start()
		defer global.App.Cron.Stop()
		select {}
	}()
}

type Algo interface {
	exec(data tron.Trc10Data)
}

type Fee struct {
	listen *ListenTRX
	resp   Response
}

func (f *Fee) exec(data tron.Trc10Data) {
	quant := float64(data.ContractData.Amount) / math.Pow10(6)
	amount := fmt.Sprintf("%.3f", quant)
	normalStatus := []int{cst.OrderStatusRunning, cst.OrderStatusReceived, cst.OrderStatusApiSuccess}
	var orders []models.Order
	global.App.DB.Find(&orders, "payments = ? AND expired = ? AND finished = ? AND status in ?", amount, false, false, normalStatus)
	if len(orders) == 0 {
		logger.Warn("[scheduler] not found energy order with payments=%f", amount)
		return
	}

	//可能会有重复订单金额，过滤一次
	var order models.Order
	for _, o := range orders {
		if !o.Expired && !o.Finished && o.Payments == amount {
			order = o
		}
	}
	if order.ToAddress == "" {
		order.ToAddress = data.OwnerAddress
	}
	resp, err := services.FeeBuyEnergy(1, order.Blocks, order.Energy, order.ToAddress)
	if err != nil {
		logger.Error("[scheduler] services.FeeBuyEnergy failed %v", err)
		return
	}
	feeData := resp.Data
	logger.Info("[scheduler] fee response: %+v", feeData)
	order.FeesAddress = feeData.PayAddress
	global.App.DB.Save(&order)

	chatId, _ := strconv.ParseInt(order.ChatId, 10, 64)
	f.resp = Response{
		ToAddress:   data.OwnerAddress,
		Energy:      order.Energy,
		Duration:    order.Duration,
		Amount:      order.Payments,
		CreatedAt:   order.CreatedAt,
		CurrentTime: time.Now(),
	}
	logger.Info("[scheduler] ready transfer energy: %d => %s", order.Energy, data.OwnerAddress)
	txId, err := grid.TransferTRX(global.App.Config.Telegram.SendAddress, order.FeesAddress, order.Amount)
	if err != nil {
		logger.Error("[scheduler] grid.TransferTRX failed %v", err)
		global.App.DB.Model(&models.Order{}).
			Where("user_id = ? AND payments = ? AND expired = ? AND finished = ?", order.UserId, amount, false, false).
			Updates(map[string]interface{}{"status": cst.OrderStatusFailure, "finished": true})
		logger.Error("[scheduler] transfer energy: %d => %s, failed %v", order.Energy, order.ToAddress, err)
		f.resp.Status = "失败"
		f.notify(chatId)
		_ = global.App.Client.DeleteMessage(tg.ChatID(chatId), order.MessageId).DoVoid(f.listen.ctx)
		return
	}
	logger.Info("[scheduler] grid transfer TRX successfully, txId: %s", txId)
	global.App.DB.Model(&models.Order{}).
		Where("user_id = ? AND payments = ? AND expired = ? AND finished = ?", order.UserId, amount, false, false).
		Updates(map[string]interface{}{"status": cst.OrderStatusSuccess, "finished": true})
	f.resp.Status = "成功"
	_ = global.App.Client.DeleteMessage(tg.ChatID(chatId), order.MessageId).DoVoid(f.listen.ctx)
	f.notify(chatId)
	return
}

func (f *Fee) notify(userId int64) {
	layout := tg.NewButtonLayout[tg.InlineKeyboardButton](2).Row()
	layout.Insert(
		tg.NewInlineKeyboardButtonURL("TG会员充值", "https://t.me/TAKOPREMIUM_Bot"),
		tg.NewInlineKeyboardButtonURL("更多定制", "https://t.me/tg_llama"),
	)
	inlineKeyboard := tg.NewInlineKeyboardMarkup(layout.Keyboard()...)
	pf, err := os.ReadFile(cst.OrderResultTemplateFile)
	if err != nil {
		logger.Error("[scheduler] read file %s, failed %v", cst.OrderResultTemplateFile, err)
		return
	}
	tmpl, err := template.New("success").Funcs(template.FuncMap{"format": utils.DateTime}).Parse(string(pf))
	if err != nil {
		logger.Error("[scheduler] parse file %s, failed %v", cst.OrderResultTemplateFile, err)
		return
	}
	buf := new(buffer.Buffer)
	err = tmpl.Execute(buf, f.resp)
	if err != nil {
		logger.Error("[scheduler] template execute file %s, failed %v", cst.OrderResultTemplateFile, err)
		return
	}

	inputFile, err := tg.NewInputFileLocal(cst.HelpTitlePNG)
	if err != nil {
		logger.Error("[scheduler] load local png failed %v", err)
	}

	fileArg := tg.NewFileArgUpload(inputFile)
	//通知用户
	chatId := tg.ChatID(userId)
	err = global.App.Client.SendPhoto(chatId, fileArg).Caption(buf.String()).ParseMode(tg.HTML).ReplyMarkup(inlineKeyboard).DoVoid(f.listen.ctx)
	if err != nil {
		logger.Error("[scheduler] send message to user failed %v", err)
	}
	//群组通知
	group := global.App.Config.App.Group
	gid := tg.Username(group)
	logger.Info("[scheduler] ready to send message to group: %+v", gid)
	err = global.App.Client.SendMessage(gid, buf.String()).ParseMode(tg.HTML).ReplyMarkup(inlineKeyboard).DisableWebPagePreview(true).DoVoid(f.listen.ctx)
	if err != nil {
		logger.Error("[scheduler] send message to group failed: %v", err)
	}

}

type Goodies struct {
	listen *ListenTRX
}

func (g *Goodies) exec(data tron.Trc10Data) {
	logger.Info("Goodies create")
	//amount := float64(data.ContractData.Amount) / math.Pow10(6)
	//normalStatus := []int{cst.OrderStatusRunning, cst.OrderStatusReceived, cst.OrderStatusApiSuccess}
	//var orders []models.Order
	//global.App.DB.Find(&orders, "amount = ? AND expired = ? AND finished = ? AND status in ?", amount, false, false, normalStatus)
	//if len(orders) == 0 {
	//	logger.Warn("not matched order with amount=%f, finished=false, status=%+v", amount, false, normalStatus)
	//	return
	//}
	////可能会有重复订单金额，过滤一次
	//var order models.Order
	//for _, item := range orders {
	//	if !item.Expired && !item.Finished && item.Amount == amount {
	//		order = item
	//	}
	//}
	//chatId, _ := strconv.ParseInt(order.ChatId, 10, 64)
	//var users []models.User
	//global.App.DB.Find(&users, "is_admin = ? OR user_id = ?", true, order.UserId)
	//response := Response{
	//	ToAddress:   order.ToAddress,
	//	Energy:      order.Energy,
	//	Duration:    order.Duration,
	//	Amount:      order.Amount,
	//	CreatedAt:   order.CreatedAt,
	//	CurrentTime: time.Now(),
	//}
	//buf := new(buffer.Buffer)
	//logger.Info("[%d %s] ready to transfer energy: %d ==> %s", chatId, order.Username, order.Energy, order.ToAddress)
	//resp, err := services.CreateTokenGoodiesCommonOrder(order.ToAddress, amount)
	//if err != nil {
	//	global.App.DB.Model(&models.Order{}).
	//		Where("user_id = ? AND amount = ? AND expired = ? AND finished = ?", order.UserId, amount, false, false).
	//		Updates(map[string]interface{}{"status": cst.OrderStatusApiFailure, "finished": true})
	//	logger.Error("[%d %s] transfer energy: %d ==> %s, failed %v", chatId, order.Username, order.Energy, order.ToAddress, err)
	//	response.Status = "失败"
	//	_ = tmpl.Execute(buf, response)
	//
	//	_ = global.App.Client.DeleteMessage(tg.ChatID(chatId), order.MessageId).DoVoid(l.ctx)
	//	//noticeRestFullApi(buf.String(), users...)
	//	//noticeGroup(buf.String(), client)
	//	return
	//}
	//if !resp.Success {
	//	global.App.DB.Model(&models.Order{}).
	//		Where("user_id = ? AND amount = ? AND expired = ? AND finished = ?", order.UserId, amount, false, false).
	//		Updates(map[string]interface{}{"status": cst.OrderStatusFailure, "finished": true})
	//	logger.Error("[%d %s] transfer energy: %d ==> %s, failed %v", chatId, order.Username, order.Energy, order.ToAddress, resp)
	//	response.Status = "失败"
	//	_ = tmpl.Execute(buf, response)
	//	_ = global.App.Client.DeleteMessage(tg.ChatID(chatId), order.MessageId).DoVoid(context.Background())
	//	//noticeRestFullApi(buf.String(), users...)
	//	//noticeGroup(buf.String(), client)
	//	return
	//}
	//global.App.DB.Model(&models.Order{}).
	//	Where("user_id = ? AND amount = ? AND expired = ? AND finished = ?", order.UserId, amount, false, false).
	//	Updates(map[string]interface{}{"finished": true, "status": cst.OrderStatusSuccess})
	//logger.Info("[%d %s] transfer energy: %d ==> %s, successfully", chatId, order.Username, order.Energy, order.ToAddress)
	//
	////通知support
	//response.Status = "成功"
	//_ = tmpl.Execute(buf, response)
	//_ = global.App.Client.DeleteMessage(tg.ChatID(chatId), order.MessageId).DoVoid(context.Background())
}

func (g *Goodies) notify(userId int64) {

}

type ListenTRX struct {
	ticker     int
	trc10Queue []string
	ctx        context.Context
	algo       Algo
}

func (l *ListenTRX) Run() {
	l.ticker++
	switch global.App.Config.Telegram.EnableApi {
	case "goodies":
		l.algo = &Goodies{listen: l}
	case "fee":
		l.algo = &Fee{listen: l}
	}
	now := time.Now()
	params := map[string]string{
		"limit":           "30",
		"start":           "0",
		"address":         global.App.Config.Telegram.ReceiveAddress,
		"type":            "trc10",
		"start_timestamp": strconv.FormatInt(now.Add(-60*time.Second).UnixMilli(), 10),
		"end_timestamp":   strconv.FormatInt(now.UnixMilli(), 10),
	}
	transfers, err := tron.TRC10Transfer(params, true, true)
	if err != nil {
		logger.Error("trc10 transfer list failed %v", err)
		return
	}
	// 首次启动并获取到的数据暂存到队列中
	if l.ticker == 1 {
		for _, transfer := range transfers {
			l.trc10Queue = append(l.trc10Queue, transfer.Hash)
		}
		//logger.Info("[scheduler] listenTRX %d times, latest hashes: %+v", l.ticker, l.trc10Queue)
		return
	}
	slice1 := l.trc10Queue
	var slice2 []string
	for _, transfer := range transfers {
		slice2 = append(slice2, transfer.Hash)
	}
	// 比对历史数据获取最新的交易号
	hashes, _ := utils.Comp(slice1, slice2)
	//logger.Info("[scheduler] listenTRX %d times, latest hashes: %+v", l.ticker, hashes)
	for _, hash := range hashes {
		for _, transfer := range transfers {
			if transfer.Hash == hash {
				l.exec(transfer)
			}
		}
	}
	//l.create(tron.Trc10Data{})
	// 历史数据过多，删除部分数据
	if len(l.trc10Queue) >= 500 {
		logger.Info("[scheduler] clean remain queue ...")
		l.trc10Queue = l.trc10Queue[400:]
	}
	// 合并历史交易号，获新的数据
	l.trc10Queue = utils.Union(l.trc10Queue, slice2)
	return
}

func (l *ListenTRX) exec(transfer tron.Trc10Data) {
	l.algo.exec(transfer)
}

type Response struct {
	ToAddress   string
	Duration    string
	Energy      int64
	Amount      string
	Status      string
	CreatedAt   time.Time
	CurrentTime time.Time
}
