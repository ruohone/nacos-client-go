package httpproxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	*http.Client
}

var defaultClient = &Client{
	Client: &http.Client{
		Timeout: time.Second,
	},
}

func NewClient(c *http.Client) *Client {
	return &Client{
		Client: c,
	}
}

func NewRequest() *Request {
	return defaultClient.NewRequest()
}

func (c *Client) NewRequest() *Request {
	r := &http.Request{
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
	}
	req := &Request{
		req:    r,
		client: c,
	}
	return req
}

func Get(url string) *Response {
	return defaultClient.Get(url)
}

func (c *Client) Get(url string) *Response {
	rsp, err := c.Client.Get(url)
	return &Response{rsp, err}
}

func PostForm(url string, params url.Values) *Response {
	return defaultClient.PostForm(url, params)
}

func (c *Client) PostForm(url string, params url.Values) *Response {
	rsp, err := c.Client.PostForm(url, params)
	return &Response{rsp, err}
}

func PostJson(url string, v interface{}) *Response {
	return defaultClient.PostJson(url, v)
}

func (c *Client) PostJson(url string, v interface{}) *Response {
	data, err := json.Marshal(v)
	if err != nil {
		return &Response{nil, err}
	}
	rsp, err := c.Client.Post(url, "application/json", bytes.NewReader(data))
	return &Response{rsp, err}
}

func Post(url string, boydType string, body io.Reader) *Response {
	return defaultClient.Post(url, boydType, body)
}

func (c *Client) Post(url string, boydType string, body io.Reader) *Response {
	rsp, err := c.Client.Post(url, boydType, body)
	return &Response{rsp, err}
}

func Do(req *http.Request) *Response {
	return defaultClient.Do(req)
}

func (c *Client) Do(req *http.Request) *Response {
	rsp, err := c.Client.Do(req)
	return &Response{rsp, err}
}

type Response struct {
	rsp *http.Response
	err error
}

func (r *Response) Error() error {
	return r.err
}

func (r *Response) ToString() (string, error) {
	if r.rsp != nil {
		defer r.rsp.Body.Close()
	}
	if r.err != nil {
		return "", r.err
	}
	if r.rsp.StatusCode != 200 {
		message, _ := ioutil.ReadAll(r.rsp.Body)
		return "", fmt.Errorf("Status:%d,Message:%s", r.rsp.Status, message)
	}
	data, err := ioutil.ReadAll(r.rsp.Body)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (r *Response) ToJson(v interface{}) error {
	if r.rsp != nil {
		defer r.rsp.Body.Close()
	}
	if r.err != nil {
		return r.err
	}
	if r.rsp.StatusCode != 200 {
		message, _ := ioutil.ReadAll(r.rsp.Body)
		return fmt.Errorf("Status:%d,Message:%s", r.rsp.Status, message)
	}
	err := json.NewDecoder(r.rsp.Body).Decode(v)
	if err != nil {
		return err
	}
	return nil
}
