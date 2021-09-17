package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type CorsConfig struct {
	Origin         string `json:"origin"`
	Methods        string `json:"methods"`
	Credentials    bool   `json:"credentials"`
	MaxAge         string `json:"maxAge"`
	AllowedHeaders string `json:"allowedHeaders"`
	CacheControl   string `json:"cacheControl"`
	Vary           string `json:"vary"`
}

type Route struct {
	Path           string            `json:"path"`
	ForwardUrl     string            `json:"forwardUrl"`
	AllowedMethods []string          `json:"allowedMethods"`
	ForwardIp      bool              `json:"forwardIp"`
	AppendPath     bool              `json:"appendPath"`
	CustomHeaders  map[string]string `json:"customHeaders"`
	SecureHeaders  bool              `json:"secureHeaders"`
	Cors           CorsConfig        `json:"cors"`
}

type Configuration struct {
	Listen      string   `json:"listen"`
	Certificate string   `json:"certificate"`
	Key         string   `json:"key"`
	Log         string   `json:"log"`
	WhiteList   []string `json:"whiteList"`
	Routes      []Route  `json:"routes"`
}

var conf = Configuration{
	Listen: ":80",
	Log:    "goginx.log",
}

func init() {
	configFileLocation := flag.String("c", "goginx.json", "Goginx configuration file location")
	help := flag.Bool("h", false, "Print this help")
	flag.Parse()
	if *help {
		flag.Usage()
		os.Exit(0)
	}
	file, err := ioutil.ReadFile(*configFileLocation)
	if err != nil {
		log.Fatalf("Error occured during unmarshaling. Error: %s", err.Error())
	}
	err = json.Unmarshal([]byte(file), &conf)
	if err != nil {
		log.Fatalf("Error occured during unmarshaling. Error: %s", err.Error())
	}
}

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

func addCorsHeaders(route Route, c *gin.Context) {
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

func handle(route Route, method string, c *gin.Context) {
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
	httpClient := http.Client{
		Timeout: time.Second * 10,
	}
	resp, err := httpClient.Do(proxyReq)
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

func main() {
	gin.SetMode(gin.ReleaseMode)
	gin.DisableConsoleColor()
	logfile, err := os.OpenFile(conf.Log, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	gin.DefaultWriter = io.MultiWriter(logfile)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {

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
	}))
	if len(conf.WhiteList) > 0 {
		r.Use(func(c *gin.Context) {
			for _, ip := range conf.WhiteList {
				if c.ClientIP() != ip {
					c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": c.ClientIP() + " is not allowed"})
					return
				}
			}
		})
	}
	r.HandleMethodNotAllowed = true

	for _, route := range conf.Routes {
		for _, method := range route.AllowedMethods {
			r.Handle(method, route.Path, func(c *gin.Context) {
				handle(route, method, c)
			})
		}
	}

	if conf.Certificate != "" && conf.Key != "" {
		log.Fatal(r.RunTLS(conf.Listen, conf.Certificate, conf.Key))
	}
	log.Fatal(r.Run(conf.Listen))
}
