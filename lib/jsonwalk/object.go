package jsonwalk

import (
	"encoding/json"
	"fmt"
	"reflect"
)

func objectAsMap(x interface{}) (map[string]interface{}, error) {
	if m, ok := x.(map[string]interface{}); ok {
		return m, nil
	}

	marshalled, err := json.Marshal(x)
	if err != nil {
		return nil, fmt.Errorf("Value %v (Go type %v) cannot be marshalled as JSON: %v", x, reflect.TypeOf(x), err)
	}

	var unmarshalled interface{}
	if err := json.Unmarshal(marshalled, &unmarshalled); err != nil {
		return nil, fmt.Errorf("Value %v (Go type %v) cannot be normalized as JSON: %v", x, reflect.TypeOf(x), err)
	}

	if m, ok := unmarshalled.(map[string]interface{}); ok {
		return m, nil
	}

	return nil, fmt.Errorf("Value %v (Go type %v) could not be converted to a map", x, reflect.TypeOf(x))
}
