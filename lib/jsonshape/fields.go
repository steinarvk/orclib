package jsonshape

import (
	"fmt"
	"reflect"
	"sort"
)

type FieldShape struct {
	Name  string
	Shape Shape
}

func shapeFields(shape Shape) ([]FieldShape, error) {
	m, ok := shape.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Invalid shape (not a map): %v (Go type %v)", shape, reflect.TypeOf(shape))
	}

	mType, ok := m["type"]
	if !ok || mType != "object" {
		return nil, fmt.Errorf("Cannot take fields of shape %v: not an object", shape)
	}

	mPropsRaw, ok := m["properties"]
	if !ok {
		return nil, fmt.Errorf("Cannot take fields of shape %v: no properties", shape)
	}

	mProps, ok := mPropsRaw.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Invalid shape (properties are not an object): %v (Go type %v)", shape, reflect.TypeOf(shape))
	}

	var ks []FieldShape
	for k, v := range mProps {
		ks = append(ks, FieldShape{Name: k, Shape: Shape(v)})
	}
	sort.Slice(ks, func(i, j int) bool { return ks[i].Name < ks[j].Name })
	return ks, nil
}

func shapeItemShape(shape Shape) (Shape, error) {
	m, ok := shape.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Invalid shape (not a map): %v (Go type %v)", shape, reflect.TypeOf(shape))
	}

	mType, ok := m["type"]
	if !ok || mType != "array" {
		return nil, fmt.Errorf("Cannot take item shape of shape %v: not an array", shape)
	}

	mItemsRaw, ok := m["items"]
	if !ok {
		return nil, fmt.Errorf("Cannot take items of shape %v: no items", shape)
	}

	return Shape(mItemsRaw), nil
}

func anyofOptions(shape Shape) ([]Shape, error) {
	m, ok := shape.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Invalid shape (not a map): %v (Go type %v)", shape, reflect.TypeOf(shape))
	}

	mAnyOfRaw, ok := m["anyOf"]
	if !ok {
		return nil, fmt.Errorf("Cannot take anyOf of shape %v: no anyOf", shape)
	}

	mAnyOf, ok := mAnyOfRaw.([]interface{})
	if !ok {
		return nil, fmt.Errorf("Cannot take anyOf of shape %v: anyOf field is not an array", shape)
	}

	var shapes []Shape
	for _, s := range mAnyOf {
		shapes = append(shapes, Shape(s))
	}

	return shapes, nil
}
