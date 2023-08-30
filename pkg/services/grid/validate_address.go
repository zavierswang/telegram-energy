package grid

import (
	"fmt"
	"telegram-energy/pkg/common/httpclient"
	"telegram-energy/pkg/core/cst"
	"telegram-energy/pkg/core/global"
	"telegram-energy/pkg/core/logger"
)

const (
	validateAddressApi = "/wallet/validateaddress"
)

type ValidateAddressResp struct {
	Result  bool   `json:"result"`
	Message string `json:"message"`
}

func ValidateAddress(address string) (*ValidateAddressResp, error) {
	uri := fmt.Sprintf("%s%s", cst.TronBaseApi, validateAddressApi)
	body := map[string]interface{}{
		"address": address,
		"visible": true,
	}
	headers := map[string]string{
		"accept":           "application/json",
		"TRON_PRO_API_KEY": global.App.Config.Telegram.GridApiKey,
	}
	var resp ValidateAddressResp
	err := httpclient.PostJson(uri, body, headers, nil, &resp)
	if err != nil {
		logger.Error("ValidateAddress request api %s failed %v", uri, err)
		return nil, err
	}
	return &resp, nil
}
