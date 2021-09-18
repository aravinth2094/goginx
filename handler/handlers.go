package handler

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/aravinth2094/goginx/types"
	"github.com/gin-gonic/gin"
)

func GetCoreHandler(route types.Route, method string) gin.HandlerFunc {
	return func(c *gin.Context) {
		body, err := ioutil.ReadAll(c.Request.Body)
		if checkAndSendError(c, err) {
			return
		}
		url := strings.TrimRight(route.ForwardUrl, "/")
		if route.AppendPath {
			url += c.Request.URL.Path
		}
		proxyReq, err := http.NewRequest(method, url+"?"+c.Request.URL.RawQuery, bytes.NewReader(body))
		if checkAndSendError(c, err) {
			return
		}
		if route.ForwardIp {
			proxyReq.Header.Add("X-Forwarded-For", c.ClientIP())
		}
		proxyReq.Header = make(http.Header)
		for h, val := range c.Request.Header {
			proxyReq.Header.Add(h, val[0])
		}
		if route.SecureHeaders {
			addSecureHeaders(c)
		}
		addCorsHeaders(route, c)
		for h, val := range route.CustomHeaders {
			proxyReq.Header.Add(h, val)
		}
		resp, err := http.DefaultClient.Do(proxyReq)
		if checkAndSendError(c, err) {
			return
		}

		respHeaders := make(map[string]string)
		for h, vals := range resp.Header {
			respHeaders[h] = vals[0]
		}
		defer resp.Body.Close()
		c.DataFromReader(resp.StatusCode, resp.ContentLength, resp.Header.Get("Content-Type"), resp.Body, respHeaders)
		for _, cookie := range resp.Cookies() {
			c.Writer.Header().Add("Set-Cookie", cookie.String())
		}
	}
}

func GetWhitelistHandler(conf types.Configuration) gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, ip := range conf.WhiteList {
			if !cidrRangeContains(ip, c.ClientIP()) {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": c.ClientIP() + " is not allowed"})
				return
			}
		}
	}
}

func GetLoggingHandler() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {

		return fmt.Sprintf("%s - [%s] %s %s %s %d %s \"%s\" <%s>\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})
}
