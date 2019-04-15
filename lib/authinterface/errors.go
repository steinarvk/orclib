package authinterface

import "net/http"

type ErrorWithFailureInfo interface {
	AuthFailureInfo() AuthFailureInfo
}

func ErrorWith(err error, info AuthFailureInfo) error {
	return errorWithInfo{err: err, info: info}
}

func DenyWith(info AuthFailureInfo) error {
	return ErrorWith(errPermissionDenied{}, info)
}

type errPermissionDenied struct{}

func (e errPermissionDenied) Error() string { return "Permission denied" }
func (e errPermissionDenied) HttpCode() int { return http.StatusUnauthorized }

type errorWithInfo struct {
	err  error
	info AuthFailureInfo
}

func (e errorWithInfo) Error() string {
	return e.err.Error()
}

func (e errorWithInfo) HttpCode() int { return http.StatusUnauthorized }

func (e errorWithInfo) AuthFailureInfo() AuthFailureInfo {
	return e.info
}
