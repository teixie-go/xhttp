package xhttp

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

var (
	// DefaultClient is the default Client
	// and is used by Get, Post, PostForm, and PostJSON.
	DefaultClient = &Client{client: http.DefaultClient}

	// Global listeners
	_listeners = make([]ListenerFunc, 0)
)

//------------------------------------------------------------------------------

type Response struct {
	err      error
	result   []byte
	Request  *http.Request
	Response *http.Response
	Duration time.Duration
}

func (r *Response) Bind(o interface{}) error {
	if r.err != nil {
		return r.err
	}
	return json.Unmarshal(r.result, o)
}

func (r *Response) Result() ([]byte, error) {
	return r.result, r.err
}

//------------------------------------------------------------------------------

type ListenerFunc func(resp *Response)

// Add global listeners
func Listen(listeners ...ListenerFunc) {
	_listeners = append(_listeners, listeners...)
}

//------------------------------------------------------------------------------

type Header map[string]string

type Options struct {
	Header Header
}

type Client struct {
	client *http.Client
}

func (c *Client) Request(method, url string, body io.Reader, options *Options) *Response {
	resp := new(Response)
	defer func() {
		// dispatch listeners
		for _, listener := range _listeners {
			listener(resp)
		}
	}()
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		resp.err = err
		return resp
	}
	if options != nil {
		for k, v := range options.Header {
			req.Header.Set(k, v)
		}
	}
	resp.Request = req
	beginTime := time.Now()
	res, err := c.client.Do(req)
	resp.Duration = time.Now().Sub(beginTime)
	if err != nil {
		resp.err = err
		return resp
	}
	defer res.Body.Close()
	resp.Response = res
	result, err := ioutil.ReadAll(res.Body)
	if err != nil {
		resp.err = err
		return resp
	}
	resp.result = result
	return resp
}

func (c *Client) Get(url string) *Response {
	return c.Request("GET", url, nil, nil)
}

func (c *Client) Post(url, data string) *Response {
	return c.Request("POST", url, strings.NewReader(data), nil)
}

func (c *Client) PostForm(url string, data string) *Response {
	return c.Request("POST", url, strings.NewReader(data), &Options{
		Header: Header{"Content-Type": "application/x-www-form-urlencoded"},
	})
}

func (c *Client) PostJSON(url string, data string) *Response {
	return c.Request("POST", url, strings.NewReader(data), &Options{
		Header: Header{"Content-Type": "application/json;charset=utf-8"},
	})
}

//------------------------------------------------------------------------------

func Request(method, url string, body io.Reader, options *Options) *Response {
	return DefaultClient.Request(method, url, body, options)
}

func Get(url string) *Response {
	return DefaultClient.Get(url)
}

func Post(url, data string) *Response {
	return DefaultClient.Post(url, data)
}

func PostForm(url, data string) *Response {
	return DefaultClient.PostForm(url, data)
}

func PostJSON(url, data string) *Response {
	return DefaultClient.PostJSON(url, data)
}
