package kms

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"strings"
	kmsmiddleware "wallet-kms/kms/middleware"
	"wallet-kms/store"
	"wallet-kms/utils"
	"wallet-kms/vault"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Service struct {
	e      *echo.Echo
	db     store.DB
	vault  vault.Vault
	config *utils.Config
}

func initService() *Service {
	var serve Service
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	// B1: enforce X-Tenant-ID forwarded by Kernel on every request
	e.Use(kmsmiddleware.TenantContext())

	db, err := store.NewBadgerDB(".wallet/db/")
	if err != nil {
		log.Panic("error initializing wallet db")
	}
	vaultSvc, err := vault.NewHashiCorpVault(os.Getenv("VAULT_URL"), os.Getenv("VAULT_TOKEN"))
	if err != nil {
		log.Panic("error initializing vault : ", err.Error())
	}

	sorBaseURL := os.Getenv("SOR_BASE_URL")
	if sorBaseURL == "" {
		log.Panic("SOR_BASE_URL environment variable not set")
	}

	serve.db = db
	serve.vault = vaultSvc
	serve.config = &utils.Config{
		SORBaseURL: sorBaseURL,
		ProxyUrl:   os.Getenv("PROXY_URL"), // optional; only needed for gasless relay
	}
	serve.e = e
	return &serve
}

func Run() {
	service := initService()
	g := service.e.Group("/wallet")

	g.GET("/health", healthCheck)

	g.POST("/createWallet", service.createWallet)
	g.POST("/submitTransaction", service.submitTransaction)
	g.POST("/signAndSubmitGaslessTxn", service.signAndSubmitGaslessTransaction)
	g.POST("/deployContract", service.deployContract)
	g.POST("/estimateGas", service.estimateGas)
	g.POST("/getBalance", service.getBalance)
	g.POST("/callContract", service.callContract)
	g.POST("/signEIP712Tx", service.signEIP712Txn)
	g.POST("/signMessage", service.signMessage)
	g.POST("/verifySignatureOffChain", service.verifySignatureOffChain)

	service.e.Logger.Fatal(service.e.Start(":8889"))
}

func healthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": "Server is up and running",
	})
}

func (serve *Service) createWallet(c echo.Context) error {
	ctx := c.Request().Context()
	tenantID := kmsmiddleware.TenantIDFrom(c)

	request := new(utils.WalletRequest)
	if err := c.Bind(request); err != nil {
		return utils.BadRequestResponse(c, "bad request", nil)
	}
	if request.Algorithm == "" || request.Name == "" {
		return utils.BadRequestResponse(c, "mandatory params missing", nil)
	}

	walletId := uuid.New()
	wallet := Wallet{
		Name:      request.Name,
		Algorithm: request.Algorithm,
		WalletId:  walletId.String(),
		VaultPath: tenantID + "/" + walletId.String(), // B2: tenant-scoped vault path
	}
	if err := wallet.generateKey(ctx, serve.vault); err != nil {
		return utils.UnexpectedFailureResponse(c, err.Error(), nil)
	}
	data, err := json.Marshal(wallet)
	if err != nil {
		return utils.UnexpectedFailureResponse(c, err.Error(), nil)
	}
	if err := serve.db.Set([]byte(utils.NAMESPACE), walletId.NodeID(), data); err != nil {
		return utils.UnexpectedFailureResponse(c, err.Error(), nil)
	}
	return utils.SendSuccessResponse(c, "wallet created successfully", utils.WalletResponse{
		WalletId: wallet.WalletId,
		Address:  wallet.Address,
	})
}

func (s *Service) submitTransaction(c echo.Context) error {
	ctx := c.Request().Context()
	tenantID := kmsmiddleware.TenantIDFrom(c)

	u := new(utils.SignAndSubmitTxn)
	if err := c.Bind(u); err != nil {
		return utils.BadRequestResponse(c, "bad request", nil)
	}
	if u.ChainId == "" {
		return utils.BadRequestResponse(c, "chainId is required", nil)
	}

	walletId, err := uuid.Parse(u.WalletId)
	if err != nil {
		return utils.UnexpectedFailureResponse(c, err.Error(), nil)
	}
	walletBytes, err := s.db.Get([]byte(utils.NAMESPACE), walletId.NodeID())
	if err != nil {
		return utils.UnauthorizedResponse(c, err.Error(), nil)
	}
	wallet := &Wallet{}
	if err := json.Unmarshal(walletBytes, wallet); err != nil {
		return utils.UnexpectedFailureResponse(c, err.Error(), nil)
	}

	// B3: route all chain calls through SOR
	client, err := utils.GetEthereumClientViaSOR(ctx, s.config.SORBaseURL, u.ChainId, tenantID)
	if err != nil {
		return utils.UnexpectedFailureResponse(c, err.Error(), nil)
	}
	chainId, err := client.ChainID(ctx)
	if err != nil {
		return utils.UnexpectedFailureResponse(c, err.Error(), nil)
	}

	var to common.Address
	if u.To != "" {
		to = common.HexToAddress(u.To)
	}

	var txnHash string
	if u.IsContractTxn {
		if u.ContractABI == "" {
			return utils.UnexpectedFailureResponse(c, "missing contract abi", nil)
		}
		contract, err := NewSmartContract(to, u.ContractABI, client)
		if err != nil {
			return utils.UnexpectedFailureResponse(c, "error initializing contract : "+err.Error(), nil)
		}
		txnOpts, err := TransactionOptionsWithKMSSigning(ctx, wallet, s.vault, chainId)
		if err != nil {
			return utils.UnexpectedFailureResponse(c, "error initializing transactor opts : "+err.Error(), nil)
		}
		if u.Value > 0 {
			txnOpts.Value = big.NewInt(u.Value)
		}
		nonce, err := client.PendingNonceAt(ctx, common.HexToAddress(wallet.Address))
		if err != nil {
			return utils.UnexpectedFailureResponse(c, err.Error(), nil)
		}
		txnOpts.Nonce = big.NewInt(int64(nonce))
		if u.Gas != 0 {
			txnOpts.GasPrice = big.NewInt(int64(u.Gas))
		}
		if u.GasLimit != 0 {
			txnOpts.GasLimit = u.GasLimit
		}
		params, err := utils.ConvertParamsAsPerTypes(u.Params)
		if err != nil {
			return utils.BadRequestResponse(c, "error converting params "+err.Error(), nil)
		}
		txn, err := contract.SmartContractTransactor.Contract.Transact(txnOpts, u.Method, params...)
		if err != nil {
			return utils.UnexpectedFailureResponse(c, "error sending txn to contract : "+err.Error(), nil)
		}
		txnHash = txn.Hash().String()
	} else {
		gasPrice, err := client.SuggestGasPrice(ctx)
		if err != nil {
			return utils.UnexpectedFailureResponse(c, "error calculating gas price : "+err.Error(), nil)
		}
		var dataBytes []byte
		if u.Data != "" {
			dataBytes, err = base64.RawStdEncoding.DecodeString(u.Data)
			if err != nil {
				return utils.UnexpectedFailureResponse(c, "error decoding data : "+err.Error(), nil)
			}
		}
		nonce, err := client.PendingNonceAt(ctx, common.HexToAddress(wallet.Address))
		if err != nil {
			return utils.UnexpectedFailureResponse(c, err.Error(), nil)
		}
		var estimatedGas uint64
		if u.Gas > 0 {
			estimatedGas = u.Gas
		} else {
			estimatedGas, err = client.EstimateGas(ctx, ethereum.CallMsg{From: common.HexToAddress(wallet.Address), To: &to, Value: big.NewInt(u.Value), Data: dataBytes, GasPrice: gasPrice})
			if err != nil {
				return utils.UnexpectedFailureResponse(c, "error estimating gas : "+err.Error(), nil)
			}
		}
		txn := types.NewTx(&types.LegacyTx{
			Nonce:    nonce,
			GasPrice: gasPrice,
			Gas:      estimatedGas,
			Value:    big.NewInt(u.Value),
			Data:     dataBytes,
			To:       &to,
		})
		signer := types.LatestSignerForChainID(chainId)
		signature, err := SignTransactionHash(ctx, wallet, s.vault, txn.Hash().Bytes())
		if err != nil {
			return utils.UnexpectedFailureResponse(c, err.Error(), nil)
		}
		txn, err = txn.WithSignature(signer, signature)
		if err != nil {
			return utils.UnexpectedFailureResponse(c, err.Error(), nil)
		}
		if err := client.SendTransaction(ctx, txn); err != nil {
			return utils.UnexpectedFailureResponse(c, "error executing txn : "+err.Error(), nil)
		}
		txnHash = txn.Hash().String()
	}
	return utils.SendSuccessResponse(c, "Signed and executed txn successfully", &utils.DeployContractResponse{TxnHash: txnHash})
}

func (s *Service) deployContract(c echo.Context) error {
	ctx := c.Request().Context()
	tenantID := kmsmiddleware.TenantIDFrom(c)

	u := new(utils.DeployContractRequest)
	if err := c.Bind(u); err != nil {
		s.e.Logger.Errorf(err.Error())
		return utils.BadRequestResponse(c, "bad request", nil)
	}
	if u.ChainId == "" {
		return utils.BadRequestResponse(c, "chainId is required", nil)
	}

	walletId, err := uuid.Parse(u.WalletId)
	if err != nil {
		return utils.UnexpectedFailureResponse(c, err.Error(), nil)
	}
	walletBytes, err := s.db.Get([]byte(utils.NAMESPACE), walletId.NodeID())
	if err != nil {
		return utils.UnauthorizedResponse(c, err.Error(), nil)
	}
	wallet := &Wallet{}
	if err := json.Unmarshal(walletBytes, wallet); err != nil {
		return utils.UnexpectedFailureResponse(c, err.Error(), nil)
	}

	client, err := utils.GetEthereumClientViaSOR(ctx, s.config.SORBaseURL, u.ChainId, tenantID)
	if err != nil {
		return utils.UnexpectedFailureResponse(c, err.Error(), nil)
	}
	chainId, err := client.ChainID(ctx)
	if err != nil {
		s.e.Logger.Errorf(err.Error())
	}
	rawDecodedText, err := base64.StdEncoding.DecodeString(u.ABI)
	if err != nil {
		s.e.Logger.Errorf(err.Error())
		return utils.UnexpectedFailureResponse(c, "ABI string error "+err.Error(), nil)
	}
	nonce, err := client.PendingNonceAt(ctx, common.HexToAddress(wallet.Address))
	if err != nil {
		return utils.UnexpectedFailureResponse(c, err.Error(), nil)
	}
	transactOpts, err := TransactionOptionsWithKMSSigning(ctx, wallet, s.vault, chainId)
	if err != nil {
		s.e.Logger.Errorf(err.Error())
		return utils.UnexpectedFailureResponse(c, "Error while fetching txn opts "+err.Error(), nil)
	}
	transactOpts.Nonce = big.NewInt(int64(nonce))
	tipCap, _ := client.SuggestGasTipCap(ctx)
	feeCap, _ := client.SuggestGasPrice(ctx)
	transactOpts.GasFeeCap = feeCap
	transactOpts.GasTipCap = tipCap
	contractMeta := ContractMetadata{
		ABI:     string(rawDecodedText),
		Bin:     u.ByteCode,
		ChainId: chainId.String(),
	}
	params, err := utils.ConvertParamsAsPerTypes(u.Params)
	if err != nil {
		return utils.BadRequestResponse(c, "error converting params "+err.Error(), nil)
	}
	contractMeta.Params = params
	address, txn, _, err := DeploySmartContract(transactOpts, client, contractMeta)
	if err != nil {
		s.e.Logger.Errorf(err.Error())
		return utils.UnexpectedFailureResponse(c, "error deploying contract "+err.Error(), nil)
	}
	return utils.SendSuccessResponse(c, "Signed and executed txn successfully", &utils.DeployContractResponse{TxnHash: txn.Hash().String(), ContractAddress: address.String()})
}

func (s *Service) estimateGas(c echo.Context) error {
	ctx := c.Request().Context()
	tenantID := kmsmiddleware.TenantIDFrom(c)

	u := new(utils.EstimateGasRequest)
	if err := c.Bind(u); err != nil {
		return utils.BadRequestResponse(c, "bad request", nil)
	}
	if u.ChainId == "" {
		return utils.BadRequestResponse(c, "chainId is required", nil)
	}

	walletId, err := uuid.Parse(u.WalletId)
	if err != nil {
		return utils.UnexpectedFailureResponse(c, err.Error(), nil)
	}
	walletBytes, err := s.db.Get([]byte(utils.NAMESPACE), walletId.NodeID())
	if err != nil {
		return utils.UnauthorizedResponse(c, err.Error(), nil)
	}
	wallet := &Wallet{}
	if err := json.Unmarshal(walletBytes, wallet); err != nil {
		return utils.UnexpectedFailureResponse(c, err.Error(), nil)
	}

	client, err := utils.GetEthereumClientViaSOR(ctx, s.config.SORBaseURL, u.ChainId, tenantID)
	if err != nil {
		return utils.UnexpectedFailureResponse(c, err.Error(), nil)
	}
	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		return utils.UnexpectedFailureResponse(c, "error calculating gas price : "+err.Error(), nil)
	}
	var to common.Address
	if u.To != "" {
		to = common.HexToAddress(u.To)
	}
	var data []byte
	if u.IsContractTxn {
		params, err := utils.ConvertParamsAsPerTypes(u.Params)
		if err != nil {
			s.e.Logger.Errorf(err.Error())
			return utils.BadRequestResponse(c, "error converting params : "+err.Error(), nil)
		}
		contractABI, err := abi.JSON(strings.NewReader(u.ContractABI))
		if err != nil {
			s.e.Logger.Errorf(err.Error())
			return utils.UnexpectedFailureResponse(c, "error reading ABI : "+err.Error(), nil)
		}
		data, err = contractABI.Pack(u.Method, params...)
		if err != nil {
			s.e.Logger.Errorf(err.Error())
			return utils.UnexpectedFailureResponse(c, "error packing ABI : "+err.Error(), nil)
		}
		if u.ByteCode != "" {
			data = append(common.FromHex(u.ByteCode), data...)
		}
	} else {
		if u.Data != "" {
			data, err = base64.RawStdEncoding.DecodeString(u.Data)
			if err != nil {
				s.e.Logger.Errorf(err.Error())
				return utils.UnexpectedFailureResponse(c, "error decoding data : "+err.Error(), nil)
			}
		}
	}
	estimatedGas, err := client.EstimateGas(ctx, ethereum.CallMsg{From: common.HexToAddress(wallet.Address), To: &to, Value: big.NewInt(u.Value), Data: data, GasPrice: gasPrice})
	if err != nil {
		s.e.Logger.Errorf(err.Error())
		return utils.UnexpectedFailureResponse(c, "error estimating gas : "+err.Error(), nil)
	}
	return utils.SendSuccessResponse(c, "", &utils.EstimatedGasResponse{Address: wallet.Address, EstimatedGas: estimatedGas})
}

func (s *Service) getBalance(c echo.Context) error {
	ctx := c.Request().Context()
	tenantID := kmsmiddleware.TenantIDFrom(c)

	u := new(utils.WalletBalanceRequest)
	if err := c.Bind(u); err != nil {
		return utils.BadRequestResponse(c, "bad request", nil)
	}
	if u.ChainId == "" {
		return utils.BadRequestResponse(c, "chainId is required", nil)
	}

	walletId, err := uuid.Parse(u.WalletId)
	if err != nil {
		return utils.UnexpectedFailureResponse(c, err.Error(), nil)
	}
	walletBytes, err := s.db.Get([]byte(utils.NAMESPACE), walletId.NodeID())
	if err != nil {
		return utils.UnauthorizedResponse(c, err.Error(), nil)
	}
	wallet := &Wallet{}
	if err := json.Unmarshal(walletBytes, wallet); err != nil {
		return utils.UnexpectedFailureResponse(c, err.Error(), nil)
	}

	client, err := utils.GetEthereumClientViaSOR(ctx, s.config.SORBaseURL, u.ChainId, tenantID)
	if err != nil {
		return utils.UnexpectedFailureResponse(c, err.Error(), nil)
	}
	balance, err := client.BalanceAt(ctx, common.HexToAddress(wallet.Address), nil)
	if err != nil {
		return utils.UnexpectedFailureResponse(c, err.Error(), nil)
	}
	return utils.SendSuccessResponse(c, "", &utils.BalanceResponse{
		Address: wallet.Address,
		Balance: balance.Uint64(),
	})
}

func (s *Service) callContract(c echo.Context) error {
	ctx := c.Request().Context()
	tenantID := kmsmiddleware.TenantIDFrom(c)

	u := new(utils.CallContractRequest)
	if err := c.Bind(u); err != nil {
		return utils.BadRequestResponse(c, "bad request", nil)
	}
	if u.To == "" {
		return utils.BadRequestResponse(c, "mandatory params missing", nil)
	}
	if u.ChainId == "" {
		return utils.BadRequestResponse(c, "chainId is required", nil)
	}

	walletId, err := uuid.Parse(u.WalletId)
	if err != nil {
		return utils.UnexpectedFailureResponse(c, err.Error(), nil)
	}
	walletBytes, err := s.db.Get([]byte(utils.NAMESPACE), walletId.NodeID())
	if err != nil {
		return utils.UnauthorizedResponse(c, err.Error(), nil)
	}
	wallet := &Wallet{}
	if err := json.Unmarshal(walletBytes, wallet); err != nil {
		return utils.UnexpectedFailureResponse(c, err.Error(), nil)
	}

	client, err := utils.GetEthereumClientViaSOR(ctx, s.config.SORBaseURL, u.ChainId, tenantID)
	if err != nil {
		return utils.UnexpectedFailureResponse(c, err.Error(), nil)
	}
	chainId, err := client.ChainID(ctx)
	if err != nil {
		return utils.UnexpectedFailureResponse(c, "error getting chain Id : "+err.Error(), nil)
	}
	to := common.HexToAddress(u.To)
	contract, err := NewSmartContract(to, u.ContractABI, client)
	if err != nil {
		return utils.UnexpectedFailureResponse(c, "error initializing contract : "+err.Error(), nil)
	}
	txnOpts, err := TransactionOptionsWithKMSSigning(ctx, wallet, s.vault, chainId)
	if err != nil {
		return utils.UnexpectedFailureResponse(c, "error initializing transactor opts : "+err.Error(), nil)
	}
	if u.Value > 0 {
		txnOpts.Value = big.NewInt(u.Value)
	}
	callOpts := &bind.CallOpts{From: txnOpts.From}
	params, err := utils.ConvertParamsAsPerTypes(u.Params)
	if err != nil {
		return utils.BadRequestResponse(c, err.Error(), nil)
	}
	var response []interface{}
	if err := contract.SmartContractCaller.Contract.Call(callOpts, &response, u.Method, params...); err != nil {
		return utils.UnexpectedFailureResponse(c, "error calling contract : "+err.Error(), nil)
	}
	return utils.SendSuccessResponse(c, "", &utils.CallContractResponse{Response: &response})
}

func (s *Service) signEIP712Txn(c echo.Context) error {
	u := new(utils.EIP712SignRequest)
	if err := c.Bind(u); err != nil {
		return utils.BadRequestResponse(c, "bad request", nil)
	}
	var data apitypes.TypedData
	if err := json.Unmarshal([]byte(u.Data), &data); err != nil {
		return utils.BadRequestResponse(c, "error unmarshalling type data : "+err.Error(), nil)
	}
	walletId, err := uuid.Parse(u.WalletId)
	if err != nil {
		return utils.UnexpectedFailureResponse(c, err.Error(), nil)
	}
	walletBytes, err := s.db.Get([]byte(utils.NAMESPACE), walletId.NodeID())
	if err != nil {
		return utils.UnauthorizedResponse(c, err.Error(), nil)
	}
	wallet := &Wallet{}
	if err := json.Unmarshal(walletBytes, wallet); err != nil {
		return utils.UnexpectedFailureResponse(c, err.Error(), nil)
	}
	signature, err := s.vault.GetEIP712Signature(data, wallet.vaultKey())
	if err != nil {
		return utils.UnexpectedFailureResponse(c, "error getting signature : "+err.Error(), nil)
	}
	return utils.SendSuccessResponse(c, "Data signed successfully", signature)
}

func (s *Service) signMessage(c echo.Context) error {
	ctx := c.Request().Context()
	u := new(utils.SignMsgRequest)
	if err := c.Bind(u); err != nil {
		return utils.BadRequestResponse(c, "bad request", nil)
	}
	if u.Message == "" {
		return utils.BadRequestResponse(c, "mandatory params missing", nil)
	}
	walletId, err := uuid.Parse(u.WalletId)
	if err != nil {
		return utils.UnexpectedFailureResponse(c, err.Error(), nil)
	}
	walletBytes, err := s.db.Get([]byte(utils.NAMESPACE), walletId.NodeID())
	if err != nil {
		return utils.UnauthorizedResponse(c, err.Error(), nil)
	}
	wallet := &Wallet{}
	if err := json.Unmarshal(walletBytes, wallet); err != nil {
		return utils.UnexpectedFailureResponse(c, err.Error(), nil)
	}
	hash := crypto.Keccak256Hash([]byte(u.Message))
	signature, err := SignTransactionHash(ctx, wallet, s.vault, hash.Bytes())
	if err != nil {
		return utils.UnexpectedFailureResponse(c, "error signing txn : "+err.Error(), nil)
	}
	return utils.SendSuccessResponse(c, "Signed message successfully", "0x"+hex.EncodeToString(signature))
}

func (s *Service) verifySignatureOffChain(c echo.Context) error {
	u := new(utils.VerifyMsgRequest)
	if err := c.Bind(u); err != nil {
		return utils.BadRequestResponse(c, "bad request", nil)
	}
	if u.Message == "" || u.Signature == "" {
		return utils.BadRequestResponse(c, "mandatory params missing", nil)
	}
	walletId, err := uuid.Parse(u.WalletId)
	if err != nil {
		return utils.UnexpectedFailureResponse(c, err.Error(), nil)
	}
	walletBytes, err := s.db.Get([]byte(utils.NAMESPACE), walletId.NodeID())
	if err != nil {
		return utils.UnauthorizedResponse(c, err.Error(), nil)
	}
	wallet := &Wallet{}
	if err := json.Unmarshal(walletBytes, wallet); err != nil {
		return utils.UnexpectedFailureResponse(c, err.Error(), nil)
	}
	hash := crypto.Keccak256Hash([]byte(u.Message))
	signature, err := hex.DecodeString(strings.Replace(u.Signature, "0x", "", 1))
	if err != nil {
		return utils.UnexpectedFailureResponse(c, "error decoding signature : "+err.Error(), nil)
	}
	pubKey, err := crypto.SigToPub(hash.Bytes(), signature)
	if err != nil {
		return utils.UnexpectedFailureResponse(c, "error getting public key from signature : "+err.Error(), nil)
	}
	if pubKey.X == nil || pubKey.Y == nil {
		return utils.UnexpectedFailureResponse(c, "error getting public key", nil)
	}
	address := crypto.PubkeyToAddress(*pubKey).Hex()
	return utils.SendSuccessResponse(c, "", &utils.VerifyMsgResponse{IsVerified: wallet.Address == address})
}

func (s *Service) signAndSubmitGaslessTransaction(c echo.Context) error {
	ctx := c.Request().Context()
	tenantID := kmsmiddleware.TenantIDFrom(c)

	u := new(utils.SignAndSubmitGSNTxnRequest)
	if err := c.Bind(u); err != nil {
		return utils.BadRequestResponse(c, "bad request", nil)
	}

	walletId, err := uuid.Parse(u.WalletId)
	if err != nil {
		return utils.UnexpectedFailureResponse(c, err.Error(), nil)
	}
	walletBytes, err := s.db.Get([]byte(utils.NAMESPACE), walletId.NodeID())
	if err != nil {
		return utils.UnauthorizedResponse(c, err.Error(), nil)
	}
	wallet := &Wallet{}
	if err := json.Unmarshal(walletBytes, wallet); err != nil {
		return utils.UnexpectedFailureResponse(c, err.Error(), nil)
	}

	chainIdStr := strconv.FormatUint(u.ChainID, 10)
	client, err := utils.GetEthereumClientViaSOR(ctx, s.config.SORBaseURL, chainIdStr, tenantID)
	if err != nil {
		return utils.UnexpectedFailureResponse(c, err.Error(), nil)
	}
	chainId, err := client.ChainID(ctx)
	if err != nil {
		s.e.Logger.Errorf(err.Error())
	}

	req := utils.GSNTxnPayloadRequest{
		ChainId:         chainId.Uint64(),
		DAppId:          u.DAppId,
		UserAddress:     wallet.Address,
		ContractAddress: u.To,
		ContractAbi:     u.ContractABI,
		Method:          u.Method,
		Args:            u.Params,
	}
	var txnPayload map[string]interface{}
	if err := utils.HttpCallWithContextHeaderJson(c, "POST", s.config.ProxyUrl+utils.GetGaslessPayloadEndpoint, req, &txnPayload); err != nil {
		return utils.UnexpectedFailureResponse(c, "error getting txn payload : "+err.Error(), nil)
	} else if txnPayload["dataToSign"] == nil || txnPayload["request"] == nil {
		return utils.UnexpectedFailureResponse(c, "Failed to get transaction payload", txnPayload)
	}

	jsonString, err := json.Marshal(txnPayload["dataToSign"])
	if err != nil {
		return utils.BadRequestResponse(c, "error unmarshalling data to sign : "+err.Error(), nil)
	}
	var typedData apitypes.TypedData
	if err = json.Unmarshal(jsonString, &typedData); err != nil {
		return utils.BadRequestResponse(c, "error unmarshalling type data : "+err.Error(), nil)
	}

	signature, err := getEIP712Signature(ctx, wallet, s.vault, typedData)
	if err != nil {
		return utils.UnexpectedFailureResponse(c, "error getting signature : "+err.Error(), nil)
	}

	sendTxRequest := &utils.GSNSendTxnRequest{
		ChainId:         chainId.Uint64(),
		UserAddress:     wallet.Address,
		DAppId:          u.DAppId,
		ContractAddress: u.To,
		Method:          u.Method,
		Request:         txnPayload["request"],
		Signature:       signature,
	}
	var sendTxResponse map[string]interface{}
	if err := utils.HttpCallWithContextHeaderJson(c, "POST", s.config.ProxyUrl+utils.SendGaslessTransactionEndpoint, sendTxRequest, &sendTxResponse); err != nil {
		return utils.UnexpectedFailureResponse(c, "error sending gasless txn : "+err.Error(), nil)
	} else if sendTxResponse["txHash"] == nil {
		return utils.UnexpectedFailureResponse(c, "Failed To Send Transaction: ", sendTxResponse)
	}

	return utils.SendSuccessResponse(c, "Transaction Complete", sendTxResponse)
}

func getEIP712Signature(ctx context.Context, w *Wallet, vault vault.Vault, data apitypes.TypedData) (string, error) {
	typedDataHash, err := data.HashStruct(data.PrimaryType, data.Message)
	if err != nil {
		return "", err
	}
	domainSeparator, err := data.HashStruct("EIP712Domain", data.Domain.Map())
	if err != nil {
		return "", err
	}

	rawData := []byte(fmt.Sprintf("\x19\x01%s%s", string(domainSeparator), string(typedDataHash)))
	challengeHash := crypto.Keccak256Hash(rawData)
	signature, err := SignTransactionHash(ctx, w, vault, challengeHash.Bytes())
	if err != nil {
		return "", err
	}
	if len(signature) != 65 {
		return "", fmt.Errorf("invalid signature length: %d", len(signature))
	}
	if signature[64] == 0 {
		signature[64] = 27
	}
	if signature[64] == 1 {
		signature[64] = 28
	}
	if signature[64] != 27 && signature[64] != 28 {
		return "", fmt.Errorf("invalid recovery id: %d", signature[64])
	}
	return "0x" + hex.EncodeToString(signature), nil
}
