package services

import (
	"encoding/json"
	"fmt"
	"telegram-energy/pkg/common/httpclient"
	"telegram-energy/pkg/core/cst"
	"telegram-energy/pkg/core/global"
	"telegram-energy/pkg/utils"
	"time"
)

type FeeBuyEnergyReq struct {
	ApiID          int     `json:"api_id"`
	Resource       int     `json:"resource,omitempty"`
	Count          int64   `json:"count"`
	PayAmount      float64 `json:"pay_amount,omitempty"`
	RentalSeconds  int64   `json:"rental_seconds,omitempty"`
	ReceiveAddress string  `json:"receive_address,omitempty"`
	CreateTime     int64   `json:"create_time"`
}

type FeeBuyEnergyResp struct {
	Code int              `json:"code"`
	Data FeeBuyEnergyData `json:"data"`
}

type FeeBuyEnergyData struct {
	Id         string `json:"id"`
	Currency   string `json:"currency"`    //货币 为trx
	Amount     int    `json:"amount"`      //订单生成金额 小数点6位
	PayAmount  int    `json:"pay_amount"`  //客户需要支付金额 小数点6位
	PayAddress string `json:"pay_address"` //收款地址
	ExpTime    int64  `json:"exp_time"`    //订单超时 时间戳
	Count      int64  `json:"count"`       //资源数量
	Resource   int    `json:"resource"`    //资源类型 1能量2宽带
	Price      int    `json:"price"`       //购买单价
	Day        int64  `json:"day"`         //订单3天 1或者3
}

func FeeBuyEnergy(resource int, rentalDuration int64, count int64, address string) (*FeeBuyEnergyResp, error) {
	body := FeeBuyEnergyReq{
		ApiID:          global.App.Config.Telegram.ApiID,
		Resource:       resource,
		Count:          count,
		RentalSeconds:  rentalDuration,
		ReceiveAddress: address,
		CreateTime:     time.Now().Unix(),
	}
	url := fmt.Sprintf("%s/%s", cst.FeeRentEnergyApi, "buyEnergy")
	js, _ := json.Marshal(body)
	sign := utils.MD5(js)                           //将data数据转MD5
	sign = global.App.Config.Telegram.AppKey + sign //appKey + md5(data)
	sign = utils.MD5([]byte(sign))                  //md5(appKey + md5(data))
	params := map[string]string{
		"sign": sign,
	}
	var response FeeBuyEnergyResp
	err := httpclient.PostJson(url, body, nil, params, &response)
	if err != nil || response.Code != 0 {
		return nil, err
	}
	//logger.Info("FeeBuyEnergy response: %+v", response)
	return &response, nil
}
