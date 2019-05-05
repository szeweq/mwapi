package mwapi

import (
	"bytes"
	"net/url"
)

func EncodeValue(bb *bytes.Buffer, v Values) {
	if v == nil {
		return
	}
	bb.WriteString("format=json&formatversion=2")
	for k, c := range v {
		bb.WriteByte('&')
		bb.WriteString(url.QueryEscape(k))
		bb.WriteByte('=')
		bb.WriteString(url.QueryEscape(c))
	}
}
func JoinValues(dst Values, srcs ...Values) {
	for _, m := range srcs {
		for k, v := range m {
			dst[k] = v
		}
	}
}

func (g Gen) Generate() Values {
	v := make(Values, len(g.Values)+1)
	v["generator"] = g.Name
	for k, x := range g.Values {
		v["g"+k] = x
	}
	return v
}
