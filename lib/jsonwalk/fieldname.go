package jsonwalk

import "strings"

func FieldName(path string) string {
	rv := strings.Split(path, ".")
	if len(rv) == 0 {
		return ""
	}
	return rv[len(rv)-1]
}
