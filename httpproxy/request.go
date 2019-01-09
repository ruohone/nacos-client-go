package httpproxy

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"strings"
)

type Request struct {
	req    *http.Request
	client *Client
	err    error
}

func (r *Request) SetUrl(l string) *Request {
	u, err := url.Parse(l)
	if err != nil {
		r.err = err
		return r
	}
	r.req.URL = u
	r.req.Host = u.Host
	return r
}

func (r *Request) Get(url string) *Request {
	if r.req.Method == "" {
		r.req.Method = http.MethodGet
	}
	return r.SetUrl(url)
}

func (r *Request) Post(url string) *Request {
	if r.req.Method == "" {
		r.req.Method = http.MethodPost
	}
	return r.SetUrl(url)
}

func (r *Request) WithJsonBody(v interface{}) *Request {
	r.WithHeader("Content-Type", "application/json")
	if v != nil {
		var data []byte
		switch vv := v.(type) {
		case []byte:
			data = vv
		case json.RawMessage:
			data = vv
		case string:
			data = []byte(vv)
		default:
			var err error
			data, err = json.Marshal(v)
			if err != nil {
				r.err = err
			}
		}
		if len(data) > 0 {
			return r.WithBody(bytes.NewReader(data))
		}
	}
	return r
}

func (r *Request) WithFormBody(body url.Values) *Request {
	r.WithHeader("Content-Type", "application/x-www-form-urlencoded")
	return r.WithBody(strings.NewReader(body.Encode()))
}

func (r *Request) WithBody(body io.Reader) *Request {
	rc, ok := body.(io.ReadCloser)
	if !ok && body != nil {
		rc = ioutil.NopCloser(body)
	}
	r.req.Body = rc
	if body != nil {
		switch v := body.(type) {
		case *bytes.Buffer:
			r.req.ContentLength = int64(v.Len())
			buf := v.Bytes()
			r.req.GetBody = func() (io.ReadCloser, error) {
				rd := bytes.NewReader(buf)
				return ioutil.NopCloser(rd), nil
			}
		case *bytes.Reader:
			r.req.ContentLength = int64(v.Len())
			snapshot := *v
			r.req.GetBody = func() (io.ReadCloser, error) {
				rd := snapshot
				return ioutil.NopCloser(&rd), nil
			}
		case *strings.Reader:
			r.req.ContentLength = int64(v.Len())
			snapshot := *v
			r.req.GetBody = func() (io.ReadCloser, error) {
				rd := snapshot
				return ioutil.NopCloser(&rd), nil
			}
		default:
		}
		if r.req.GetBody != nil && r.req.ContentLength == 0 {
			r.req.Body = http.NoBody
			r.req.GetBody = func() (io.ReadCloser, error) { return http.NoBody, nil }
		}
	}
	return r
}

func (r *Request) WithRequestID(requestID string) *Request {
	return r.WithHeader("X-Request-Id", requestID)
}

func (r *Request) WithCookie(c *http.Cookie) *Request {
	r.req.AddCookie(c)
	return r
}

func (r *Request) WithQuery(k, v string) *Request {
	if k != "" && v != "" {
		q := r.req.URL.Query()
		q.Set(k, v)
		r.req.URL.RawQuery = q.Encode()
	}
	return r
}

func (r *Request) WithQuerys(querys url.Values) *Request {
	if len(querys) > 0 {
		q := r.req.URL.Query()
		for k, vv := range querys {
			for _, v := range vv {
				q.Add(k, v)
			}
		}
		r.req.URL.RawQuery = q.Encode()
	}
	return r
}

func (r *Request) WithTrace(trace *httptrace.ClientTrace) *Request {
	if trace != nil {
		ctx := httptrace.WithClientTrace(r.req.Context(), trace)
		r.req = r.req.WithContext(ctx)
	}
	return r
}

func (r *Request) WithHeader(k, v string) *Request {
	if k != "" && v != "" {
		r.req.Header.Set(k, v)
	}
	return r
}

func (r *Request) Build() (*http.Request, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.req, nil
}

func (r *Request) Execute() *Response {
	req, err := r.Build()
	if err != nil {
		return &Response{nil, err}
	}
	return r.client.Do(req)
}

func (r *Request) ExecuteAsJson(v interface{}) error {
	rsp := r.Execute()
	return rsp.ToJson(v)
}

func (r *Request) ExecuteAsString() (string, error) {
	rsp := r.Execute()
	return rsp.ToString()
}
