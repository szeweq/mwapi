package mwapi

import (
	"bytes"
	"strconv"
	"sync"
	"time"
)

type (
	Query struct {
		c       *Client
		lists   []string
		props   []string
		titles  []string
		pageids []int64
		gen     []Gen
		v       Values
		cont    string
		r       *Response
	}
	QueryPage struct {
		PageID  int64  `json:"pageid"`
		NS      int    `json:"ns"`
		Title   string `json:"title"`
		Missing bool   `json:"missing,omitempty"`
	}
	QueryPropInfo struct {
		ContentModel string    `json:"contentmodel"`
		Touched      time.Time `json:"touched"`
		LastRevID    int64     `json:"lastrevid"`
		Length       uint64    `json:"length"`
	}
	QueryRevision struct {
		Slots struct {
			Main QueryRevisionContent `json:"main"`
		} `json:"slots"`
	}
	QueryRevisionContent struct {
		ContentModel string `json:"contentmodel"`
		Content      string `json:"content"`
	}
	ReadPage struct {
		QueryPage
		Revisions []QueryRevision
	}
)

var (
	poolQuery = sync.Pool{
		New: func() interface{} {
			return &Query{v: Values{"action": "query"}}
		},
	}
)

func (rp *ReadPage) Content() string {
	return rp.Revisions[0].Slots.Main.Content
}

func (mw *Client) Query() *Query {
	q := poolQuery.Get().(*Query)
	q.c = mw
	return q
}
func (q *Query) ReturnToPool() {
	q.c = nil
	q.lists = q.lists[:0]
	q.props = q.props[:0]
	q.titles = q.titles[:0]
	q.pageids = q.pageids[:0]
	q.gen = q.gen[:0]
	q.v = Values{"action": "query"}
	q.r = nil
	poolQuery.Put(q)
}
func (q *Query) With(v ...interface{}) *Query {
	for _, a := range v {
		switch x := a.(type) {
		case string:
			q.titles = append(q.titles, x)
		case []string:
			q.titles = append(q.titles, x...)
		case int64:
			q.pageids = append(q.pageids, x)
		case []int64:
			q.pageids = append(q.pageids, x...)
		case Gen:
			q.gen = append(q.gen, x)
		}
	}
	return q
}
func (q *Query) List(list string, v Values) *Query {
	q.lists = append(q.lists, list)
	if v != nil && len(v) > 0 {
		for k, a := range v {
			q.v[k] = a
		}
	}
	return q
}
func (q *Query) Prop(prop string, v Values) *Query {
	q.props = append(q.props, prop)
	if v != nil && len(v) > 0 {
		for k, a := range v {
			q.v[k] = a
		}
	}
	return q
}
func (q *Query) Continue(cont ...string) bool {
	if q.r == nil {
		q.v["rawcontinue"] = ""
		return true
	}
	for _, c := range cont {
		var s map[string]string
		e := q.r.Get(&s, "query-continue", c)
		if e == nil && len(s) > 0 {
			for k, v := range s {
				q.v[k] = v
			}
			return true
		}
	}
	return false
}

func (q *Query) Do() (e error) {
	xv := q.v
	var bb bytes.Buffer
	if len(q.titles) > 0 {
		xv["titles"] = appendStrings(&bb, q.titles)
		bb.Reset()
	}
	if len(q.pageids) > 0 {
		b := make([]byte, 0, 16)
		for i, p := range q.pageids {
			if i > 0 {
				b = append(b, '|')
			}
			b = strconv.AppendInt(b, p, 10)
		}
		xv["pageids"] = string(b)
		b = b[:0]
	}
	if len(q.props) > 0 {
		xv["prop"] = appendStrings(&bb, q.props)
		bb.Reset()
	}
	if len(q.lists) > 0 {
		xv["list"] = appendStrings(&bb, q.lists)
		bb.Reset()
	}
	if len(q.gen) > 0 {
		for i, g := range q.gen {
			if i > 0 {
				bb.WriteByte('|')
			}
			bb.WriteString(g.Name)
			for k, v := range g.Values {
				xv["g"+k] = v
			}
		}
		xv["generator"] = bb.String()
		bb.Reset()
	}
	q.r, e = q.c.Post(xv)
	return
}

func (q *Query) Get(v interface{}, path ...interface{}) error {
	path = append([]interface{}{"query"}, path...)
	return q.r.Get(v, path...)
}
func (q *Query) Pages(v interface{}) error {
	return q.r.Get(v, "query", "pages")
}
