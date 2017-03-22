package foreman

import (
	"net/http"
	"net/url"
)

// Query represents a search query.
// Used to get resource list such as hosts and architectures.
//
// parmas := url.Values{}
// params.Add("search", "host1.local.io")
// q := Query{"hosts", params}
// resp, err := client.Index(q)
type Query struct {
	Resource   string
	Parameters url.Values
}

// Index returns the HTTP response from Foreman regarding the resource
// as specified in the query `query`.
func (c Client) Index(query Query) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, query.Resource, nil)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = query.Parameters.Encode()
	return c.Do(req)
}
