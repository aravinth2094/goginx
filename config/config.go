package config

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/aravinth2094/goginx/handler"
)

func readFileFromLocal(fileLocation string) ([]byte, error) {
	return ioutil.ReadFile(fileLocation)
}

func readFileFromUrl(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func readFile(fileLocation string) ([]byte, error) {
	if fileLocation == "" {
		return nil, nil
	}
	if fileLocation[0:4] == "http" {
		return readFileFromUrl(fileLocation)
	}
	return readFileFromLocal(fileLocation)
}

func ParseConfig(configFileLocation string) (*handler.Configuration, error) {
	conf := &handler.Configuration{
		Listen: ":80",
		Log:    "goginx.log",
	}
	file, err := readFile(configFileLocation)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal([]byte(file), &conf)
	if err != nil {
		return nil, err
	}
	return conf, nil
}
