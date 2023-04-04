package proxy

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
	"net/http/httputil"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type ResponseCache struct {
	sync.Mutex
	cache map[KeyRequest]ResponseInfo
}

func NewCacheResponse() *ResponseCache {
	return &ResponseCache{
		cache: make(map[KeyRequest]ResponseInfo)}
}

func (self *ResponseCache) Save(key KeyRequest, resp *http.Response) error {
	body, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return fmt.Errorf("http_cache response: save: %w", err)
	}
	self.Lock()
	defer self.Unlock()
	self.cache[key] = ResponseInfo{
		Time: time.Now().Unix(),
		Body: body,
	}
	return nil
}

func (self *ResponseCache) Find(request KeyRequest) (*http.Response, bool, error) {
	self.Lock()
	defer self.Unlock()
	if data, exist := self.cache[request]; exist {
		logrus.Info("fetched from http_cache ↓")
		respCache := bufio.NewReader(bytes.NewReader(data.Body))
		resp, err := http.ReadResponse(respCache, nil)
		if err != nil {
			return nil, false, fmt.Errorf("http_cache response: get response: %w", err)
		}
		return resp, true, nil
	}
	return nil, false, nil
}

func (self *ResponseCache) Update() {
	self.Lock()
	defer self.Unlock()

	for key, data := range self.cache {
		if time.Since(time.Unix(data.Time, 0)).Minutes() < 10 {
			delete(self.cache, key)
		}
	}
}

type KeyRequest struct {
	URL    string
	Method string
	Body   string
}

type ResponseInfo struct {
	Time int64
	Body []byte
}
