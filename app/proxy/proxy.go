package proxy

import (
	"io"
	"net/http"
	"net/http/httputil"

	"github.com/sirupsen/logrus"
)

type ServerProxy struct {
	Host  string
	Cache *ResponseCache
}

func NewServerProxy(cache *ResponseCache) *ServerProxy {
	return &ServerProxy{
		Cache: cache,
	}
}

// ProxyRequestHandler handles the http request using proxy
func (self *ServerProxy) ProxyRequestHandler(proxy *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		keyRequest := self.makeKeyUnique(r)
		resp, present, err := self.Cache.Find(keyRequest)
		if err != nil {
			logrus.Errorf("server proxy: proxy request handler: %s", err.Error())
		}

		defer func() {
			if err := resp.Body.Close(); err != nil {
				logrus.Errorf("server proxy: serve http: body close: %s", err.Error())
			}
		}()
		if present {
			logrus.Info(keyRequest.URL, " ", resp.Status)

			w.WriteHeader(resp.StatusCode)
			if _, err := io.Copy(w, resp.Body); err != nil {
				logrus.Errorf("server proxy: serve http: io.copy: %s", err.Error())
			}
			return
		}

		proxy.ServeHTTP(w, r)
	}
}
