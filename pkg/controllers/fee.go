package controllers

import (
	"context"
	"fmt"
	"github.com/mr-linch/go-tg"
	"github.com/mr-linch/go-tg/tgb"
	"go.uber.org/zap/buffer"
	"html/template"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
	"telegram-energy/pkg/core/cst"
	"telegram-energy/pkg/core/global"
	"telegram-energy/pkg/core/logger"
	"telegram-energy/pkg/middleware"
	"telegram-energy/pkg/models"
	"telegram-energy/pkg/utils"
	"time"
)

type FeeRender struct{}

func (r *FeeRender) render(ctx context.Context, update *tgb.MessageUpdate) error {
	userId := update.Message.From.ID.PeerID()
	username := strings.ReplaceAll(update.Message.From.Username.PeerID(), "@", "")
	logger.Info("[%s %s] trigger action [render] update", userId, username)
	sess := middleware.SessionManager.Get(ctx)
	//address := sess.RentAddress
	//logger.Info("[%s %s] rent address %s", userId, username, address)
	sess.RentEnergy = 32000
	sess.RentBlocks = 3600
	rentSun := sess.RentEnergy
	if sess.RentBlocks == 3600*24*3 {
		rentSun = sess.RentEnergy * 3
	}
	var sun int64 = 100
	if rentSun > 100000 {
		sun = 60
	}
	total := rentSun * sun
	sess.Amount = float64(total) / math.Pow10(6)
	point := utils.RandPoint()
	logger.Info("[%s %s] energy amount: %f and point: %.3f", userId, username, sess.Amount, point)
	sess.Payments = fmt.Sprintf("%.3f", sess.Amount*(1.0+global.App.Config.Telegram.Ratio)+point)
	logger.Info("[%s %s] total of payment: %s", userId, username, sess.Payments)
	duration := utils.DurationSec(sess.RentBlocks)
	inputFile, _ := tg.NewInputFileLocal(cst.HelpTitlePNG)
	fileArg := tg.NewFileArgUpload(inputFile)
	tpl := RentTmpl{
		//RentAddress:  sess.RentAddress,
		RentDuration: duration,
		RentEnergy:   sess.RentEnergy,
		MinEnergy:    32000,
		CastTRX:      sess.Payments,
		CurrentTime:  time.Now(),
	}
	pf, err := os.ReadFile(cst.FeeTemplateFile)
	if err != nil {
		logger.Error("read template file %s, failed %v", cst.FeeTemplateFile, err)
		return err
	}
	tmpl, err := template.New("rent_receiver").
		Funcs(template.FuncMap{
			"format": utils.DateTime,
			"count":  utils.EnergyCount,
		}).Parse(string(pf))
	if err != nil {
		logger.Error("[%s %s] template parse file %s, failed %v", userId, username, cst.FeeTemplateFile, err)
		return err
	}
	buf := new(buffer.Buffer)
	err = tmpl.Execute(buf, tpl)
	if err != nil {
		logger.Error("[%s %s] template execute file %s, failed %v", userId, username, cst.GoodiesTemplateFile, err)
		return err
	}
	return update.AnswerPhoto(fileArg).
		Caption(buf.String()).
		ParseMode(tg.HTML).
		ReplyMarkup(feeInlineKeyboard(sess)).
		DoVoid(ctx)
}

func (r *FeeRender) reset(ctx context.Context, callback *tgb.CallbackQueryUpdate) error {
	userId := callback.Message.From.ID.PeerID()
	chatId := callback.Message.Chat.ID
	messageId := callback.Message.ID
	username := strings.ReplaceAll(callback.Message.From.Username.PeerID(), "@", "")
	logger.Info("[%s %s] trigger action [reset] update", userId, username)
	sess := middleware.SessionManager.Get(ctx)
	//address := sess.RentAddress
	//logger.Info("[%s %s] rent address %s", userId, username, address)
	sess.RentEnergy = 32000
	sess.RentBlocks = 3600
	rentSun := sess.RentEnergy
	if sess.RentBlocks == 3600*24*3 {
		rentSun = sess.RentEnergy * 3
	}
	var sun int64 = 100
	if rentSun > 100000 {
		sun = 60
	}
	total := rentSun * sun
	sess.Amount = float64(total) / math.Pow10(6)
	point := utils.RandPoint()
	logger.Info("[%s %s] energy amount: %f and point: %.3f", userId, username, sess.Amount, point)
	sess.Payments = fmt.Sprintf("%.3f", sess.Amount*(1.0+global.App.Config.Telegram.Ratio)+point)
	logger.Info("[%s %s] total of payment: %f", userId, username, sess.Payments)
	duration := utils.DurationSec(sess.RentBlocks)
	tpl := RentTmpl{
		//RentAddress:  sess.RentAddress,
		RentDuration: duration,
		RentEnergy:   sess.RentEnergy,
		MinEnergy:    32000,
		CastTRX:      sess.Payments,
		CurrentTime:  time.Now(),
	}
	pf, err := os.ReadFile(cst.FeeTemplateFile)
	if err != nil {
		logger.Error("read template file %s, failed %v", cst.FeeTemplateFile, err)
		return err
	}
	tmpl, err := template.New("rent_receiver").
		Funcs(template.FuncMap{
			"format": utils.DateTime,
			"count":  utils.EnergyCount,
		}).Parse(string(pf))
	if err != nil {
		logger.Error("[%s %s] template parse file %s, failed %v", userId, username, cst.FeeTemplateFile, err)
		return err
	}
	buf := new(buffer.Buffer)
	err = tmpl.Execute(buf, tpl)
	if err != nil {
		logger.Error("[%s %s] template execute file %s, failed %v", userId, username, cst.GoodiesTemplateFile, err)
		return err
	}
	return callback.Client.EditMessageText(chatId, messageId, buf.String()).
		ParseMode(tg.HTML).
		ReplyMarkup(feeInlineKeyboard(sess)).
		DisableWebPagePreview(true).
		DoVoid(ctx)
}

func (r *FeeRender) choose(ctx context.Context, callback *tgb.CallbackQueryUpdate) error {
	chatId := callback.Message.Chat.ID
	userId := callback.Message.From.ID.PeerID()
	username := strings.Replace(callback.From.Username.PeerID(), "@", "", 1)
	messageId := callback.Message.ID
	logger.Info("[%s %s] trigger action [choose] controller", userId, username)
	sess := middleware.SessionManager.Get(ctx)
	compile, err := regexp.Compile(`^rent\s+(?P<type>[a-z]+)\s+(?P<action>[a-z]+)?\s?(?P<number>(?:0|[1-9]\d*)(?:\.\d*)?)`)
	if err != nil {
		logger.Error("[%s %s] compile RentChoose failed %v", userId, username, err)
		return err
	}
	groups := utils.FindGroups(compile, callback.Data)
	number, _ := strconv.ParseInt(groups["number"], 10, 64)
	action := groups["action"]
	sess.RentChoose = callback.Data
	logger.Info("[%s %s] groups: %+v", userId, username, groups)
	switch groups["type"] {
	case "count":
		if action == "plus" {
			sess.RentEnergy += number * 32000
		} else if action == "minus" {
			sess.RentEnergy -= number * 32000
			if sess.RentEnergy <= 32000 {
				sess.RentEnergy = 32000
			}
		}
	case "energy":
		if action == "plus" {
			sess.RentEnergy += number
		} else if action == "minus" {
			sess.RentEnergy -= number
			if sess.RentEnergy <= 32000 {
				sess.RentEnergy = 32000
			}
		}
	case "days":
		sess.RentBlocks = number
	}
	logger.Info("[%s %s] sess: %+v", userId, username, sess)
	rentEnergy := sess.RentEnergy
	if sess.RentBlocks == 3600*24*3 {
		rentEnergy = sess.RentEnergy * 3
	}
	var sun int64 = 100
	if rentEnergy > 100000 {
		sun = 60
	}
	total := rentEnergy * sun
	sess.Amount = float64(total) / math.Pow10(6)
	point := utils.RandPoint()
	logger.Info("[%s %s] energy amount: %f and point: %.3f", userId, username, sess.Amount, point)
	sess.Payments = fmt.Sprintf("%.3f", sess.Amount*(1.0+global.App.Config.Telegram.Ratio)+point)
	logger.Info("[%s %s] total payment: %f", userId, username, sess.Payments)
	tpl := RentTmpl{
		//RentAddress:  sess.RentAddress,
		RentDuration: utils.DurationSec(sess.RentBlocks),
		RentEnergy:   sess.RentEnergy,
		MinEnergy:    32000,
		CastTRX:      sess.Payments,
		CurrentTime:  time.Now(),
	}
	pf, err := os.ReadFile(cst.FeeTemplateFile)
	if err != nil {
		logger.Error("read template file %s, failed %v", cst.FeeTemplateFile, err)
		return err
	}
	tmpl, err := template.New("rent_receiver").
		Funcs(template.FuncMap{
			"format": utils.DateTime,
			"count":  utils.EnergyCount,
		}).Parse(string(pf))
	if err != nil {
		logger.Error("[%s %s] template parse file %s, failed %v", userId, username, cst.FeeTemplateFile, err)
		return err
	}
	buf := new(buffer.Buffer)
	err = tmpl.Execute(buf, tpl)
	if err != nil {
		logger.Error("[%s %s] template execute file %s, failed %v", userId, username, cst.GoodiesTemplateFile, err)
		return err
	}
	return callback.Client.EditMessageCaption(chatId, messageId, buf.String()).
		ParseMode(tg.HTML).
		ReplyMarkup(feeInlineKeyboard(sess)).
		DoVoid(ctx)
}

func (r *FeeRender) confirm(ctx context.Context, callback *tgb.CallbackQueryUpdate) error {
	chatId := callback.Message.Chat.ID
	userId := callback.Message.From.ID.PeerID()
	username := strings.Replace(callback.From.Username.PeerID(), "@", "", 1)
	messageId := callback.Message.ID
	logger.Info("[%s %s] trigger action [confirm] controller", userId, username)
	sess := middleware.SessionManager.Get(ctx)
	duration := utils.DurationSec(sess.RentBlocks)
	logger.Info("[%s %s] finally rent duration %s amount: %.3f TRX, payments: %.3f TRX", userId, username, duration, sess.Amount, sess.Payments)
	tpl := FeeConfirmTmpl{
		//RentAddress:  sess.RentAddress,
		RentDuration: duration,
		RentEnergy:   sess.RentEnergy,
		CastTRX:      sess.Payments,
		CurrentTime:  time.Now(),
		ExpiredTime:  time.Unix(sess.ExpiredTime, 0).Format(cst.DateTimeFormatter),
	}
	pf, err := os.ReadFile(cst.FeeConfirmTemplateFile)
	if err != nil {
		logger.Error("[%s %s] read template file %s, failed %v", userId, username, cst.FeeConfirmTemplateFile, err)
		return err
	}
	tmpl, err := template.New("rent_confirm").Funcs(template.FuncMap{"format": utils.DateTime, "count": utils.EnergyCount}).Parse(string(pf))
	if err != nil {
		logger.Error("[%s %s] template parse file %s, failed %v", userId, username, cst.FeeConfirmTemplateFile, err)
		return err
	}
	buf := new(buffer.Buffer)
	err = tmpl.Execute(buf, tpl)
	if err != nil {
		logger.Error("[%s %s] template execute file %s, failed %v", userId, username, cst.FeeConfirmTemplateFile, err)
		return err
	}
	layout := tg.NewButtonLayout[tg.InlineKeyboardButton](2).Row()
	layout.Insert(
		tg.NewInlineKeyboardButtonCallback("取消订单", Callback.Close),
		tg.NewInlineKeyboardButtonCallback("确认支付", Callback.RentPayment),
	)
	inlineKeyboard := tg.NewInlineKeyboardMarkup(layout.Keyboard()...)
	return callback.Client.EditMessageCaption(chatId, messageId, buf.String()).ParseMode(tg.HTML).ReplyMarkup(inlineKeyboard).DoVoid(ctx)
}

func (r *FeeRender) payment(ctx context.Context, callback *tgb.CallbackQueryUpdate) error {
	chatId := callback.Message.Chat.ID
	messageId := callback.Message.ID
	userId := callback.CallbackQuery.From.ID.PeerID()
	username := strings.Replace(callback.From.Username.PeerID(), "@", "", 1)
	logger.Info("[%s %s] trigger action [payment] callback", userId, username)
	sess := middleware.SessionManager.Get(ctx)
	logger.Info("[%s %s] rent %d energy, will be pay %.3 TRX", userId, username, sess.RentEnergy, sess.Amount)
	//resp, err := services.FeeBuyEnergy(1, sess.RentBlocks, sess.RentEnergy, "")
	//if err != nil {
	//	logger.Error("[%s %s] services.FeeBuyEnergy failed %v", userId, username, err)
	//	return err
	//}
	//data := resp.Data
	//sess.PayAddress = data.PayAddress
	//sess.ExpiredTime = data.ExpTime
	//logger.Info("[%s %s] fee response: %+v", userId, username, data)
	order := models.Order{
		UserId:      userId,
		Username:    strings.ToLower(username),
		CreatedAt:   time.Now(),
		Energy:      sess.RentEnergy,
		Amount:      sess.Amount,
		Payments:    sess.Payments,
		Blocks:      sess.RentBlocks,
		Duration:    utils.DurationSec(sess.RentBlocks),
		FeesAddress: sess.PayAddress,
		ChatId:      chatId.PeerID(),
		MessageId:   messageId,
		Status:      cst.OrderStatusRunning,
	}
	logger.Info("[%s %s] rent order details: %+v", userId, username, order)
	global.App.DB.Save(&order)
	buf := new(buffer.Buffer)
	tpl := RentPaymentTmpl{
		Address:   global.App.Config.Telegram.ReceiveAddress,
		Amount:    sess.Payments,
		ExpiredAt: time.Unix(sess.ExpiredTime, 0),
	}
	pf, err := os.ReadFile(cst.RentPaymentTemaplteFile)
	if err != nil {
		logger.Error("[%s %s] read template file %s, failed %v", userId, username, cst.RentPaymentTemaplteFile, err)
		return err
	}
	tmpl, err := template.New("rent_payment").Funcs(template.FuncMap{"format": utils.DateTime}).Parse(string(pf))
	if err != nil {
		logger.Error("[%s %s] template parse file %s, failed %v", userId, username, cst.RentPaymentTemaplteFile, err)
		return err
	}
	err = tmpl.Execute(buf, tpl)
	if err != nil {
		logger.Error("[%s %s] template execute file %s, failed %v", userId, username, cst.RentPaymentTemaplteFile, err)
		return err
	}
	layout := tg.NewButtonLayout[tg.InlineKeyboardButton](2).Row()
	layout.Insert(
		tg.NewInlineKeyboardButtonCallback("关闭/取消", Callback.Cancel),
		tg.NewInlineKeyboardButtonURL("联系客服", fmt.Sprintf("https://t.me/%s", global.App.Config.App.Support)),
	)
	inlineKeyboard := tg.NewInlineKeyboardMarkup(layout.Keyboard()...)
	go tickerTimeout(ctx, callback)
	return callback.Client.EditMessageCaption(chatId, messageId, buf.String()).ParseMode(tg.HTML).ReplyMarkup(inlineKeyboard).DoVoid(ctx)
}

func feeInlineKeyboard(sess *middleware.Session) tg.InlineKeyboardMarkup {
	infoLayout := tg.NewButtonLayout[tg.InlineKeyboardButton](1).Row()
	infoLayout.Insert(
		tg.NewInlineKeyboardButtonURL("能量介绍", "https://t.me/TAKO_HOME"),
	)
	layout := tg.NewButtonLayout[tg.InlineKeyboardButton](3).Row()
	layout.Insert(
		tg.NewInlineKeyboardButtonCallback("♻️重置选择", Callback.RentReset),
		tg.NewInlineKeyboardButtonCallback("立即租用", Callback.RentConfirm),
	)
	inlineKeyboard := tg.NewInlineKeyboardMarkup(layout.Keyboard()...)

	layoutDuration := tg.NewButtonLayout[tg.InlineKeyboardButton](3).Row()
	durations := []int64{3600, 3600 * 24 * 3}
	for _, duration := range durations {
		df := utils.DurationSec(duration)
		var isChecked bool
		if sess.RentBlocks == duration {
			isChecked = true
		}
		if isChecked {
			df = fmt.Sprintf("✅%s", df)
		}
		layoutDuration.Insert(tg.NewInlineKeyboardButtonCallback(df, fmt.Sprintf("rent days %d", duration)))
	}
	inlineKeyboardDuration := tg.NewInlineKeyboardMarkup(layoutDuration.Keyboard()...)

	//layoutEmptyComputeCount := tg.NewButtonLayout[tg.InlineKeyboardButton](1).Row()
	//layoutEmptyComputeCount.Insert(tg.NewInlineKeyboardButtonCallback("👇调整可兑换次数捷区👇", Callback.Empty))
	//inlineKeyboardEmptyComputeCount := tg.NewInlineKeyboardMarkup(layoutEmptyComputeCount.Keyboard()...)

	layoutComputeCount := tg.NewButtonLayout[tg.InlineKeyboardButton](3).Row()
	layoutComputeCount.Insert(
		tg.NewInlineKeyboardButtonCallback("➕1次", "rent count plus 1"),
		tg.NewInlineKeyboardButtonCallback("➕3次", "rent count plus 3"),
		tg.NewInlineKeyboardButtonCallback("➕5次", "rent count plus 5"),
		tg.NewInlineKeyboardButtonCallback("➕10次", "rent count plus 10"),
		tg.NewInlineKeyboardButtonCallback("➕30次", "rent count plus 30"),
		tg.NewInlineKeyboardButtonCallback("➖1次", "rent count minus 1"),
		tg.NewInlineKeyboardButtonCallback("➖3次", "rent count minus 3"),
		tg.NewInlineKeyboardButtonCallback("➖5次", "rent count minus 5"),
		tg.NewInlineKeyboardButtonCallback("➖10次", "rent count minus 10"),
	)
	inlineKeyboardComputeCount := tg.NewInlineKeyboardMarkup(layoutComputeCount.Keyboard()...)

	//layoutEmptyComputeEnergy := tg.NewButtonLayout[tg.InlineKeyboardButton](1).Row()
	//layoutEmptyComputeEnergy.Insert(tg.NewInlineKeyboardButtonCallback("👇调整能量数快捷区👇", Callback.Empty))
	//inlineKeyboardEmptyComputeEnergy := tg.NewInlineKeyboardMarkup(layoutEmptyComputeEnergy.Keyboard()...)
	//
	//layoutComputeEnergy := tg.NewButtonLayout[tg.InlineKeyboardButton](4).Row()
	//layoutComputeEnergy.Insert(
	//	tg.NewInlineKeyboardButtonCallback("➕1万", "rent energy plus 10000"),
	//	tg.NewInlineKeyboardButtonCallback("➕5万", "rent energy plus 50000"),
	//	tg.NewInlineKeyboardButtonCallback("➕10万", "rent energy plus 100000"),
	//	tg.NewInlineKeyboardButtonCallback("➕100万", "rent energy plus 1000000"),
	//	tg.NewInlineKeyboardButtonCallback("➖1万", "rent energy minus 10000"),
	//	tg.NewInlineKeyboardButtonCallback("➖5万", "rent energy minus 50000"),
	//	tg.NewInlineKeyboardButtonCallback("➖10万", "rent energy minus 100000"),
	//	tg.NewInlineKeyboardButtonCallback("➖100万", "rent energy minus 1000000"),
	//)
	//inlineKeyboardComputeEnergy := tg.NewInlineKeyboardMarkup(layoutComputeEnergy.Keyboard()...)

	layoutBase := tg.NewButtonLayout[tg.InlineKeyboardButton](2).Row()
	layoutBase.Insert(
		tg.NewInlineKeyboardButtonCallback("关闭", Callback.Close),
		tg.NewInlineKeyboardButtonURL("更多定制", "https://t.me/tg_llama"),
	)
	inlineKeyboardBase := tg.NewInlineKeyboardMarkup(layoutBase.Keyboard()...)

	inlineKeyboard.InlineKeyboard = append(inlineKeyboard.InlineKeyboard, inlineKeyboardDuration.InlineKeyboard...)
	//inlineKeyboard.InlineKeyboard = append(inlineKeyboard.InlineKeyboard, inlineKeyboardEmptyComputeCount.InlineKeyboard...)
	inlineKeyboard.InlineKeyboard = append(inlineKeyboard.InlineKeyboard, inlineKeyboardComputeCount.InlineKeyboard...)
	//inlineKeyboard.InlineKeyboard = append(inlineKeyboard.InlineKeyboard, inlineKeyboardEmptyComputeEnergy.InlineKeyboard...)
	//inlineKeyboard.InlineKeyboard = append(inlineKeyboard.InlineKeyboard, inlineKeyboardComputeEnergy.InlineKeyboard...)
	inlineKeyboard.InlineKeyboard = append(inlineKeyboard.InlineKeyboard, inlineKeyboardBase.InlineKeyboard...)
	return inlineKeyboard
}

type FeeConfirmTmpl struct {
	//RentAddress  string
	RentDuration string
	RentEnergy   int64
	CastTRX      string
	CurrentTime  time.Time
	ExpiredTime  string
}
