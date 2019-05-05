package mwapi

import "fmt"

type LoginType int

const (
	LoginLegacy LoginType = iota
	LoginToken
	LoginClient
)

//Login handles user login via an API
func (mw *Client) Login() error {
	var r *Response
	var e error
	if mw.LoginType == LoginClient {
		tkm, e := mw.Token("login")
		if e != nil {
			return e
		}
		v := Values{
			"action":         "clientlogin",
			"username":       mw.uname,
			"password":       mw.paswd,
			"logintoken":     tkm["logintoken"],
			"loginreturnurl": "http://localhost/",
		}
		if r, e = mw.Post(v); e != nil {
			return e
		}
		var s string
		if e = r.Get(&s, "clientlogin", "status"); e != nil {
			return e
		}
		if s != "PASS" {
			v := s
			_ = r.Get(&s, "clientlogin", "messagecode")
			return fmt.Errorf("unexpected clientlogin status: %s - %s", v, s)
		}
		return nil
	}
	var a loginInfo
	mwv := Values{
		"action":     "login",
		"lgname":     mw.uname,
		"lgpassword": mw.paswd,
	}
	if mw.LoginType == LoginToken {
		tkm, e := mw.Token("login")
		if e != nil {
			return e
		}
		mwv["lgtoken"] = tkm["logintoken"]
	} else {
		r, e = mw.Post(mwv)
		if e != nil {
			return e
		}
		_ = r.Get(&a, "login")
		if a.Result == "" {
			return fmt.Errorf("LOGIN NO RESULT %#v", a)
		}
		mwv["lgtoken"] = a.Token
	}
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
