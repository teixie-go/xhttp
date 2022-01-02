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
	Get(url string) *Response
	Post(url string, body io.Reader) *Response
	PostForm(url string, body io.Reader) *Response
	PostJSON(url string, body io.Reader) *Response
	Request(resolver func() (*http.Request, error)) *Response
}

type client struct {
	serve Serve
}

func (c *client) Request(resolver func() (*http.Request, error)) (resp *Response) {
	resp = &Response{}
	req, err := resolver()
	if err != nil {
		resp.Err = err
		return
	}
	w := httptest.NewRecorder()
	startTime := time.Now()
	c.serve.ServeHTTP(w, req)
	resp.Duration = time.Now().Sub(startTime)
	resp.Val = w.Body.Bytes()
	resp.Request = req
	resp.ResponseRecorder = w
	resp.RawResponse = w.Result()
	return
}

func (c *client) request(method, url string, body io.Reader, header http.Header) *Response {
	return c.Request(func() (*http.Request, error) {
		req, err := http.NewRequest(method, url, body)
		if err != nil {
			return nil, err
		}
		req.Header = header
		return req, nil
	})
}

func (c *client) Get(url string) *Response {
	return c.request("GET", url, nil, nil)
}

func (c *client) Post(url string, body io.Reader) *Response {
	return c.request("POST", url, body, nil)
}

func (c *client) PostForm(url string, body io.Reader) *Response {
	return c.request("POST", url, body, http.Header{
		"Content-Type": []string{xhttp.MIMEPOSTForm},
	})
}

func (c *client) PostJSON(url string, body io.Reader) *Response {
	return c.request("POST", url, body, http.Header{
		"Content-Type": []string{xhttp.MIMEJSON + ";charset=utf-8"},
	})
}

func NewClient(s Serve) Client {
	return &client{serve: s}
}
