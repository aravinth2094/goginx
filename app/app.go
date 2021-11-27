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
	validate := flag.Bool("V", false, "Validate configuration file")
	help := flag.Bool("h", false, "Print this help")
	flag.Parse()
	if *help {
		flag.Usage()
		os.Exit(0)
	}
	if *validate {
		conf, err := getConfigurationFromFile(*configFileLocation)
		if err != nil {
			log.Fatalln(err)
		}
		if err := conf.Validate(); err != nil {
			log.Fatalln(err)
		}
		os.Exit(0)
	}
	gin.SetMode(gin.ReleaseMode)
	gin.DisableConsoleColor()
	return *configFileLocation
}

func initLogFile(conf *handler.Configuration) error {
	logfile, err := os.OpenFile(conf.Log, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	gin.DefaultWriter = io.MultiWriter(logfile)
	return nil
}

func getConfigurationFromFile(configurationFile string) (*handler.Configuration, error) {
	conf, err := config.ParseConfig(configurationFile)
	if err != nil {
		return nil, err
	}
	return conf, nil
}

func StartWithConfig(conf *handler.Configuration) error {
	if err := conf.Validate(); err != nil {
		return err
	}
	err := initLogFile(conf)
	if err != nil {
		return err
	}
	r := gin.New()
	logger, _ := zap.NewProduction()
	r.Use(conf.GetLoggingHandler())
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
		r.Use(conf.GetWhitelistHandler())
	}
	var discoveryHandler gin.HandlerFunc
	var discoveryService *handler.DiscoveryService
	if conf.Discovery {
		discoveryHandler, discoveryService = conf.GetDiscoveryHandler()
		r.POST("/discovery", discoveryHandler)
	}
	r.HandleMethodNotAllowed = true
	var store *persistence.InMemoryStore

	for _, route := range conf.Routes {
		if route.ForwardUrl[:7] == "file://" {
			r.StaticFS(route.Path, http.Dir(route.ForwardUrl[7:]))
			continue
		}
		for _, method := range route.AllowedMethods {
			handlerFunction := route.GetCoreHandler(conf, method, discoveryService)
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
	conf, err := getConfigurationFromFile(initialize())
	if err != nil {
		return err
	}
	return StartWithConfig(conf)
}
