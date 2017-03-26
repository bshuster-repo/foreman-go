package foreman

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"reflect"
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

func TestCreate(t *testing.T) {
	tt := []struct {
		name       string
		parameters map[string]interface{}
		expected   map[string]interface{}
	}{
		{"architectures", map[string]interface{}{"name": "arch"}, map[string]interface{}{"name": "arch"}},
		{"media", map[string]interface{}{"name": "Arch"}, map[string]interface{}{"name": "Arch"}},
	}
	c := New(Options{})

	fURL := defaultAddress + "/api/" + defaultAPIVersion
	for _, te := range tt {
		c.httpClient = newTestingHTTPClient(func(req *http.Request) (*http.Response, error) {
			if req.URL.String() != fURL+"/"+te.name {
				t.Errorf("expected request URL to be '%s' but got '%s'", fURL+"/"+te.name, req.URL)
			}
			if req.Method != http.MethodPost {
				t.Errorf("expected request Method to be '%s' but got '%s'", http.MethodPost, req.Method)
			}
			defer req.Body.Close()
			var jsonMap map[string]interface{}
			if err := json.NewDecoder(req.Body).Decode(&jsonMap); err != nil {
				t.Errorf("expected json.NewDecoder.Decode to not return error: %s", err)
			}
			if !reflect.DeepEqual(jsonMap, te.expected) {
				t.Errorf("expected body to be '%s' but got '%s'", te.expected, jsonMap)
			}
			return &http.Response{}, nil
		})
		c.Create(Resource{Name: te.name, Parameters: te.parameters})
	}
}

type testFailedMarshaler map[string]string

func (tfm testFailedMarshaler) MarshalJSON() ([]byte, error) {
	return nil, errors.New("")
}

func TestAPIError(t *testing.T) {
	c := New(Options{})
	tt := []struct {
		resource Resource
		expErr   string
		apiF     func(Resource) (*http.Response, error)
	}{
		{Resource{Name: ":bla"}, "missing protocol scheme", c.Create},
		{Resource{Name: ":bla"}, "missing protocol scheme", c.Update},
		{Resource{Name: ":bla"}, "missing protocol scheme", c.Delete},
		{Resource{Name: "bla", Parameters: testFailedMarshaler(map[string]string{"h": "h"})}, "json: error calling MarshalJSON", c.Create},
		{Resource{Name: "bla", Parameters: testFailedMarshaler(map[string]string{"h": "h"})}, "json: error calling MarshalJSON", c.Update},
	}
	c.httpClient = newTestingHTTPClient(func(req *http.Request) (*http.Response, error) { return &http.Response{}, nil })
	for _, te := range tt {
		_, err := te.apiF(te.resource)
		if err == nil {
			t.Error("expected api function to return error")
			t.FailNow()
		}
		if !strings.Contains(err.Error(), te.expErr) {
			t.Errorf("expected error to contain '%s' but got '%s'", te.expErr, err)
		}
	}
}

func TestUpdate(t *testing.T) {
	tt := []struct {
		name       string
		id         string
		parameters map[string]interface{}
		expected   map[string]interface{}
	}{
		{"architectures", "1", map[string]interface{}{"name": "arch"}, map[string]interface{}{"name": "arch"}},
		{"media", "Arch", map[string]interface{}{"name": "Arch2016"}, map[string]interface{}{"name": "Arch2016"}},
	}
	c := New(Options{})

	fURL := defaultAddress + "/api/" + defaultAPIVersion
	for _, te := range tt {
		c.httpClient = newTestingHTTPClient(func(req *http.Request) (*http.Response, error) {
			if req.URL.String() != fURL+"/"+te.name+"/"+te.id {
				t.Errorf("expected request URL to be '%s' but got '%s'", fURL+"/"+te.name+"/"+te.id, req.URL)
			}
			if req.Method != http.MethodPut {
				t.Errorf("expected request Method to be '%s' but got '%s'", http.MethodPut, req.Method)
			}
			defer req.Body.Close()
			var jsonMap map[string]interface{}
			if err := json.NewDecoder(req.Body).Decode(&jsonMap); err != nil {
				t.Errorf("expected json.NewDecoder.Decode to not return error: %s", err)
			}
			if !reflect.DeepEqual(jsonMap, te.expected) {
				t.Errorf("expected body to be '%s' but got '%s'", te.expected, jsonMap)
			}
			return &http.Response{}, nil
		})
		c.Update(Resource{Name: te.name, ID: te.id, Parameters: te.parameters})
	}
}

func TestDelete(t *testing.T) {
	tt := []struct {
		name string
		id   string
	}{
		{"architectures", "1"},
		{"media", "Arch"},
	}
	c := New(Options{})

	fURL := defaultAddress + "/api/" + defaultAPIVersion
	for _, te := range tt {
		c.httpClient = newTestingHTTPClient(func(req *http.Request) (*http.Response, error) {
			if req.URL.String() != fURL+"/"+te.name+"/"+te.id {
				t.Errorf("expected request URL to be '%s' but got '%s'", fURL+"/"+te.name+"/"+te.id, req.URL)
			}
			if req.Method != http.MethodDelete {
				t.Errorf("expected request Method to be '%s' but got '%s'", http.MethodDelete, req.Method)
			}
			return &http.Response{}, nil
		})
		c.Delete(Resource{Name: te.name, ID: te.id})
	}
}
