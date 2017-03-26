package foreman

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
)

// Query represents a search query.
// Used to get resource list such as hosts and architectures.
//
// parmas := url.Values{}
// params.Add("search", "host1.local.io")
// q := foreman.Query{"hosts", params}
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
// Used for creating, updating or removing  a specific resource such as hosts.
//
// This is an example for creating a new architecture called "ARM"
// that is associated with operating system ids: 1 and 2:
// params := map[string]interface{}{
// 		"name": "ARM",
//      "operatingsystem_ids": int[]{1,2},
// }
// r := foreman.Resource{Name: "architectures", Parameters: params}
// resp, err := client.Create(r)
//
// This example updates the architecture from above:
// params := map[string]interface{}{
// 		"name": "ARM",
//      "operatingsystem_ids": int[]{1},
// }
// r := foreman.Resource{Name: "architectures", ID: "ARM", Parameters: params}
// resp, err := client.Update(r)
//
// This example removes the architecture "ARM":
// r := foreman.Resource{Name: "architectures", ID: "ARM"}
// resp, err := client.Delete(r)
type Resource struct {
	Name       string
	ID         string
	Parameters interface{}
}

// Create returns the HTTP response from Foreman regarding resource
// creation that is specified in the resource argument `item`.
func (c Client) Create(item Resource) (*http.Response, error) {
	jsonBytes, err := json.Marshal(item.Parameters)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, item.Name, bytes.NewReader(jsonBytes))
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

// Update returns the HTTP response from Foreman regarding updating
// a resource which is specified in the resource argument `item`.
func (c Client) Update(item Resource) (*http.Response, error) {
	jsonBytes, err := json.Marshal(item.Parameters)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPut, item.Name+"/"+item.ID, bytes.NewReader(jsonBytes))
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

// Delete returns the HTTP response from Foreman regarding removing
// a resource which is specified in the resource argument `item`.
func (c Client) Delete(item Resource) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodDelete, item.Name+"/"+item.ID, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}
