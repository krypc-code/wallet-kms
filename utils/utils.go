package utils

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"log"
	"math/big"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/labstack/echo/v4"
)

const (
	AddNcWallet       = "/ncWallet/addWallet"
	GetWalletNonce    = "/ncWallet/getNonce"
	UpdateWalletNonce = "/ncWallet/updateNonce"
)

type ResponseBody struct {
	Status  string
	Message string
	Data    interface{}
}

func BadRequestResponse(c echo.Context, message string, data interface{}) error {
	RespBody := ResponseBody{
		Status:  "FAILURE",
		Message: message,
		Data:    data,
	}
	return c.JSON(http.StatusBadRequest, RespBody)
}

func UnauthorizedResponse(c echo.Context, message string, data interface{}) error {
	RespBody := ResponseBody{
		Status:  "FAILURE",
		Message: message,
		Data:    data,
	}
	return c.JSON(http.StatusUnauthorized, RespBody)
}

func UnexpectedFailureResponse(c echo.Context, message string, data interface{}) error {
	RespBody := ResponseBody{
		Status:  "FAILURE",
		Message: message,
		Data:    data,
	}
	return c.JSON(http.StatusExpectationFailed, RespBody)
}

func SendFailureResponse(c echo.Context, message string, data interface{}) error {
	RespBody := ResponseBody{
		Status:  "FAILURE",
		Message: message,
		Data:    data,
	}
	return c.JSON(http.StatusOK, RespBody)
}

func SendSuccessResponse(c echo.Context, message string, data interface{}) error {
	RespBody := ResponseBody{
		Status:  "SUCCESS",
		Message: message,
		Data:    data,
	}
	return c.JSON(http.StatusOK, RespBody)
}

func ConvertParamsAsPerTypes(params []Param) ([]interface{}, error) {
	var response []interface{}
	for _, param := range params {
		switch param.Type {
		case "string":
			response = append(response, param.Value)
		case "uint", "uint8", "uint16", "uint32", "uint64", "uint128", "uint256":
			response = append(response, big.NewInt(int64(param.Value.(float64))))
		case "bool":
			response = append(response, param.Value)
		case "address":
			response = append(response, common.HexToAddress(param.Value.(string)))
		case "int", "int8", "int16", "int32", "int64", "int128", "int256":
			response = append(response, big.NewInt(int64(param.Value.(float64))))
		default:
			response = append(response, param.Value)
		}
	}
	return response, nil
}

func HttpCall(method string, url string, body interface{}, restype interface{}, header map[string]string) error {
	var databytes []byte
	var err error
	if nil != body {
		databytes, err = json.Marshal(body)
		if err != nil {
			return err
		}
	} else {
		databytes = nil
	}
	req, err := http.NewRequest(method, url, bytes.NewBuffer(databytes))
	if err != nil {
		return err
	}
	for key, val := range header {
		req.Header.Set(key, val)
	}
	client := &http.Client{}
	client.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// Use json.Decode for reading streams of JSON data
	if err := json.NewDecoder(resp.Body).Decode(&restype); err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func GetEthereumClient(ctx context.Context, config *Config) (*ethclient.Client, error) {
	header := make(http.Header)
	header.Set("Content-Type", "application/json")
	header.Set("Authorization", config.AuthToken)
	header.Set("instanceId", config.InstanceId)
	options := rpc.WithHeaders(header)
	rpcClient, err := rpc.DialOptions(ctx, config.Endpoint, options)
	if err != nil {
		return nil, err
	}
	client := ethclient.NewClient(rpcClient)
	return client, nil
}

func ValidateAlgorithm(algorithm string) (string, error) {
	switch algorithm {
	case "ed25519":
		return "ed25519", nil
	case "secp256k1":
		return "ecdsa-p256", nil
	default:
		return "ecdsa-p256", nil
	}
}
