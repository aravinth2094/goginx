package handler

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	roundrobin "github.com/hlts2/round-robin"
)

func checkAndSendError(c *gin.Context, err error) bool {
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		c.Abort()
		return true
	}
	return false
}

func (route Route) addSecureHeaders(c *gin.Context) {
	c.Writer.Header().Add("X-Frame-Options", "DENY")
	c.Writer.Header().Add("X-XSS-Protection", "1; mode=block")
	c.Writer.Header().Add("X-Content-Type-Options", "nosniff")
	c.Writer.Header().Add("Content-Security-Policy", "default-src 'self'")
	c.Writer.Header().Add("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
}

func (route Route) addCorsHeaders(c *gin.Context) {
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

func (conf *Configuration) getLoadBalancer(route Route) (roundrobin.RoundRobin, error) {
	urls := make([]*url.URL, 0)
	upstreams := conf.Upstreams[route.ForwardUrl[0:strings.Index(route.ForwardUrl, ":")]]
	if len(upstreams) == 0 {
		urls = append(urls, &url.URL{
			Host: route.ForwardUrl,
		})
	}
	for _, upstream := range upstreams {
		urls = append(urls, &url.URL{
			Host: upstream + route.ForwardUrl[strings.Index(route.ForwardUrl, ":")+1:],
		})
	}
	return roundrobin.New(urls...)
}

func cidrRangeContains(cidrRange string, checkIP string) bool {
	_, network, err := net.ParseCIDR(cidrRange)
	if err != nil {
		return false
	}
	ip := net.ParseIP(checkIP)
	return network.Contains(ip)
}

func (conf *Configuration) Validate() error {
	if conf.Listen == "" {
		return errors.New("listen address is not set")
	}
	host, port, err := net.SplitHostPort(conf.Listen)
	if err != nil {
		return err
	}
	if host != "" && net.ParseIP(host) == nil {
		if _, err := net.LookupHost(host); err != nil {
			return err
		}
	}
	if port == "0" || port == "" {
		return errors.New("port invalid")
	}
	if port == "80" && (conf.Certificate != "" || conf.Key != "") {
		log.Println("WARNING: You are attempting to run HTTPS server on port 80. Port 443 is recommended.")
	}
	if port == "443" && (conf.Certificate == "" || conf.Key == "") {
		log.Println("WARNING: You are attempting to run HTTP server on port 443. Port 80 is recommended.")
	}
	if conf.Log == "" {
		return errors.New("log file is not set")
	}
	if len(conf.Routes) == 0 {
		return errors.New("no routes are set")
	}
	for _, route := range conf.Routes {
		if route.ForwardUrl == "" || !strings.Contains(route.ForwardUrl, ":") {
			return fmt.Errorf("%s invalid forwardUrl", route.Path)
		}
		if len(route.AllowedMethods) == 0 && route.ForwardUrl[0:strings.Index(route.ForwardUrl, ":")] != "file" {
			return fmt.Errorf("%s must contain atleast one allowedMethod", route.Path)
		}
		if route.ForwardUrl[0:strings.Index(route.ForwardUrl, ":")] != "http" && route.ForwardUrl[0:strings.Index(route.ForwardUrl, ":")] != "file" && route.ForwardUrl[0:strings.Index(route.ForwardUrl, ":")] != "https" {
			if _, ok := conf.Upstreams[route.ForwardUrl[0:strings.Index(route.ForwardUrl, ":")]]; !ok {
				return fmt.Errorf("%s forwardUrl not in upstream", route.ForwardUrl)
			}
		}
	}
	return nil
}
