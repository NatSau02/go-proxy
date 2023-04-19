package proxy

import (
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

var validHosts = []string{
	".com",
	".org",
	".edu",
	".net",
	".fn",
	".info",
	".biz",
	".ru",
}

func isValidHost(host string) bool {
	for _, suff := range validHosts {
		if strings.HasSuffix(host, suff) {
			return true
		}

	}
	return false
}

func (self *ServerProxy) makeKeyUnique(req *http.Request) KeyRequest {

	var body string
	if req.Body != nil {
		bytes, err := io.ReadAll(req.Body)
		if err != nil {
			log.Fatal(err)
		}

		body = hex.EncodeToString(bytes)
		if err != nil {
			log.Fatal(err)
		}
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
	if isValidHost(host) {
		self.Host = host
		return fmt.Sprintf("https://%s/%s", self.Host, path)
	}
	return fmt.Sprintf("https://%s/%s", self.Host, url)
}
