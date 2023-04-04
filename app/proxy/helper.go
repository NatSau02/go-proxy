package proxy

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

type ValidHost []string

var hostsArray = ValidHost{
	".com",
	".org",
	".edu",
	".net",
	".fn",
	".info",
	".biz",
	".ru",
}

func (self ValidHost) isValid(host string) bool {
	for _, suff := range self {
		if strings.HasSuffix(host, suff) {
			return true
		}
	}
	return false
}

func (self *ServerProxy) makeKeyUnique(req *http.Request) KeyRequest {

	bytes, err := io.ReadAll(req.Body)
	if err != nil {
		log.Fatal(err)
	}

	body := hex.EncodeToString(bytes)
	if err != nil {
		log.Fatal(err)
	}

	return KeyRequest{
		URL:    self.makeURL(req),
		Method: req.Method,
		Body:   body,
	}

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

func (self *ServerProxy) doRequest(url string, req *http.Request) (*http.Response, error) {

	client := &http.Client{}
	reqTarget, err := http.NewRequestWithContext(context.Background(),
		req.Method, url, req.Body)
	if err != nil {
		return nil, fmt.Errorf(
			"do request: new request with context: %s", err.Error())
	}

	resp, err := client.Do(reqTarget)
	if err != nil {
		return nil, fmt.Errorf("do request: %s", err)
	}

	return resp, nil
}
