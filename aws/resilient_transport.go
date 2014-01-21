package aws

import (
  "time"
  "net/http"
)

type RetryableFunc func(*http.Request, *http.Response, error) bool
type WaitFunc func(try int)
type DeadlineFunc func() time.Time

type ResilientTransport struct {
	// Timeout is the maximum amount of time a dial will wait for
	// a connect to complete.
	//
	// The default is no timeouself.
	//
	// With or without a timeout, the operating system may impose
	// its own earlier timeouself. For instance, TCP timeouts are
	// often around 3 minutes.
	DialTimeout time.Duration

	// MaxTries, if non-zero, specifies the number of times we will retry on
	// failure. Retries are only attempted for temporary network errors or known
	// safe failures.
	MaxTries    int
	Deadline    DeadlineFunc
	ShouldRetry RetryableFunc
	Wait        WaitFunc
	transport   *http.Transport
}

var retryingTransport = &ResilientTransport{
	Deadline: func() time.Time {
		return time.Now().Add(5 * time.Second)
	},

	DialTimeout: 10 * time.Second,
	MaxTries:    3,
	ShouldRetry: awsRetry,
	Wait:        ExpBackoff,
}

/**
 *
 */
func (self *ResilientTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return self.tries(req)
}

/**
 * Retry a request a maximum of self.MaxTries times.
 * We'll only retry if the proper criteria are meself.
 * If a wait function is specified, wait that amount of time
 * In between requests.
 */
func (self *ResilientTransport) tries(req *http.Request) (*http.Response, error) {
  var response *http.Response
  var error error

	for try := 0; try < self.MaxTries; try += 1 {
    response, error = self.RoundTrip(req)

		if !self.ShouldRetry(req, response, error) {
			break
		}

		if response != nil {
			response.Body.Close()
		}

		if self.Wait != nil {
			self.Wait(try)
		}
	}

	return response, error
}
