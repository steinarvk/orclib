package jsonshape

import (
	"fmt"
	"reflect"

	"github.com/steinarvk/orclib/lib/jsonwalk"
)

func ShapeOf(value interface{}, opts ...Option) (Shape, error) {
	builtOpts, err := buildOptions(opts)
	if err != nil {
		return nil, err
	}

	return shapeOf(builtOpts, value)
}

func shapeOf(opts options, rawValue interface{}) (Shape, error) {
	rv, err := jsonwalk.WalkWithValues(rawValue, func(parent interface{}, path string, rawValue interface{}) (interface{}, bool, error) {
		var fieldValue Shape

		attachValue := func(fieldShape Shape) error {
			if parent == nil {
				// nothing to attach to. the return value will do it.
				return nil
			}

			parentObj := parent.(*objectShape)
			parentObj.fields = append(parentObj.fields, FieldShape{
				Name:  jsonwalk.FieldName(path),
				Shape: fieldShape,
			})

			return nil
		}

		switch value := rawValue.(type) {
		case nil:
			fieldValue = nullShapeValue
		case map[string]interface{}:
			newParent := &objectShape{}
			if err := attachValue(newParent); err != nil {
				return nil, false, err
			}
			return newParent, true, nil
		case string:
			fieldValue = stringShapeValue
		case float64:
			fieldValue = numberShapeValue
		case bool:
			fieldValue = boolShapeValue
		case []interface{}:
			v, err := arrayShapeFromValues(opts, value)
			if err != nil {
				return nil, false, fmt.Errorf("At %q: %v", path, err)
			}

			fieldValue = v

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
	return rv.(Shape), err
}
