package utils

import "fmt"

func AddWalletToPlatform(config *Config, request *AddWalletRequest) error {
	res := ResponseBody{}
	header := make(map[string]string)
	header["Content-Type"] = "application/json"
	header["Authorization"] = config.AuthToken
	header["InstanceId"] = config.InstanceId
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
	if err := HttpCall("POST", config.ProxyUrl+UpdateWalletNonce, request, &res, header); err != nil {
		return err
	}
	if res.Status == "FAILURE" {
		return fmt.Errorf(res.Message)
	}
	return nil
}
