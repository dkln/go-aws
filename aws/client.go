package aws

import (
	"math"
	"net"
	"net/http"
	"time"
)

// Convenience method for creating an http client
func NewClient(rt *ResilientTransport) *http.Client {
	rt.transport = &http.Transport{
		Dial: func(netw, addr string) (net.Conn, error) {
			c, err := net.DialTimeout(netw, addr, rt.DialTimeout)
			if err != nil {
				return nil, err
			}
			c.SetDeadline(rt.Deadline())
			return c, nil
		},
		Proxy: http.ProxyFromEnvironment,
	}
	// TODO: Would be nice is ResilientTransport allowed clients to initialize
	// with http.Transport attributes.
	return &http.Client{
		Transport: rt,
	}
}


// Exported default client
var RetryingClient = NewClient(retryingTransport)

func ExpBackoff(try int) {
	time.Sleep(100 * time.Millisecond *
		time.Duration(math.Exp2(float64(try))))
}

func LinearBackoff(try int) {
	time.Sleep(time.Duration(try*100) * time.Millisecond)
}

// Decide if we should retry a request.
// In general, the criteria for retrying a request is described here
// http://docs.aws.amazon.com/general/latest/gr/api-retries.html
func awsRetry(req *http.Request, res *http.Response, err error) bool {
	retry := false

	// Don't retry if we got a result and no error.
	if err == nil && res != nil {
		retry = false
	}

	// Retry if there's a temporary network error.
	if neterr, ok := err.(net.Error); ok {
		if neterr.Temporary() {
			retry = true
		}
	}

	// Retry if we get a 5xx series error.
	if res != nil {
		if 500 <= res.StatusCode && res.StatusCode < 600 {
			retry = true
		}
	}
	return retry
}

