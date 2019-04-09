package jsonwalk

import (
	"encoding/json"
	"fmt"
	"reflect"
)

func arrayAsSlice(x interface{}) ([]interface{}, error) {
	if xs, ok := x.([]interface{}); ok {
		return xs, nil
	}

	marshalled, err := json.Marshal(x)
	if err != nil {
		return nil, fmt.Errorf("Value %v (Go type %v) cannot be marshalled as JSON: %v", x, reflect.TypeOf(x), err)
	}

	var unmarshalled interface{}
	if err := json.Unmarshal(marshalled, &unmarshalled); err != nil {
		return nil, fmt.Errorf("Value %v (Go type %v) cannot be normalized as JSON: %v", x, reflect.TypeOf(x), err)
	}

	if xs, ok := unmarshalled.([]interface{}); ok {
		return xs, nil
	}

	return nil, fmt.Errorf("Value %v (Go type %v) could not be converted to a slice", x, reflect.TypeOf(x))
}
