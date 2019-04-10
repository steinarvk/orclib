package jsonshape

import (
	"fmt"
	"strings"

	"github.com/steinarvk/orclib/lib/jsonwalk"
)

type Shape interface {
	AsJSONSchema() interface{}
	CompactShapeDescription() (string, bool)
}

type FieldShape struct {
	Name     string
	Optional bool
	Shape    Shape
}

type anyOfShape struct {
	options []Shape
}

func (s *anyOfShape) AsJSONSchema() interface{} {
	var schemas []interface{}
	for _, opt := range s.options {
		schemas = append(schemas, opt.AsJSONSchema())
	}
	return jsonSchema{
		AnyOf: schemas,
	}
}

func (s *anyOfShape) CompactShapeDescription() (string, bool) {
	suffix := ""
	var descs []string
	for _, opt := range s.options {
		if opt == nullShapeValue {
			suffix = "?"
			continue
		}
		desc, ok := opt.CompactShapeDescription()
		if !ok {
			return "", false
		}
		descs = append(descs, desc)
	}
	return strings.Join(descs, "|") + suffix, true
}

type primitiveShape struct {
	t *jsonwalk.Type
}

func (s *primitiveShape) AsJSONSchema() interface{} {
	return jsonSchema{
		Type: s.t.String(),
	}
}

func (s *primitiveShape) CompactShapeDescription() (string, bool) {
	return s.t.String(), true
}

type objectShape struct {
	fields []FieldShape
}

func (s *objectShape) AsJSONSchema() interface{} {
	propSchemas := map[string]interface{}{}
	for _, field := range s.fields {
		propSchemas[field.Name] = field.Shape.AsJSONSchema()
	}
	return jsonSchema{
		Type:       "object",
		Properties: propSchemas,
	}
}

func (s *objectShape) CompactShapeDescription() (string, bool) {
	if len(s.fields) > 1 {
		return "", false
	}
	if len(s.fields) == 0 {
		return "{}", true
	}
	if s.fields[0].Optional {
		return "", false
	}
	subshape, ok := s.fields[0].Shape.CompactShapeDescription()
	if !ok {
		return "", false
	}
	return fmt.Sprintf("{%q:%s}", s.fields[0].Name, subshape), true
}

type arrayShape struct {
	itemShape Shape
}

func (s *arrayShape) AsJSONSchema() interface{} {
	return jsonSchema{
		Type:  "array",
		Items: s.itemShape.AsJSONSchema(),
	}
}

func (s *arrayShape) CompactShapeDescription() (string, bool) {
	subshape, ok := s.itemShape.CompactShapeDescription()
	if !ok {
		return "", false
	}
	return "[" + subshape + "]", true
}

type emptyArrayShape struct{}

func (s emptyArrayShape) AsJSONSchema() interface{} {
	return jsonSchema{
		Type:     "array",
		MaxItems: 0,
	}
}

func (s emptyArrayShape) CompactShapeDescription() (string, bool) {
	return "[]", true
}

var stringShapeValue = &primitiveShape{jsonwalk.String}
var numberShapeValue = &primitiveShape{jsonwalk.Number}
var boolShapeValue = &primitiveShape{jsonwalk.Bool}
var nullShapeValue = &primitiveShape{jsonwalk.Null}
var emptyArrayShapeValue = emptyArrayShape{}
