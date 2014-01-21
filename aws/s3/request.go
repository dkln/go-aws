package s3

import (
  "net/url"
  "net/http"
  "io"
  "fmt"
)

type request struct {
	method   string
	bucket   string
	path     string
	signpath string
	params   url.Values
	headers  http.Header
	baseurl  string
	payload  io.Reader
	prepared bool
}

/**
 *
 */
func (self *request) url() (*url.URL, error) {
	u, err := url.Parse(self.baseurl)

	if err != nil {
		return nil, fmt.Errorf("bad S3 endpoint URL %q: %v", self.baseurl, err)
	}

	u.RawQuery = self.params.Encode()
	u.Path = self.path

	return u, nil
}
