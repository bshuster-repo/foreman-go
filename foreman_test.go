package foreman

import (
	"net/http"
	"strings"
	"testing"
)

func TestNewDefaults(t *testing.T) {
	c := New(Options{})
	if c.httpClient != defaultHTTPClient {
		t.Errorf("expected the default used http client to be %p %#v but got %p %#v", defaultHTTPClient, defaultHTTPClient, c.httpClient, c.httpClient)
	}
	if c.mods == nil {
		t.Errorf("expected mods not to be nil")
	}
}

func setTestRequest(t *testing.T) *http.Request {
	req, err := http.NewRequest(http.MethodGet, "", nil)
	if err != nil {
		t.Errorf("expected the request creation not to return error: %s", err)
	}
	return req
}

func assertModifyNotNil(t *testing.T, c Client) {
	if c.mods == nil {
		t.Errorf("expected mods not to be nil")
		t.FailNow()
	}
}

func TestNewSetsHeadersCorrectly(t *testing.T) {
	tt := []struct {
		key string
		val string
	}{
		{"Content-Type", "application/json"},
		{"Agent", "GoForemanAPIClient"},
	}
	c := New(Options{})
	assertModifyNotNil(t, c)
	req := c.mods.Modify(setTestRequest(t))
	for _, te := range tt {
		if req.Header.Get(te.key) != te.val {
			t.Errorf("expected request %s to be '%s' but got '%s'", te.key, te.val, req.Header.Get(te.key))
		}
	}
}

func TestNewDefaultsIsCorrect(t *testing.T) {
	c := New(Options{})
	assertModifyNotNil(t, c)
	req := c.mods.Modify(setTestRequest(t))
	expURL := strings.TrimPrefix(defaultAddress, "http://")
	if req.URL.Host != expURL {
		t.Errorf("expected request URL Host to be '%s' but got '%s'", expURL, req.URL.Host)
	}
	if req.URL.Scheme != "http" {
		t.Errorf("expected request URL Scheme to be 'http' but got '%s'", req.URL.Scheme)
	}
	if req.URL.Path != defaultAPIVersion+"/" {
		t.Errorf("expected request URL Path to be '%s' but got '%s'", defaultAPIVersion+"/", req.URL.Path)
	}

	_, _, k := req.BasicAuth()
	if k {
		t.Error("expected request to not have BasicAuth")
	}
}

func TestNewSetsForemanURL(t *testing.T) {
	tt := []struct {
		address string
		apiVer  string
		scheme  string
		host    string
		path    string
	}{
		{"http://10.20.30.40:8989", "", "http", "10.20.30.40:8989", "v2/"},
		{"https://172.20.30.40:8989", "", "https", "172.20.30.40:8989", "v2/"},
		{"10.20.30.40:8989", "", "", "10.20.30.40:8989", "v2/"},
		{"tcp://foreman.internal.acme.io:8989", "", "", "tcp://foreman.internal.acme.io:8989", "v2/"},
		{"https://foreman.acme.io:80", "v5", "https", "foreman.acme.io:80", "v5/"},
	}
	for _, te := range tt {
		c := New(Options{
			Address:    te.address,
			APIVersion: te.apiVer,
		})
		assertModifyNotNil(t, c)
		req := c.mods.Modify(setTestRequest(t))
		if req.URL.Scheme != te.scheme {
			t.Errorf("expected request URL Scheme to be '%s' but got '%s'", te.scheme, req.URL.Scheme)
		}
		if req.URL.Host != te.host {
			t.Errorf("expected request URL Host to be '%s' but got '%s'", te.host, req.URL.Host)
		}
		if req.URL.Path != te.path {
			t.Errorf("expected request URL Host to be '%s' but got '%s'", te.path, req.URL.Path)
		}
	}

}

func TestHasBasicAuth(t *testing.T) {
	tt := []Options{
		{Username: "dave"},
		{Username: "admin", Password: "helloworld!"},
	}
	for _, te := range tt {
		c := New(te)
		assertModifyNotNil(t, c)
		req := c.mods.Modify(setTestRequest(t))
		u, p, k := req.BasicAuth()
		if !k {
			t.Error("expected request to have BasicAuth")
		}
		if u != te.Username {
			t.Errorf("expected Username to be '%s' but got '%s'", te.Username, u)
		}
		if p != te.Password {
			t.Errorf("expected Username to be '%s' but got '%s'", te.Password, p)
		}
	}
}

func TestNewSetHTTPClient(t *testing.T) {
	opts := Options{
		HTTPClient: &http.Client{},
	}
	c := New(opts)
	if c.httpClient == defaultHTTPClient {
		t.Errorf("expected HTTP client(%p) '%v' not to be equal to the default one(%p) '%v'", c.httpClient, c.httpClient, defaultHTTPClient, defaultHTTPClient)
	}
}

type funcRoundTripper func(*http.Request) (*http.Response, error)

func (frt funcRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return frt(req)
}

func newTestingHTTPClient(ops func(*http.Request) (*http.Response, error)) *http.Client {
	return &http.Client{
		Transport: funcRoundTripper(ops),
	}
}

func TestClientHead(t *testing.T) {
	tt := []struct {
		resource  string
		username  string
		password  string
		basicAuth bool
		address   string
		apiVer    string
		scheme    string
		host      string
		path      string
	}{
		{"hosts", "admin", "newpass1", true, "http://foreman.acme.io", "", "http", "foreman.acme.io", "v2/hosts"},
		{"", "admin", "", true, "https://foreman.acme.io", "v3", "https", "foreman.acme.io", "v3/"},
		{"", "", "", false, "https://foreman.acme.io", "v3", "https", "foreman.acme.io", "v3/"},
		{"http://hello.world.com/hostl", "", "", false, "https://foreman.acme.io", "v3", "https", "foreman.acme.io", "v3/hostl"},
	}
	for _, te := range tt {
		c := New(Options{
			Username:   te.username,
			Password:   te.password,
			Address:    te.address,
			APIVersion: te.apiVer,
			HTTPClient: newTestingHTTPClient(func(req *http.Request) (*http.Response, error) {
				u, p, k := req.BasicAuth()
				if k != te.basicAuth {
					t.Errorf("expected the request BasicAuth to be set to '%v' but got '%v'", te.basicAuth, k)
				}
				if u != te.username {
					t.Errorf("expected the Basic Auth username to be '%s' but got '%s'", te.username, u)
				}
				if p != te.password {
					t.Errorf("expected the Basic Auth password to be '%s' but got '%s'", te.password, p)
				}
				if req.URL.Scheme != te.scheme {
					t.Errorf("expected the request URL Scheme to be '%s' but got '%s'", te.scheme, req.URL.Scheme)
				}
				if req.URL.Host != te.host {
					t.Errorf("expected the request URL Host to be '%s' but got '%s'", te.host, req.URL.Host)
				}
				if req.URL.Path != te.path {
					t.Errorf("expected the request URL Path to be '%s' but got '%s'", te.path, req.URL.Path)
				}
				if req.Method != http.MethodHead {
					t.Errorf("expected the request Method to be '%s' but got '%s'", http.MethodHead, req.Method)
				}

				return &http.Response{}, nil
			}),
		})
		assertModifyNotNil(t, c)
		_, err := c.Head(te.resource)
		if err != nil {
			t.Errorf("expected Head not to return errors: %s", err)
		}
	}
}

func TestFailedHead(t *testing.T) {
	c := New(Options{})
	_, err := c.Head(":hehehe")
	if err == nil {
		t.Error("expected Head to return error")
	}
	expErr := "missing protocol scheme"
	if !strings.Contains(err.Error(), expErr) {
		t.Errorf("expected error message(%s) to contain '%s'", err.Error(), expErr)
	}
}
