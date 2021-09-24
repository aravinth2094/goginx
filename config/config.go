package config

import (
	"encoding/json"
	"io/ioutil"

	"github.com/aravinth2094/goginx/handler"
)

func ParseConfig(configFileLocation string) (handler.Configuration, error) {
	conf := handler.Configuration{
		Listen: ":80",
		Log:    "goginx.log",
	}
	file, err := ioutil.ReadFile(configFileLocation)
	if err != nil {
		return handler.Configuration{}, err
	}
	err = json.Unmarshal([]byte(file), &conf)
	if err != nil {
		return handler.Configuration{}, err
	}
	return conf, nil
}
