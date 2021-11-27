package handler

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
	Cache          int               `json:"cache"`
	Timeout        int               `json:"timeout"`
}

type Configuration struct {
	Listen      string              `json:"listen"`
	Certificate string              `json:"certificate"`
	Key         string              `json:"key"`
	Log         string              `json:"log"`
	WhiteList   []string            `json:"whiteList"`
	Compression bool                `json:"compression"`
	Upstreams   map[string][]string `json:"upstreams"`
	Routes      []Route             `json:"routes"`
	Discovery   bool                `json:"discovery"`
}

type DiscoveryClient struct {
	Service string `json:"service"`
	Host    string `json:"host"`
	Port    int    `json:"port"`
	Active  bool
}

type DiscoveryService struct {
	services            map[string][]DiscoveryClient
	serviceCurrentIndex map[string]int
}
