package mwapi

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"
)

type (
	//Response contains json-encoded API result
	Response struct {
		path []interface{}
		v    jsoniter.Any
	}
)

//Get saves decoded response to a value with specified prefixed path
func (r *Response) Get(to interface{}, path ...interface{}) error {
	return r.GetRaw(to, append(r.path, path...)...)
}

//Get saves decoded response to a value with raw path
func (r *Response) GetRaw(to interface{}, path ...interface{}) error {
	x := r.v.Get(path...)
	if e := r.v.LastError(); e != nil {
		return e
	}
	if x.ValueType() == jsoniter.InvalidValue {
		var b jsoniter.RawMessage
		r.v.ToVal(&b)
		return fmt.Errorf("invalid value %s on %s", path, b)
	}
	x.ToVal(to)
	return nil
}

//ReturnToPool sets all contained data to default end adds a response to a pool
func (r *Response) ReturnToPool() {
	r.v = nil
	r.path = r.path[:0]
	poolResponse.Put(r)
}
