package jsonshape

import (
	"fmt"
	"reflect"

	"github.com/steinarvk/orclib/lib/jsonwalk"
)

type Shape interface{}

func ShapeOf(value interface{}, opts ...Option) (Shape, error) {
	builtOpts, err := buildOptions(opts)
	if err != nil {
		return nil, err
	}

	return shapeOf(builtOpts, value)
}

func shapeOf(opts options, rawValue interface{}) (Shape, error) {
	rv, err := jsonwalk.WalkWithValues(rawValue, func(parent interface{}, path string, rawValue interface{}) (interface{}, bool, error) {
		var fieldValue interface{}

		attachValue := func(fieldValue interface{}) error {
			if parent == nil {
				// nothing to attach to. the return value will do it.
				return nil
			}

			fieldName := jsonwalk.FieldName(path)
			parentMap, ok := parent.(map[string]interface{})
			if !ok {
				return fmt.Errorf("Internal error: at %q, expected parent (%v, Go type %v) to be object", path, parent, reflect.TypeOf(parent))
			}
			rawParentProps, ok := parentMap["properties"]
			if !ok {
				return fmt.Errorf("Internal error: at %q, expected parent (%v, Go type %v) to have property 'properties'", path, parent, reflect.TypeOf(parent))
			}
			parentProps, ok := rawParentProps.(map[string]interface{})
			if !ok {
				return fmt.Errorf("Internal error: at %q, expected parent.properties (%v, Go type %v) to be object", path, rawParentProps, reflect.TypeOf(rawParentProps))
			}

			parentProps[fieldName] = fieldValue

			return nil
		}

		switch value := rawValue.(type) {
		case nil:
			fieldValue = nullSchema
		case map[string]interface{}:
			newParent := map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			}
			if err := attachValue(newParent); err != nil {
				return nil, false, err
			}
			return newParent, true, nil
		case string:
			fieldValue = stringSchema
		case float64:
			fieldValue = numberSchema
		case bool:
			fieldValue = boolSchema
		case []interface{}:
			m := map[string]interface{}{}
			for _, v := range value {
				schema, err := shapeOf(opts, v)
				if err != nil {
					return nil, false, fmt.Errorf("At %q: %v", path, err)
				}
				h, err := hashSchema(schema)
				if err != nil {
					return nil, false, fmt.Errorf("At %q: %v", path, err)
				}
				m[h] = schema
			}
			var schemas []interface{}
			for _, schema := range m {
				schemas = append(schemas, schema)
			}
			if len(schemas) == 0 {
				fieldValue = emptyArraySchema
			} else if len(schemas) == 1 {
				fieldValue = arrayOfType(schemas[0])
			} else {
				fieldValue = arrayOfType(anyOf(schemas...))
			}
		default:
			return nil, false, fmt.Errorf("Internal error: value at %q (%v, Go type %v) was not normalized", path, rawValue, reflect.TypeOf(rawValue))
		}

		if parent == nil {
			return fieldValue, false, nil
		}

		if err := attachValue(fieldValue); err != nil {
			return nil, false, err
		}

		return nil, false, nil
	})
	return rv, err
}
