package s3

// The ListResp type holds the results of a List bucket operation.
type ListResp struct {
	Name       string
	Prefix     string
	Delimiter  string
	Marker     string
	NextMarker string
	MaxKeys    int
	// IsTruncated is true if the results have been truncated because
	// there are more keys and prefixes than can fit in MaxKeys.
	// N.B. this is the opposite sense to that documented (incorrectly) in
	// http://goo.gl/YjQTc
	IsTruncated    bool
	Contents       []Key
	CommonPrefixes []string `xml:">Prefix"`
}
