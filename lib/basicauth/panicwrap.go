package basicauth

import "errors"

var unknownError = errors.New("Unknown error")

func wrappingPanic(f func() string) (rv string, returnedErr error) {
	defer func() {
		if r := recover(); r != nil {
			err, ok := r.(error)
			if !ok {
				err = unknownError
			}
			rv = ""
			returnedErr = err
		}
	}()
	rv = f()
	return
}
