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

	// 监听器方法，可用于日志处理、上报监控平台等操作
	_listeners = make([]ListenerFunc, 0)
)

//------------------------------------------------------------------------------

type Response struct {
	Err      error
	Val      []byte
	Request  *http.Request
	Response *http.Response
	Duration time.Duration
}

func (r *Response) responseHeader(key string) string {
	if r.Response == nil {
		return ""
	}
	return r.Response.Header.Get(key)
}

func (r *Response) Result() ([]byte, error) {
	if r.Err != nil {
		return nil, r.Err
	}
	if r.Response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("StatusCode=%v", r.Response.StatusCode)
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

type ListenerFunc func(method, url string, body io.Reader, resp *Response)

func Listen(listeners ...ListenerFunc) {
	_listeners = append(_listeners, listeners...)
}

//------------------------------------------------------------------------------

type Header map[string]string

type Options struct {
	Header Header
}

type Client interface {
	Request(method, url string, body io.Reader, options *Options) *Response
	Head(url string) *Response
	Get(url string) *Response
	Post(url string, body io.Reader) *Response
	PostForm(url string, body io.Reader) *Response
	PostJSON(url string, body io.Reader) *Response
}

type client struct {
	client *http.Client
}

func (c *client) Request(method, url string, body io.Reader, options *Options) (resp *Response) {
	resp = &Response{}
	defer func() {
		for _, listener := range _listeners {
			listener(method, url, body, resp)
		}
	}()
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		resp.Err = err
		return
	}
	if options != nil {
		for k, v := range options.Header {
			req.Header.Set(k, v)
		}
	}
	resp.Request = req
	beginTime := time.Now()
	retres, err := c.client.Do(req)
	resp.Duration = time.Now().Sub(beginTime)
	if err != nil {
		resp.Err = err
		return
	}
	defer retres.Body.Close()
	resp.Response = retres
	val, err := ioutil.ReadAll(retres.Body)
	if err != nil {
		resp.Err = err
		return
	}
	resp.Val = val
	return
}

func (c *client) Head(url string) *Response {
	return c.Request("HEAD", url, nil, nil)
}

func (c *client) Get(url string) *Response {
	return c.Request("GET", url, nil, nil)
}

func (c *client) Post(url string, body io.Reader) *Response {
	return c.Request("POST", url, body, nil)
}

func (c *client) PostForm(url string, body io.Reader) *Response {
	return c.Request("POST", url, body, &Options{
		Header: Header{"Content-Type": MIMEPOSTForm},
	})
}

func (c *client) PostJSON(url string, body io.Reader) *Response {
	return c.Request("POST", url, body, &Options{
		Header: Header{"Content-Type": MIMEJSON + ";charset=utf-8"},
	})
}

func NewClient(cli http.Client) Client {
	return &client{client: &cli}
}

//------------------------------------------------------------------------------

func Request(method, url string, body io.Reader, options *Options) *Response {
	return DefaultClient.Request(method, url, body, options)
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
