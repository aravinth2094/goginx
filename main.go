package main

import (
	"flag"
	"io"
	"log"
	"os"

	"github.com/aravinth2094/goginx/config"
	"github.com/aravinth2094/goginx/handler"
	"github.com/aravinth2094/goginx/types"
	"github.com/gin-gonic/gin"
)

var conf types.Configuration

func init() {
	configFileLocation := flag.String("c", "goginx.json", "Goginx configuration file location")
	help := flag.Bool("h", false, "Print this help")
	flag.Parse()
	if *help {
		flag.Usage()
		os.Exit(0)
	}
	var err error
	conf, err = config.ParseConfig(*configFileLocation)
	if err != nil {
		log.Fatalf("Error parsing configuration file: %s", err)
	}
	gin.SetMode(gin.ReleaseMode)
	gin.DisableConsoleColor()
	logfile, err := os.OpenFile(conf.Log, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	gin.DefaultWriter = io.MultiWriter(logfile)
}

func main() {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(handler.GetLoggingHandler())
	if len(conf.WhiteList) > 0 {
		r.Use(handler.GetWhitelistHandler(conf))
	}
	r.HandleMethodNotAllowed = true

	for _, route := range conf.Routes {
		for _, method := range route.AllowedMethods {
			r.Handle(method, route.Path, handler.GetCoreHandler(route, method))
		}
	}

	if conf.Certificate != "" && conf.Key != "" {
		log.Fatal(r.RunTLS(conf.Listen, conf.Certificate, conf.Key))
	}
	log.Fatal(r.Run(conf.Listen))
}
