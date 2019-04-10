package jsonshape

import "fmt"

func shapeFields(s Shape) ([]FieldShape, error) {
	rv, ok := s.(*objectShape)
	if !ok {
		return nil, fmt.Errorf("Not an object shape: %v", s)
	}
	return rv.fields, nil
}

func shapeItemShape(s Shape) (Shape, error) {
	rv, ok := s.(*arrayShape)
	if !ok {
		return nil, fmt.Errorf("Not a non-empty array shape: %v", s)
	}
	return rv.itemShape, nil
}

func anyofOptions(s Shape) ([]Shape, error) {
	rv, ok := s.(*anyOfShape)
	if !ok {
		return nil, fmt.Errorf("Not an anyOf shape: %v", s)
	}
	return rv.options, nil
}
