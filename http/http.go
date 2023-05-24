package http

import (
	"crypto/tls"
	"net/http"
	"net/url"
	"time"
)

var proxyURL *url.URL
var DefaultHttpTimeout = 6 * time.Second

func SetGlobalProxy(proxyServer string) {
	proxyURL, _ = url.Parse(proxyServer)
}

func RemoveGlobalProxy() {
	proxyURL = nil
}

func NewHttpClient(timeout ...time.Duration) *http.Client {
	transCfg := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		Proxy:           http.ProxyURL(proxyURL),
	}
	if len(timeout) == 1 {
		return &http.Client{
			Transport: transCfg,   // disable tls verify
			Timeout:   timeout[0], //必须设置一个超时，不然程序会抛出非自定义错误
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			}, //不跳转302
		}
	}
	return &http.Client{
		Transport: transCfg,           // disable tls verify
		Timeout:   DefaultHttpTimeout, //必须设置一个超时，不然程序会抛出非自定义错误
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}, //不跳转302
	}
}
