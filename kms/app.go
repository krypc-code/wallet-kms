package kms

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"math/big"
	"net/http"
	"os"
	"wallet-kms/store"
	"wallet-kms/utils"
	"wallet-kms/vault"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
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

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	db, err := store.NewBadgerDB(".wallet/db/")
	if err != nil {
		log.Panic("error initializing wallet db")
	}
	vault, err := vault.NewHashiCorpVault(os.Getenv("VAULT_URL"), os.Getenv("VAULT_TOKEN"))
	if err != nil {
		log.Panic("error initializing vault : ", err.Error())
	}
	serve.db = db
	serve.vault = vault
	serve.config = &utils.Config{
		AuthToken:      os.Getenv("AUTH_TOKEN"),
		ProxyUrl:       os.Getenv("PROXY_URL"),
		Endpoint:       os.Getenv("ENDPOINT"),
		InstanceId:     os.Getenv("WALLET_INSTANCE_ID"),
		SubscriptionId: os.Getenv("SUBSCRIPTION_ID"),
	}
	serve.e = e
	go serve.ScheduleService()
	return &serve
}

func Run() {
	service := initService()
	g := service.e.Group("/wallet")

	//healthcheck
	g.GET("/health", healthCheck)

	//wallet API
	g.POST("/createWallet", service.createWallet)
	g.POST("/submitTransaction", service.submitTransaction)
	g.POST("/deployContract", service.deployContract)

	service.e.Logger.Fatal(service.e.Start(":8889"))
}

func healthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": "Server is up and running",
	})
}

func (serve *Service) createWallet(c echo.Context) error {
	ctx := c.Request().Context()
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
	if err := utils.AddWalletToPlatform(serve.config, &utils.AddWalletRequest{
		WalletId:  wallet.WalletId,
		Address:   wallet.Address,
		Name:      wallet.Name,
		Algorithm: wallet.Algorithm,
	}); err != nil {
		return utils.UnexpectedFailureResponse(c, err.Error(), nil)
	}
	return utils.SendSuccessResponse(c, "wallet created successfully", utils.WalletResponse{
		WalletId: wallet.WalletId,
	})
}

func (s *Service) submitTransaction(c echo.Context) error {
	ctx := c.Request().Context()
	u := new(utils.SignAndSubmitTxn)
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
	var to common.Address
	if u.To != "" {
		to = common.HexToAddress(u.To)
	}
	client, err := utils.GetEthereumClient(ctx, s.config)
	if err != nil {
		return utils.UnexpectedFailureResponse(c, err.Error(), nil)
	}
	chainId, err := client.ChainID(ctx)
	if err != nil {
		s.e.Logger.Errorf(err.Error())
	}
	nonce, err := utils.GetNonceFromPlatform(s.config, &utils.NonceRequest{WalletId: wallet.WalletId, ChainId: chainId.String()})
	if err != nil {
		return utils.UnexpectedFailureResponse(c, err.Error(), nil)
	}
	var txnHash string
	if u.IsContractTxn {
		var contractABI string
		if u.ContractABI != "" {
			contractABI = u.ContractABI
		} else {
			return utils.UnexpectedFailureResponse(c, "missing contract abi", nil)
		}
		contract, err := NewSmartContract(to, contractABI, client)
		if err != nil {
			return utils.UnexpectedFailureResponse(c, "error initializing contract : "+err.Error(), nil)
		}
		txnOpts, err := wallet.TransactionOptionsWithKMSSigning(ctx, s.vault, chainId)
		if err != nil {
			return utils.UnexpectedFailureResponse(c, "error initializing transactor opts : "+err.Error(), nil)
		}
		if u.Value > 0 {
			txnOpts.Value = big.NewInt(u.Value)
		}
		txnOpts.Nonce = big.NewInt(int64(nonce))
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
		signature, err := wallet.SignTransactionHash(ctx, s.vault, txn.Hash().Bytes())
		txn, err = txn.WithSignature(signer, signature)
		if err != nil {
			return utils.UnexpectedFailureResponse(c, err.Error(), nil)
		}
		if err := client.SendTransaction(ctx, txn); err != nil {
			return utils.UnexpectedFailureResponse(c, "error executing txn : "+err.Error(), nil)
		}
		txnHash = txn.Hash().String()
	}
	err = utils.UpdatePlatformNonce(s.config, &utils.NonceRequest{WalletId: wallet.WalletId, ChainId: chainId.String()})
	if err != nil {
		return utils.UnexpectedFailureResponse(c, err.Error(), nil)
	}
	return utils.SendSuccessResponse(c, "Signed and executed txn successfully", &utils.DeployContractResponse{TxnHash: txnHash})
}

func (s *Service) deployContract(c echo.Context) error {
	ctx := c.Request().Context()
	u := new(utils.DeployContractRequest)
	if err := c.Bind(u); err != nil {
		s.e.Logger.Errorf(err.Error())
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
	client, err := utils.GetEthereumClient(ctx, s.config)
	if err != nil {
		return utils.UnexpectedFailureResponse(c, err.Error(), nil)
	}
	chainId, err := client.ChainID(ctx)
	if err != nil {
		s.e.Logger.Errorf(err.Error())
	}
	var txnHash string
	rawDecodedText, err := base64.StdEncoding.DecodeString(u.ABI)
	if err != nil {
		s.e.Logger.Errorf(err.Error())
		return utils.UnexpectedFailureResponse(c, "ABI string error "+err.Error(), nil)
	}
	nonce, err := utils.GetNonceFromPlatform(s.config, &utils.NonceRequest{WalletId: wallet.WalletId, ChainId: chainId.String()})
	if err != nil {
		return utils.UnexpectedFailureResponse(c, err.Error(), nil)
	}
	transactOpts, err := wallet.TransactionOptionsWithKMSSigning(ctx, s.vault, chainId)
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
	err = utils.UpdatePlatformNonce(s.config, &utils.NonceRequest{WalletId: wallet.WalletId, ChainId: chainId.String()})
	if err != nil {
		return utils.UnexpectedFailureResponse(c, err.Error(), nil)
	}
	txnHash = txn.Hash().String()
	return utils.SendSuccessResponse(c, "Signed and executed txn successfully", &utils.DeployContractResponse{TxnHash: txnHash, ContractAddress: address.String()})
}
