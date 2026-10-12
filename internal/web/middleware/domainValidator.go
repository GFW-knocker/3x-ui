// Package middleware provides HTTP middleware functions for the 3x-ui web panel,
// including domain validation utilities.
package middleware

import (
	"net"
	"strings"

	"github.com/gin-gonic/gin"
)

// DomainValidatorMiddleware returns a Gin middleware that validates the request domain.
// It extracts the host from the request, strips any port number, and compares it
// against the configured domain. Requests for any other host get the generic
// nginx-style 404, so a scan by IP learns nothing about the panel.
func DomainValidatorMiddleware(domain string) gin.HandlerFunc {
	return func(c *gin.Context) {
		host := c.Request.Host
		if colonIndex := strings.LastIndex(host, ":"); colonIndex != -1 {
			host, _, _ = net.SplitHostPort(c.Request.Host)
		}

		if host != domain {
			NotFound(c)
			return
		}

		c.Next()
	}
}
