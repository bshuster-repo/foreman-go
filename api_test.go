package foreman

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestIndex(t *testing.T) {
	tt := []struct {
		resource string
		query    string
	}{
		{},
		{"hosts", "search=myhost.local"},
		{"bla", "a=1&b=2&c=10"},
		{"", "a=1&b=2&c=10"},
		{"blabla", ""},
	}
	opts := Options{
		Address:    "http://foreman.acme.io:89",
		APIVersion: "v3",
	}
	c := New(opts)
	address := opts.Address + "/api/" + opts.APIVersion

	for _, te := range tt {
		c.httpClient = newTestingHTTPClient(func(req *http.Request) (*http.Response, error) {
			fURL := address + "/" + te.resource
			if te.query != "" {
				fURL = fURL + "?" + te.query
			}
			if req.URL.String() != fURL {
				t.Errorf("expected request URL to be '%s' but got '%s'", fURL, req.URL)
			}
			if req.URL.Scheme != "http" {
				t.Errorf("expected request URL Scheme to be 'http' but got '%s'", req.URL.Scheme)
			}
			if req.URL.Host != strings.TrimPrefix(opts.Address, "http://") {
				t.Errorf("expected request URL Host to be '%s' but got '%s'", fURL, req.URL.Host)
			}
			if req.URL.Path != "/api/v3/"+te.resource {
				t.Errorf("expected request URL Path to be '%s' but got '%s'", "/api/v3/"+te.resource, req.URL.Path)
			}
			if req.URL.RawQuery != te.query {
				t.Errorf("expected request URL RawQuery to be '%s' but got '%s'", te.query, req.URL.RawQuery)
			}
			if req.Method != http.MethodGet {
				t.Errorf("expected request Method to be '%s' but got '%s'", http.MethodGet, req.Method)
			}
			return &http.Response{}, nil
		})
		qVal, err := url.ParseQuery(te.query)
		if err != nil {
			t.Errorf("expected ParseQuery not to return error: %s", err)
		}
		_, err = c.Index(Query{te.resource, qVal})
		if err != nil {
			t.Errorf("expected Index not to return error: %s", err)
		}
	}
}

func TestIndexError(t *testing.T) {
	c := New(Options{})
	if _, err := c.Index(Query{":wrongURL", url.Values{}}); err == nil {
		t.Error("expected Index to return error")
	}
}
