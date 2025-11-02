package rest

import (
	"net/http"
	"net/url"
	"strings"
)

type IClient interface {
	Verb(verb string) *Request
	Get() *Request
	Post() *Request
	Put() *Request
	Delete() *Request
}

func NewRESTClient(baseURL *url.URL, client *http.Client) (*RESTClient, error) {
	base := *baseURL
	if !strings.HasSuffix(base.Path, "/") {
		base.Path += "/"
	}
	base.RawQuery = ""
	base.Fragment = ""

	return &RESTClient{
		base:   &base,
		Client: client,
	}, nil
}

type RESTClient struct {
	base *url.URL

	Client *http.Client
}

func (c *RESTClient) Verb(verb string) *Request {
	return NewRequest(c).Verb(verb)
}

func (c *RESTClient) Post() *Request {
	return c.Verb("POST")
}

func (c *RESTClient) Put() *Request {
	return c.Verb("PUT")
}

func (c *RESTClient) Get() *Request {
	return c.Verb("GET")
}

func (c *RESTClient) Delete() *Request {
	return c.Verb("DELETE")
}
