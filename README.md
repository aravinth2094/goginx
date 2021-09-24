# Goginx
A simpler version of Nginx.

## Installation
* Install golang
* Install goreleaser
* Install make
```shell
go get
make build
```

## Goginx Configuration
Run using
```shell
goginx -c config.json
```
```shell
goginx -c https://<fileuploadserver.io>/config.json
```
or if you have ```goginx.json``` in the current directory
```shell
goginx
```
Help Menu
```shell
Usage of goginx:
  -V    Validate configuration file
  -c string
        Goginx configuration file location (default "goginx.json")
  -h    Print this help
```
Basic Sample goginx.json file
```json
{
    "routes" : [
        {
            "path" : "/search",
            "forwardUrl" : "https://httpbin.org/anything",
            "allowedMethods": [ "GET", "POST" ]
        }
    ]
}
```

Advanced Sample goginx.json file
```json
{
    "listen" : ":443",
    "certificate" : "cert.pem",
    "key" : "key.pem",
    "log" : "goginx.log",
    "whitelist": [
        "127.0.0.1",
        "192.168.1.0/24"
    ],
    "compression" : true,
    "upstreams" : {
        "httpbin" : [
            "https://httpbin.org"
        ]
    },
    "routes" : [
        {
            "path" : "/search",
            "forwardUrl" : "httpbin:/anything",
            "allowedMethods": [ "GET", "POST" ],
            "forwardIp": true,
            "appendPath": false,
            "customHeaders" : {
                "X-Custom-Header1" : "Custom-Header1-Value",
                "X-Custom-Header2" : "Custom-Header2-Value"
            },
            "secureHeaders" : true,
            "cors" : {
                "origin" : "*",
                "methods" : "GET, POST",
                "credentials": true,
                "maxAge": "86400",
                "allowedHeaders": "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization",
                "cacheControl": "no-cache",
                "vary": "Accept-Encoding"
            },
            "cache" : 60,
            "timeout" : 5000
        },
        {
            "path" : "/downloads",
            "forwardUrl" : "file://dist"
        }
    ]
}
```