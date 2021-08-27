package httptest

import (
	"io"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/teixie-go/xhttp"
)

var (
	_ Client = (*client)(nil)
)

type Serve interface {
	ServeHTTP(w http.ResponseWriter, req *http.Request)
}

type Response struct {
	xhttp.Response
	ResponseRecorder *httptest.ResponseRecorder
}

type Client interface {
	Request(method, url string, body io.Reader, options *xhttp.Options) *Response
	Get(url string) *Response
	Post(url string, body io.Reader) *Response
	PostForm(url string, body io.Reader) *Response
	PostJSON(url string, body io.Reader) *Response
}

type client struct {
	serve Serve
}

func (c *client) Request(method, url string, body io.Reader, options *xhttp.Options) (resp *Response) {
	resp = &Response{}
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
	w := httptest.NewRecorder()
	startTime := time.Now()
	c.serve.ServeHTTP(w, req)
	resp.Duration = time.Now().Sub(startTime)
	resp.Val = w.Body.Bytes()
	resp.Request = req
	resp.ResponseRecorder = w
	resp.Response.Response = w.Result()
	return
}

func (c *client) Get(url string) *Response {
	return c.Request("GET", url, nil, nil)
}

func (c *client) Post(url string, body io.Reader) *Response {
	return c.Request("POST", url, body, nil)
}

func (c *client) PostForm(url string, body io.Reader) *Response {
	return c.Request("POST", url, body, &xhttp.Options{
		Header: xhttp.Header{"Content-Type": xhttp.MIMEPOSTForm},
	})
}

func (c *client) PostJSON(url string, body io.Reader) *Response {
	return c.Request("POST", url, body, &xhttp.Options{
		Header: xhttp.Header{"Content-Type": xhttp.MIMEJSON + ";charset=utf-8"},
	})
}

func NewClient(s Serve) Client {
	return &client{serve: s}
}
