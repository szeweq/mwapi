package mwapi

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"sync"

	jsoniter "github.com/json-iterator/go"
)

type (
	Client struct {
		url   *url.URL
		uname string
		paswd string
		htcl  *http.Client
		login bool
	}
	Response struct {
		v jsoniter.Any
	}
	Values map[string]string
)

var (
	poolRequest = sync.Pool{
		New: func() interface{} {
			return &http.Request{
				Proto:      "HTTP/1.1",
				ProtoMajor: 1,
				ProtoMinor: 1,
			}
		},
	}
	poolBuffer = sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(nil)
		},
	}
)

func borrowRequest(method, host string) *http.Request {
	r := poolRequest.Get().(*http.Request)
	r.Method = method
	r.Host = host
	return r
}
func returnRequest(r *http.Request) {
	r.ContentLength = 0
	r.Body = nil
	r.Header = nil
	poolRequest.Put(r)
}

func borrowBuffer() *bytes.Buffer {
	return poolBuffer.Get().(*bytes.Buffer)
}
func returnBuffer(b *bytes.Buffer) {
	b.Reset()
	poolBuffer.Put(b)
}

func (mr *Response) Get(to interface{}, path ...interface{}) error {
	x := mr.v.Get(path...)
	if e := mr.v.LastError(); e != nil {
		return e
	}
	if x.ValueType() == jsoniter.InvalidValue {
		var b jsoniter.RawMessage
		mr.v.ToVal(&b)
		return fmt.Errorf("Invalid value %s on %s", path, b)
	}
	x.ToVal(to)
	return nil
}

func (mw *Client) request(rq *http.Request) (*Response, error) {
	rs, re := mw.htcl.Do(rq)
	if re != nil {
		panic(re)
	}
	defer rs.Body.Close()
	bb, be := ioutil.ReadAll(rs.Body)
	rs.Body.Close()
	if be != nil {
		return nil, be
	}
	ja := jsoniter.Get(bb)
	if e := ja.LastError(); e != nil {
		return nil, e
	}
	je := ja.Get("error")
	if je.ValueType() == jsoniter.ObjectValue {
		var mae MWAPIError
		je.ToVal(&mae)
		return &Response{jsoniter.Wrap(nil)}, mae
	}
	return &Response{ja}, nil
}

func (mw *Client) Get(v Values) (*Response, error) {
	bb := borrowBuffer()
	defer returnBuffer(bb)
	EncodeValue(bb, v)
	rq := &http.Request{
		Method: "GET",
		URL: &url.URL{
			Scheme:     mw.url.Scheme,
			Host:       mw.url.Host,
			Path:       mw.url.Path,
			ForceQuery: true,
			RawQuery:   bb.String(),
		},
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
		Host:       mw.url.Host,
	}
	return mw.request(rq)
}

func (mw *Client) Post(v Values) (*Response, error) {
	bb := borrowBuffer()
	defer returnBuffer(bb)
	EncodeValue(bb, v)
	rq := borrowRequest("POST", mw.url.Host)
	defer returnRequest(rq)
	rq.URL = mw.url
	rq.Header = http.Header{
		"Content-Type": {"application/x-www-form-urlencoded"},
	}
	rq.ContentLength = int64(bb.Len())
	rq.Body = ioutil.NopCloser(bb)
	return mw.request(rq)
}

func NewClient(apiphp string, user string, pass string) (*Client, error) {
	urlx, e := url.Parse(apiphp)
	if e != nil {
		return nil, e
	}
	cj, ce := cookiejar.New(nil)
	if ce != nil {
		return nil, e
	}
	return &Client{urlx, user, pass, &http.Client{Jar: cj}, false}, nil
}
