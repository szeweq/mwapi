package mwapi

import (
	"bytes"
	"fmt"
)

type (
	loginInfo struct {
		Result string `json:"result,omitempty"`
		Token  string `json:"token,omitempty"`
	}
	//Error handles error information returned from MediaWiki API
	Error struct {
		Code string `json:"code"`
		Info string `json:"info"`
	}
	//Gen is used as a generator in query requests
	Gen struct {
		Name   string
		Values Values
	}
)

func (e Error) Error() string {
	return fmt.Sprintf("MWAPI Error (%s): %s", e.Code, e.Info)
}

func appendStrings(b *bytes.Buffer, as []string) string {
	for i, s := range as {
		if i > 0 {
			b.WriteByte('|')
		}
		b.WriteString(s)
	}
	return b.String()
}
