package app

import (
	"flag"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/aravinth2094/goginx/config"
	"github.com/aravinth2094/goginx/handler"
	"github.com/aravinth2094/goginx/types"
	"github.com/gin-gonic/gin"
)

func initialize() string {
	configFileLocation := flag.String("c", "goginx.json", "Goginx configuration file location")
	help := flag.Bool("h", false, "Print this help")
	flag.Parse()
	if *help {
		flag.Usage()
		os.Exit(0)
	}
	gin.SetMode(gin.ReleaseMode)
	gin.DisableConsoleColor()
	return *configFileLocation
}

func initLogFile(conf types.Configuration) {
	logfile, err := os.OpenFile(conf.Log, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	gin.DefaultWriter = io.MultiWriter(logfile)
}

func getConfigurationFromFile(configurationFile string) types.Configuration {
	conf, err := config.ParseConfig(configurationFile)
	if err != nil {
		log.Fatalf("Error parsing configuration file: %s", err)
	}
	return conf
}

func StartWithConfig(conf types.Configuration) error {
	initLogFile(conf)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(handler.GetLoggingHandler())
	if len(conf.WhiteList) > 0 {
		r.Use(handler.GetWhitelistHandler(conf))
	}
	r.HandleMethodNotAllowed = true

	for _, route := range conf.Routes {
		if route.ForwardUrl[:7] == "file://" {
			r.StaticFS(route.Path, http.Dir(route.ForwardUrl[7:]))
			continue
		}
		for _, method := range route.AllowedMethods {
			r.Handle(method, route.Path, handler.GetCoreHandler(route, method))
		}
	}

	if conf.Certificate != "" && conf.Key != "" {
		return r.RunTLS(conf.Listen, conf.Certificate, conf.Key)
	}
	return r.Run(conf.Listen)
}

func Start() error {
	return StartWithConfig(getConfigurationFromFile(initialize()))
}
