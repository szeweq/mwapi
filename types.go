package mwapi

import (
	"bytes"
	"fmt"
)

type (
	MWError struct {
		Error string `json:"error,omitempty"`
	}
	loginResponse struct {
		MWError
		Login *loginInfo `json:"login,omitempty"`
	}
	loginInfo struct {
		Result string `json:"result,omitempty"`
		Token  string `json:"token,omitempty"`
	}
	MWAPIError struct {
		Code string `json:"code"`
		Info string `json:"info"`
	}
	PipedStrings []string
	Gen          struct {
		Name   string
		Values Values
	}
)

func (e MWAPIError) Error() string {
	return fmt.Sprintf("MWAPI Error (%s): %s", e.Code, e.Info)
}

func (ps PipedStrings) Append(b *bytes.Buffer) {
	for i, s := range ps {
		if i > 0 {
			b.WriteByte('|')
		}
		b.WriteString(s)
	}
}
