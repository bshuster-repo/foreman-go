package foreman

import (
	"bytes"
	"encoding/json"
	"errors"
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

// Resource represents a resource in Foreman.
// Used to create, update or delete a specific resource such as hosts.
type Resource struct {
	Name       string
	ID         string
	Parameters interface{}
}

// Create returns the HTTP response from Foreman regarding resource
// creation that is specified in the resource argument `item`.
func (c Client) Create(item Resource) (*http.Response, error) {
	if item.Name == "" {
		return nil, errors.New("Name is mandatory")
	}
	// According to the Foreman API guide the JSON
	// have to start with a main key that is the singular form.
	mainKey := item.Name[:len(item.Name)-1]
	if item.Name == "media" {
		mainKey = "medium"
	}
	jsonBytes, err := json.Marshal(map[string]interface{}{
		mainKey: item.Parameters,
	})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, item.Name, bytes.NewReader(jsonBytes))
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}
