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
	"telegram-energy/pkg/services"
	"telegram-energy/pkg/utils"
	"time"
)

type GoodiesRender struct{}

func (r *GoodiesRender) render(ctx context.Context, update *tgb.MessageUpdate) error {
	chatId := update.Chat.ID
	userId := update.Message.From.ID.PeerID()
	username := strings.ReplaceAll(update.From.Username.PeerID(), "@", "")
	pf, err := os.ReadFile(cst.GoodiesTemplateFile)
	if err != nil {
		logger.Error("read template file %s, failed %v", cst.GoodiesTemplateFile, err)
		return err
	}
	tmpl, err := template.New("rent_receiver").
		Funcs(template.FuncMap{
			"format": utils.DateTime,
			"count":  utils.EnergyCount,
		}).Parse(string(pf))
	if err != nil {
		logger.Error("[%s %s] template parse file %s, failed %v", userId, username, cst.GoodiesTemplateFile, err)
		return err
	}
	sess := middleware.SessionManager.Get(ctx)
	sess.RentAddress = update.Message.Text
	var minEnergy int64
	goodies, err := services.GetGoodiesBaseInfo()
	if err != nil {
		logger.Error("[%s %s] service.GetMinEnergyOrder() failed %v", chatId, username, err)
		return update.Answer("服务异常，请联系管理~").DoVoid(ctx)
	}
	logger.Info("[%s %s] token_goodies api response: %+v", userId, username, goodies)
	defaultDuration := goodies.RentalDurations[0]
	min := float64(goodies.MinOrderAmountInSun) / defaultDuration.Multiplier / float64(goodies.MinEnergyPriceInSun)
	minEnergy, _ = strconv.ParseInt(fmt.Sprintf("%.f", min), 10, 64)
	duration := utils.Duration(defaultDuration.Blocks)
	sess.RentBlocks = defaultDuration.Blocks
	sess.RentMultiplier = defaultDuration.Multiplier
	sess.FeesAddress = goodies.OrderFeesAddress
	sess.RentEnergy = 64000
	sess.MinEnergyPriceInSun = goodies.MinEnergyPriceInSun
	sess.MinOrderAmountInSun = goodies.MinOrderAmountInSun
	logger.Info("[%s %s] default goodies duration: %+v", chatId.PeerID(), username, defaultDuration)
	payments := float64(sess.RentEnergy) * float64(sess.MinEnergyPriceInSun) * sess.RentMultiplier
	payments = payments / math.Pow10(6)
	amount := payments * (1.0 + global.App.Config.Telegram.Ratio)
	amount, _ = strconv.ParseFloat(fmt.Sprintf("%.3f", amount), 64)
	point := utils.RandPoint()
	amount += point
	logger.Info("[%s %s] default rent duration %s and payment amount: %.3f", chatId.PeerID(), username, duration, amount)
	sess.Payments = fmt.Sprintf("%.3f", payments)
	sess.Amount, _ = strconv.ParseFloat(fmt.Sprintf("%.3f", amount), 64)

	logger.Info("[%s %s] default cast TRX: %.3f", userId, username, amount)
	tpl := RentTmpl{
		RentAddress:  sess.RentAddress,
		RentDuration: duration,
		RentEnergy:   sess.RentEnergy,
		MinEnergy:    minEnergy,
		CastTRX:      sess.Payments,
		CurrentTime:  time.Now(),
	}
	buf := new(buffer.Buffer)
	err = tmpl.Execute(buf, tpl)
	if err != nil {
		logger.Error("[%s %s] template execute file %s, failed %v", userId, username, cst.GoodiesTemplateFile, err)
		return err
	}
	return update.Answer(buf.String()).
		ParseMode(tg.HTML).
		ReplyMarkup(goodiesRentInlineKeyboard(sess.RentBlocks, goodies.RentalDurations)).
		DoVoid(ctx)
}

func (r *GoodiesRender) reset(ctx context.Context, callback *tgb.CallbackQueryUpdate) error {
	chatId := callback.Message.Chat.ID
	userId := callback.Message.From.ID.PeerID()
	username := strings.Replace(callback.From.Username.PeerID(), "@", "", 1)
	messageId := callback.Message.ID
	logger.Info("[%s %s] trigger action [reset] controller", userId, username)
	sess := middleware.SessionManager.Get(ctx)
	pf, err := os.ReadFile(cst.GoodiesTemplateFile)
	if err != nil {
		logger.Error("[%s %s] read template file %s, failed %v", userId, username, cst.GoodiesTemplateFile, err)
		return err
	}
	tmpl, err := template.New("rent_reception").Funcs(template.FuncMap{"format": utils.DateTime, "count": utils.EnergyCount}).Parse(string(pf))
	if err != nil {
		logger.Error("[%s %s] template parse file %s, failed %v", userId, username, cst.GoodiesTemplateFile, err)
		return err
	}
	var minEnergy int64
	goodies, err := services.GetGoodiesBaseInfo()
	if err != nil {
		logger.Error("[%s %s] service.GetMinEnergyOrder() failed %v", chatId, username, err)
		return callback.Client.EditMessageText(chatId, messageId, "服务异常，请联系管理~").DoVoid(ctx)
	}
	logger.Info("[%s %s] token_goodies api response: %+v", userId, username, goodies)
	defaultDuration := goodies.RentalDurations[0]
	min := float64(goodies.MinOrderAmountInSun) / defaultDuration.Multiplier / float64(goodies.MinEnergyPriceInSun)
	minEnergy, _ = strconv.ParseInt(fmt.Sprintf("%.f", min), 10, 64)
	duration := utils.Duration(defaultDuration.Blocks)
	sess.RentBlocks = defaultDuration.Blocks
	sess.RentMultiplier = defaultDuration.Multiplier
	sess.RentEnergy = 64000
	sess.MinEnergyPriceInSun = goodies.MinEnergyPriceInSun
	sess.MinOrderAmountInSun = goodies.MinOrderAmountInSun
	logger.Info("[%s %s] default goodies duration: %+v", userId, username, defaultDuration)
	payments := float64(sess.RentEnergy) * float64(sess.MinEnergyPriceInSun) * sess.RentMultiplier
	payments = payments / math.Pow10(6)
	amount := payments * (1.0 + global.App.Config.Telegram.Ratio)
	amount, _ = strconv.ParseFloat(fmt.Sprintf("%.3f", amount), 64)
	logger.Info("[%s %s] default cast TRX: %.3f", userId, username, amount)
	tpl := RentTmpl{
		RentAddress:  sess.RentAddress,
		RentDuration: duration,
		RentEnergy:   sess.RentEnergy,
		MinEnergy:    minEnergy,
		CastTRX:      sess.Payments,
		CurrentTime:  time.Now(),
	}
	buf := new(buffer.Buffer)
	err = tmpl.Execute(buf, tpl)
	if err != nil {
		logger.Error("[%s %s] template execute file %s, failed %v", userId, username, cst.GoodiesTemplateFile, err)
		return err
	}
	return callback.Client.EditMessageText(chatId, messageId, buf.String()).
		ParseMode(tg.HTML).
		ReplyMarkup(goodiesRentInlineKeyboard(sess.RentBlocks, goodies.RentalDurations)).
		DoVoid(ctx)
}

func (r *GoodiesRender) choose(ctx context.Context, callback *tgb.CallbackQueryUpdate) error {
	chatId := callback.Message.Chat.ID
	userId := callback.Message.From.ID.PeerID()
	username := strings.Replace(callback.From.Username.PeerID(), "@", "", 1)
	messageId := callback.Message.ID
	logger.Info("[%s %s] trigger action [choose] controller", userId, username)
	sess := middleware.SessionManager.Get(ctx)
	goodies, err := services.GetGoodiesBaseInfo()
	if err != nil {
		logger.Error("[%s %s] service.GetMinEnergyOrder() failed %v", userId, username, err)
		return callback.Client.EditMessageText(chatId, messageId, "服务异常，请联系管理~").DoVoid(ctx)
	}
	compile, err := regexp.Compile(`^rent\s+(?P<type>[a-z]+)\s+(?P<action>[a-z]+)?\s?(?P<number>(?:0|[1-9]\d*)(?:\.\d*)?)_?(?P<multiplier>\d+\.?\d+?)?`)
	if err != nil {
		logger.Error("[%s %s] compile RentChoose failed %v", userId, username, err)
		return err
	}
	groups := utils.FindGroups(compile, callback.Data)
	number, _ := strconv.ParseFloat(groups["number"], 64)
	action := groups["action"]
	sess.RentChoose = callback.Data
	min := math.Ceil(float64(sess.MinOrderAmountInSun) / sess.RentMultiplier / float64(sess.MinEnergyPriceInSun))
	minEnergy, _ := strconv.ParseInt(fmt.Sprintf("%.f", min), 10, 64)
	logger.Info("[%s %s] minimum value of energy: %.3f(float) =>%d (int64)", userId, username, min, minEnergy)
	switch groups["type"] {
	case "energy":
		if action == "plus" {
			sess.RentEnergy += int64(number)
		} else if action == "minus" {
			sess.RentEnergy -= int64(number)
			if sess.RentEnergy <= minEnergy {
				sess.RentEnergy = minEnergy
			}
		}
	case "count":
		nu := number * cst.PerCountEnergy
		if action == "plus" {
			sess.RentEnergy += int64(nu)
		} else if action == "minus" {
			sess.RentEnergy -= int64(nu)
			if sess.RentEnergy <= minEnergy {
				sess.RentEnergy = minEnergy
			}
		}
	case "blocks":
		multiplier, _ := strconv.ParseFloat(groups["multiplier"], 64)
		sess.RentBlocks, _ = strconv.ParseInt(groups["number"], 10, 64)
		sess.RentMultiplier = multiplier
	}
	duration := utils.Duration(sess.RentBlocks)
	payments := float64(sess.RentEnergy) * float64(sess.MinEnergyPriceInSun) * sess.RentMultiplier
	payments = payments / math.Pow10(6)
	point := utils.RandPoint()
	amount := payments * (1.0 + global.App.Config.Telegram.Ratio)
	amount, _ = strconv.ParseFloat(fmt.Sprintf("%.3f", amount), 64)
	amount += point
	logger.Info("[%s %s] choice [blocks=%d, duration=%s, energy=%d, payments=%.3f, amount=%.3f]", userId, username, sess.RentBlocks, duration, sess.RentEnergy, payments, amount)
	sess.Payments = fmt.Sprintf("%.3f", payments)
	sess.Amount, _ = strconv.ParseFloat(fmt.Sprintf("%.3f", amount), 64)
	pf, err := os.ReadFile(cst.GoodiesTemplateFile)
	if err != nil {
		logger.Error("[%s %s] read template file %s, failed %v", userId, username, cst.GoodiesTemplateFile, err)
		return err
	}
	tmpl, err := template.New("rent_choose").Funcs(template.FuncMap{"format": utils.DateTime, "count": utils.EnergyCount}).Parse(string(pf))
	if err != nil {
		logger.Error("[%s %s] template parse file %s, failed %v", userId, username, cst.GoodiesTemplateFile, err)
		return err
	}
	buf := new(buffer.Buffer)
	tpl := RentTmpl{
		RentAddress:  sess.RentAddress,
		RentDuration: duration,
		RentEnergy:   sess.RentEnergy,
		MinEnergy:    minEnergy,
		CastTRX:      sess.Payments,
		CurrentTime:  time.Now(),
	}
	err = tmpl.Execute(buf, tpl)
	if err != nil {
		logger.Error("[%s %s] template execute file %s, failed %v", userId, username, cst.GoodiesTemplateFile, err)
		return err
	}
	return callback.Update.Reply(
		ctx,
		callback.Client.EditMessageText(chatId, messageId, buf.String()).ParseMode(tg.HTML).
			ReplyMarkup(goodiesRentInlineKeyboard(sess.RentBlocks, goodies.RentalDurations)),
	)
}

func (r *GoodiesRender) confirm(ctx context.Context, callback *tgb.CallbackQueryUpdate) error {
	chatId := callback.Message.Chat.ID
	userId := callback.Message.From.ID.PeerID()
	username := strings.Replace(callback.From.Username.PeerID(), "@", "", 1)
	messageId := callback.Message.ID
	logger.Info("[%s %s] trigger action [confirm] controller", userId, username)
	sess := middleware.SessionManager.Get(ctx)
	duration := utils.Duration(sess.RentBlocks)
	logger.Info("[%s %s] finally rent duration %s and payment amount: %.3f", userId, username, duration, sess.Amount)
	tpl := RentTmpl{
		RentAddress:  sess.RentAddress,
		RentDuration: duration,
		RentEnergy:   sess.RentEnergy,
		CastTRX:      sess.Payments,
		CurrentTime:  time.Now(),
	}
	logger.Info("[%s %s] rent confirm sess: %+v", userId, username, sess)
	pf, err := os.ReadFile(cst.GoodiesConfirmTemplateFile)
	if err != nil {
		logger.Error("[%s %s] read template file %s, failed %v", userId, username, cst.GoodiesConfirmTemplateFile, err)
		return err
	}
	tmpl, err := template.New("rent_confirm").Funcs(template.FuncMap{"format": utils.DateTime, "count": utils.EnergyCount}).Parse(string(pf))
	if err != nil {
		logger.Error("[%s %s] template parse file %s, failed %v", userId, username, cst.GoodiesConfirmTemplateFile, err)
		return err
	}
	buf := new(buffer.Buffer)
	err = tmpl.Execute(buf, tpl)
	if err != nil {
		logger.Error("[%s %s] template execute file %s, failed %v", userId, username, cst.GoodiesTemplateFile, err)
		return err
	}
	layout := tg.NewButtonLayout[tg.InlineKeyboardButton](2).Row()
	layout.Insert(
		tg.NewInlineKeyboardButtonCallback("取消订单", Callback.Close),
		tg.NewInlineKeyboardButtonCallback("确认支付", Callback.RentPayment),
	)
	inlineKeyboard := tg.NewInlineKeyboardMarkup(layout.Keyboard()...)
	err = callback.Update.Reply(ctx, callback.Client.EditMessageText(chatId, messageId, buf.String()).ParseMode(tg.HTML).ReplyMarkup(inlineKeyboard))
	if err != nil {
		logger.Error("[%s %s] reply message failed %v", userId, username, err)
	}
	return nil
}

func (r *GoodiesRender) payment(ctx context.Context, callback *tgb.CallbackQueryUpdate) error {
	chatId := callback.Message.Chat.ID
	messageId := callback.Message.ID
	userId := callback.CallbackQuery.From.ID.PeerID()
	username := strings.Replace(callback.From.Username.PeerID(), "@", "", 1)
	logger.Info("[%s %s] trigger action [payment] controller", userId, username)
	sess := middleware.SessionManager.Get(ctx)
	logger.Info("[%s %s] rent %d energy, will be pay %.2f TRX", userId, username, sess.RentEnergy, sess.Amount)
	now := time.Now()
	order := models.Order{
		UserId:      userId,
		Username:    strings.ToLower(username),
		CreatedAt:   time.Now(),
		Energy:      sess.RentEnergy,
		ToAddress:   sess.RentAddress,
		Payments:    sess.Payments,
		Amount:      sess.Amount,
		Duration:    utils.Duration(sess.RentBlocks),
		Blocks:      sess.RentBlocks,
		Multiplier:  sess.RentMultiplier,
		FeesAddress: sess.FeesAddress,
		ChatId:      chatId.PeerID(),
		MessageId:   messageId,
		Status:      cst.OrderStatusRunning,
	}
	logger.Info("[%s %s] rent order details: %+v", userId, username, order)
	global.App.DB.Save(&order)
	buf := new(buffer.Buffer)
	tpl := RentPaymentTmpl{
		Address:   global.App.Config.Telegram.ReceiveAddress,
		Amount:    fmt.Sprintf("%.3f", sess.Amount),
		ExpiredAt: now.Add(time.Minute * 10),
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
	return callback.Client.EditMessageText(chatId, messageId, buf.String()).ParseMode(tg.HTML).ReplyMarkup(inlineKeyboard).DoVoid(ctx)
}

func goodiesRentInlineKeyboard(blocks int64, durations []services.RentalDuration) tg.InlineKeyboardMarkup {
	layout := tg.NewButtonLayout[tg.InlineKeyboardButton](3).Row()
	layout.Insert(
		tg.NewInlineKeyboardButtonCallback("↩️重置数据", Callback.RentReset),
		tg.NewInlineKeyboardButtonCallback("确认租用", Callback.RentConfirm),
	)
	inlineKeyboard := tg.NewInlineKeyboardMarkup(layout.Keyboard()...)

	layoutDuration := tg.NewButtonLayout[tg.InlineKeyboardButton](3).Row()
	for _, duration := range durations {
		df := utils.Duration(duration.Blocks)
		var isChecked bool
		if blocks == duration.Blocks {
			isChecked = true
		}
		if isChecked {
			df = fmt.Sprintf("✅%s", df)
		}
		layoutDuration.Insert(tg.NewInlineKeyboardButtonCallback(df, fmt.Sprintf("rent blocks %d_%f", duration.Blocks, duration.Multiplier)))
	}
	inlineKeyboardDuration := tg.NewInlineKeyboardMarkup(layoutDuration.Keyboard()...)

	layoutEmptyComputeCount := tg.NewButtonLayout[tg.InlineKeyboardButton](1).Row()
	layoutEmptyComputeCount.Insert(tg.NewInlineKeyboardButtonCallback("👇调整可兑换次数捷区👇", Callback.Empty))
	inlineKeyboardEmptyComputeCount := tg.NewInlineKeyboardMarkup(layoutEmptyComputeCount.Keyboard()...)

	layoutComputeCount := tg.NewButtonLayout[tg.InlineKeyboardButton](4).Row()
	layoutComputeCount.Insert(
		tg.NewInlineKeyboardButtonCallback("➕0.1次", "rent count plus 0.1"),
		tg.NewInlineKeyboardButtonCallback("➕1次", "rent count plus 1"),
		tg.NewInlineKeyboardButtonCallback("➕3次", "rent count plus 3"),
		tg.NewInlineKeyboardButtonCallback("➕5次", "rent count plus 5"),
		tg.NewInlineKeyboardButtonCallback("➕10次", "rent count plus 10"),
		tg.NewInlineKeyboardButtonCallback("➕30次", "rent count plus 30"),
		tg.NewInlineKeyboardButtonCallback("➖0.1次", "rent count minus 0.1"),
		tg.NewInlineKeyboardButtonCallback("➖1次", "rent count minus 1"),
		tg.NewInlineKeyboardButtonCallback("➖3次", "rent count minus 3"),
		tg.NewInlineKeyboardButtonCallback("➖5次", "rent count minus 5"),
		tg.NewInlineKeyboardButtonCallback("➖10次", "rent count minus 10"),
	)
	inlineKeyboardComputeCount := tg.NewInlineKeyboardMarkup(layoutComputeCount.Keyboard()...)

	layoutEmptyComputeEnergy := tg.NewButtonLayout[tg.InlineKeyboardButton](1).Row()
	layoutEmptyComputeEnergy.Insert(tg.NewInlineKeyboardButtonCallback("👇调整能量数快捷区👇", Callback.Empty))
	inlineKeyboardEmptyComputeEnergy := tg.NewInlineKeyboardMarkup(layoutEmptyComputeEnergy.Keyboard()...)

	layoutComputeEnergy := tg.NewButtonLayout[tg.InlineKeyboardButton](4).Row()
	layoutComputeEnergy.Insert(
		tg.NewInlineKeyboardButtonCallback("➕1万", "rent energy plus 10000"),
		tg.NewInlineKeyboardButtonCallback("➕5万", "rent energy plus 50000"),
		tg.NewInlineKeyboardButtonCallback("➕10万", "rent energy plus 100000"),
		tg.NewInlineKeyboardButtonCallback("➕100万", "rent energy plus 1000000"),
		tg.NewInlineKeyboardButtonCallback("➖1万", "rent energy minus 10000"),
		tg.NewInlineKeyboardButtonCallback("➖5万", "rent energy minus 50000"),
		tg.NewInlineKeyboardButtonCallback("➖10万", "rent energy minus 100000"),
		tg.NewInlineKeyboardButtonCallback("➖100万", "rent energy minus 1000000"),
	)
	inlineKeyboardComputeEnergy := tg.NewInlineKeyboardMarkup(layoutComputeEnergy.Keyboard()...)

	layoutBase := tg.NewButtonLayout[tg.InlineKeyboardButton](2).Row()
	layoutBase.Insert(
		tg.NewInlineKeyboardButtonCallback("关闭", Callback.Close),
		tg.NewInlineKeyboardButtonURL("联系客服", fmt.Sprintf("https://t.me/%s", global.App.Config.App.Support)),
	)
	inlineKeyboardBase := tg.NewInlineKeyboardMarkup(layoutBase.Keyboard()...)

	inlineKeyboard.InlineKeyboard = append(inlineKeyboard.InlineKeyboard, inlineKeyboardDuration.InlineKeyboard...)
	inlineKeyboard.InlineKeyboard = append(inlineKeyboard.InlineKeyboard, inlineKeyboardEmptyComputeCount.InlineKeyboard...)
	inlineKeyboard.InlineKeyboard = append(inlineKeyboard.InlineKeyboard, inlineKeyboardComputeCount.InlineKeyboard...)
	inlineKeyboard.InlineKeyboard = append(inlineKeyboard.InlineKeyboard, inlineKeyboardEmptyComputeEnergy.InlineKeyboard...)
	inlineKeyboard.InlineKeyboard = append(inlineKeyboard.InlineKeyboard, inlineKeyboardComputeEnergy.InlineKeyboard...)
	inlineKeyboard.InlineKeyboard = append(inlineKeyboard.InlineKeyboard, inlineKeyboardBase.InlineKeyboard...)
	return inlineKeyboard
}
