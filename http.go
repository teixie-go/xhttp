package xhttp

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

var (
	// DefaultClient is the default Client
	// and is used by Get, Post, PostForm, and PostJSON.
	DefaultClient = &Client{client: http.DefaultClient}

	// 监听器方法，可用于日志处理、上报监控平台等操作
	_listeners = make([]ListenerFunc, 0)
)

//------------------------------------------------------------------------------

type Response struct {
	val         []byte
	Err         error
	Method      string
	Url         string
	RequestBody string
	Request     *http.Request
	Response    *http.Response
	Duration    time.Duration
}

func (r *Response) Bind(obj interface{}) error {
	if r.Err != nil {
		return r.Err
	}
	return json.Unmarshal(r.val, obj)
}

func (r *Response) Result() ([]byte, error) {
	return r.val, r.Err
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

func (c *Client) Request(method, url, body string, options *Options) (resp *Response) {
	resp = &Response{Method: method, Url: url, RequestBody: body}
	defer func() {
		// dispatch listeners
		for _, listener := range _listeners {
			listener(resp)
		}
	}()
	req, err := http.NewRequest(method, url, strings.NewReader(body))
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
	if retres.StatusCode != http.StatusOK {
		resp.Err = fmt.Errorf("http StatusCode(%d)", retres.StatusCode)
		return
	}
	val, err := ioutil.ReadAll(retres.Body)
	if err != nil {
		resp.Err = err
		return
	}
	resp.val = val
	return
}

func (c *Client) Get(url string) *Response {
	return c.Request("GET", url, "", nil)
}

func (c *Client) Post(url, data string) *Response {
	return c.Request("POST", url, data, nil)
}

func (c *Client) PostForm(url string, data string) *Response {
	return c.Request("POST", url, data, &Options{
		Header: Header{"Content-Type": "application/x-www-form-urlencoded"},
	})
}

func (c *Client) PostJSON(url string, data string) *Response {
	return c.Request("POST", url, data, &Options{
		Header: Header{"Content-Type": "application/json;charset=utf-8"},
	})
}

//------------------------------------------------------------------------------

func Request(method, url, body string, options *Options) *Response {
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
