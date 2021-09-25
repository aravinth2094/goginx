package handler

import (
	"testing"
)

func TestValidateRegular(t *testing.T) {
	conf := &Configuration{
		Listen: ":80",
		Log:    "./log",
		Routes: []Route{
			{
				Path:           "/",
				AllowedMethods: []string{"GET"},
				ForwardUrl:     "http://localhost/",
			},
		},
	}
	if err := conf.Validate(); err != nil {
		t.Errorf("Validate error: %s", err.Error())
	}
}

func TestValidateUpstream(t *testing.T) {
	conf := &Configuration{
		Listen: ":80",
		Log:    "./log",
		Routes: []Route{
			{
				Path:           "/",
				AllowedMethods: []string{"GET"},
				ForwardUrl:     "local:/",
			},
		},
	}
	if err := conf.Validate(); err == nil {
		t.Errorf("Validate error: Upstream validation")
	}
}

func TestValidationNoRoutes(t *testing.T) {
	conf := &Configuration{
		Listen: ":80",
		Log:    "./log",
	}
	if err := conf.Validate(); err == nil {
		t.Errorf("Validate error: Route validation")
	}
}

func TestValidateForwardUrl(t *testing.T) {
	conf := &Configuration{
		Listen: ":80",
		Log:    "./log",
		Routes: []Route{
			{
				Path:           "/",
				AllowedMethods: []string{"GET"},
				ForwardUrl:     "",
			},
		},
	}
	if err := conf.Validate(); err == nil {
		t.Errorf("Validate error: Upstream validation")
	}
}

func TestValidateNoAllowedMethods(t *testing.T) {
	conf := &Configuration{
		Listen: ":80",
		Log:    "./log",
		Routes: []Route{
			{
				Path:       "/",
				ForwardUrl: "https://localhost",
			},
		},
	}
	if err := conf.Validate(); err == nil {
		t.Errorf("Validate error: Upstream validation")
	}
}

func TestValidateNoListen(t *testing.T) {
	conf := &Configuration{
		Listen: "",
		Log:    "./log",
		Routes: []Route{
			{
				Path:           "/",
				AllowedMethods: []string{"GET"},
				ForwardUrl:     "local:/",
			},
		},
	}
	if err := conf.Validate(); err == nil {
		t.Errorf("Validate error: Listen validation")
	}
}

func TestValidateNoLog(t *testing.T) {
	conf := &Configuration{
		Listen: ":80",
		Log:    "",
		Routes: []Route{
			{
				Path:           "/",
				AllowedMethods: []string{"GET"},
				ForwardUrl:     "local:/",
			},
		},
	}
	if err := conf.Validate(); err == nil {
		t.Errorf("Validate error: Log validation")
	}
}

func TestValidateInvalidListen(t *testing.T) {
	conf := &Configuration{
		Listen: "devlocal",
		Log:    "./log",
		Routes: []Route{
			{
				Path:           "/",
				AllowedMethods: []string{"GET"},
				ForwardUrl:     "local:/",
			},
		},
	}
	if err := conf.Validate(); err == nil {
		t.Errorf("Validate error: Listen validation")
	}
}

func TestValidateNoPath(t *testing.T) {
	conf := &Configuration{
		Listen: ":80",
		Log:    "./log",
		Routes: []Route{
			{
				Path:           "",
				AllowedMethods: []string{"GET"},
				ForwardUrl:     "local:/",
			},
		},
	}
	if err := conf.Validate(); err == nil {
		t.Errorf("Validate error: Path validation")
	}
}

func TestValidatePort(t *testing.T) {
	conf := &Configuration{
		Listen: ":",
		Log:    "./log",
		Routes: []Route{
			{
				Path:           "/",
				AllowedMethods: []string{"GET"},
				ForwardUrl:     "local:/",
			},
		},
	}
	if err := conf.Validate(); err == nil {
		t.Errorf("Validate error: Port validation")
	}
}

func TestGetLoadBalancer(t *testing.T) {
	urls := []string{
		"http://localhost:8080",
		"http://localhost:8081",
		"http://localhost:8082",
		"http://localhost:8083",
		"http://localhost:8084",
	}
	conf := Configuration{
		Upstreams: map[string][]string{
			"test": urls,
		},
		Routes: []Route{
			{
				Path:       "/",
				ForwardUrl: "test:/",
				AllowedMethods: []string{
					"GET",
				},
			},
		},
	}
	route, err := conf.getLoadBalancer(conf.Routes[0])
	if err != nil {
		t.Error(err)
	}
	if route == nil {
		t.Error("route is nil")
	}
	for i := 0; i < len(urls)*2; i++ {
		host := route.Next().Host
		if host != urls[i%len(urls)]+"/" {
			t.Errorf("host must be %s", host)
		}
	}
}

func TestGetLoadBalancerNoUpstream(t *testing.T) {
	urls := []string{
		"https://localhost:8443",
	}
	conf := Configuration{
		Routes: []Route{
			{
				Path:       "/",
				ForwardUrl: "https://localhost:8443/",
				AllowedMethods: []string{
					"GET",
				},
			},
		},
	}
	route, err := conf.getLoadBalancer(conf.Routes[0])
	if err != nil {
		t.Error(err)
	}
	if route == nil {
		t.Error("route is nil")
	}
	for i := 0; i < len(urls)*2; i++ {
		host := route.Next().Host
		if host != urls[i%len(urls)]+"/" {
			t.Errorf("host must be %s", host)
		}
	}
}

func TestCidrRangeContains(t *testing.T) {
	if !cidrRangeContains("192.168.1.0/24", "192.168.1.180") {
		t.Error("cidrRangeContains failed")
	}
}

func TestCidrRangeNotContains(t *testing.T) {
	if cidrRangeContains("192.168.1.0/24", "192.168.2.180") {
		t.Error("cidrRangeContains failed")
	}
}
