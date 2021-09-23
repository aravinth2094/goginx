package app

import (
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aravinth2094/goginx/config"
	"github.com/aravinth2094/goginx/handler"
	"github.com/aravinth2094/goginx/types"
	"github.com/gin-contrib/cache"
	"github.com/gin-contrib/cache/persistence"
	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/timeout"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/penglongli/gin-metrics/ginmetrics"
	"go.uber.org/zap"
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
	logger, _ := zap.NewProduction()
	r.Use(handler.GetLoggingHandler())
	r.Use(ginzap.Ginzap(logger, time.RFC3339, true))
	r.Use(ginzap.RecoveryWithZap(logger, true))
	m := ginmetrics.GetMonitor()
	m.SetMetricPath("/metrics")
	m.SetSlowTime(10)
	m.SetDuration([]float64{0.1, 0.3, 1.2, 5, 10})
	m.Use(r)
	if conf.Compression {
		r.Use(gzip.Gzip(gzip.DefaultCompression))
	}
	if len(conf.WhiteList) > 0 {
		r.Use(handler.GetWhitelistHandler(conf))
	}
	r.HandleMethodNotAllowed = true
	var store *persistence.InMemoryStore

	for _, route := range conf.Routes {
		if route.ForwardUrl[:7] == "file://" {
			r.StaticFS(route.Path, http.Dir(route.ForwardUrl[7:]))
			continue
		}
		for _, method := range route.AllowedMethods {
			handlerFunction := handler.GetCoreHandler(conf, route, method)
			if route.Cache > 0 {
				if store == nil {
					store = persistence.NewInMemoryStore(time.Minute)
				}
				handlerFunction = cache.CachePage(store, time.Duration(route.Cache)*time.Second, handlerFunction)
			}
			if route.Timeout > 0 {
				handlerFunction = timeout.New(
					timeout.WithTimeout(time.Duration(route.Timeout)*time.Millisecond),
					timeout.WithHandler(handlerFunction),
				)
			}
			r.Handle(method, route.Path, handlerFunction)
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
