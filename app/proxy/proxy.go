package proxy

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"
)

var hopHeaders = []string{
	"Connection",
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te", // canonicalized version of "TE"
	"Trailers",
	"Transfer-Encoding",
	"Upgrade",
}

type ServerProxy struct {
	Host  string
	Cache map[string]CacheData
}

type CacheData struct {
	time int64
	body []byte
}

func NewServerProxy() *ServerProxy {
	return &ServerProxy{
		Cache: make(map[string]CacheData),
	}
}

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

func (s *ServerProxy) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
	url := s.makeURL(req)
	if data, exist := s.Cache[url]; exist {
		if time.Since(time.Unix(data.time, 0)).Minutes() < 10 {
			fmt.Println("Нашел в кэше!!!")
			respCache := bufio.NewReader(bytes.NewReader(data.body))
			resp, err := http.ReadResponse(respCache, nil)
			log.Println(url, " ", resp.Status)
			if err != nil {
				log.Fatal(err)
			}
			io.Copy(wr, resp.Body)
			return
		}

	}

	var err error
	client := &http.Client{}

	req, err = http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return
	}

	delHopHeaders(req.Header)

	if clientIP, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
		appendHostToXForwardHeader(req.Header, clientIP)
	}

	resp, err := client.Do(req)
	if err != nil {
		http.Error(wr, "Server Error", http.StatusInternalServerError)
		log.Fatal("ServeHTTP:", err)
	}
	defer resp.Body.Close()

	log.Println(req.URL.String(), " ", resp.Status)

	cache, err := httputil.DumpResponse(resp, true)
	if err != nil {

	}
	s.Cache[req.URL.String()] = CacheData{
		time: time.Now().Unix(),
		body: cache,
	}

	delHopHeaders(resp.Header)

	copyHeader(wr.Header(), resp.Header)
	wr.WriteHeader(resp.StatusCode)
	io.Copy(wr, resp.Body)
}

func (s *ServerProxy) makeURL(req *http.Request) string {
	url := req.RequestURI
	if len(url) > 0 {
		url = url[1:]
	}
	host, path, _ := strings.Cut(url, "/")
	if isValidHost(host) {
		s.Host = host
		return fmt.Sprintf("https://%s/%s", s.Host, path)
	}
	return fmt.Sprintf("https://%s/%s", s.Host, url)
}

func isValidHost(host string) bool {
	for _, suff := range validHosts {
		if strings.HasSuffix(host, suff) {
			return true
		}

	}
	return false
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func delHopHeaders(header http.Header) {
	for _, h := range hopHeaders {
		header.Del(h)
	}
}

func appendHostToXForwardHeader(header http.Header, host string) {
	if prior, ok := header["X-Forwarded-For"]; ok {
		host = strings.Join(prior, ", ") + ", " + host
	}
	header.Set("X-Forwarded-For", host)
}
