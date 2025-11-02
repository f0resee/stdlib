package rest

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	"golang.org/x/net/http2"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func NewRequest(c *RESTClient) *Request {
	r := &Request{
		c: c,
	}
	return r
}

type Request struct {
	c *RESTClient

	verb       string
	pathPrefix string
	params     url.Values
	headers    http.Header
	body       io.Reader

	err error
}

func (r *Request) Verb(verb string) *Request {
	r.verb = verb
	return r
}

func (r *Request) AbsPath(segments ...string) *Request {
	if r.err != nil {
		return r
	}
	r.pathPrefix = path.Join(r.c.base.Path, path.Join(segments...))
	if len(segments) == 1 && (len(r.c.base.Path) > 1 || len(segments[0]) > 1 && strings.HasPrefix(segments[0], "/")) {
		r.pathPrefix += "/"
	}
	return r
}

func (r *Request) Param(paramName, s string) *Request {
	if r.err != nil {
		return r
	}
	return r.setParam(paramName, s)
}

func (r *Request) setParam(paramName, value string) *Request {
	if r.params == nil {
		r.params = make(url.Values)
	}
	r.params[paramName] = append(r.params[paramName], value)
	return r
}

func (r *Request) SetHeader(key string, values ...string) *Request {
	if r.headers == nil {
		r.headers = http.Header{}
	}
	r.headers.Del(key)
	for _, value := range values {
		r.headers.Add(key, value)
	}
	return r
}

func (r *Request) URL() *url.URL {
	p := r.pathPrefix

	finalURL := &url.URL{}
	if r.c.base != nil {
		*finalURL = *r.c.base
	}
	finalURL.Path = p

	query := url.Values{}
	for key, values := range r.params {
		for _, value := range values {
			query.Add(key, value)
		}
	}
	finalURL.RawQuery = query.Encode()
	return finalURL
}

func (r *Request) newHTTPRequest(ctx context.Context) (*http.Request, error) {
	url := r.URL().String()
	req, err := http.NewRequestWithContext(ctx, r.verb, url, r.body)
	if err != nil {
		return nil, err
	}
	req.Header = r.headers
	return req, nil
}

func (r *Request) request(ctx context.Context, fn func(*http.Request, *http.Response)) error {
	client := r.c.Client
	if client == nil {
		client = http.DefaultClient
	}

	req, err := r.newHTTPRequest(ctx)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	f := func(req *http.Request, resp *http.Response) {
		if resp == nil {
			return
		}
		fn(req, resp)
	}
	f(req, resp)
	return nil
}

type Result struct {
	body        []byte
	contentType string
	err         error
	statusCode  int
}

func (r *Request) transformResponse(ctx context.Context, resp *http.Response, req *http.Request) Result {
	var body []byte
	if resp.Body != nil {
		data, err := io.ReadAll(resp.Body)
		switch err.(type) {
		case nil:
			body = data
		case http2.StreamError:
			streamErr := fmt.Errorf("Original error: %w", err)
			return Result{
				err: streamErr,
			}
		default:
			unexpectedErr := fmt.Errorf("Original error: %w", err)
			return Result{
				err: unexpectedErr,
			}
		}
	}

	if resp.StatusCode != http.StatusOK {
		return Result{
			body:        body,
			contentType: resp.Header.Get("Content-Type"),
			statusCode:  resp.StatusCode,
			err:         fmt.Errorf("status not ok"),
		}
	}

	return Result{
		body:        body,
		contentType: resp.Header.Get("Content-Type"),
		statusCode:  resp.StatusCode,
	}
}

func (r *Request) Do(ctx context.Context) Result {
	var result Result
	err := r.request(ctx, func(req *http.Request, resp *http.Response) {
		result = r.transformResponse(ctx, resp, req)
	})
	if err != nil {
		return Result{err: err}
	}
	return result
}
