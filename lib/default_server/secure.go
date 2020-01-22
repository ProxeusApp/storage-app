package default_server

import (
	"strings"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

var allowedSites = []string{
	"https://mercuriahelvetica.com/",
	"https://cif.unibas.ch",
}

// Secure uses modified security config, without XFrameOptions, for selected referrers
func Secure(next echo.HandlerFunc) echo.HandlerFunc {
	defaultSecure := middleware.Secure()(next)
	c := middleware.DefaultSecureConfig
	c.XFrameOptions = ""
	weakerSecure := middleware.SecureWithConfig(c)(next)

	return func(c echo.Context) error {
		referer := c.Request().Header.Get("Referer")
		for _, allowedSite := range allowedSites {
			if strings.HasPrefix(referer, allowedSite) {
				return weakerSecure(c)
			}
		}
		return defaultSecure(c)
	}
}
