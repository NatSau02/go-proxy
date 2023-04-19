package proxy

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func (self *ServerProxy) ConfigureProxy(targetHost string) (*httputil.ReverseProxy, error) {
	URL, err := url.Parse(targetHost)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(URL)

	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		self.modifyRequest(req)
	}

	proxy.ModifyResponse = self.modifyResponse()
	proxy.ErrorHandler = errorHandler()
	return proxy, nil
}

func (self *ServerProxy) modifyRequest(req *http.Request) {
	newURL := self.makeURL(req)

	u, err := url.Parse(newURL)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Add("X-Forwarded-Host", req.Host)
	req.Header.Add("X-Origin-Host", u.Host)
	req.Host = u.Host
	req.URL = u
}

func errorHandler() func(http.ResponseWriter, *http.Request, error) {
	return func(w http.ResponseWriter, req *http.Request, err error) {
		if err != nil {
			fmt.Printf("Got error while modifying response: %v \n", err)
		}
		return
	}
}

func (self *ServerProxy) modifyResponse() func(*http.Response) error {
	return func(resp *http.Response) error {
		log.Println(resp.Request.URL.String(), " ", resp.Status)

		keyRequest := self.makeKeyUnique(resp.Request)

		if err := self.Cache.Save(keyRequest, resp); err != nil {
			logrus.Errorf("server proxy: serve http: %s", err.Error())
		}

		return nil
	}
}
