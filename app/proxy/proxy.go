package proxy

import (
	"io"
	"net/http"

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

func (self *ServerProxy) ServeHTTP(wr http.ResponseWriter, req *http.Request) {

	keyRequest := self.makeKeyUnique(req)

	resp, existCache, err := self.Cache.Find(keyRequest)
	if err != nil {
		logrus.Errorf("server proxy: serve http: %s", err.Error())
	}

	if !existCache {
		resp, err = self.doRequest(keyRequest.URL, req)
		if err != nil {
			http.Error(wr, "Server Error", http.StatusInternalServerError)
			logrus.Errorf("server proxy: serve http: %s", err.Error())
			return
		}
		if err := self.Cache.Save(keyRequest, resp); err != nil {
			logrus.Errorf("server proxy: serve http: %s", err.Error())
		}
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			logrus.Errorf("server proxy: serve http: body close: %s", err.Error())
		}
	}()

	logrus.Info(keyRequest.URL, " ", resp.Status)

	copyHeader(wr.Header(), resp.Header)
	wr.WriteHeader(resp.StatusCode)
	if _, err := io.Copy(wr, resp.Body); err != nil {
		logrus.Errorf("server proxy: serve http: io.copy: %s", err.Error())
	}
}
