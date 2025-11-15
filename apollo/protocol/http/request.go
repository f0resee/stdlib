package http

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/f0resee/stdlib/apollo/component/log"
	"github.com/f0resee/stdlib/apollo/env"
	"github.com/f0resee/stdlib/apollo/env/config"
	"github.com/f0resee/stdlib/apollo/env/server"
	"github.com/f0resee/stdlib/apollo/extension"
	"github.com/f0resee/stdlib/apollo/utils"
)

var (
	onErrorRetryInterval = 2 * time.Second

	connectTimeout = 1 * time.Second

	maxRetries = 5

	defaultMaxConnsPerHost = 512
	defaultTimeoutBySecond = 1 * time.Second
	defaultKeepAliveSecond = 60 * time.Second
	once                   sync.Once
	defaultTransport       *http.Transport
)

func getDefaultTransport(insecureSkipVerify bool) *http.Transport {
	once.Do(func() {
		defaultTransport = &http.Transport{
			Proxy:               http.ProxyFromEnvironment,
			MaxIdleConns:        defaultMaxConnsPerHost,
			MaxIdleConnsPerHost: defaultMaxConnsPerHost,
			DialContext: (&net.Dialer{
				KeepAlive: defaultKeepAliveSecond,
				Timeout:   defaultTimeoutBySecond,
			}).DialContext,
		}
		if insecureSkipVerify {
			defaultTransport.TLSClientConfig = &tls.Config{
				InsecureSkipVerify: insecureSkipVerify,
			}
		}
	})
	return defaultTransport
}

type CallBack struct {
	SuccessCallBack   func([]byte, CallBack) (interface{}, error)
	NotModifyCallBack func() error
	AppConfigFunc     func() config.AppConfig
	Namespace         string
}

func Request(requestURL string, connectionConfig *env.ConnectConfig, callBack *CallBack) (interface{}, error) {
	client := &http.Client{}
	if connectionConfig != nil && connectionConfig.Timeout != 0 {
		client.Timeout = connectionConfig.Timeout
	} else {
		client.Timeout = connectTimeout
	}

	var err error
	url, err := url.Parse(requestURL)
	if err != nil {
		log.Errorf("request Apollo Server url: %q is invalid: %v", requestURL, err)
		return nil, err
	}

	var insecureSkipVerify bool
	if strings.HasPrefix(url.Scheme, "https") {
		insecureSkipVerify = true
	}
	client.Transport = getDefaultTransport(insecureSkipVerify)

	retry := 0
	var retries = maxRetries
	if connectionConfig != nil && !connectionConfig.IsRetry {
		retries = 1
	}
	for {
		retry++

		if retry > retries {
			break
		}
		var req *http.Request
		req, err := http.NewRequest("GET", requestURL, nil)
		if req == nil || err != nil {
			log.Errorf("Generate connect Apollo request Fail, url: %s, error: %v", requestURL, err)
			return nil, errors.New("generate connect Apollo request fail")
		}

		httpAuth := extension.GetHTTPAuth()
		if httpAuth != nil {
			headers := httpAuth.HTTPHeaders(requestURL, connectionConfig.AppID, connectionConfig.Secret)
			if len(headers) > 0 {
				req.Header = headers
			}
			host := req.Header.Get("Host")
			if len(host) > 0 {
				req.Host = host
			}
		}

		var res *http.Response
		res, err = client.Do(req)
		if res != nil {
			defer res.Body.Close()
		}

		if res == nil || err != nil {
			log.Errorf("Connect Apollo Server Fail, url: %s, error: %v", requestURL, err)
			time.Sleep(onErrorRetryInterval)
			continue
		}

		switch res.StatusCode {
		case http.StatusOK:
			var resonseBody []byte
			resonseBody, err = ioutil.ReadAll(res.Body)
			if err != nil {
				log.Errorf("Read Apollo Server Response Fail, url: %s, error: %v", requestURL, err)
				time.Sleep(onErrorRetryInterval)
				continue
			}

			if callBack != nil && callBack.SuccessCallBack != nil {
				return callBack.SuccessCallBack(resonseBody, *callBack)
			}
			return nil, nil
		case http.StatusNotModified:
			log.Debugf("Config Not Modified, error: %v", err)
			if callBack != nil && callBack.NotModifyCallBack != nil {
				return nil, callBack.NotModifyCallBack()
			}
			return nil, nil
		case http.StatusBadRequest, http.StatusUnauthorized, http.StatusNotFound, http.StatusMethodNotAllowed:
			log.Errorf("Connect Apollo Server Fail, url: %s, StatusCode: %d", requestURL, res.StatusCode)
			return nil, fmt.Errorf("Connect Apollo Server Fail, url: %s, StatusCode: %d", requestURL, res.StatusCode)
		default:
			log.Errorf("Connect Apollo Server Fail, url: %s, StatusCode: %d", requestURL, res.StatusCode)
			time.Sleep(onErrorRetryInterval)
			continue
		}
	}
	log.Errorf("Over Max Retry Still Error, error: %v", err)
	if retry > retries {
		err = errors.New("over Max Retry Still Error")
	}
	return nil, err
}

func RequestRecovery(appConfig config.AppConfig, connectConfig *env.ConnectConfig, callback *CallBack) (interface{}, error) {
	format := "%s%s"
	var err error
	var response interface{}

	for {
		host := loadBalance(appConfig)
		if host == "" {
			return nil, err
		}

		retquestURL := fmt.Sprintf(format, host, connectConfig.URI)
		response, err = Request(retquestURL, connectConfig, callback)
		if err == nil {
			return response, nil
		}

		server.SetDownNode(appConfig.GetHost(), host)
	}
}

func loadBalance(appConfig config.AppConfig) string {
	if !server.IsConnectDirectly(appConfig.GetHost()) {
		return appConfig.GetHost()
	}

	serverInfo := extension.GetLoadBalance().Load(server.GetServers(appConfig.GetHost()))
	if serverInfo == nil {
		return utils.Empty
	}
	return serverInfo.HomepageURL
}
