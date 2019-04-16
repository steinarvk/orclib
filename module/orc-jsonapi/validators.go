package jsonapi

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"

	"github.com/gorilla/mux"
)

var (
	numericRE = regexp.MustCompile(`^[0-9]+$`)
)

func Numeric(s string) error {
	if s == "" {
		return errors.New("empty")
	}
	if !numericRE.MatchString(s) {
		return errors.New("not numeric")
	}
	return nil
}

func Nonempty(s string) error {
	if s == "" {
		return errors.New("empty string")
	}
	return nil
}

func ValidateVar(varName string, check func(string) error) EndpointWrapper {
	return EndpointWrapper(func(next genericHandler) genericHandler {
		return func(w http.ResponseWriter, req *http.Request) (interface{}, error) {
			value, ok := mux.Vars(req)[varName]
			if !ok {
				return nil, BadRequest(fmt.Sprintf("Missing URL component %q", varName))
			}
			if err := check(value); err != nil {
				return nil, BadRequest(fmt.Sprintf("Bad URL component %q=%q: %v", varName, value, err))
			}
			return next(w, req)
		}
	})
}
