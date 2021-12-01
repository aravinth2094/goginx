package handler

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func (route Route) GetCoreHandler(conf *Configuration, method string, discoveryService *DiscoveryService) gin.HandlerFunc {
	rr, _ := conf.getLoadBalancer(route)
	ds := func() *DiscoveryClient {
		client, _ := discoveryService.GetService(route.ForwardUrl[0:strings.Index(route.ForwardUrl, ":")])
		return client
	}
	next := func() (*DiscoveryClient, string) {
		if conf.Discovery {
			ds := ds()
			host, port := ds.Host, ds.Port
			return ds, net.JoinHostPort(host, strconv.Itoa(port))
		} else {
			return nil, rr.Next().Host
		}
	}
	return func(c *gin.Context) {
		body, err := ioutil.ReadAll(c.Request.Body)
		if checkAndSendError(c, err) {
			return
		}
		ds, host := next()
		url := strings.TrimRight(host, "/")
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
			route.addSecureHeaders(c)
		}
		route.addCorsHeaders(c)
		for h, val := range route.CustomHeaders {
			proxyReq.Header.Add(h, val)
		}
		resp, err := http.DefaultClient.Do(proxyReq)
		if checkAndSendError(c, err) {
			if ds != nil {
				discoveryService.MarkInactive(ds)
			}
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

func (conf Configuration) GetWhitelistHandler() gin.HandlerFunc {
	whiteList := make(map[string]bool)
	for _, ip := range conf.WhiteList {
		whiteList[ip] = true
	}
	return func(c *gin.Context) {
		for _, ip := range conf.WhiteList {
			if whiteList[c.ClientIP()] || cidrRangeContains(ip, c.ClientIP()) {
				return
			}
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": c.ClientIP() + " is not allowed"})
		}
	}
}

func (conf Configuration) GetLoggingHandler() gin.HandlerFunc {
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

func (conf Configuration) GetDiscoveryHandler() (gin.HandlerFunc, *DiscoveryService) {
	service := &DiscoveryService{}
	go service.HeartBeatServices()
	return func(c *gin.Context) {
		client := &DiscoveryClient{}
		if err := c.Bind(&client); err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
		if client.Service == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "service name is required"})
			return
		}
		if client.Host == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "service host is required"})
			return
		}
		if client.Port < 1 || client.Port > 65535 {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "service port is invalid"})
			return
		}
		client.Active = true
		service.AppendService(client)
	}, service
}
