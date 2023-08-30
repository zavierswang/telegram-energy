package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fbsobreira/gotron-sdk/pkg/common"
	"go.uber.org/zap/buffer"
	"io"
	"math"
	"net/http"
	"os/exec"
	"telegram-energy/pkg/common/httpclient"
	"telegram-energy/pkg/core/cst"
	"telegram-energy/pkg/core/global"
	"telegram-energy/pkg/core/logger"
	"telegram-energy/pkg/models"
)

type TronResource struct {
	MinOrderAmountInSun      int64            `json:"min_order_amount_in_sun"`
	MinEnergyPriceInSun      int64            `json:"min_energy_price_in_sun"`
	MinBandwidthPriceInSun   int64            `json:"min_bandwidth_price_in_sun"`
	ContractAddress          string           `json:"contract_address"`
	OrderFeesAddress         string           `json:"order_fees_address"`
	CreateOrderFeeLimitInSun int64            `json:"create_order_fee_limit_in_sun"`
	RentalDurations          []RentalDuration `json:"rental_durations"`
}

type RentalDuration struct {
	Blocks     int64   `json:"blocks"`
	Multiplier float64 `json:"multiplier"`
}

type TokenGoodiesResp struct {
	Success       bool   `json:"success"`
	Action        string `json:"action"`
	Type          string `json:"type"`
	RequestId     string `json:"requestid"`
	Message       string `json:"message"`
	OrderId       int64  `json:"orderid"`
	CreateTxnHash string `json:"createtxnhash"`
	LogType       string `json:"logtype"`
}

type ListOrderResp struct {
	Total  int     `json:"total"`
	Orders []Order `json:"orders"`
}

type Order struct {
	ResourceId             int    `json:"resourceid"`
	OrderId                int64  `json:"orderid"`
	TokenId                int64  `json:"tokenid"`
	PriceInSun             int    `json:"priceinsun"`
	ResourceAmount         int    `json:"resourceamount"`
	TotalPayoutAmountInSun int64  `json:"totalpayoutamountinsun"`
	TotalTrxToFreeze       int64  `json:"totaltrxtofreeze"`
	MinResourceAmount      int    `json:"minresourceamount"`
	MinPayoutAmountInSun   int    `json:"minpayoutamountinsun"`
	MinTrxToFreeze         int    `json:"mintrxtofreeze"`
	FreezePeriod           int    `json:"freezeperiod"`
	FreezeTo               string `json:"freezeto"`
}

func GetGoodiesBaseInfo() (*TronResource, error) {
	//获取基本参数
	body := map[string]string{
		"action":  cst.GoodiesGetBaseAction,
		"type":    cst.GoodiesCreateApiBase,
		"api_key": global.App.Config.Telegram.AppKey,
	}
	b, _ := json.Marshal(&body)
	resp, err := http.Post(cst.GoodiesEnergyRentApi, "application/json", bytes.NewReader(b))
	if err != nil {
		logger.Error("GoodiesEnergyBaseInfo post failed %v", err)
		return nil, err
	}
	defer resp.Body.Close()
	buf := new(buffer.Buffer)
	_, _ = io.Copy(buf, resp.Body)
	var resource TronResource
	err = json.Unmarshal(buf.Bytes(), &resource)
	if err != nil {
		logger.Error("GoodiesEnergyBaseInfo resource unmarshal failed %v", err)
		return nil, err
	}
	return &resource, nil
}

func CreateGoodiesOrder(toAddress string, amount float64) (*TokenGoodiesResp, error) {
	res, err := GetGoodiesBaseInfo()
	if err != nil {
		return nil, err
	}
	logger.Info("token_goodies of base info: %+v", res)
	var orders []models.Order
	global.App.DB.Find(&orders, "amount = ? AND finished = ? AND expired = ?", amount, false, false)
	if len(orders) == 0 {
		logger.Error("not match rent energy order of token_goodies")
		return nil, errors.New("not found token_goodies order")
	}
	order := orders[0]
	logger.Info("token_goodies order: %+v", order)
	// The number of days you need the energy/bandwidth for. Always pass 3
	freezeDays := 3
	// This value should be >= min_energy_price_in_sun variable above for energy orders and >= min_bandwidth_price_in_sun variable above for bandwidth orders
	var priceInSun int64
	if priceInSun <= res.MinEnergyPriceInSun {
		priceInSun = res.MinEnergyPriceInSun
	}
	// IMPORTANT: Make sure this value is >= min_order_amount_in_sun variable above
	orderFeesAmountInSun := float64(order.Energy) * float64(priceInSun) * order.Multiplier
	logger.Info("fees amount in sun: %f", orderFeesAmountInSun)
	if int64(orderFeesAmountInSun) < res.MinOrderAmountInSun {
		//orderFeesAmountInSun = resource.MinOrderAmountInSun
		return nil, errors.New(fmt.Sprintf("amount * priceInSun * freezeDays value must >= %d", res.MinOrderAmountInSun))
	}
	orderFeesAmountInSunStr := fmt.Sprintf("%.f", orderFeesAmountInSun)
	logger.Info("this order will be pay %.3f TRX", orderFeesAmountInSun/math.Pow10(6))
	//TODO: replace with golang complement
	logger.Info("will execute nodejs sendTrx(%s, %s)", order.FeesAddress, orderFeesAmountInSunStr)
	cmd := exec.Command("node", "third/index.js", order.FeesAddress, orderFeesAmountInSunStr)
	outByte, err := cmd.CombinedOutput()
	if err != nil {
		logger.Error("transfer TRX failed %s, reason:%v", string(outByte), err)
		return nil, err
	}
	logger.Info("execute nodejs third/index.js finished")
	param := map[string]interface{}{
		"action":     cst.GoodiesCreateAction,
		"resourceid": cst.GoodiesEnergyResourceId,
		"signedtxn":  string(outByte),
		"type":       cst.GoodiesType,
		"api_key":    global.App.Config.Telegram.AppKey,
		"order": map[string]interface{}{
			"freezeto":             toAddress,
			"amount":               order.Energy,
			"freezeperiod":         freezeDays,
			"freezeperiodinblocks": order.Blocks,
			"priceinsun":           float64(res.MinEnergyPriceInSun) * order.Multiplier,
			"priceinsunabsolute":   res.MinEnergyPriceInSun,
		},
	}
	var goodiesRes TokenGoodiesResp
	err = httpclient.PostJson(cst.GoodiesEnergyRentApi, param, nil, nil, &goodiesRes)
	if err != nil {
		logger.Error("CreateGoodiesOrder failed %v", err)
		return nil, err
	}
	return &goodiesRes, nil
}

func CreateTokenGoodiesCommonOrder(toAddress string, amount float64) (*TokenGoodiesResp, error) {
	res, err := GetGoodiesBaseInfo()
	if err != nil {
		return nil, err
	}
	logger.Info("token_goodies of base info: %+v", res)
	var orders []models.Order
	global.App.DB.Find(&orders, "amount = ? AND finished = ? AND expired = ?", amount, false, false)
	if len(orders) == 0 {
		logger.Error("not match rent energy order of token_goodies")
		return nil, errors.New("not found token_goodies order")
	}
	order := orders[0]
	logger.Info("token_goodies order: %+v", order)
	// The number of days you need the energy/bandwidth for. Always pass 3
	freezeDays := 3
	// This value should be >= min_energy_price_in_sun variable above for energy orders and >= min_bandwidth_price_in_sun variable above for bandwidth orders
	var priceInSun int64
	if priceInSun <= res.MinEnergyPriceInSun {
		priceInSun = res.MinEnergyPriceInSun
	}
	// IMPORTANT: Make sure this value is >= min_order_amount_in_sun variable above
	orderFeesAmountInSun := float64(order.Energy) * float64(priceInSun) * order.Multiplier
	logger.Info("fees amount in sun: %f", orderFeesAmountInSun)
	if int64(orderFeesAmountInSun) < res.MinOrderAmountInSun {
		//orderFeesAmountInSun = resource.MinOrderAmountInSun
		return nil, errors.New(fmt.Sprintf("amount * priceInSun * freezeDays value must >= %d", res.MinOrderAmountInSun))
	}
	orderFeesAmountInSunStr := fmt.Sprintf("%.f", orderFeesAmountInSun)
	logger.Info("this order will be pay %.3f TRX", orderFeesAmountInSun/math.Pow10(6))
	//TODO: replace with golang complement
	logger.Info("will execute nodejs sendTrx(%s, %s)", order.FeesAddress, orderFeesAmountInSunStr)
	cmd := exec.Command("node", "third/index.js", order.FeesAddress, orderFeesAmountInSunStr)
	outByte, err := cmd.CombinedOutput()
	if err != nil {
		logger.Error("transfer TRX failed %s, reason:%v", string(outByte), err)
		return nil, err
	}
	logger.Info("execute nodejs third/index.js finished")
	param := map[string]interface{}{
		"action":     cst.GoodiesCreateAction,
		"resourceid": cst.GoodiesEnergyResourceId,
		"signedtxn":  string(outByte),
		"type":       cst.GoodiesType,
		"api_key":    global.App.Config.Telegram.AppKey,
		"order": map[string]interface{}{
			"freezeto":             toAddress,
			"amount":               order.Energy,
			"freezeperiod":         freezeDays,
			"freezeperiodinblocks": order.Blocks,
			"priceinsun":           float64(res.MinEnergyPriceInSun) * order.Multiplier,
			"priceinsunabsolute":   res.MinEnergyPriceInSun,
		},
	}
	b, _ := json.Marshal(&param)
	fmt.Println(common.JSONPrettyFormat(string(b)))
	resp, err := http.Post(cst.GoodiesEnergyRentApi, "application/json", bytes.NewReader(b))
	if err != nil {
		logger.Error("%s post failed %v", cst.GoodiesEnergyRentApi, err)
		return nil, err
	}
	defer resp.Body.Close()
	bf := new(buffer.Buffer)
	_, _ = io.Copy(bf, resp.Body)
	var goodiesRes TokenGoodiesResp
	err = json.Unmarshal(bf.Bytes(), &goodiesRes)
	if err != nil {
		logger.Error("goodies response Unmarshal failed %v", err)
		return nil, err
	}
	logger.Info("goodies create order response: %+v", goodiesRes)
	return &goodiesRes, nil
}
