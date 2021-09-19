package handler

import (
	"net"
	"net/http"

	"github.com/aravinth2094/goginx/types"
	"github.com/gin-gonic/gin"
)

func checkAndSendError(c *gin.Context, err error) bool {
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		c.Abort()
		return true
	}
	return false
}

func addSecureHeaders(c *gin.Context) {
	c.Writer.Header().Add("X-Frame-Options", "DENY")
	c.Writer.Header().Add("X-XSS-Protection", "1; mode=block")
	c.Writer.Header().Add("X-Content-Type-Options", "nosniff")
	c.Writer.Header().Add("Content-Security-Policy", "default-src 'self'")
	c.Writer.Header().Add("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
}

func addCorsHeaders(route types.Route, c *gin.Context) {
	if route.Cors.Origin != "" {
		c.Writer.Header().Add("Access-Control-Allow-Origin", route.Cors.Origin)
	}
	if route.Cors.Methods != "" {
		c.Writer.Header().Add("Access-Control-Allow-Methods", route.Cors.Methods)
	}
	if route.Cors.Credentials {
		c.Writer.Header().Add("Access-Control-Allow-Credentials", "true")
	}
	if route.Cors.MaxAge != "" {
		c.Writer.Header().Add("Access-Control-Max-Age", route.Cors.MaxAge)
	}
	if route.Cors.AllowedHeaders != "" {
		c.Writer.Header().Add("Access-Control-Allow-Headers", route.Cors.AllowedHeaders)
	}
	if route.Cors.CacheControl != "" {
		c.Writer.Header().Add("Access-Control-Allow-Cache", route.Cors.CacheControl)
	}
	if route.Cors.Vary != "" {
		c.Writer.Header().Add("Access-Control-Allow-Vary", route.Cors.Vary)
	}
}

func cidrRangeContains(cidrRange string, checkIP string) bool {
	_, network, err := net.ParseCIDR(cidrRange)
	if err != nil {
		return false
	}
	ip := net.ParseIP(checkIP)
	return network.Contains(ip)
}
