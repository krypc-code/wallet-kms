package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

const (
	CtxTenantID  = "tenant_id"
	CtxEntityID  = "entity_id"
	CtxRequestID = "request_id"
)

// TenantContext is an Echo middleware that enforces the X-Tenant-ID header
// forwarded by Kernel after JWT auth and entitlement checks.
// Returns 401 if the header is absent.
func TenantContext() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			tenantID := c.Request().Header.Get("X-Tenant-ID")
			if tenantID == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "X-Tenant-ID header is required",
				})
			}
			c.Set(CtxTenantID, tenantID)
			c.Set(CtxEntityID, c.Request().Header.Get("X-Entity-ID"))
			c.Set(CtxRequestID, c.Request().Header.Get("X-Request-ID"))
			return next(c)
		}
	}
}

// TenantIDFrom returns the tenant ID stored in the Echo context by TenantContext middleware.
func TenantIDFrom(c echo.Context) string {
	v, _ := c.Get(CtxTenantID).(string)
	return v
}
