package jsonwalk

import (
	"encoding/json"
	"fmt"
	"reflect"
)

type internalType string

type Type struct {
	t internalType
}

func (t *Type) String() string {
	return string(t.t)
}

var (
	Object = &Type{internalType("object")}
	Array  = &Type{internalType("array")}
	String = &Type{internalType("string")}
	Number = &Type{internalType("number")}
	Bool   = &Type{internalType("bool")}
	Null   = &Type{internalType("null")}
)

func TypeOf(x interface{}) (*Type, error) {
	return naiveTypeOf(x)
}

// Normalize JSON object so that it is (shallowly) one of the basic JSON types.
func Normalize(x interface{}) (interface{}, error) {
	switch v := x.(type) {
	case nil:
		return nil, nil
	case string:
		return v, nil
	case float64:
		return v, nil
	case bool:
		return v, nil
	case map[string]interface{}:
		return v, nil
	case []interface{}:
		return v, nil
	}

	marshalled, err := json.Marshal(x)
	if err != nil {
		return nil, fmt.Errorf("Value %v (Go type %v) cannot be marshalled as JSON: %v", x, reflect.TypeOf(x), err)
	}

	var unmarshalled interface{}
	if err := json.Unmarshal(marshalled, &unmarshalled); err != nil {
		return nil, fmt.Errorf("Value %v (Go type %v) cannot be normalized as JSON: %v", x, reflect.TypeOf(x), err)
	}

	return unmarshalled, nil
}

func naiveTypeOf(x interface{}) (*Type, error) {
	// A naive implementation. Both for a quick-and-dirty proof of concept,
	// and for usage in tests as a reference implementation.
	// This should be "obviously correct" at great performance cost.
	norm, err := Normalize(x)
	if err != nil {
		return nil, fmt.Errorf("Unable to normalize value %v (Go type %v): %v", x, reflect.TypeOf(x), err)
	}

	switch norm.(type) {
	case map[string]interface{}:
		return Object, nil
	case []interface{}:
		return Array, nil
	case bool:
		return Bool, nil
	case float64:
		return Number, nil
	case string:
		return String, nil
	case nil:
		return Null, nil
	}

	return nil, fmt.Errorf("Cannot determine JSON type of %v (Go type %v)", x, reflect.TypeOf(x))
}
