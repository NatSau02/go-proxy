package proxy

import "strings"

type ResponseInfo struct {
	Time int64
	Body []byte
}

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
