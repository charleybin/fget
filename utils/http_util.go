package utils

import (
	"encoding/json"
	"fmt"
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

	ERROR_MSG_INTERNAL_SERVER = "服务器内部错误"
	ERROR_MSG_PARAM_NULL      = "没有 '%s' 参数"
	ERROR_MSG_PARAM_ILLEGAL   = "参数 '%s' 不合法"
)

type ResponseJson struct {
	Code int
	Msg  string
	Data interface{}
}

type UploadResult struct {
}

func MakeResponseJson(code int, msg string, data interface{}) *ResponseJson {
	return &ResponseJson{code, msg, data}
}

func MakeSuccJson(data interface{}) *ResponseJson {
	return MakeResponseJson(ERROR_CODE_SUCCESS, ERROR_MSG_SUCCESS, data)
}

func ResponseSucc(w http.ResponseWriter, data interface{}) {
	response_json := MakeSuccJson(data)
	response, err := json.Marshal(response_json)
	if err != nil {
		httpLog.Error("http_util", err.Error())
		return
	}

	w.Write(response)
}

func MakeInternalErrJson() *ResponseJson {
	return MakeResponseJson(ERROR_CODE_INTERNAL_SERVER, ERROR_MSG_INTERNAL_SERVER, "")
}

func ResponseInternalErr(w http.ResponseWriter) {
	response_json := MakeInternalErrJson()
	response, err := json.Marshal(response_json)
	if err != nil {
		httpLog.Error("http_util", err.Error())
		return
	}

	w.Write(response)
}

func MakeParamIllegalErrJson(param_name string) *ResponseJson {
	return MakeResponseJson(ERROR_CODE_PARAM_ILLEGAL, fmt.Sprintf(ERROR_MSG_PARAM_ILLEGAL, param_name), "")
}

func ResponseParamIllegalErr(w http.ResponseWriter, param_name string) {
	response_json := MakeParamIllegalErrJson(param_name)
	response, err := json.Marshal(response_json)
	if err != nil {
		httpLog.Error("http_util", err.Error())
		return
	}

	w.Write(response)
}

func MakeParamNullErrJson(param_name string) *ResponseJson {
	return MakeResponseJson(ERROR_CODE_PARAM_NULL, fmt.Sprintf(ERROR_MSG_PARAM_NULL, param_name), "")
}

func ResponseParamNullErr(w http.ResponseWriter, param_name string) {
	response_json := MakeParamNullErrJson(param_name)
	response, err := json.Marshal(response_json)
	if err != nil {
		httpLog.Error("http_util", err.Error())
		return
	}

	w.Write(response)
}

func ResponseNotFound(w http.ResponseWriter) {
	w.WriteHeader(404)
}

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
