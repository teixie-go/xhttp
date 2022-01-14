package xhttp

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/teixie-go/xhttp/binding"
)

const (
	MIMEJSON     = "application/json"
	MIMEXML      = "application/xml"
	MIMEXML2     = "text/xml"
	MIMEPOSTForm = "application/x-www-form-urlencoded"
)

var (
	_ Client = (*client)(nil)

	// The default Client and is used by Head, Get, Post, PostForm, and PostJSON.
	DefaultClient = &client{client: http.DefaultClient}

	// The global middleware.
	_middleware Middleware
)

//------------------------------------------------------------------------------

type Response struct {
	Err         error
	Val         []byte
	Request     *http.Request
	RawResponse *http.Response
	Duration    time.Duration
}

func (r *Response) responseHeader(key string) string {
	if r.RawResponse == nil {
		return ""
	}
	return r.RawResponse.Header.Get(key)
}

func (r *Response) Result() ([]byte, error) {
	if r.Err != nil {
		return nil, r.Err
	}
	if r.RawResponse.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("StatusCode=%v", r.RawResponse.StatusCode)
	}
	return r.Val, nil
}

func (r *Response) BindWith(obj interface{}, b binding.Binding) error {
	val, err := r.Result()
	if err != nil {
		return err
	}
	return b.Bind(val, obj)
}

func (r *Response) BindJSON(obj interface{}) error {
	return r.BindWith(obj, binding.JSON)
}

func (r *Response) BindXML(obj interface{}) error {
	return r.BindWith(obj, binding.XML)
}

func (r *Response) Bind(obj interface{}) error {
	contentType := stripFlags(r.responseHeader("Content-Type"))
	switch contentType {
	case MIMEXML, MIMEXML2:
		return r.BindXML(obj)
	}
	return r.BindJSON(obj)
}

func stripFlags(str string) string {
	for i, char := range str {
		if char == ' ' || char == ';' {
			return str[:i]
		}
	}
	return str
}

//------------------------------------------------------------------------------

type Handler func(method, url string, body io.Reader, request *http.Request) *Response

type Middleware func(Handler) Handler

type Client interface {
	Head(url string) *Response
	Get(url string) *Response
	Post(url string, body io.Reader) *Response
	PostForm(url string, body io.Reader) *Response
	PostJSON(url string, body io.Reader) *Response
	Request(method, url string, body io.Reader, resolver func() (*http.Request, error)) *Response
}

type client struct {
	client     *http.Client
	middleware Middleware
}

type Option func(*client)

func (c *client) Request(method, url string, body io.Reader, resolver func() (*http.Request, error)) *Response {
	req, err := resolver()
	resp := &Response{Err: err}
	h := func(method, url string, body io.Reader, request *http.Request) *Response {
		if resp.Err != nil {
			return resp
		}
		resp.Request = request
		beginTime := time.Now()
		retres, err := c.client.Do(request)
		resp.Duration = time.Now().Sub(beginTime)
		if err != nil {
			resp.Err = err
			return resp
		}
		defer retres.Body.Close()
		resp.RawResponse = retres
		val, err := ioutil.ReadAll(retres.Body)
		if err != nil {
			resp.Err = err
			return resp
		}
		resp.Val = val
		return resp
	}
	if c.middleware != nil {
		h = c.middleware(h)
	}
	if _middleware != nil {
		h = _middleware(h)
	}
	return h(method, url, body, req)
}

func (c *client) request(method, url string, body io.Reader, header http.Header) *Response {
	return c.Request(method, url, body, func() (*http.Request, error) {
		req, err := http.NewRequest(method, url, body)
		if err != nil {
			return nil, err
		}
		req.Header = header
		return req, nil
	})
}

func (c *client) Head(url string) *Response {
	return c.request(http.MethodHead, url, nil, nil)
}

func (c *client) Get(url string) *Response {
	return c.request(http.MethodGet, url, nil, nil)
}

func (c *client) Post(url string, body io.Reader) *Response {
	return c.request(http.MethodPost, url, body, nil)
}

func (c *client) PostForm(url string, body io.Reader) *Response {
	return c.request(http.MethodPost, url, body, http.Header{
		"Content-Type": []string{MIMEPOSTForm},
	})
}

func (c *client) PostJSON(url string, body io.Reader) *Response {
	return c.request(http.MethodHead, url, body, http.Header{
		"Content-Type": []string{MIMEJSON + ";charset=utf-8"},
	})
}

func NewClient(cli http.Client, opts ...Option) Client {
	client := &client{client: &cli}
	for _, o := range opts {
		o(client)
	}
	return client
}

func withMiddlewareChain(chain Middleware, middleware ...Middleware) Middleware {
	if len(middleware) == 0 {
		return chain
	}
	if chain == nil {
		return withMiddlewareChain(middleware[0], middleware[1:]...)
	}
	return func(next Handler) Handler {
		for i := len(middleware) - 1; i >= 0; i-- {
			next = middleware[i](next)
		}
		return chain(next)
	}
}

func WithMiddleware(middleware ...Middleware) Option {
	return func(c *client) {
		c.middleware = withMiddlewareChain(c.middleware, middleware...)
	}
}

//------------------------------------------------------------------------------

func Request(method, url string, body io.Reader, resolver func() (*http.Request, error)) *Response {
	return DefaultClient.Request(method, url, body, resolver)
}

func Head(url string) *Response {
	return DefaultClient.Head(url)
}

func Get(url string) *Response {
	return DefaultClient.Get(url)
}

func Post(url string, body io.Reader) *Response {
	return DefaultClient.Post(url, body)
}

func PostForm(url string, body io.Reader) *Response {
	return DefaultClient.PostForm(url, body)
}

func PostJSON(url string, body io.Reader) *Response {
	return DefaultClient.PostJSON(url, body)
}

func Use(middleware ...Middleware) {
	_middleware = withMiddlewareChain(_middleware, middleware...)
}
