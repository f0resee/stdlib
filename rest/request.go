package rest

import (
	"context"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"golang.org/x/net/http2"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func NewRequest(c *RESTClient) *Request {
	var timeout time.Duration
	if c.Client != nil {
		timeout = c.Client.Timeout
	}
	r := &Request{
		c:       c,
		timeout: timeout,
	}
	return r
}

type Request struct {
	c *RESTClient

	timeout time.Duration

	verb       string
	pathPrefix string
	subpath    string
	params     url.Values
	headers    http.Header

	body      io.Reader
	bodyBytes []byte

	err error
}

func (r *Request) Verb(verb string) *Request {
	r.verb = verb
	return r
}

func (r *Request) Prefix(segments ...string) *Request {
	if r.err != nil {
		return r
	}
	r.pathPrefix = path.Join(r.pathPrefix, path.Join(segments...))
	return r
}

func (r *Request) Suffix(segments ...string) *Request {
	if r.err != nil {
		return r
	}
	r.subpath = path.Join(r.subpath, path.Join(segments...))
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

func (r *Request) RequestURI(uri string) *Request {
	if r.err != nil {
		return r
	}
	locator, err := url.Parse(uri)
	if err != nil {
		r.err = err
		return r
	}
	r.AbsPath(locator.Path)
	if len(locator.Query()) > 0 {
		r.params = make(url.Values)
		maps.Copy(r.params, locator.Query())
	} else {
		r.params = nil
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

func (r *Request) Timeout(d time.Duration) *Request {
	if r.err != nil {
		return r
	}
	r.timeout = d
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

func (r *Request) Body(obj interface{}) *Request {
	if r.err != nil {
		return r
	}
	switch t := obj.(type) {
	case string:
		data, err := os.ReadFile(t)
		if err != nil {
			r.err = err
			return r
		}
		r.body = nil
		r.bodyBytes = data
	case []byte:
		r.body = nil
		r.bodyBytes = t
	case io.Reader:
		r.body = t
		r.bodyBytes = nil
	default:
		r.err = fmt.Errorf("unknown type used for body: %+v", obj)
	}
	return r
}

func (r *Request) Error() error {
	return r.err
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

	if r.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, r.timeout)
		defer cancel()
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

func (r *Request) transformResponse(_ context.Context, resp *http.Response, _ *http.Request) Result {
	var body []byte
	if resp.Body != nil {
		data, err := io.ReadAll(resp.Body)
		switch err.(type) {
		case nil:
			body = data
		case http2.StreamError:
			streamErr := fmt.Errorf("original error: %w", err)
			return Result{
				err: streamErr,
			}
		default:
			unexpectedErr := fmt.Errorf("original error: %w", err)
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
