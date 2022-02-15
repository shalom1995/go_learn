package util

import (
	"bytes"
	"encoding/json"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	u "net/url"
	"sync"
	"time"
)

const (
	retryTime = 5
)

var (
	Pool = sync.Pool{
		New: func() interface{} {
			Client := http.Client{
				Transport: &http.Transport{
					DisableKeepAlives: true,
				},
				Timeout: time.Duration(time.Second * 10),
			}
			return Client
		},
	}
)

func init() {
	InitClient(10)
}

func InitClient(PoolSize int) {
	for i := 0; i < PoolSize; i++ {
		Client := http.Client{
			Transport: &http.Transport{
				DisableKeepAlives: true,
			},
		}

		Pool.Put(Client)
	}
}

func PostUrlRetry(url string, params map[string]string, body interface{}, headers map[string]string) ([]byte, error) {
	i := 0
	data, err := PostUrl(url, params, body, headers)
	for err != nil && i < retryTime {
		data, err = PostUrl(url, params, body, headers)
		i++
	}
	if err != nil {
		return nil, err
	}

	return data, nil
}

func PostUrl(url string, params map[string]string, body interface{}, headers map[string]string) ([]byte, error) {
	var (
		bodyJson []byte
		req      *http.Request
		err      error
	)

	client := Pool.Get().(http.Client)
	defer Pool.Put(client)

	if body != nil {
		bodyJson, err = json.Marshal(body)
		if err != nil {
			return nil, errors.Wrap(err, "json marshal request body error")
		}
	}

	req, err = http.NewRequest(http.MethodPost, url, bytes.NewBuffer(bodyJson))
	if err != nil {
		return nil, errors.Wrap(err, "NewRequest error")
	}

	contentType := "Content-type"
	req.Header.Set(contentType, headers[contentType])
	//add params
	q := req.URL.Query()
	if params != nil {
		for key, val := range params {
			q.Add(key, val)
		}
		req.URL.RawQuery = q.Encode()
	}

	//add headers
	if headers != nil {
		for key, val := range headers {
			req.Header.Add(key, val)
		}
	}

	response, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "client.Do error")
	}
	defer response.Body.Close()
	d, err_ := ioutil.ReadAll(response.Body)
	if err_ != nil {
		return nil, errors.Wrap(err, "ioutil.ReadAll error")
	}

	return d, nil
}

func GetURLRetry(url string, params map[string]string) ([]byte, error) {
	i := 0
	data, err := GetUrl(url, params)
	for err != nil && i < retryTime {
		data, err = GetUrl(url, params)
		i++
	}
	if err != nil {
		return nil, err
	}
	return data, nil
}

/**
 * @parameter:[url 要请求的路由][data 路由后面跟着的参数]
 * @return: 返回相应的数据
 * @Description: 向目标地址发起Get请求
 * @author: shalom
 * @date: 2020/12/9 23:29
 */
func GetUrl(url string, data map[string]string) ([]byte, error) {
	params := u.Values{}
	ur, err := u.Parse(url)
	if err != nil {
		return nil, errors.Wrap(err, "url.Parse error")
	}

	for key, value := range data {
		params.Set(key, value)
	}

	//	生成请求路径
	ur.RawQuery = params.Encode()
	urlPath := ur.String()

	res, err := http.Get(urlPath)
	if err != nil {
		return nil, errors.Wrap(err, "http.Get error")
	}
	defer res.Body.Close()

	//	将json
	d, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "ioutil.ReadAll error")
	}
	return d, nil
}
