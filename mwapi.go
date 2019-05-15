package mwapi

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"sync"

	jsoniter "github.com/json-iterator/go"
)

type (
	//Client handles all requests to MediaWiki API
	Client struct {
		url   *url.URL
		uname string
		paswd string
		htcl  *http.Client
		login bool
		LoginType
	}

	//Values hold query arguments for an API call
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

func (mw *Client) request(rq *http.Request) (*Response, error) {
	rs, re := mw.htcl.Do(rq)
	if re != nil {
		panic(re)
	}
	b := borrowBuffer()
	defer returnBuffer(b)
	if rs.ContentLength > 0 {
		b.Grow(int(rs.ContentLength))
	}
	_, be := b.ReadFrom(rs.Body)
	_ = rs.Body.Close()
	_ = rs.Body.Close()
	if be != nil {
		return nil, be
	}
	ja := jsoniter.Get(b.Bytes())
	if e := ja.LastError(); e != nil {
		return nil, e
	}
	je := ja.Get("error")
	if je.ValueType() == jsoniter.ObjectValue {
		var mae Error
		je.ToVal(&mae)
		return &Response{v: jsoniter.Wrap(nil)}, mae
	}
	return &Response{v: ja}, nil
}

//Get handles GET request with specified values
func (mw *Client) Get(v Values) (*Response, error) {
	bb := borrowBuffer()
	defer returnBuffer(bb)
	encodeValue(bb, v)
	rq := borrowRequest("GET", mw.url.Host)
	rq.URL = &url.URL{
		Scheme:     mw.url.Scheme,
		Host:       mw.url.Host,
		Path:       mw.url.Path,
		ForceQuery: true,
		RawQuery:   bb.String(),
	}
	rq.Header = make(http.Header)
	return mw.request(rq)
}

//Post handles POST request with specified values
func (mw *Client) Post(v Values) (*Response, error) {
	bb := borrowBuffer()
	defer returnBuffer(bb)
	encodeValue(bb, v)
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

//NewClient creates a new Client
func NewClient(apiphp string, user string, pass string) (*Client, error) {
	urlx, e := url.Parse(apiphp)
	if e != nil {
		return nil, e
	}
	cj, ce := cookiejar.New(nil)
	if ce != nil {
		return nil, e
	}
	return &Client{urlx, user, pass, &http.Client{Jar: cj}, false, LoginLegacy}, nil
}
