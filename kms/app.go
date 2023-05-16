package kms

import (
	"context"
	"crypto/ecdsa"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Service struct {
	e        *echo.Echo
	db       DB
	password string
}

func initService() *Service {

	var serve Service
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	db, err := NewBadgerDB(".wallet/db/")
	if err != nil {
		log.Panic("error initializing wallet db")
	}
	serve.db = db
	serve.password = os.Getenv("WALLET_PASSWORD")
	serve.e = e

	go serve.FetchAndExecuteTxn()

	return &serve
}

func Run() {
	service := initService()
	g := service.e.Group("/wallet")

	//healthcheck
	g.GET("/health", healthCheck)

	//wallet API
	g.POST("/createWallet", service.createWallet)

	service.e.Logger.Fatal(service.e.Start(":8888"))
}

func healthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": "Server is up and running",
	})
}

func (serve *Service) createWallet(c echo.Context) error {
	request := new(WalletRequest)
	if err := c.Bind(request); err != nil {
		return c.JSON(http.StatusInternalServerError, ResponseBody{Status: "FAILURE", Message: "bad request", Data: nil})
	}
	key, err := generateKey(serve.password, request.Algorithm)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ResponseBody{Status: "FAILURE", Message: err.Error(), Data: nil})
	}
	dbKey := uuid.New()
	data, err := json.Marshal(key)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ResponseBody{Status: "FAILURE", Message: err.Error(), Data: nil})
	}
	if err := serve.db.Set([]byte(NAMESPACE), dbKey.NodeID(), data); err != nil {
		return c.JSON(http.StatusInternalServerError, ResponseBody{Status: "FAILURE", Message: err.Error(), Data: nil})
	}
	RespBody := ResponseBody{
		Status:  "SUCCESS",
		Message: "wallet created successfully",
		Data: WalletResponse{
			WalletId: dbKey.String(),
		},
	}
	return c.JSON(http.StatusOK, RespBody)
}

func (serve *Service) FetchAndExecuteTxn() {
	for {
		log.Println("running cron job")
		time.Sleep(10 * time.Minute)
		var pendingTxns []TransactionResponse

		// TODO API call to fetch pending txns from platform

		for _, txn := range pendingTxns {
			if err := serve.SignAndSubmitTxn(&txn); err != nil {
				log.Println("error submitting txn : " + err.Error())
			}
		}
	}
}

func (serve *Service) SignAndSubmitTxn(tx *TransactionResponse) error {
	ctx := context.Background()
	txnBytes, err := base64.StdEncoding.DecodeString(tx.Transaction)
	if err != nil {
		return err
	}
	var txn types.Transaction
	if err := txn.UnmarshalBinary(txnBytes); err != nil {
		return err
	}
	txHash := txn.Hash().Bytes()
	key, err := uuid.Parse(tx.WalletId)
	if err != nil {
		return err
	}
	val, err := serve.db.Get([]byte(NAMESPACE), key.NodeID())
	if err != nil {
		return err
	}
	wallet := &Key{}
	if err := json.Unmarshal(val, wallet); err != nil {
		return err
	}
	privKey, err := getDecodedPrivateKey(serve.password, []byte(wallet.PrivateKey))
	if err != nil {
		return err
	}
	var signature []byte
	switch wallet.Algorithm {
	case "secp256k1":
		signature, err = secp256k1.Sign(txHash, math.PaddedBigBytes(privKey.(*ecdsa.PrivateKey).D, 32))
		if err != nil {
			return err
		}
	case "ed25519":
		signature = ed25519.Sign(privKey.(ed25519.PrivateKey), txHash)
	}
	client, err := ethclient.DialContext(ctx, tx.ProxyUrl)
	if err != nil {
		return err
	}
	chainId, err := client.ChainID(ctx)
	if err != nil {
		return err
	}
	signer := types.LatestSignerForChainID(chainId)
	signedTxn, err := txn.WithSignature(signer, signature)
	if err != nil {
		return err
	}
	if err := client.SendTransaction(ctx, signedTxn); err != nil {
		return err
	}
	log.Println("Successfully submitted txn with Id : " + signedTxn.Hash().String())
	return nil
}
