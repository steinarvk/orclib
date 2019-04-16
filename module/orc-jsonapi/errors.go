package jsonapi

import (
	"fmt"
	"net/http"
)

type HttpCode int

func (e HttpCode) Error() string { return http.StatusText(int(e)) }
func (e HttpCode) HttpCode() int { return int(e) }

type ErrorWithCode interface {
	Error() string
	HttpCode() int
}

type WithCode struct {
	Code    int
	Message string
}

func (c WithCode) Error() string {
	return c.Message
}

func (c WithCode) HttpCode() int {
	return c.Code
}

func BadRequest(s string) WithCode {
	return WithCode{http.StatusBadRequest, fmt.Sprintf("Bad request: %s", s)}
}

func MissingField(name string) WithCode {
	return BadRequest(fmt.Sprintf("missing field %q", name))
}

func BadField(name string, why error) WithCode {
	return BadRequest(fmt.Sprintf("invalid value for field %q: %v", name, why))
}

func getErrorMessage(err error) string {
	return err.Error()
}

func getErrorCode(err error) int {
	code := http.StatusOK
	if err != nil {
		code = http.StatusInternalServerError

		if unwrapped, ok := err.(ErrorWithCode); ok {
			code = unwrapped.HttpCode()
		}
	}
	return code
}

type BasicResponse struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}
