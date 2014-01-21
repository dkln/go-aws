package s3

import (
  "io"
  "net/url"
  "net/http"
  "time"
  "io/ioutil"
  "bytes"
  "strconv"
)

// The Bucket type encapsulates operations with an S3 bucket.
type Bucket struct {
	*S3
	Name string
}

// PutBucket creates a new bucket.
//
// See http://goo.gl/ndjnR for details.
func (self *Bucket) PutBucket(perm ACL) error {
	headers := map[string][]string{
		"x-amz-acl": {string(perm)},
	}
	req := &request{
		method:  "PUT",
		bucket:  self.Name,
		path:    "/",
		headers: headers,
		payload: self.locationConstraint(),
	}
	return self.S3.query(req, nil)
}

// DelBucket removes an existing S3 bucket. All objects in the bucket must
// be removed before the bucket itself can be removed.
//
// See http://goo.gl/GoBrY for details.
func (self *Bucket) DelBucket() (err error) {
	req := &request{
		method: "DELETE",
		bucket: self.Name,
		path:   "/",
	}
	for attempt := attempts.Start(); attempt.Next(); {
		err = self.S3.query(req, nil)
		if !shouldRetry(err) {
			break
		}
	}
	return err
}

// Get retrieves an object from an S3 bucket.
//
// See http://goo.gl/isCO7 for details.
func (self *Bucket) Get(path string) (data []byte, err error) {
	body, err := self.GetReader(path)
	if err != nil {
		return nil, err
	}
	data, err = ioutil.ReadAll(body)
	body.Close()
	return data, err
}

// GetReader retrieves an object from an S3 bucket.
// It is the caller's responsibility to call Close on rc when
// finished reading.
func (self *Bucket) GetReader(path string) (rc io.ReadCloser, err error) {
	resp, err := self.GetResponse(path)
	if resp != nil {
		return resp.Body, err
	}
	return nil, err
}

// GetResponse retrieves an object from an S3 bucket returning the http response
// It is the caller's responsibility to call Close on rc when
// finished reading.
func (self *Bucket) GetResponse(path string) (*http.Response, error) {
	req := &request{
		bucket: self.Name,
		path:   path,
	}
	err := self.S3.prepare(req)
	if err != nil {
		return nil, err
	}
	for attempt := attempts.Start(); attempt.Next(); {
		resp, err := self.S3.run(req, nil)
		if shouldRetry(err) && attempt.HasNext() {
			continue
		}
		if err != nil {
			return nil, err
		}
		return resp, nil
	}
	panic("unreachable")
}

// Put inserts an object into the S3 bucket.
//
// See http://goo.gl/FEBPD for details.
func (self *Bucket) Put(path string, data []byte, contType string, perm ACL) error {
	body := bytes.NewBuffer(data)
	return self.PutReader(path, body, int64(len(data)), contType, perm)
}

/*
PutHeader - like Put, inserts an object into the S3 bucket.
Instead of Content-Type string, pass in custom headers to override defaults.
*/
func (self *Bucket) PutHeader(path string, data []byte, customHeaders map[string][]string, perm ACL) error {
	body := bytes.NewBuffer(data)
	return self.PutReaderHeader(path, body, int64(len(data)), customHeaders, perm)
}

// PutReader inserts an object into the S3 bucket by consuming data
// from r until EOF.
func (self *Bucket) PutReader(path string, r io.Reader, length int64, contType string, perm ACL) error {
	headers := map[string][]string{
		"Content-Length": {strconv.FormatInt(length, 10)},
		"Content-Type":   {contType},
		"x-amz-acl":      {string(perm)},
	}
	req := &request{
		method:  "PUT",
		bucket:  self.Name,
		path:    path,
		headers: headers,
		payload: r,
	}
	return self.S3.query(req, nil)
}

/*
PutReaderHeader - like PutReader, inserts an object into S3 from a reader.
Instead of Content-Type string, pass in custom headers to override defaults.
*/
func (self *Bucket) PutReaderHeader(path string, r io.Reader, length int64, customHeaders map[string][]string, perm ACL) error {
	// Default headers
	headers := map[string][]string{
		"Content-Length": {strconv.FormatInt(length, 10)},
		"Content-Type":   {"application/text"},
		"x-amz-acl":      {string(perm)},
	}

	// Override with custom headers
	for key, value := range customHeaders {
		headers[key] = value
	}

	req := &request{
		method:  "PUT",
		bucket:  self.Name,
		path:    path,
		headers: headers,
		payload: r,
	}
	return self.S3.query(req, nil)
}

// Del removes an object from the S3 bucket.
//
// See http://goo.gl/APeTt for details.
func (self *Bucket) Del(path string) error {
	req := &request{
		method: "DELETE",
		bucket: self.Name,
		path:   path,
	}
	return self.S3.query(req, nil)
}

// List returns information about objects in an S3 bucket.
//
// The prefix parameter limits the response to keys that begin with the
// specified prefix.
//
// The delim parameter causes the response to group all of the keys that
// share a common prefix up to the next delimiter in a single entry within
// the CommonPrefixes field. You can use delimiters to separate a bucket
// into different groupings of keys, similar to how folders would work.
//
// The marker parameter specifies the key to start with when listing objects
// in a bucket. Amazon S3 lists objects in alphabetical order and
// will return keys alphabetically greater than the marker.
//
// The max parameter specifies how many keys + common prefixes to return in
// the response. The default is 1000.
//
// For example, given these keys in a bucket:
//
//     index.html
//     index2.html
//     photos/2006/January/sample.jpg
//     photos/2006/February/sample2.jpg
//     photos/2006/February/sample3.jpg
//     photos/2006/February/sample4.jpg
//
// Listing this bucket with delimiter set to "/" would yield the
// following result:
//
//     &ListResp{
//         Name:      "sample-bucket",
//         MaxKeys:   1000,
//         Delimiter: "/",
//         Contents:  []Key{
//             {Key: "index.html", "index2.html"},
//         },
//         CommonPrefixes: []string{
//             "photos/",
//         },
//     }
//
// Listing the same bucket with delimiter set to "/" and prefix set to
// "photos/2006/" would yield the following result:
//
//     &ListResp{
//         Name:      "sample-bucket",
//         MaxKeys:   1000,
//         Delimiter: "/",
//         Prefix:    "photos/2006/",
//         CommonPrefixes: []string{
//             "photos/2006/February/",
//             "photos/2006/January/",
//         },
//     }
//
// See http://goo.gl/YjQTc for details.
func (self *Bucket) List(prefix, delim, marker string, max int) (result *ListResp, err error) {
	params := map[string][]string{
		"prefix":    {prefix},
		"delimiter": {delim},
		"marker":    {marker},
	}
	if max != 0 {
		params["max-keys"] = []string{strconv.FormatInt(int64(max), 10)}
	}
	req := &request{
		bucket: self.Name,
		params: params,
	}
	result = &ListResp{}
	for attempt := attempts.Start(); attempt.Next(); {
		err = self.S3.query(req, result)
		if !shouldRetry(err) {
			break
		}
	}
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Returns a mapping of all key names in this bucket to Key objects
func (self *Bucket) GetBucketContents() (*map[string]Key, error) {
	bucket_contents := map[string]Key{}
	prefix := ""
	path_separator := ""
	marker := ""
	for {
		contents, err := self.List(prefix, path_separator, marker, 1000)
		if err != nil {
			return &bucket_contents, err
		}
		for _, key := range contents.Contents {
			bucket_contents[key.Key] = key
		}
		if contents.IsTruncated {
			marker = contents.NextMarker
		} else {
			break
		}
	}

	return &bucket_contents, nil
}

// URL returns a non-signed URL that allows retriving the
// object at path. It only works if the object is publicly
// readable (see SignedURL).
func (self *Bucket) URL(path string) string {
	req := &request{
		bucket: self.Name,
		path:   path,
	}
	err := self.S3.prepare(req)
	if err != nil {
		panic(err)
	}
	u, err := req.url()
	if err != nil {
		panic(err)
	}
	u.RawQuery = ""
	return u.String()
}

// SignedURL returns a signed URL that allows anyone holding the URL
// to retrieve the object at path. The signature is valid until expires.
func (self *Bucket) SignedURL(path string, expires time.Time) string {
	req := &request{
		bucket: self.Name,
		path:   path,
		params: url.Values{"Expires": {strconv.FormatInt(expires.Unix(), 10)}},
	}

	err := self.S3.prepare(req)

	if err != nil {
		panic(err)
	}
	u, err := req.url()
	if err != nil {
		panic(err)
	}
	return u.String()
}
