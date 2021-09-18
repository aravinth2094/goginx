package config

import (
	"encoding/json"
	"io/ioutil"

	"github.com/aravinth2094/goginx/types"
)

func ParseConfig(configFileLocation string) (types.Configuration, error) {
	conf := types.Configuration{
		Listen: ":80",
		Log:    "goginx.log",
	}
	file, err := ioutil.ReadFile(configFileLocation)
	if err != nil {
		return types.Configuration{}, err
	}
	err = json.Unmarshal([]byte(file), &conf)
	if err != nil {
		return types.Configuration{}, err
	}
	return conf, nil
}
