package kms

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

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
