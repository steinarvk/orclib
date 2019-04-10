package jsonshape

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"sort"

	"github.com/steinarvk/orclib/lib/orchash"
)

func hashSchema(s Shape) (string, error) {
	return orchash.ComputeJSONHash(s.AsJSONSchema())
}

func uniqueShapes(originalShapes []Shape) ([]Shape, error) {
	m := map[string]Shape{}
	for _, shape := range originalShapes {
		h, err := hashSchema(shape)
		if err != nil {
			return nil, err
		}
		m[h] = shape
	}

	var shapes []Shape
	for _, v := range m {
		shapes = append(shapes, v)
	}

	return shapes, nil
}

func uniqueShapesFromValues(opts options, values []interface{}) ([]Shape, error) {
	var shapes []Shape
	for _, value := range values {
		shape, err := shapeOf(opts, value)
		if err != nil {
			return nil, err
		}
		shapes = append(shapes, shape)
	}
	return uniqueShapes(shapes)
}

func stringsetHash(m map[string]struct{}) string {
	var ks []string
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	h := sha256.New()
	for _, k := range ks {
		h.Write([]byte(base64.RawStdEncoding.EncodeToString([]byte(k))))
		h.Write([]byte{0})
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

func objectKeysetHash(obj *objectShape) string {
	m := map[string]struct{}{}
	for _, field := range obj.fields {
		m[field.Name] = struct{}{}
	}
	return stringsetHash(m)
}

func anyofShapeOfSimilarObjects(opts options, objs []*objectShape) (Shape, error) {
	if len(objs) == 0 {
		return nil, fmt.Errorf("No objects provided")
	}
	if len(objs) == 1 {
		return objs[0], nil
	}
	allFieldShapes := map[string][]Shape{}
	optional := map[string]bool{}
	for _, obj := range objs {
		for _, field := range obj.fields {
			allFieldShapes[field.Name] = append(allFieldShapes[field.Name], field.Shape)
			optional[field.Name] = false
		}
	}
	var ks []string
	for k := range allFieldShapes {
		for _, obj := range objs {
			hasField := false
			for _, field := range obj.fields {
				if field.Name == k {
					hasField = true
					break
				}
			}
			if !hasField {
				optional[k] = true
			}
		}
		ks = append(ks, k)
	}
	sort.Strings(ks)

	var newFields []FieldShape
	for _, k := range ks {
		combinedShape, err := anyofShapeFromShapes(opts, allFieldShapes[k])
		if err != nil {
			return nil, err
		}

		newFields = append(newFields, FieldShape{
			Name:     k,
			Shape:    combinedShape,
			Optional: optional[k],
		})
	}

	return &objectShape{newFields}, nil
}

func arrayShapeFromValues(opts options, values []interface{}) (Shape, error) {
	if len(values) == 0 {
		return emptyArrayShapeValue, nil
	}

	valueshape, err := anyofShapeFromValues(opts, values)
	if err != nil {
		return nil, err
	}

	return &arrayShape{valueshape}, nil
}

func anyofShapeFromValues(opts options, values []interface{}) (Shape, error) {
	shapes, err := uniqueShapesFromValues(opts, values)
	if err != nil {
		return nil, err
	}
	return anyofShapeFromShapes(opts, shapes)
}

func anyofShapeFromShapes(opts options, shapes []Shape) (Shape, error) {
	shapes, err := uniqueShapes(shapes)
	if err != nil {
		return nil, err
	}

	if len(shapes) == 0 {
		return nil, fmt.Errorf("No options provided")
	}

	if len(shapes) == 1 {
		return shapes[0], nil
	}

	// Find out which are objects and non-empty arrays; only these
	// potentially require special treatment.
	// Arrays: [X] | [Y] ==> [X|Y]
	// Objects: {"x": X, "y": XY} | {"x": X, "y": YY}
	//          ==> {"x": X, "y": XX | XY}
	// Arrays always, objects only if the key set is exactly equal.

	var nonspecialShapes []Shape
	var arrayShapes []*arrayShape
	var objectShapes []*objectShape

	for _, shape := range shapes {
		switch cast := shape.(type) {
		case *objectShape:
			objectShapes = append(objectShapes, cast)
		case *arrayShape:
			arrayShapes = append(arrayShapes, cast)
		default:
			nonspecialShapes = append(nonspecialShapes, shape)
		}
	}

	finalShapes := nonspecialShapes

	// array shapes are simple. Just merge them all.
	if len(arrayShapes) > 0 {
		var allArrayOptions []Shape
		for _, ar := range arrayShapes {
			allArrayOptions = append(allArrayOptions, ar.itemShape)
		}
		itemshape, err := anyofShapeFromShapes(opts, allArrayOptions)
		if err != nil {
			return nil, err
		}
		finalShapes = append(finalShapes, &arrayShape{itemshape})
	}

	// object shapes are subtler, need to partition into key sets
	// or do we? might look nicer without
	if len(objectShapes) > 0 {
		/*
			byKeyset := map[string][]*objectShape{}
			for _, shape := range objectShapes {
				k := objectKeysetHash(shape)
				byKeyset[k] = append(byKeyset[k], shape)
			}
			for _, similarObjs := range byKeyset {
				newShape, err := anyofShapeOfSimilarObjects(opts, similarObjs)
				if err != nil {
					return nil, err
				}
				finalShapes = append(finalShapes, newShape)
			}
		*/
		newShape, err := anyofShapeOfSimilarObjects(opts, objectShapes)
		if err != nil {
			return nil, err
		}
		finalShapes = append(finalShapes, newShape)
	}

	if len(finalShapes) == 1 {
		return finalShapes[0], nil
	}

	return &anyOfShape{finalShapes}, nil
}
