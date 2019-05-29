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
		cli   *http.Client
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
	poolResponse = sync.Pool{
		New: func() interface{} {
			return &Response{}
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

func returnBuffer(b *bytes.Buffer) {
	b.Reset()
	poolBuffer.Put(b)
}

func (mw *Client) request(rq *http.Request) (r *Response, e error) {
	rs, e := mw.cli.Do(rq)
	returnRequest(rq)
	if e != nil {
		return
	}
	b := poolBuffer.Get().(*bytes.Buffer)
	defer returnBuffer(b)
	if rs.ContentLength > 0 {
		b.Grow(int(rs.ContentLength))
	}
	_, e = b.ReadFrom(rs.Body)
	_ = rs.Body.Close()
	if e != nil {
		return
	}
	ja := jsoniter.Get(b.Bytes())
	if e = ja.LastError(); e != nil {
		return
	}
	r = poolResponse.Get().(*Response)
	je := ja.Get("error")
	if je.ValueType() == jsoniter.ObjectValue {
		var mae Error
		je.ToVal(&mae)
		r.v = jsoniter.Wrap(nil)
		e = mae
	} else {
		r.v = ja
	}
	return
}

//Get handles GET request with specified values
func (mw *Client) Get(v Values) (*Response, error) {
	bb := poolBuffer.Get().(*bytes.Buffer)
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
	bb := poolBuffer.Get().(*bytes.Buffer)
	defer returnBuffer(bb)
	encodeValue(bb, v)
	rq := borrowRequest("POST", mw.url.Host)
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
	cj, e := cookiejar.New(nil)
	if e != nil {
		return nil, e
	}
	return &Client{urlx, user, pass, &http.Client{Jar: cj}, false, LoginLegacy}, nil
}
