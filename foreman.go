package foreman

import (
	"net/http"
	"net/url"
	"strings"
)

var (
	defaultHTTPClient = &http.Client{}
	defaultModifier   = mdfyFunc(func(req *http.Request) (*http.Request, error) {
		return req, nil
	})
	defaultAddress    = "http://localhost:3000"
	defaultAPIVersion = "v2"
)

// Client represents the Foreman client.
type Client struct {
	httpClient *http.Client
	mods       modifier
}

// Options represents options to configure a Foreman client.
type Options struct {
	// Foreman address and API version
	Address    string
	APIVersion string

	// Foreman basic auth credentials
	Username string
	Password string

	// Use a specific HTTP Client
	HTTPClient *http.Client
}

// New setups a new Foreman client from the given options `opts` and returns it.
func New(opts Options) Client {
	client := Client{httpClient: defaultHTTPClient}
	if opts.HTTPClient != nil {
		client.httpClient = opts.HTTPClient
	}
	if opts.Address == "" {
		opts.Address = defaultAddress
	}
	if opts.APIVersion == "" {
		opts.APIVersion = defaultAPIVersion
	}
	mod := modifier(defaultModifier)
	mods := []modifierDecorator{
		setURLHostModifier(opts.Address, opts.APIVersion),
		addHeaderModifier("Accept", "application/json"),
		addHeaderModifier("Content-Type", "application/json"),
		addHeaderModifier("User-Agent", "GoForemanAPIClient"),
	}
	if opts.Username != "" {
		mods = append(mods, setBasicAuthModifier(opts.Username, opts.Password))
	}
	for _, m := range mods {
		mod = m(mod)
	}
	client.mods = mod
	return client
}

// Do sends an HTTP request and returns its HTTP response.
// If there was a problem during that process, an error is returned.
func (c Client) Do(req *http.Request) (*http.Response, error) {
	nReq, err := c.mods.Modify(req)
	if err != nil {
		return &http.Response{}, err
	}
	return c.httpClient.Do(nReq)
}

type modifier interface {
	Modify(*http.Request) (*http.Request, error)
}

type mdfyFunc func(*http.Request) (*http.Request, error)

func (m mdfyFunc) Modify(req *http.Request) (*http.Request, error) {
	return m(req)
}

type modifierDecorator func(modifier) modifier

func newModifierDecorator(ops func(*http.Request) (*http.Request, error)) modifierDecorator {
	return func(m modifier) modifier {
		return mdfyFunc(func(req *http.Request) (*http.Request, error) {
			newReq, err := ops(req)
			if err != nil {
				return nil, err
			}
			return m.Modify(newReq)
		})
	}
}

func addHeaderModifier(key, value string) modifierDecorator {
	return newModifierDecorator(func(req *http.Request) (*http.Request, error) {
		req.Header.Add(key, value)
		return req, nil
	})
}

func setURLHostModifier(address, apiVersion string) modifierDecorator {
	return newModifierDecorator(func(req *http.Request) (*http.Request, error) {
		url, err := url.Parse(address + "/api/" + apiVersion + "/" + strings.TrimPrefix(req.URL.Path, "/"))
		if err != nil {
			return req, err
		}
		url.RawQuery = req.URL.RawQuery
		req.URL = url
		return req, nil
	})
}

func setBasicAuthModifier(username, password string) modifierDecorator {
	return newModifierDecorator(func(req *http.Request) (*http.Request, error) {
		req.SetBasicAuth(username, password)
		return req, nil
	})
}
