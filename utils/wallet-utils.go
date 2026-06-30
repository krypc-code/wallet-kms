package utils

import (
	"github.com/labstack/echo/v4"
)

// HttpCallWithContextHeaderJson forwards the echo request context headers to an
// outbound HTTP call. Used exclusively by the gasless transaction relay path.
func HttpCallWithContextHeaderJson(c echo.Context, method, url string, body interface{}, restype interface{}) error {
	header := make(map[string]string)
	header["Authorization"] = c.Request().Header.Get("Authorization")
	header["requestId"] = c.Request().Header.Get("requestId")
	header["hopCount"] = c.Request().Header.Get("hopCount")
	header["Content-Type"] = "application/json"
	return HttpCall(method, url, body, restype, header)
}
