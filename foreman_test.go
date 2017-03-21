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

func setRequest(t *testing.T, method, resource string) *http.Request {
	req, err := http.NewRequest(method, resource, nil)
	if err != nil {
		t.Errorf("expected the request creation not to return error: %s", err)
	}
	return req
}

func setTestRequest(t *testing.T) *http.Request {
	return setRequest(t, http.MethodGet, "")
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
		{"User-Agent", "GoForemanAPIClient"},
	}
	c := New(Options{})
	assertModifyNotNil(t, c)
	req, err := c.mods.Modify(setTestRequest(t))
	if err != nil {
		t.Errorf("expected Modify to return error: %s", err)
	}
	for _, te := range tt {
		if req.Header.Get(te.key) != te.val {
			t.Errorf("expected request %s to be '%s' but got '%s'", te.key, te.val, req.Header.Get(te.key))
		}
	}
}

func TestNewDefaultsIsCorrect(t *testing.T) {
	c := New(Options{})
	assertModifyNotNil(t, c)
	req, err := c.mods.Modify(setTestRequest(t))
	if err != nil {
		t.Errorf("expected Modify to return error: %s", err)
	}
	if err != nil {
		t.Errorf("expected Modify to return error: %s", err)
	}
	expURL := strings.TrimPrefix(defaultAddress, "http://")
	if req.URL.Host != expURL {
		t.Errorf("expected request URL Host to be '%s' but got '%s'", expURL, req.URL.Host)
	}
	if req.URL.Scheme != "http" {
		t.Errorf("expected request URL Scheme to be 'http' but got '%s'", req.URL.Scheme)
	}
	if req.URL.Path != "/api/"+defaultAPIVersion+"/" {
		t.Errorf("expected request URL Path to be '%s' but got '%s'", "/api/"+defaultAPIVersion+"/", req.URL.Path)
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
		{"http://10.20.30.40:8989", "", "http", "10.20.30.40:8989", "/api/v2/"},
		{"https://172.20.30.40:8989", "", "https", "172.20.30.40:8989", "/api/v2/"},
		{"https://foreman.internal.acme.io:8989", "", "https", "foreman.internal.acme.io:8989", "/api/v2/"},
		{"https://foreman.acme.io:80", "v5", "https", "foreman.acme.io:80", "/api/v5/"},
	}
	for _, te := range tt {
		c := New(Options{
			Address:    te.address,
			APIVersion: te.apiVer,
		})
		assertModifyNotNil(t, c)
		req, err := c.mods.Modify(setTestRequest(t))
		if err != nil {
			t.Errorf("expected Modify to not return error: %s", err)
			t.FailNow()
		}
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

func TestBadAddress(t *testing.T) {
	opts := Options{
		Address: "10.20.30.40:8989",
	}
	c := New(opts)
	assertModifyNotNil(t, c)
	if _, err := c.mods.Modify(setTestRequest(t)); err == nil {
		t.Error("expected Modify to return error")
	}

	if _, err := c.Do(setTestRequest(t)); err == nil {
		t.Error("expected Modify to return error")
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
		req, err := c.mods.Modify(setTestRequest(t))
		if err != nil {
			t.Errorf("expected Modify to return error: %s", err)
		}
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

func TestClientDo(t *testing.T) {
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
		{"hosts", "admin", "newpass1", true, "http://foreman.acme.io", "", "http", "foreman.acme.io", "/api/v2/hosts"},
		{"", "admin", "", true, "https://foreman.acme.io", "v3", "https", "foreman.acme.io", "/api/v3/"},
		{"", "", "", false, "https://foreman.acme.io", "v3", "https", "foreman.acme.io", "/api/v3/"},
		{"http://hello.world.com/hostl", "", "", false, "https://foreman.acme.io", "v3", "https", "foreman.acme.io", "/api/v3/hostl"},
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
		_, err := c.Do(setRequest(t, http.MethodHead, te.resource))
		if err != nil {
			t.Errorf("expected Head not to return errors: %s", err)
		}
	}
}
