package mwapi

import (
	"errors"
	"fmt"
	"strings"
)

var (
	errNoCSRF = errors.New("no CSRF Token provided")
)

func (mw *Client) Login() error {
	var a loginInfo
	mwv := Values{
		"action":     "login",
		"lgname":     mw.uname,
		"lgpassword": mw.paswd,
	}
	r, e := mw.Post(mwv)
	if e != nil {
		return e
	}
	_ = r.Get(&a, "login")
	if a.Result == "" {
		return fmt.Errorf("LOGIN NO RESULT %#v", a)
	}
	mwv["lgtoken"] = a.Token
	r, e = mw.Post(mwv)
	if e != nil {
		return e
	}
	_ = r.Get(&a, "login")
	if a.Result != "Success" {
		return fmt.Errorf("LOGIN TOKEN NO RESULT %#v", a)
	}
	mw.login = true
	return nil
}

func (mw *Client) Token(tkn ...string) (map[string]string, error) {
	tokens := map[string]string{}
	v := Values{
		"action": "query",
		"meta":   "tokens",
	}
	if len(tkn) > 0 {
		v["type"] = strings.Join(tkn, "|")
	}
	r, e := mw.Post(v)
	if e != nil {
		return nil, e
	}
	_ = r.Get(&tokens, "query", "tokens")
	return tokens, nil
}

func (mw *Client) Read() *Query {
	return mw.Query().Prop("revisions", Values{"rvprop": "content", "rvslots": "main"})
}

func (mw *Client) actionWithToken(action string, v Values, to interface{}) error {
	tkm, e := mw.Token()
	if e != nil {
		return e
	}
	tk, ok := tkm["csrftoken"]
	if !ok {
		return errNoCSRF
	}
	v["action"] = action
	v["token"] = tk
	r, e := mw.Post(v)
	if e != nil {
		return e
	}
	if to == nil {
		return nil
	}
	return r.Get(to, action)
}

func (mw *Client) Edit(tit string, txt string, summ string, minor, create bool, to interface{}) error {
	v := Values{
		"title":   tit,
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
	return mw.actionWithToken("edit", v, to)
}

func (mw *Client) Move(from, to, reason string, v Values, x interface{}) error {
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
	return mw.actionWithToken("move", v, x)
}

func (mw *Client) Delete(title, reason string, v Values, x interface{}) error {
	if v == nil {
		v = make(Values, 4)
	}
	JoinValues(v, Values{
		"title":  title,
		"reason": reason,
	})
	return mw.actionWithToken("delete", v, x)
}

func (mw *Client) Block(user, expiry, reason string, v Values, x interface{}) error {
	if v == nil {
		v = make(Values, 4)
	}
	JoinValues(v, Values{
		"user":   user,
		"expiry": expiry,
		"reason": reason,
	})
	return mw.actionWithToken("block", v, x)
}
