package utils

import (
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	MOBILE_UA = "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/48.0.2564.23 Mobile Safari/537.36"
)

var fileNameEscaper = strings.NewReplacer("\\", "\\\\", "\"", "\\\"")
var httpLog = GetLogger("http_utils")

//错误码
const (
	ERROR_CODE_SUCCESS = 0

	ERROR_CODE_INTERNAL_SERVER = (iota + 1000)
	ERROR_CODE_PARAM_NULL
	ERROR_CODE_PARAM_ILLEGAL
)

const (
	ERROR_MSG_SUCCESS = "success"
)

func HttpGet(url string, headers map[string]string) (*http.Response, error) {

	client := &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives:   true,
			MaxIdleConnsPerHost: 1024,
			Dial: func(newtw, addr string) (net.Conn, error) {
				deadline := time.Now().Add(45 * time.Second)
				c, err := net.DialTimeout(newtw, addr, time.Second*45)
				if err != nil {
					//fmt.Printf("---dialTimeout err:%v \n", err)
					return nil, err
				}
				c.SetDeadline(deadline)
				return c, nil
			},
		},
	}

	// New request
	req, err := http.NewRequest("GET", url, nil)
	if nil != err {
		return nil, err
	}

	// Set Header
	for key, val := range headers {
		req.Header.Set(key, val)
	}

	// Send request
	resp, err := client.Do(req)

	return resp, err
}

func HttpProxyGet(urlAddr string, proxyAddr string, headers map[string]string) (*http.Response, error) {

	// New request
	req, err := http.NewRequest("GET", urlAddr, nil)
	if nil != err {
		return nil, err
	}
	proxy, err := url.Parse(proxyAddr)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxy),
		},
	}
	// Set Header
	//	for key, val := range headers {
	//		req.Header.Set(key, val)
	//	}

	// Send request
	resp, err := client.Do(req)

	return resp, err
}

func HttpPost(url string, headers map[string]string, data map[string]string) (*http.Response, error) {

	client := &http.Client{}

	// Build post string
	pairs := make([]string, 0)
	for key, val := range data {
		pairs = append(pairs, key+"="+val)
	}
	postData := strings.Join(pairs, "&")

	// New request
	req, err := http.NewRequest("POST", url, strings.NewReader(postData))
	if nil != err {
		return nil, err
	}

	// Set Header
	for key, val := range headers {
		req.Header.Set(key, val)
	}

	// Send request
	resp, err := client.Do(req)

	return resp, err
}

func URLDecode(url string) string {
	url = strings.Replace(url, "%3A", ":", -1)
	url = strings.Replace(url, "%2F", "/", -1)
	return url
}
