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
or if you have ```goginx.json``` in the current directory
```shell
goginx
```
Sample goginx.json file
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
    "routes" : [
        {
            "path" : "/search",
            "forwardUrl" : "https://httpbin.org/anything",
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
            }
        },
        {
            "path" : "/downloads",
            "forwardUrl" : "file://dist"

        }
    ]
}
```