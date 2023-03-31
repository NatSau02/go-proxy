package proxy

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/sirupsen/logrus"
)

type CacheResponse struct {
	cache map[string]ResponseInfo
}

func NewCacheResponse() *CacheResponse {
	return &CacheResponse{
		cache: make(map[string]ResponseInfo)}
}

func (self *CacheResponse) Save(url string, resp *http.Response) error {
	body, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return fmt.Errorf("cache response: save: %w", err)
	}
	self.cache[url] = ResponseInfo{
		Time: time.Now().Unix(),
		Body: body,
	}
	return nil
}

func (self *CacheResponse) Find(url string) (*http.Response, bool, error) {
	if data, exist := self.cache[url]; exist {
		if time.Since(time.Unix(data.Time, 0)).Minutes() < 10 {
			logrus.Info("fetched from cache ↓")
			respCache := bufio.NewReader(bytes.NewReader(data.Body))
			resp, err := http.ReadResponse(respCache, nil)
			if err != nil {
				return nil, false, fmt.Errorf("cache response: get response: %w", err)
			}
			return resp, true, nil
		}

	}
	return nil, false, nil
}
