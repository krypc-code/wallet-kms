package kms

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Service struct {
	e *echo.Echo
}

type ResponseBody struct {
	Status  string
	Message string
	Data    interface{}
}

func initService() *Service {

	var serve Service
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	serve.e = e
	return &serve
}

func Run() {
	service := initService()
	g := service.e.Group("/wallet")

	//healthcheck
	g.GET("/health", healthCheck)

	//wallet API
	// g.POST("/createWallet", service.createWallet)

	service.e.Logger.Fatal(service.e.Start(":8888"))
}

func healthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": "Server is up and running",
	})
}
