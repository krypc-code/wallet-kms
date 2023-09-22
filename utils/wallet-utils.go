package utils

import (
	"encoding/json"
	"fmt"

	"github.com/labstack/echo/v4"
)

func AddWalletToPlatform(config *Config, request *AddWalletRequest) error {
	res := ResponseBody{}
	header := make(map[string]string)
	header["Content-Type"] = "application/json"
	header["Authorization"] = config.AuthToken
	header["InstanceId"] = config.InstanceId
	header["SubscriptionId"] = config.SubscriptionId
	if err := HttpCall("POST", config.ProxyUrl+AddNcWallet, request, &res, header); err != nil {
		return err
	}
	if res.Status == "FAILURE" {
		return fmt.Errorf(res.Message)
	}
	return nil
}

func GetNonceFromPlatform(config *Config, request *NonceRequest) (uint64, error) {
	res := ResponseBody{}
	header := make(map[string]string)
	header["Content-Type"] = "application/json"
	header["Authorization"] = config.AuthToken
	header["InstanceId"] = config.InstanceId
	header["SubscriptionId"] = config.SubscriptionId
	if err := HttpCall("POST", config.ProxyUrl+GetWalletNonce, request, &res, header); err != nil {
		return 0, err
	}
	if res.Status == "FAILURE" {
		return 0, fmt.Errorf(res.Message)
	}
	nonce, ok := res.Data.(map[string]interface{})["nonce"].(float64)
	if !ok {
		return 0, fmt.Errorf("invalid nonce")
	}
	return uint64(nonce), nil
}

func UpdatePlatformNonce(config *Config, request *NonceRequest) error {
	res := ResponseBody{}
	header := make(map[string]string)
	header["Content-Type"] = "application/json"
	header["Authorization"] = config.AuthToken
	header["InstanceId"] = config.InstanceId
	header["SubscriptionId"] = config.SubscriptionId
	if err := HttpCall("POST", config.ProxyUrl+UpdateWalletNonce, request, &res, header); err != nil {
		return err
	}
	if res.Status == "FAILURE" {
		return fmt.Errorf(res.Message)
	}
	return nil
}

func GetAllDeployRecords(config *Config) ([]Deploy, error) {
	req := Request{}
	res := ResponseBody{}
	header := make(map[string]string)
	header["Content-Type"] = "application/json"
	header["Authorization"] = config.AuthToken
	header["InstanceId"] = config.InstanceId
	header["SubscriptionId"] = config.SubscriptionId
	if err := HttpCall("POST", config.ProxyUrl+FetchDeployRecords, req, &res, header); err != nil {
		return nil, err
	}
	if res.Status == "FAILURE" {
		return nil, fmt.Errorf(res.Message)
	}
	resBytes, err := json.Marshal(res.Data)
	if err != nil {
		return nil, err
	}
	var records []Deploy
	if err := json.Unmarshal(resBytes, &records); err != nil {
		return nil, err
	}
	return records, nil
}

func GetAllTransactionRecords(config *Config) ([]Transaction, error) {
	req := Request{}
	res := ResponseBody{}
	header := make(map[string]string)
	header["Content-Type"] = "application/json"
	header["Authorization"] = config.AuthToken
	header["InstanceId"] = config.InstanceId
	header["SubscriptionId"] = config.SubscriptionId
	if err := HttpCall("POST", config.ProxyUrl+FetchTransactionsRecords, req, &res, header); err != nil {
		return nil, err
	}
	if res.Status == "FAILURE" {
		return nil, fmt.Errorf(res.Message)
	}
	resBytes, err := json.Marshal(res.Data)
	if err != nil {
		return nil, err
	}
	var records []Transaction
	if err := json.Unmarshal(resBytes, &records); err != nil {
		return nil, err
	}
	return records, nil
}

func GetBalanceByChainId(config *Config, address, chainId string) (*BalanceResponse, error) {
	request := BalanceRequest{
		Address: address,
		ChainId: chainId,
	}
	res := ResponseBody{}
	header := make(map[string]string)
	header["Content-Type"] = "application/json"
	header["Authorization"] = config.AuthToken
	header["InstanceId"] = config.InstanceId
	header["SubscriptionId"] = config.SubscriptionId
	if err := HttpCall("POST", config.ProxyUrl+GetBalance, request, &res, header); err != nil {
		return nil, err
	}
	if res.Status == "FAILURE" {
		return nil, fmt.Errorf(res.Message)
	}
	resBytes, err := json.Marshal(res.Data)
	if err != nil {
		return nil, err
	}
	var balanceRes BalanceResponse
	if err := json.Unmarshal(resBytes, &balanceRes); err != nil {
		return nil, err
	}
	return &balanceRes, nil
}

func GetPendingWalletRecords(config *Config) ([]PendingWalletResponse, error) {
	req := Request{}
	res := ResponseBody{}
	header := make(map[string]string)
	header["Content-Type"] = "application/json"
	header["Authorization"] = config.AuthToken
	header["InstanceId"] = config.InstanceId
	header["SubscriptionId"] = config.SubscriptionId
	if err := HttpCall("POST", config.ProxyUrl+FetchPendingWallets, req, &res, header); err != nil {
		return nil, err
	}
	if res.Status == "FAILURE" {
		return nil, fmt.Errorf(res.Message)
	}
	resBytes, err := json.Marshal(res.Data)
	if err != nil {
		return nil, err
	}
	var records []PendingWalletResponse
	if err := json.Unmarshal(resBytes, &records); err != nil {
		return nil, err
	}
	return records, nil
}

func UpdatePendingWallet(config *Config, request *AddWalletRequest) error {
	res := ResponseBody{}
	header := make(map[string]string)
	header["Content-Type"] = "application/json"
	header["Authorization"] = config.AuthToken
	header["InstanceId"] = config.InstanceId
	header["SubscriptionId"] = config.SubscriptionId
	if err := HttpCall("POST", config.ProxyUrl+UpdatePendingWallets, request, &res, header); err != nil {
		return err
	}
	if res.Status == "FAILURE" {
		return fmt.Errorf(res.Message)
	}
	return nil
}

func HttpCallWithContextHeaderJson(c echo.Context, method, url string, body interface{}, restype interface{}) error {
	header := make(map[string]string)
	header["Authorization"] = c.Request().Header.Get("Authorization")
	header["requestId"] = c.Request().Header.Get("requestId")
	header["hopCount"] = c.Request().Header.Get("hopCount")
	header["Content-Type"] = "application/json"
	return HttpCall(method, url, body, restype, header)
}
