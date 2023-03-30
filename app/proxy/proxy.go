package proxy

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
)

type ServerProxy struct {
	Host  string
	Cache *CacheResponse
}

func NewServerProxy(cache *CacheResponse) *ServerProxy {
	return &ServerProxy{
		Cache: cache,
	}
}

func (self *ServerProxy) ServeHTTP(wr http.ResponseWriter, req *http.Request) {

	url := self.makeURL(req)

	resp, existCache, err := self.Cache.Find(url)
	if err != nil {
		logrus.Errorf("server proxy: serve http: %s", err.Error())
	}

	if !existCache {
		resp, err = doRequest(url, req)
		if err != nil {
			http.Error(wr, "Server Error", http.StatusInternalServerError)
			logrus.Errorf("server proxy: serve http: %s", err.Error())
			return
		}
		if err := self.Cache.Save(url, resp); err != nil {
			logrus.Errorf("server proxy: serve http: %s", err.Error())
		}
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			logrus.Errorf("server proxy: serve http: body close: %s", err.Error())
		}
	}()

	logrus.Info(url, " ", resp.Status)

	copyHeader(wr.Header(), resp.Header)
	wr.WriteHeader(resp.StatusCode)
	if _, err := io.Copy(wr, resp.Body); err != nil {
		logrus.Errorf("server proxy: serve http: io.copy: %s", err.Error())
	}
}

func doRequest(url string, req *http.Request) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequestWithContext(context.Background(),
		http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf(
			"do request: new request with context: %s", err.Error())
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %s", err)
	}

	return resp, nil
}

func (self *ServerProxy) makeURL(req *http.Request) string {
	url := req.RequestURI
	if len(url) > 0 {
		url = url[1:]
	}
	host, path, _ := strings.Cut(url, "/")
	if hostsArray.isValid(host) {
		self.Host = host
		return fmt.Sprintf("https://%s/%s", self.Host, path)
	}
	return fmt.Sprintf("https://%s/%s", self.Host, url)
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}
