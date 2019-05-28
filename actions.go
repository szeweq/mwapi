package mwapi

import (
	"errors"
	"strings"
)

var (
	errNoCSRF = errors.New("no CSRF Token provided")
)

//Token returns tokens for specified actions
func (mw *Client) Token(names ...string) (tokens map[string]string, e error) {
	v := Values{
		"action": "query",
		"meta":   "tokens",
	}
	if len(names) > 0 {
		v["type"] = strings.Join(names, "|")
	}
	r, e := mw.Post(v)
	if e == nil {
		e = r.GetRaw(&tokens, "query", "tokens")
	}
	return
}

//Read creates a query which allows user to obtain page content
func (mw *Client) Read() *Query {
	return mw.Query().Prop("revisions", Values{"rvprop": "content", "rvslots": "main"})
}

func (mw *Client) actionWithToken(action string, v Values) (r *Response, e error) {
	tkm, e := mw.Token()
	if e != nil {
		return
	}
	tk, ok := tkm["csrftoken"]
	if !ok {
		return nil, errNoCSRF
	}
	v["action"] = action
	v["token"] = tk
	r, e = mw.Post(v)
	if e == nil {
		r.path = []interface{}{action}
	}
	return
}

//Edit sends an edit action
func (mw *Client) Edit(title string, txt string, summ string, minor, create bool) (*Response, error) {
	v := Values{
		"title":   title,
		"text":    txt,
		"bot":     "",
		"summary": summ,
	}
	if minor {
		v["minor"] = ""
	}
	if !create {
		v["nocreate"] = ""
	}
	return mw.actionWithToken("edit", v)
}

//Move sends a move action
func (mw *Client) Move(from, to, reason string, v Values) (*Response, error) {
	if v == nil {
		v = make(Values, 4)
	}
	JoinValues(v, Values{
		"from":       from,
		"to":         to,
		"reason":     reason,
		"movetalk":   "",
		"noredirect": "",
	})
	return mw.actionWithToken("move", v)
}

//Delete sends a delete action
func (mw *Client) Delete(title, reason string, v Values) (*Response, error) {
	if v == nil {
		v = make(Values, 4)
	}
	JoinValues(v, Values{
		"title":  title,
		"reason": reason,
	})
	return mw.actionWithToken("delete", v)
}

//Block sends a block action
func (mw *Client) Block(user, expiry, reason string, v Values) (*Response, error) {
	if v == nil {
		v = make(Values, 4)
	}
	JoinValues(v, Values{
		"user":   user,
		"expiry": expiry,
		"reason": reason,
	})
	return mw.actionWithToken("block", v)
}
